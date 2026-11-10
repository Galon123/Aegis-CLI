package main

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/klauspost/compress/zstd"
)

func ZipDirectory(srcDir string) ([]byte, error){

	var buf bytes.Buffer

	enc, err := zstd.NewWriter(&buf)
	if err != nil{
		return nil, err
	}
	defer enc.Close()

	tarWriter := tar.NewWriter(enc)
	defer tarWriter.Close()

	baseDir, err := filepath.Abs(srcDir)
	if err != nil{
		return nil, err
	}

	err = filepath.WalkDir(baseDir, func(path string, d fs.DirEntry, err error) error {

		if err != nil{
			return err
		}

		relPath, err := filepath.Rel(baseDir, path)
		if err != nil{
			return err
		}

		if relPath == "."{
			return nil
		}

		info, err := d.Info()
		if err != nil{
			return err
		}

		headers, err := tar.FileInfoHeader(info, info.Name())
		if err != nil{
			return err
		}

		headers.Name = filepath.ToSlash(relPath)

		if d.IsDir(){
			headers.Name += "/"
		}

		err = tarWriter.WriteHeader(headers)
		if err != nil{
			return err
		}

		if d.IsDir(){
			return nil
		}

		file, err := os.Open(path)
		if err != nil{
			return err
		}

		defer file.Close()

		_, err = io.Copy(tarWriter, file)
		return err
	})

	if err != nil{
		return nil, err
	}

	if err = enc.Close(); err != nil{
		return nil, err
	}

	if err = tarWriter.Close(); err != nil{
		return nil, err
	}

	return buf.Bytes(), nil
}

func UnZipDirectory(archivedData []byte, targetDir string) error{

	reader := bytes.NewReader(archivedData)

	dec, err := zstd.NewReader(reader)
	if err != nil{
		return err
	}

	tarReader := tar.NewReader(dec)

	absTargetDir, err := filepath.Abs(targetDir)
	if err != nil{
		return err
	}

	err = os.MkdirAll(absTargetDir, 0700)
	if err != nil{
		return fmt.Errorf("Error in creating target directory: %w", err)
	}

	for {
		header, err := tarReader.Next()
		if err == io.EOF{
			break
		}
		if err != nil{
			return fmt.Errorf("Error in reading Headers: %w", err)
		}

		cleanPath := filepath.Clean(header.Name)
		targetPath := filepath.Join(absTargetDir, cleanPath)

		//Zip-Slip Fix
		expectedPrefix := absTargetDir + string(filepath.Separator)
		if !strings.HasPrefix(targetPath, expectedPrefix){
			return fmt.Errorf("Nice Try...But not happening: %w", err)
		}

		switch header.Typeflag{
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, 0700); err != nil{
				return fmt.Errorf("Couldnt Create Directory: %w", err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(targetPath), 0700); err != nil{
				return fmt.Errorf("Couldnt Create Parent Directories: %w", err)
			}

			file, err := os.OpenFile(targetPath, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0777)
			if err != nil{
				return fmt.Errorf("Couldnt create File : %w", err)
			}

			_, err = io.Copy(file, tarReader)
			file.Close()
			if err != nil{
				return fmt.Errorf("Couldnt Copy data into file: %w", err)
			}
		}
	}
	return nil
}