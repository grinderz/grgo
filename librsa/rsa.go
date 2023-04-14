package librsa

import (
	"crypto/rsa"
	"encoding/binary"
	"math/big"
	"unsafe"
)

func ParsePublicKey(ns, es []byte) *rsa.PublicKey {
	e := binary.BigEndian.Uint32(es)
	n := new(big.Int)

	n.SetBytes(ns)

	return &rsa.PublicKey{N: n, E: int(e)}
}

func MarshalPublicKey(pubKey *rsa.PublicKey) ([]byte, []byte) {
	n := pubKey.N.Bytes()
	e := make([]byte, unsafe.Sizeof(uint32(0)))

	binary.BigEndian.PutUint32(e, uint32(pubKey.E))

	return n, e
}

func ParsePrivateKey(pubKey *rsa.PublicKey, ds []byte) *rsa.PrivateKey {
	d := new(big.Int)
	d.SetBytes(ds)

	return &rsa.PrivateKey{PublicKey: *pubKey, D: d}
}

func MarshalPrivateKey(privatekey *rsa.PrivateKey) []byte {
	return privatekey.D.Bytes()
}

// https://www.openssl.org/docs/man1.1.0/crypto/RSA_public_decrypt.html

func PublicDecrypt(pubKey *rsa.PublicKey, data []byte) []byte {
	c := new(big.Int) //nolint:varnamelen
	m := new(big.Int) //nolint:varnamelen

	m.SetBytes(data)

	e := big.NewInt(int64(pubKey.E))

	c.Exp(m, e, pubKey.N)

	out := c.Bytes()
	skip := 0
	step := 2

	for ind := step; ind < len(out); ind++ {
		if ind+1 >= len(out) {
			break
		}

		if out[ind] == 0xff && out[ind+1] == 0 {
			skip = ind + step
			break
		}
	}

	return out[skip:]
}
