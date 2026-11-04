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

func EncryptData(plaintext []byte, passphrase string) ([]byte, []byte, []byte, error){

	salt := make([]byte, 16)
	if _, err  := io.ReadFull(rand.Reader, salt); err != nil{
		return nil, nil, nil, err
	}

	key := DeriveKey(passphrase, salt)

	block, err := aes.NewCipher(key)
	if err != nil{
		return nil, nil, nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil{
		return nil, nil, nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err  := io.ReadFull(rand.Reader, nonce); err != nil{
		return nil, nil, nil, err
	}

	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	return salt, nonce, ciphertext, nil
}

func DecryptData(salt, nonce, ciphertext []byte, passphrase string) ([]byte, error){
	key := DeriveKey(passphrase, salt)

	block, err := aes.NewCipher(key)
	if err != nil{
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil{
		return nil, err
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil{
		return nil, err
	}

	return plaintext, nil
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
