package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
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

type EncryptedMessage struct {
	Data string `json:"Data"`
	Key  string `json:"Key"`
	IV   string `json:"IV"`
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
	// log.Println(len(PrivateKey.PublicKey.N.Bytes()))
	// log.Println(JWKPublicKey.N)
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
		Type:  "PUBLIC KEY",
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

func EncryptAES(key []byte, data []byte) (ciphered, nonce []byte) {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}

	nonce = make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	ciphered = aesgcm.Seal(nil, nonce, data, nil)
	return
}

func DecryptAES(key, ciphered, iv []byte) (decoded []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		err = errors.New("Cannot create new cipher -> " + err.Error())
		return
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		err = errors.New("Cannot create new GCM -> " + err.Error())
		return
	}
	decoded, err = aesgcm.Open(nil, iv, ciphered, nil)
	if err != nil {
		err = errors.New("Cannot decrypt -> " + err.Error())
		return
	}
	return
}

func Encrypt(data []byte, pub *rsa.PublicKey) ([]byte, error) {
	hash := sha256.New()
	cipherdata, err := rsa.EncryptOAEP(hash, rand.Reader, pub, data, nil)
	if err != nil {
		return cipherdata, errors.New("Cannot encrypt data - " + err.Error())
		// log.Panic(err)
	}
	return cipherdata, nil
}

func Decrypt(ciphered []byte, priv *rsa.PrivateKey) ([]byte, error) {
	hash := sha256.New()
	decrypteddata, err := rsa.DecryptOAEP(hash, rand.Reader, priv, ciphered, nil)
	if err != nil {
		return decrypteddata, err
	}
	return decrypteddata, nil
}

func RsaPublicKeyByModulusAndExponent(nStr string, eStr string) *rsa.PublicKey {
	decN, err := base64.RawURLEncoding.DecodeString(nStr)
	if err != nil {
		log.Println(err)
		return nil
	}
	// log.Println("RSA Gen - ", decN, nStr)
	n := big.NewInt(0)
	n.SetBytes(decN)

	decE, err := base64.RawURLEncoding.DecodeString(eStr)
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

func DecryptMessage(Key, IV, Data string) (string, error) {
	iv, err := base64.StdEncoding.DecodeString(IV)
	if err != nil {
		return "", errors.New("Cannot decode IV -> " + err.Error())
	}
	key, err := base64.StdEncoding.DecodeString(Key)
	if err != nil {
		return "", errors.New("Cannot decode Key -> " + err.Error())
	}
	data, _ := hex.DecodeString(Data)
	decryptedKey, err := Decrypt(key, PrivateKey)
	if err != nil {
		return "", errors.New("Cannot decrypt AES Key -> " + err.Error())
	}
	decodedDecryptedKey := make([]byte, 32)
	// Key Value is base64url encoded -> https://tools.ietf.org/html/rfc7518#page-32
	_, err = base64.RawURLEncoding.Decode(decodedDecryptedKey, decryptedKey)
	if err != nil {
		return "", errors.New("Cannot decode decrypted AES Key -> " + err.Error())
	}
	decryptedData, err := DecryptAES(decodedDecryptedKey, data, iv)
	if err != nil {
		return "", errors.New("Cannot DecryptAES -> " + err.Error())
	}
	return string(decryptedData), nil
}

func EncryptMessage(data []byte, key *rsa.PublicKey) (answer EncryptedMessage, err error) {

	// var answer EncryptedMessage
	aeskey := make([]byte, 32)
	if _, err = io.ReadFull(rand.Reader, aeskey); err != nil {
		err = errors.New("Cannot fill key -> " + err.Error())
		return
	}
	// log.Println(aeskey)
	c, err := aes.NewCipher(aeskey)

	if err != nil {
		err = errors.New("Cannot create cipher -> " + err.Error())
		return
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		err = errors.New("Cannot create new GCM -> " + err.Error())
		return
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		err = errors.New("Cannot fill nonce -> " + err.Error())
		return
	}
	encrypted := gcm.Seal(nil, nonce, data, nil)
	answer.Data = string(encrypted)
	encodedAESKey := []byte(base64.RawURLEncoding.EncodeToString(aeskey))

	encyptedAESKey, err := Encrypt(encodedAESKey, key)
	if err != nil {
		err = errors.New("Cannot encrypt aeskey -> " + err.Error())
		return
	}

	answer.Data = hex.EncodeToString(encrypted)
	answer.Key = base64.StdEncoding.EncodeToString(encyptedAESKey)
	answer.IV = base64.StdEncoding.EncodeToString(nonce)
	return
}

func Test() {
	key := RsaPublicKeyByModulusAndExponent(JWKPublicKey.N, JWKPublicKey.E)
	log.Printf("%v", PrivateKey.PublicKey.N.Bytes())
	encrypted, err := Encrypt([]byte("Hello"), key)
	decrypted, err := Decrypt(encrypted, PrivateKey)
	if err != nil {
		log.Println("TEST Error 2 - ", err.Error())
		return
	}
	if string(decrypted) != "Hello" {
		log.Println("TEST Error 3 - ", decrypted)
	} else {
		log.Println("TEST ALL FINE")
	}

}
