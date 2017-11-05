package main

// Blockchain represents the actual blockchain holding all its blocks
type Blockchain struct {
	blocks []*Block
}

// AddBlock adds a block to the Blockchain
func (bc *Blockchain) AddBlock(data string) {
	prevBlock := bc.blocks[len(bc.blocks)-1]
	newBlock := NewBlock(data, prevBlock.Hash)
	bc.blocks = append(bc.blocks, newBlock)
}
