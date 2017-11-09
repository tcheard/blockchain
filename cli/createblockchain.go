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
	defer bc.DB.Close()

	UTXOSet := blockchain.UTXOSet{
		Blockchain: bc,
	}
	err = UTXOSet.Reindex()
	if err != nil {
		fmt.Printf("Failed to reindex blockchain: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Done!")
}
