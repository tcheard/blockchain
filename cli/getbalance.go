package cli

import (
	"fmt"
	"os"

	"github.com/tcheard/blockchain/pkg/blockchain"
	"github.com/tcheard/blockchain/pkg/util"
)

func (cli *CLI) getBalance(address string) {
	if !blockchain.ValidateAddress(address) {
		fmt.Printf("Address is not valid")
		os.Exit(1)
	}

	bc, err := blockchain.NewBlockchain()
	if err != nil {
		fmt.Printf("Failed to retrieve blockchain: %v\n", err)
		os.Exit(1)
	}
	defer bc.DB.Close()

	balance := 0
	pubKeyHash := util.Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs, err := bc.FindUTXO(pubKeyHash)
	if err != nil {
		fmt.Printf("Failed to find unspent transaction outputs: %v\n", err)
		os.Exit(1)
	}

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance for '%s': %d\n", address, balance)
}
