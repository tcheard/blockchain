package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
)

const targetBits = 24

// ProofOfWork stores data for creating a proof of work for a block
type ProofOfWork struct {
	block  *Block
	target *big.Int
}

// NewProofOfWork creates a new proof of work for a given block
func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))

	pow := &ProofOfWork{
		block:  b,
		target: target,
	}

	return pow
}

func (pow *ProofOfWork) prepareData(nonce int64) []byte {
	return bytes.Join(
		[][]byte{
			pow.block.PrevBlockHash,
			pow.block.HashTransactions(),
			IntToBytes(pow.block.Timestamp),
			IntToBytes(int64(targetBits)),
			IntToBytes(nonce),
		},
		[]byte{},
	)
}

var maxNonce int64 = math.MaxInt64

// Run mines the block as a proof of work
func (pow *ProofOfWork) Run() (int64, []byte) {
	var hashInt big.Int
	var hash [32]byte
	var nonce int64

	fmt.Printf("Mining a new block")
	for nonce <= maxNonce {
		data := pow.prepareData(nonce)
		hash = sha256.Sum256(data)
		fmt.Printf("\r%x", hash)
		hashInt.SetBytes(hash[:])
		if hashInt.Cmp(pow.target) == -1 {
			break
		} else {
			nonce++
		}
	}
	fmt.Print("\n\n")

	return nonce, hash[:]
}

// Validate validates the proof of work
func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int

	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	return hashInt.Cmp(pow.target) == -1
}
