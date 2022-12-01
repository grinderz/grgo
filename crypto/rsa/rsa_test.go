package rsa

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"os"
	"testing"
)

func TestRsa(t *testing.T) {
	plain1 := "test"

	key1, err := rsa.GenerateKey(rand.Reader, 2048)
	checkError(err)
	checkError(key1.Validate())

	enc1, err := rsa.SignPKCS1v15(nil, key1, crypto.Hash(0), []byte(plain1))
	checkError(err)

	if plain1 != string(PublicDecrypt(&key1.PublicKey, enc1)) {
		t.Fatal("plain1 != decrypt(enc1)")
	}

	d := MarshalPrivateKey(key1)
	n, e := MarshalPublicKey(&key1.PublicKey)

	key2 := ParsePrivateKey(ParsePublicKey(n, e), d)

	enc2, err := rsa.SignPKCS1v15(nil, key2, crypto.Hash(0), []byte(plain1))
	checkError(err)

	if !bytes.Equal(enc1, enc2) {
		t.Fatal("enc1 != enc2")
	}

	if plain1 != string(PublicDecrypt(&key2.PublicKey, enc2)) {
		t.Fatal("plain1 != decrypt(enc2)")
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Println("fatal error ", err.Error())
		os.Exit(1)
	}
}
