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

	UTXOSet := blockchain.UTXOSet{
		Blockchain: bc,
	}

	tx, err := blockchain.NewUTXOTransaction(from, to, amount, &UTXOSet)
	if err != nil {
		fmt.Printf("Failed to create UTXO transaction: %v\n", err)
		os.Exit(1)
	}
	cb, err := blockchain.NewCoinbaseTransaction(from, "") // For simplicity make the sender the miner
	if err != nil {
		fmt.Printf("Failed to create coinbase transaction: %v\n", err)
	}

	newBlock, err := bc.MineBlock([]*blockchain.Transaction{tx, cb})
	if err != nil {
		fmt.Printf("Failed to mine block: %v\n", err)
		os.Exit(1)
	}

	UTXOSet.Update(newBlock)

	fmt.Println("Success!")
}
