package librsa_test

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/grinderz/grgo/librsa"
)

func TestRsa(t *testing.T) {
	t.Parallel()

	plain1 := "test"

	key1, err := rsa.GenerateKey(rand.Reader, 2048)
	checkError(t, err)
	checkError(t, key1.Validate())

	enc1, err := rsa.SignPKCS1v15(nil, key1, crypto.Hash(0), []byte(plain1))
	checkError(t, err)

	if plain1 != string(librsa.PublicDecrypt(&key1.PublicKey, enc1)) {
		t.Fatal("plain1 != decrypt(enc1)")
	}

	d := librsa.MarshalPrivateKey(key1)
	n, e := librsa.MarshalPublicKey(&key1.PublicKey)

	key2 := librsa.ParsePrivateKey(librsa.ParsePublicKey(n, e), d)

	enc2, err := rsa.SignPKCS1v15(nil, key2, crypto.Hash(0), []byte(plain1))
	checkError(t, err)

	if !bytes.Equal(enc1, enc2) {
		t.Fatal("enc1 != enc2")
	}

	if plain1 != string(librsa.PublicDecrypt(&key2.PublicKey, enc2)) {
		t.Fatal("plain1 != decrypt(enc2)")
	}
}

func checkError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatal(err)
	}
}
