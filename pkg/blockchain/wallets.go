package blockchain

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"os"
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
	return wallets, err
}

// CreateWallet adds a new wallet to wallets
func (ws *Wallets) CreateWallet() (string, error) {
	wallet, err := NewWallet()
	if err != nil {
		return "", err
	}

	wAddr, err := wallet.GetAddress()
	if err != nil {
		return "", err
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
		return err
	}

	fileContent, err := ioutil.ReadFile(walletFile)
	if err != nil {
		return err
	}

	var wallets Wallets
	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	if err = decoder.Decode(&wallets); err != nil {
		return err
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
		return err
	}

	return ioutil.WriteFile(walletFile, content.Bytes(), 0644)
}
