package cli

import (
	"fmt"
	"os"

	"github.com/tcheard/blockchain/pkg/blockchain"
)

func (cli *CLI) createBlockchain(address string) {
	bc, err := blockchain.CreateBlockchain(address)
	if err != nil {
		fmt.Printf("Failed to create blockchain: %v\n", err)
		os.Exit(1)
	}
	bc.DB.Close()

	fmt.Println("Done!")
}
