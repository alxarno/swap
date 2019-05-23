package crypto

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
)

const (
	bitSize = 2048
)

type JWKkey struct {
	Kty string `json:"kty"`
	E   string `json:"e"`
	N   string `json:"n"`
}

var (
	PrivateKey       *rsa.PrivateKey
	EncodedPublicKey []byte
	JWKPublicKey     JWKkey
)

func GenerateKeys() {
	key, err := generatePrivateKey()
	if err != nil {
		log.Panic(err)
	}
	PrivateKey = key
	EncodedPublicKey = encodePublicKeyToPEM(&PrivateKey.PublicKey)
	JWKPublicKey.Kty = "RSA"
	JWKPublicKey.E = base64.RawURLEncoding.EncodeToString(big.NewInt(int64(PrivateKey.PublicKey.E)).Bytes())
	JWKPublicKey.N = base64.RawURLEncoding.EncodeToString(PrivateKey.PublicKey.N.Bytes())
}

func generatePrivateKey() (*rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return nil, err
	}

	// Validate Private Key
	err = privateKey.Validate()
	if err != nil {
		return nil, err
	}

	// log.Println("RSA keys generated")
	return privateKey, nil
}

func encodePrivateKeyToPEM(key *rsa.PrivateKey) []byte {
	privDER := x509.MarshalPKCS1PrivateKey(key)
	privBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privDER,
	}

	privatePEM := pem.EncodeToMemory(&privBlock)
	return privatePEM
}

func encodePublicKeyToPEM(key *rsa.PublicKey) []byte {
	pubASN1, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		log.Panic(err)
	}

	pubBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubASN1,
	})

	return pubBytes
}

func BytesToPrivateKey(priv []byte) *rsa.PrivateKey {
	block, _ := pem.Decode(priv)
	enc := x509.IsEncryptedPEMBlock(block)
	b := block.Bytes
	var err error
	if enc {
		log.Println("is encrypted pem block")
		b, err = x509.DecryptPEMBlock(block, nil)
		if err != nil {
			log.Panic(err)
		}
	}
	key, err := x509.ParsePKCS1PrivateKey(b)
	if err != nil {
		log.Panic(err)
	}
	return key
}

func BytesToPublicKey(pub []byte) *rsa.PublicKey {
	block, _ := pem.Decode(pub)
	enc := x509.IsEncryptedPEMBlock(block)
	b := block.Bytes
	var err error
	if enc {
		log.Println("is encrypted pem block")
		b, err = x509.DecryptPEMBlock(block, nil)
		if err != nil {
			log.Panic(err)
		}
	}
	ifc, err := x509.ParsePKIXPublicKey(b)
	if err != nil {
		log.Panic(err)
	}
	key, ok := ifc.(*rsa.PublicKey)
	if !ok {
		log.Panic("not ok")
	}
	return key
}

func Encrypt(data []byte, pub *rsa.PublicKey) []byte {
	hash := sha512.New()
	cipherdata, err := rsa.EncryptOAEP(hash, rand.Reader, pub, data, nil)
	if err != nil {
		log.Panic(err)
	}
	return cipherdata
}

func Decrypt(ciphered []byte, priv *rsa.PrivateKey) []byte {
	hash := sha512.New()
	decrypteddata, err := rsa.DecryptOAEP(hash, rand.Reader, priv, ciphered, nil)
	if err != nil {
		log.Panic(err)
	}
	return decrypteddata
}

func RsaPublicKeyByModulusAndExponent(nStr string, eStr string) *rsa.PublicKey {
	decN, err := base64.StdEncoding.DecodeString(nStr)
	if err != nil {
		log.Println(err)
		return nil
	}
	n := big.NewInt(0)
	n.SetBytes(decN)

	decE, err := base64.StdEncoding.DecodeString(eStr)
	if err != nil {
		log.Println(err)
		return nil
	}

	var eBytes []byte
	if len(decE) < 8 {
		eBytes = make([]byte, 8-len(decE), 8)
		eBytes = append(eBytes, decE...)
	} else {
		eBytes = decE
	}
	eReader := bytes.NewReader(eBytes)
	var e uint64
	err = binary.Read(eReader, binary.BigEndian, &e)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	pKey := rsa.PublicKey{N: n, E: int(e)}
	return &pKey
}
