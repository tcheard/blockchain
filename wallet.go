package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"

	"golang.org/x/crypto/ripemd160"

	perrors "github.com/pkg/errors"
)

const (
	version            = byte(0x00)
	addressChecksumLen = 4
)

// Wallet stores private and public keys
type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

// NewWallet creates and returns a Wallet
func NewWallet() (*Wallet, error) {
	private, public, err := newKeyPair()
	if err != nil {
		return nil, perrors.Wrap(err, "failed to generate keypair")
	}

	return &Wallet{PrivateKey: private, PublicKey: public}, nil
}

// GetAddress returns a wallet's address
func (w Wallet) GetAddress() ([]byte, error) {
	pubKeyHash, err := HashPublicKey(w.PublicKey)
	if err != nil {
		return nil, perrors.Wrap(err, "failed to hash public key")
	}

	versionedPayload := append([]byte{version}, pubKeyHash...)
	checksum := checksum(versionedPayload)

	fullPayload := append(versionedPayload, checksum...)

	return Base58Encode(fullPayload), nil
}

// HashPublicKey hashes a public key
func HashPublicKey(pubKey []byte) ([]byte, error) {
	publicSHA := sha256.Sum256(pubKey)

	hasher := ripemd160.New()
	_, err := hasher.Write(publicSHA[:])
	if err != nil {
		return nil, perrors.Wrap(err, "failed to hash the public key")
	}

	return hasher.Sum(nil), nil
}

func newKeyPair() (ecdsa.PrivateKey, []byte, error) {
	c := elliptic.P256()
	private, err := ecdsa.GenerateKey(c, rand.Reader)
	if err != nil {
		return ecdsa.PrivateKey{}, nil, perrors.Wrap(err, "failed to generate key")
	}
	pubKey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)

	return *private, pubKey, nil
}

func checksum(payload []byte) []byte {
	first := sha256.Sum256(payload)
	second := sha256.Sum256(first[:])

	return second[:addressChecksumLen]
}

// Wallets stores a collection of wallets
type Wallets struct {
	Wallets map[string]*Wallet
}
