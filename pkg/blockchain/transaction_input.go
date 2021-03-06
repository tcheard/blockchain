package blockchain

import (
	"bytes"
)

// TXInput represents a transaction input
type TXInput struct {
	Txid      []byte
	Vout      int
	Signature []byte
	PubKey    []byte
}

// UsesKey checks whether the address initiated the transaction
func (in *TXInput) UsesKey(pubKeyHash []byte) (bool, error) {
	lockingHash, err := HashPublicKey(in.PubKey)
	if err != nil {
		return false, err
	}
	return bytes.Compare(lockingHash, pubKeyHash) == 0, nil
}
