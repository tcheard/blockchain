package cli

import (
	"fmt"
	"os"
	"strconv"

	"github.com/tcheard/blockchain/pkg/blockchain"
)

func (cli *CLI) printChain() {
	bc, err := blockchain.NewBlockchain()
	if err != nil {
		fmt.Printf("Failed to get blockchain: %v\n", err)
		os.Exit(1)
	}
	defer bc.DB.Close()

	bci := bc.Iterator()

	for {
		block, err := bci.Next()
		if err != nil {
			fmt.Printf("Failed to retrieve next block: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Hash: %x\n", block.Hash)
		pow := blockchain.NewProofOfWork(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}
