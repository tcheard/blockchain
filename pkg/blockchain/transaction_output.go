package blockchain

import (
	"bytes"
	"encoding/gob"

	"github.com/tcheard/blockchain/pkg/util"
)

// TXOutput represents a transaction output
type TXOutput struct {
	Value      int
	PubKeyHash []byte
}

// Lock signs the output
func (out *TXOutput) Lock(address []byte) {
	pubKeyHash := util.Base58Decode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	out.PubKeyHash = pubKeyHash
}

// IsLockedWithKey checks if the output can be used by the owner of the pubKey
func (out *TXOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}

// NewTXOutput creates a new TXOutput
func NewTXOutput(value int, address string) *TXOutput {
	txo := &TXOutput{value, nil}
	txo.Lock([]byte(address))

	return txo
}

// TXOutputs collects TXOutput
type TXOutputs struct {
	Outputs []*TXOutput
}

// Serialize serializes TXOutputs
func (outs TXOutputs) Serialize() ([]byte, error) {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	if err := enc.Encode(outs); err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

// DeserializeOutputs deserializes TXOutputs
func DeserializeOutputs(data []byte) (TXOutputs, error) {
	var outputs TXOutputs

	dec := gob.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(&outputs); err != nil {
		return TXOutputs{}, err
	}

	return outputs, nil
}
