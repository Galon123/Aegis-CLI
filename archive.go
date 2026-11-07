package main

import (
	"archive/tar"
	"bytes"
	"io"
	"io/fs"
	"os"
	"path/filepath"

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

	if err := enc.Close(); err != nil{
		return nil, err
	}

	if err := tarWriter.Close(); err != nil{
		return nil, err
	}

	return buf.Bytes(), nil
}