package main

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ncruces/zenity"
)

func main() {
	inDir, err := zenity.SelectFile(
		zenity.Filename(""),
		zenity.Directory(),
		zenity.DisallowEmpty(),
		zenity.Title("Select input directory."),
	)
	if err != nil {
		zenity.Error(
			err.Error(),
			zenity.Title("Error"),
			zenity.ErrorIcon,
		)
		log.Fatal(err)
	}

	// find folders
	files, err := os.ReadDir(inDir)
	if err != nil{
		zenity.Error(
			err.Error(),
			zenity.Title("Error"),
			zenity.ErrorIcon,
		)
		log.Fatal(err)
	}
	for _, file := range files {
		ext := filepath.Ext(filepath.Join(inDir,file.Name()))


		if ext == ".zip" {
			fmt.Println("")
			fmt.Println("----------------------------------------------------------------------------------------")
			fmt.Println(file.Name())
			fmt.Println("----------------------------------------------------------------------------------------")

			// open the zip file
			reader, err := zip.OpenReader(filepath.Join(inDir,file.Name()))
			if err != nil {
				zenity.Error(
					err.Error(),
					zenity.Title("Error"),
					zenity.ErrorIcon,
				)
				log.Fatal(err)
			}
			defer reader.Close()

			// extract each file
			for _, f := range reader.File {

				// check for ._ mac file
				if !(filepath.Base(f.Name)[:2] == "._" && f.FileHeader.UncompressedSize64 == 0){
					size, err := unzipFile(f, inDir)
					if err != nil {
						zenity.Error(
							err.Error(),
							zenity.Title("Error"),
							zenity.ErrorIcon,
						)
						log.Fatal(err)
					}
					if size != int64(f.FileHeader.UncompressedSize64){
						err = fmt.Errorf("validation error: size mismatch\n name: %s\n original: %d\n new: %d",f.Name, int64(f.FileHeader.UncompressedSize64), size)
						zenity.Error(
							err.Error(),
							zenity.Title("Error"),
							zenity.ErrorIcon,
						)
						log.Fatal(err)
					}
				}
			}

			// remove zip
			err = reader.Close()
			if err != nil {
				log.Fatal(err)
			}
			err = os.Remove(filepath.Join(inDir,file.Name()))
			if err != nil {
				zenity.Error(
					err.Error(),
					zenity.Title("Error"),
					zenity.ErrorIcon,
				)
				log.Fatal(err)
			}
		}
	}
}

func unzipFile(f *zip.File, destination string) (int64, error) {
	filePath := filepath.Join(destination, f.Name)
	if !strings.HasPrefix(filePath, filepath.Clean(destination)+string(os.PathSeparator)) {
		return 0, fmt.Errorf("invalid file path: %s", filePath)
	}

	if f.FileInfo().IsDir() {
		if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
			return 0, err
		}
		return 0, nil
	}

	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		return 0, err
	}

	destinationFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return 0, err
	}
	defer destinationFile.Close()

	zippedFile, err := f.Open()
	if err != nil {
		return 0, err
	}
	defer zippedFile.Close()

	if _, err := io.Copy(destinationFile, zippedFile); err != nil {
		return 0, err
	}

	// get uncompressed size in bytes
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}

	return fileInfo.Size(), nil
 }
