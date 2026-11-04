package main

import "fmt"

func main(){
	passphrase := "super-secret-password"

	plaintext := "This shit works"

	fmt.Printf("Original Sentence : %s\n", plaintext)
	salt, nonce, ciphertext, err := EncryptData([]byte(plaintext), passphrase)
	if err != nil{
		panic(err)
	}

	fmt.Printf("Encrypted data : %x\n", ciphertext)

	decryptText, err := DecryptData(salt, nonce, ciphertext, passphrase)
	if err != nil{
		panic(err)
	}
	fmt.Printf("Decrypted Data : %s\n", decryptText)
}