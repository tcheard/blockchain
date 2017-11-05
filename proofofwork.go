package main

import "math/big"

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
