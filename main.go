package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

func main(){
	
	if len(os.Args) < 2{
		os.Exit(0)
	}

	// Encrypt flags
	encFlag := flag.NewFlagSet("encrypt", flag.ExitOnError)
	encFile := encFlag.String("file", "", "File or folder to encrypt")
	encOut := encFlag.String("out", "", "Output encrypted file path")
	encRm := encFlag.Bool("rm", false, "Remove source after success")

	// Decrypt flags
	decFlag := flag.NewFlagSet("decrypt", flag.ExitOnError)
	decFile := decFlag.String("file", "", "Encrypted file to restore")
	decOut := decFlag.String("out", "", "Output target directory")

	switch os.Args[1]{
	case "encrypt":
		err := encFlag.Parse(os.Args[2:])
		if err != nil{
			fmt.Fprintf(os.Stderr, "Invalid Flags Present: %v\n", err)
			os.Exit(1)
		}

		if *encFile == ""{
			fmt.Fprintf(os.Stderr, "No file specified.\nUse the -file flag to specify a file to encrypt.\n")
			os.Exit(1)
		}

		fmt.Print("Enter passphrase: ")
		pass,err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
		if err != nil{
			fmt.Fprintf(os.Stderr, "Password Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Print("Confirm passphrase: ")
		confirmPass, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
		if err != nil{
			fmt.Fprintf(os.Stderr, "Password Error: %v\n", err)
			os.Exit(1)
		}

		if !bytes.Equal(pass, confirmPass){
			fmt.Fprintf(os.Stderr, "Passwords do  not match")
			os.Exit(1)
		}

		info, err := os.Stat(*encFile)
		if err != nil{
			fmt.Fprintf(os.Stderr, "File Access Error: %v\n", err)
			os.Exit(1)
		}

		outFileName := *encOut
		if outFileName == ""{
			outFileName = *encFile + ".enc"
		}

		file, err := os.Create(outFileName)
		if err != nil{
			fmt.Fprintf(os.Stderr, "File Creation Error: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()

		salt, err := GenerateSalt()
		if err != nil{
			fmt.Fprintf(os.Stderr, "Salt Creation Error: %v\n", err)
			os.Exit(1)
		}

		err = WriteHeader(file, salt)
		if err != nil{
			fmt.Fprintf(os.Stderr, "Header Error: %v\n", err)
			os.Exit(1)
		}

		gcmBlock, err := NewEncryptionBlock(string(pass), salt)
		if err != nil{
			fmt.Fprintf(os.Stderr, "GCM Block Error: %v\n", err)
			os.Exit(1)
		}

		var srcReader io.Reader

		if info.IsDir(){
			archiveFile, err := ZipDirectory(*encFile)
			if err != nil{
				fmt.Fprintf(os.Stderr, "Archive Error: %v\n", err)
				os.Exit(1)
			}

			srcReader = bytes.NewReader(archiveFile)
		} else {
			inputFile, err := os.Open(*encFile)
			if err != nil{
				fmt.Fprintf(os.Stderr, "File Error: %v\n", err)
				os.Exit(1)
			}
			defer inputFile.Close()

			srcReader = inputFile
		}

		if err = EncryptStream(srcReader, file, gcmBlock);err != nil{
			fmt.Fprintf(os.Stderr, "Encryption Error: %v\n", err)
			os.Exit(1)			
		}

		if *encRm{
			if err = os.RemoveAll(*encFile); err != nil{
				fmt.Fprintf(os.Stderr, "File Remove Error: %v\n", err)
				os.Exit(1)
			} 
		}

	case "decrypt":
		err := decFlag.Parse(os.Args[2:])
		if err != nil{
			fmt.Fprintf(os.Stderr, "Invalid Flags Present: %v\n", err)
			os.Exit(1)
		}

		fmt.Print("Enter passphrase: ")
		pass, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
		if err != nil{
			fmt.Fprintf(os.Stderr, "Password Error: %v\n", err)
			os.Exit(1)
		}

		outFileName := *decOut
		if outFileName == ""{
			outFileName = strings.TrimSuffix(*decFile, ".enc")
		}

		inputFile, err := os.Open(*decFile)
		if err != nil{
			fmt.Fprintf(os.Stderr, "File Error: %v\n", err)
			os.Exit(1)
		}
		defer inputFile.Close()

		
		salt, err := ReadHeader(inputFile)
		if err != nil{
			fmt.Fprintf(os.Stderr, "File Error: %v\n", err)
			os.Exit(1)
		}

		gcmBlock, err := NewEncryptionBlock(string(pass), salt)
		if err != nil{
			fmt.Fprintf(os.Stderr, "GCM Block Error: %v\n", err)
			os.Exit(1)
		}

		pr, pw := io.Pipe()
		go func() {
			defer pw.Close()
			if err := DecryptStream(inputFile, pw, gcmBlock); err != nil{
				pw.CloseWithError(err)
			}
		} ()

		bufReader := bufio.NewReader(pr)

		plaintextBytes, err := bufReader.Peek(4)
		if err != nil{
			fmt.Fprintf(os.Stderr, "Byte Inspection Error: %v\n", err)
			os.Exit(1)
		}

		isZstdArchive := len(plaintextBytes) >= 4 && 
						plaintextBytes[0] == 0x28 && 
						plaintextBytes[1] == 0xB5 && 
						plaintextBytes[2] == 0x2F && 
						plaintextBytes[3] == 0xFD
		
		if isZstdArchive{
			archiveData, err := io.ReadAll(bufReader)
			if err != nil{
				fmt.Fprintf(os.Stderr, "Buffer Read Error: %v\n", err)
				os.Exit(1)
			}
			err = UnZipDirectory(archiveData, outFileName)
			if err != nil{
				fmt.Fprintf(os.Stderr, "Unzipping Error: %v\n", err)
				os.Exit(1)
			}
		} else {
			outFile, err := os.Create(outFileName)
			if err != nil{
				fmt.Fprintf(os.Stderr, "File Error: %v\n", err)
				os.Exit(1)
			}
			defer outFile.Close()

			if _, err = io.Copy(outFile, bufReader); err != nil{
				fmt.Fprintf(os.Stderr, "File Write Error: %v\n", err)
				os.Exit(1)
			}
		}

	}
}