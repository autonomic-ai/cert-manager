package pki

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

var ellipticCurves = map[int]elliptic.Curve{
	224: elliptic.P224(),
	256: elliptic.P256(),
	384: elliptic.P384(),
	521: elliptic.P521(),
}

func GeneratePrivateKey(algorithm string, keySize int) (crypto.PrivateKey, error) {
	switch algorithm {
	case "rsa":
		return GenerateRSAPrivateKey(keySize)
	case "ecdsa":
		return GenerateECPrivateKey(keySize)
	default:
		return nil, fmt.Errorf("unsupported private key algorithm specified: %s", algorithm)
	}
}

func GenerateRSAPrivateKey(keySize int) (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, keySize)
}

func GenerateECPrivateKey(keySize int) (*ecdsa.PrivateKey, error) {
	ecCurve, ok := ellipticCurves[keySize]
	if !ok {
		return nil, fmt.Errorf("unsupported ecdsa key size specified: %d", keySize)
	}

	return ecdsa.GenerateKey(ecCurve, rand.Reader)
}

func EncodePKCS1PrivateKey(pk *rsa.PrivateKey) []byte {
	block := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(pk)}

	return pem.EncodeToMemory(block)
}

func EncodeSEC1PrivateKey(pk *ecdsa.PrivateKey) ([]byte, error) {
	asnBytes, err := x509.MarshalECPrivateKey(pk)
	if err != nil {
		return nil, err
	}

	block := &pem.Block{Type: "EC PRIVATE KEY", Bytes: asnBytes}
	return pem.EncodeToMemory(block), nil
}

func EncodePrivateKey(pk crypto.PrivateKey) ([]byte, error) {
	switch k := pk.(type) {
	case *rsa.PrivateKey:
		return EncodePKCS1PrivateKey(k), nil
	case *ecdsa.PrivateKey:
		return EncodeSEC1PrivateKey(k)
	default:
		return nil, fmt.Errorf("unknown key type: %T", pk)
	}
}
