package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/koshatul/afero-static/src/afstatic"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func mainCommand(cmd *cobra.Command, args []string) {
	var cmp afstatic.CompressionType
	switch strings.ToLower(viper.GetString("compression")) {
	case string(afstatic.NoCompression):
		cmp = afstatic.NoCompression
	case string(afstatic.DeflateCompress):
		cmp = afstatic.DeflateCompress
	case string(afstatic.GZipCompress):
		cmp = afstatic.GZipCompress
	case string(afstatic.LzwCompress):
		cmp = afstatic.LzwCompress
	case string(afstatic.SnappyCompress):
		cmp = afstatic.SnappyCompress
	case string(afstatic.ZlibCompress):
		cmp = afstatic.ZlibCompress
	default:
		logrus.Errorf("Invalid compression type: %s", strings.ToLower(viper.GetString("compression")))
		cmd.Help()
		return
	}
	logrus.Infof("Using Compression: %s", cmp)

	j := afstatic.NewBuilder(cmp, viper.GetString("package"))
	j.Init()

	for _, path := range args {
		absPath, err := filepath.Abs(path)
		if err != nil {
			log.Fatal(err)
		}

		absStat, err := os.Stat(absPath)
		if err != nil {
			log.Fatal(err)
		}

		if absStat.IsDir() {
			logrus.Infof("Processing directory: %s", absPath)
			convertFs := afero.NewBasePathFs(afero.NewOsFs(), fmt.Sprintf("%s/", absPath))

			afero.Walk(convertFs, "/", func(path string, info os.FileInfo, err error) error {
				if !info.IsDir() {
					logrus.WithField("file", path).Infof("Adding file: %s", path)
					f, err := convertFs.Open(path)
					if err != nil {
						logrus.WithField("file", path).Errorf("Unable to open file: %s", err)
						return err
					}
					err = j.AddFile(path, f)
					if err != nil {
						logrus.WithField("file", path).Errorf("Unable to add file: %s", err)
						return err
					}
				}
				return nil
			})
		} else {
			path := filepath.Base(absPath)
			logrus.WithField("file", path).Infof("Processing file: %s", path)
			f, err := os.Open(absPath)
			if err != nil {
				logrus.WithField("file", path).Errorf("Unable to open file: %s", err)
			}
			err = j.AddFile(path, f)
			if err != nil {
				logrus.WithField("file", path).Errorf("Unable to add file: %s", err)
			}

		}
	}

	var out io.Writer
	switch viper.GetString("file") {
	case "-":
		out = os.Stdout
	default:
		f, err := os.Create(viper.GetString("file"))
		if err != nil {
			log.Fatal(err)
		}
		out = f
		defer f.Close()
	}

	err := j.Render(out)
	if err != nil {
		log.Fatal(err)
	}

}
