package main

import (
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
)

const ChunkSize = 64 * 1024

func EncryptStream(src io.Reader, dst io.Writer, gcm cipher.AEAD) error{

	buf := make([]byte, ChunkSize)

	for{
		n, err := src.Read(buf)
		
		if n > 0{
			chunkData := buf[:n]

			nonce := make([]byte, gcm.NonceSize())
			if _, err := io.ReadFull(rand.Reader, nonce); err != nil{
				return err
			}

			ciphertext := gcm.Seal(nil, nonce, chunkData, nil)

			chunkLen := uint32(len(nonce) + len(ciphertext))
			if err := binary.Write(dst, binary.BigEndian, chunkLen); err != nil{
				return err
			}

			if _, err := dst.Write(nonce); err != nil{
				return err
			}
			
			if _, err := dst.Write(ciphertext); err != nil{
				return err
			}
		}

		if err == io.EOF{
			break
		}

		if err != nil{
			return err
		}
	}
	return nil
}

func DecryptStream(src io.Reader, dst io.Writer, gcm cipher.AEAD) error{

	for{
		var chunkLen uint32;
		err := binary.Read(src, binary.BigEndian, &chunkLen)

		if err == io.EOF{
			return nil
		}

		if err != nil{
			return err
		}

		nonce := make([]byte, gcm.NonceSize())
		if _, err := io.ReadFull(src, nonce); err != nil{
			return err
		}

		cipherLen := int(chunkLen) - len(nonce)
		if cipherLen < 0{
			return fmt.Errorf("Invalid Stream Decrypt")
		}

		ciphertext := make([]byte, cipherLen)
		if _, err := io.ReadFull(src, ciphertext); err != nil{
			return err
		}

		plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
		if err != nil{
			return fmt.Errorf("Invalid Password (probably)")
		}

		if _, err := dst.Write(plaintext); err != nil{
			return err
		}
	}
	return nil
}