package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/spf13/viper"

	"github.com/koshatul/afero-static/src/afstatic"
	"github.com/spf13/cobra"
)

var cmdTest = &cobra.Command{
	Use:    "test",
	Short:  "Test Command",
	Run:    testCommand,
	Args:   cobra.MinimumNArgs(1),
	Hidden: true,
}

func init() {
	rootCmd.AddCommand(cmdTest)
}

func testCommand(cmd *cobra.Command, args []string) {
	var cmp afstatic.CompressionType
	switch strings.ToLower(viper.GetString("compression")) {
	case "snappy":
		cmp = afstatic.SnappyCompress
	case "none":
		cmp = afstatic.NoCompression
	default:
		cmp = afstatic.SnappyCompress
	}

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

	// f, err := os.Open("artifacts/test.css")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// j.AddFile("/test.css", f)

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
	// fmt.Println()

	// fmt.Printf("Val: %s", reflect.TypeOf(0x01))
	// k := jen.NewFile("main")
	// k.HeaderComment("This file is generated - do not edit.")
	// k.Line()
	// k.Var().Id("file1").Op("=").Index().Byte().Values(
	// 	jen.Lit(0x01),
	// 	jen.Lit(0x02),
	// 	jen.Lit(0x03),
	// 	jen.Lit(0x04),
	// 	jen.Lit(0x05),
	// )

	// f.Var().Id("file2").Op("=").Index().Byte().Values(
	// 	jen.Lit(0x06),
	// 	jen.Lit(0x07),
	// 	jen.Lit(0x08),
	// 	jen.Lit(0x09),
	// 	jen.Lit(0x10),
	// )

	// fmt.Printf("%#v\n", f)

}
