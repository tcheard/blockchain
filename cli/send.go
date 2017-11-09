package cli

import (
	"fmt"
	"os"

	"github.com/tcheard/blockchain/pkg/blockchain"
)

func (cli *CLI) send(from, to string, amount int) {
	if !blockchain.ValidateAddress(from) {
		fmt.Printf("Address is not valid")
		os.Exit(1)
	}
	if !blockchain.ValidateAddress(to) {
		fmt.Printf("Address is not valid")
		os.Exit(1)
	}

	bc, err := blockchain.NewBlockchain()
	if err != nil {
		fmt.Printf("Failed to get blockchain: %v\n", err)
		os.Exit(1)
	}
	defer bc.DB.Close()

	tx, err := blockchain.NewUTXOTransaction(from, to, amount, bc)
	if err != nil {
		fmt.Printf("Failed to create transaction: %v\n", err)
		os.Exit(1)
	}

	err = bc.MineBlock([]*blockchain.Transaction{tx})
	if err != nil {
		fmt.Printf("Failed to mine block: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Success!")
}
