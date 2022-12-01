package rsa

import (
	"crypto/rsa"
	"encoding/binary"
	"math/big"
)

func ParsePublicKey(ns, es []byte) *rsa.PublicKey {
	n := new(big.Int)
	n.SetBytes(ns)
	e := binary.BigEndian.Uint32(es)
	return &rsa.PublicKey{N: n, E: int(e)}
}

func MarshalPublicKey(pubKey *rsa.PublicKey) (n, e []byte) {
	e = make([]byte, 4)
	binary.BigEndian.PutUint32(e, uint32(pubKey.E))
	n = pubKey.N.Bytes()
	return
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
	c := new(big.Int)
	m := new(big.Int)

	m.SetBytes(data)
	e := big.NewInt(int64(pubKey.E))

	c.Exp(m, e, pubKey.N)

	out := c.Bytes()
	skip := 0
	for i := 2; i < len(out); i++ {
		if i+1 >= len(out) {
			break
		}
		if out[i] == 0xff && out[i+1] == 0 {
			skip = i + 2
			break
		}
	}
	return out[skip:]
}
