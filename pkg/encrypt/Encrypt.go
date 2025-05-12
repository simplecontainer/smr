package encrypt

import (
	"bytes"
	"compress/flate"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
)

func Compress(b []byte) *bytes.Buffer {
	var buf bytes.Buffer
	fw, err := flate.NewWriter(&buf, flate.BestCompression)
	if err != nil {
		log.Fatal(err)
	}

	_, err = fw.Write(b)
	if err != nil {
		log.Fatal(err)
	}

	err = fw.Close()

	if err != nil {
		return &bytes.Buffer{}
	}

	return &buf
}

func Decompress(b []byte) string {
	var buf = bytes.NewBuffer(b)
	fr := flate.NewReader(buf)
	defer fr.Close()

	data, err := ioutil.ReadAll(fr)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	return string(data)
}

func Encrypt(stringToEncrypt string, keyString string) (string, error) {
	key, _ := hex.DecodeString(keyString)
	plaintext := []byte(stringToEncrypt)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	return fmt.Sprintf("%x", ciphertext), nil
}

func Decrypt(encryptedString string, keyString string) (string, error) {
	key, _ := hex.DecodeString(keyString)
	enc, _ := hex.DecodeString(encryptedString)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()

	if nonceSize > len(enc) {
		return "", errors.New("invalid key")
	}

	nonce, ciphertext := enc[:nonceSize], enc[nonceSize:]

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s", plaintext), nil
}
