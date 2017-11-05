package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

func AddBlock(bc *Blockchain, data string) {
	err := bc.AddBlock(data)
	if err != nil {
		log.Fatal("Failed to add block")
		os.Exit(1)
	}
}

func main() {
	bc, err := NewBlockchain()
	if err != nil {
		log.Fatal("Failed to get blockchain", err)
		os.Exit(1)
	}

	AddBlock(bc, "Send 1 BTC to Nikkii")
	AddBlock(bc, "Send another 2 BTC to Nikkii")

	iter := bc.Iterator()

	for {
		block, err := iter.Next()
		if err != nil {
			log.Fatal("Failed to retrieve next block", err)
			os.Exit(1)
		}

		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		pow := NewProofOfWork(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}
