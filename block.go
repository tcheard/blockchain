package main

import (
	"bytes"
	"encoding/gob"
	"time"

	"github.com/pkg/errors"
)

// Block stores the information for a block in the blockchain
type Block struct {
	Timestamp     int64
	Data          []byte
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int64
}

// NewBlock creates a new block
func NewBlock(data string, prevBlockHash []byte) *Block {
	block := &Block{
		Timestamp:     time.Now().Unix(),
		Data:          []byte(data),
		PrevBlockHash: prevBlockHash,
		Hash:          []byte{},
	}

	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()
	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

// NewGenesisBlock creates a Block for the first block in a blockchain
func NewGenesisBlock() *Block {
	return NewBlock("Genesis Block", []byte{})
}

// Serialize serializes the block using the gob encoding
func (b *Block) Serialize() ([]byte, error) {
	var result bytes.Buffer

	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(b)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to encode block")
	}

	return result.Bytes(), nil
}

// DeserializeBlock takes a gob encoded byte slice and decodes it into a Block
func DeserializeBlock(d []byte) (*Block, error) {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to decode block")
	}

	return &block, nil
}
