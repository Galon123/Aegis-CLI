package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"

	"golang.org/x/crypto/argon2"
)

func DeriveKey(passphrase string, salt []byte) []byte{
	return argon2.IDKey([]byte(passphrase), salt, 1, 64*1024, 4, 32)
}

func GenerateSalt() ([]byte, error){
	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil{
		return nil, err
	}
	return salt, nil
}

func NewEncryptionBlock(passphrase string, salt []byte) (cipher.AEAD, error){
	key := DeriveKey(passphrase, salt)

	block, err := aes.NewCipher(key)
	if err != nil{
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil{
		return nil, err
	}

	return gcm, nil
}

//Custom Header for Aegis Encryptions
func WriteHeader(w io.Writer, salt []byte) error{

	magicNo := []byte{'A', 'E', 'G', '2'}
	if _, err := w.Write(magicNo); err != nil{
		return err
	}
	if _, err := w.Write(salt); err != nil{
		return err
	}
	return nil
}

func ReadHeader(r io.Reader) ([]byte, error){

	initBytes := make([]byte, 4)
	if _, err := io.ReadFull(r, initBytes); err != nil{
		return nil, err
	}

	if string(initBytes) != "AEG2"{
		return nil, fmt.Errorf("Invalid File Format\n")
	}

	salt := make([]byte, 16)
	if _, err := io.ReadFull(r, salt); err != nil{
		return nil, err
	}
	return salt, nil
}
