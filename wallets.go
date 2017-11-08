package main

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"os"

	perrors "github.com/pkg/errors"
)

const walletFile = "wallet.dat"

// Wallets stores a collection of wallets
type Wallets struct {
	Wallets map[string]*Wallet
}

// NewWallets creates Wallets and fills it from a file if it exists
func NewWallets() (*Wallets, error) {
	wallets := &Wallets{}
	wallets.Wallets = make(map[string]*Wallet)
	err := wallets.LoadFromFile()
	return wallets, perrors.Wrap(err, "failed to load wallets from file")
}

// CreateWallet adds a new wallet to wallets
func (ws *Wallets) CreateWallet() (string, error) {
	wallet, err := NewWallet()
	if err != nil {
		return "", perrors.Wrap(err, "failed to create new wallet")
	}

	wAddr, err := wallet.GetAddress()
	if err != nil {
		return "", perrors.Wrap(err, "failed to get address from wallet")
	}

	address := fmt.Sprintf("%s", wAddr)

	ws.Wallets[address] = wallet

	return address, nil
}

// GetAddresses returns an array of addresses stored in the wallet file
func (ws *Wallets) GetAddresses() []string {
	var addresses []string

	for address := range ws.Wallets {
		addresses = append(addresses, address)
	}

	return addresses
}

// GetWallet returns a wallet by its address
func (ws Wallets) GetWallet(address string) *Wallet {
	return ws.Wallets[address]
}

// LoadFromFile loads wallets from the file
func (ws *Wallets) LoadFromFile() error {
	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		return perrors.Wrap(err, "wallet file doesn't exist")
	}

	fileContent, err := ioutil.ReadFile(walletFile)
	if err != nil {
		return perrors.Wrap(err, "failed to read content from wallet file")
	}

	var wallets Wallets
	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	if err = decoder.Decode(&wallets); err != nil {
		return perrors.Wrap(err, "failed to decode wallets")
	}

	ws.Wallets = wallets.Wallets
	return nil
}

// SaveToFile saves wallets to a file
func (ws *Wallets) SaveToFile() error {
	var content bytes.Buffer

	gob.Register(elliptic.P256())

	encoder := gob.NewEncoder(&content)
	if err := encoder.Encode(&ws); err != nil {
		return perrors.Wrap(err, "failed to encode wallets")
	}

	if err := ioutil.WriteFile(walletFile, content.Bytes(), 0644); err != nil {
		return perrors.Wrap(err, "failed to write wallet file")
	}

	return nil
}
