package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"log"
	"math/big"

	"golang.org/x/crypto/ripemd160"
)

const version = byte(0x00)
const addressChecksumLen = 4

// Wallet stores private and public keys
type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

// NewWallet creates and returns a Wallet
func NewWallet() *Wallet {
	private, public := newKeyPair()
	wallet := Wallet{private, public}

	return &wallet
}

// GetAddress returns wallet address
func (w Wallet) GetAddress() []byte {
	pubKeyHash := HashPubKey(w.PublicKey)

	versionedPayload := append([]byte{version}, pubKeyHash...)
	checksum := checksum(versionedPayload)

	fullPayload := append(versionedPayload, checksum...)
	address := Base58Encode(fullPayload)

	return address
}

// HashPubKey hashes public key
func HashPubKey(pubKey []byte) []byte {
	publicSHA256 := sha256.Sum256(pubKey)

	RIPEMD160Hasher := ripemd160.New()
	_, err := RIPEMD160Hasher.Write(publicSHA256[:])
	if err != nil {
		log.Panic(err)
	}
	publicRIPEMD160 := RIPEMD160Hasher.Sum(nil)

	return publicRIPEMD160
}

// ValidateAddress check if address if valid
func ValidateAddress(address string) bool {
	pubKeyHash := Base58Decode([]byte(address))
	actualChecksum := pubKeyHash[len(pubKeyHash)-addressChecksumLen:]
	version := pubKeyHash[0]
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-addressChecksumLen]
	targetChecksum := checksum(append([]byte{version}, pubKeyHash...))

	return bytes.Equal(actualChecksum, targetChecksum)
}

// Checksum generates a checksum for a public key
func checksum(payload []byte) []byte {
	firstSHA := sha256.Sum256(payload)
	secondSHA := sha256.Sum256(firstSHA[:])

	return secondSHA[:addressChecksumLen]
}

func newKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err)
	}
	pubKey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)

	return *private, pubKey
}
func (w *Wallet) GobEncode() ([]byte, error) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)

	// Encode the curve name
	curveName := "P256"
	if err := encoder.Encode(curveName); err != nil {
		return nil, err
	}

	// Encode the private key components
	privateKeyBytes := w.PrivateKey.D.Bytes()
	if err := encoder.Encode(privateKeyBytes); err != nil {
		return nil, err
	}

	// Encode the public key components
	xBytes := w.PrivateKey.PublicKey.X.Bytes()
	yBytes := w.PrivateKey.PublicKey.Y.Bytes()
	if err := encoder.Encode(xBytes); err != nil {
		return nil, err
	}
	if err := encoder.Encode(yBytes); err != nil {
		return nil, err
	}

	// Encode the public key
	if err := encoder.Encode(w.PublicKey); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (w *Wallet) GobDecode(data []byte) error {
	buffer := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buffer)

	// Decode the curve name
	var curveName string
	if err := decoder.Decode(&curveName); err != nil {
		return err
	}

	// Set the curve based on the name
	var curve elliptic.Curve
	switch curveName {
	case "P256":
		curve = elliptic.P256()
	default:
		return fmt.Errorf("unsupported curve: %s", curveName)
	}

	// Decode the private key components
	var privateKeyBytes []byte
	if err := decoder.Decode(&privateKeyBytes); err != nil {
		return err
	}

	w.PrivateKey.PublicKey.Curve = curve
	w.PrivateKey.D = new(big.Int).SetBytes(privateKeyBytes)

	// Decode the public key components
	var xBytes, yBytes []byte
	if err := decoder.Decode(&xBytes); err != nil {
		return err
	}
	if err := decoder.Decode(&yBytes); err != nil {
		return err
	}

	w.PrivateKey.PublicKey.X = new(big.Int).SetBytes(xBytes)
	w.PrivateKey.PublicKey.Y = new(big.Int).SetBytes(yBytes)

	// Decode the public key
	if err := decoder.Decode(&w.PublicKey); err != nil {
		return err
	}

	return nil
}
