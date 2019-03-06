package afstatic

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/lzw"
	"compress/zlib"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/golang/snappy"
	"github.com/sirupsen/logrus"

	"github.com/dave/jennifer/jen"
)

type Builder struct {
	Compression CompressionType
	File        *jen.File
	files       map[string]string
}

func NewBuilder(compressionType CompressionType, packageName string) *Builder {
	f := jen.NewFile(packageName)
	f.HeaderComment("This file is generated - do not edit.")
	f.Line()
	return &Builder{
		Compression: compressionType,
		File:        f,
		files:       map[string]string{},
	}
}

func (b *Builder) Init() error {
	b.File.ImportName("github.com/spf13/afero", "afero")
	b.File.Var().Id("Fs").Id("afero.Fs")

	return nil
}

func (b *Builder) AddFile(filename string, file io.Reader) error {
	v := []jen.Code{}

	var rdr io.Reader
	switch b.Compression {
	case SnappyCompress:
		src, err := ioutil.ReadAll(file)
		if err != nil {
			return err
		}
		encoded := snappy.Encode(nil, src)
		logrus.Debugf("Copied %d bytes into compressor", len(src))
		rdr = bytes.NewBuffer(encoded)
	case DeflateCompress:
		cmpOut := new(bytes.Buffer)
		rawIn, err := flate.NewWriter(cmpOut, -1)
		if err != nil {
			return err
		}
		n, err := io.Copy(rawIn, file)
		rawIn.Close()
		if err != nil {
			return err
		}
		logrus.Debugf("Copied %d bytes into compressor", n)
		rdr = cmpOut
	case GZipCompress:
		cmpOut := new(bytes.Buffer)
		rawIn := gzip.NewWriter(cmpOut)
		n, err := io.Copy(rawIn, file)
		rawIn.Close()
		if err != nil {
			return err
		}
		logrus.Debugf("Copied %d bytes into compressor", n)
		rdr = cmpOut
	case LzwCompress:
		cmpOut := new(bytes.Buffer)
		rawIn := lzw.NewWriter(cmpOut, lzw.LSB, 8)
		n, err := io.Copy(rawIn, file)
		rawIn.Close()
		if err != nil {
			return err
		}
		logrus.Debugf("Copied %d bytes into compressor", n)
		rdr = cmpOut
	case ZlibCompress:
		cmpOut := new(bytes.Buffer)
		rawIn := zlib.NewWriter(cmpOut)
		n, err := io.Copy(rawIn, file)
		rawIn.Close()
		if err != nil {
			return err
		}
		logrus.Debugf("Copied %d bytes into compressor", n)
		rdr = cmpOut
	default:
		logrus.Debug("No compressor, so no bytes copied")
		rdr = file
	}

	buf := make([]byte, 1)
	var err error
	for {
		_, err = rdr.Read(buf)
		if err != nil {
			break
		}
		v = append(v, jen.Lit(int(buf[0])))
	}

	logrus.Debugf("Wrote %d bytes to static asset", len(v))

	b64filename := base64.RawStdEncoding.EncodeToString([]byte(filename))

	fileid := fmt.Sprintf("file_%s_%s", b64filename, b.Compression)

	b.files[filename] = fileid

	b.File.Var().Id(fileid).Op("=").Index().Byte().Values(
		v...,
	)

	return nil
}

func (b *Builder) Render(w io.Writer) error {
	v := []jen.Code{}

	switch b.Compression {
	case SnappyCompress:
		// Add SnappyDecompress buffer
		v = append(v, jen.Var().Id("o").Index().Byte())
	case DeflateCompress, GZipCompress, LzwCompress, ZlibCompress:
		v = append(
			v,
			jen.Var().Id("bufIn").Op("*").Qual("bytes", "Buffer"),
			jen.Var().Id("bufOut").Op("*").Qual("bytes", "Buffer"),
			jen.Var().Id("cmpOut").Qual("io", "ReadCloser"),
		)
	}

	for filename, file := range b.files {
		s := []jen.Code{}
		switch b.Compression {
		case SnappyCompress:
			// Do SnappyDecompress.
			s = []jen.Code{
				jen.List(
					jen.Id("o"), jen.Id("_"),
				).Op("=").Qual("github.com/golang/snappy", "Decode").Call(jen.Nil(), jen.Id(file)),
				jen.Qual("github.com/spf13/afero", "WriteFile").Call(
					jen.Id("Fs"),
					jen.Lit(filename),
					jen.Id("o"),
					jen.Qual("os", "ModePerm"),
				),
			}
		case DeflateCompress:
			s = []jen.Code{
				jen.Id("bufIn").Op("=").Qual("bytes", "NewBuffer").Call(jen.Id(file)),
				jen.Id("cmpOut").Op("=").Qual("compress/flate", "NewReader").Call(jen.Id("bufIn")),
				jen.Qual("io", "Copy").Call(jen.Id("bufOut"), jen.Id("cmpOut")),
				jen.Qual("github.com/spf13/afero", "WriteFile").Call(
					jen.Id("Fs"),
					jen.Lit(filename),
					jen.Id("bufOut").Dot("Bytes").Call(),
					jen.Qual("os", "ModePerm"),
				),
				jen.Id("cmpOut").Dot("Close").Call(),
			}
		case GZipCompress:
			s = []jen.Code{
				jen.Id("bufIn").Op("=").Qual("bytes", "NewBuffer").Call(jen.Id(file)),
				jen.Id("cmpOut").Op("=").Qual("compress/gzip", "NewReader").Call(jen.Id("bufIn")),
				jen.Qual("io", "Copy").Call(jen.Id("bufOut"), jen.Id("cmpOut")),
				jen.Qual("github.com/spf13/afero", "WriteFile").Call(
					jen.Id("Fs"),
					jen.Lit(filename),
					jen.Id("bufOut").Dot("Bytes").Call(),
					jen.Qual("os", "ModePerm"),
				),
				jen.Id("cmpOut").Dot("Close").Call(),
			}
		case LzwCompress:
			s = []jen.Code{
				jen.Id("bufIn").Op("=").Qual("bytes", "NewBuffer").Call(jen.Id(file)),
				jen.Id("cmpOut").Op("=").Qual("compress/lzw", "NewReader").Call(jen.Id("bufIn"), jen.Qual("compress/lzw", "LSB"), jen.Lit(8)),
				jen.Qual("io", "Copy").Call(jen.Id("bufOut"), jen.Id("cmpOut")),
				jen.Qual("github.com/spf13/afero", "WriteFile").Call(
					jen.Id("Fs"),
					jen.Lit(filename),
					jen.Id("bufOut").Dot("Bytes").Call(),
					jen.Qual("os", "ModePerm"),
				),
				jen.Id("cmpOut").Dot("Close").Call(),
			}
		case ZlibCompress:
			s = []jen.Code{
				jen.Id("bufIn").Op("=").Qual("bytes", "NewBuffer").Call(jen.Id(file)),
				jen.List(jen.Id("cmpOut"), jen.Id("_")).Op("=").Qual("compress/zlib", "NewReader").Call(jen.Id("bufIn")),
				jen.Qual("io", "Copy").Call(jen.Id("bufOut"), jen.Id("cmpOut")),
				jen.Qual("github.com/spf13/afero", "WriteFile").Call(
					jen.Id("Fs"),
					jen.Lit(filename),
					jen.Id("bufOut").Dot("Bytes").Call(),
					jen.Qual("os", "ModePerm"),
				),
				jen.Id("cmpOut").Dot("Close").Call(),
			}
		case NoCompression:
			// No Decompression needed.
			s = []jen.Code{
				jen.Qual("github.com/spf13/afero", "WriteFile").Call(
					jen.Id("Fs"),
					jen.Lit(filename),
					jen.Id(file),
					jen.Qual("os", "ModePerm"),
				),
			}
		}
		v = append(
			v,
			s...,
		)
	}

	v = append(
		[]jen.Code{jen.Id("Fs").Op("=").Qual("github.com/spf13/afero", "NewMemMapFs").Call()},
		v...,
	)

	b.File.Func().Id("init").Params().Block(
		v...,
	)

	return b.File.Render(w)
}
