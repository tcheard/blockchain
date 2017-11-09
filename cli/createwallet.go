package cli

import (
	"fmt"
	"os"

	"github.com/tcheard/blockchain/pkg/blockchain"
)

func (cli *CLI) createWallet() {
	wallets, _ := blockchain.NewWallets()
	address, err := wallets.CreateWallet()
	if err != nil {
		fmt.Printf("Failed to create wallet: %v\n", err)
		os.Exit(1)
	}
	if err = wallets.SaveToFile(); err != nil {
		fmt.Printf("Failed to save wallets: %v\n", err)
	}

	fmt.Printf("Your new address: %s\n", address)
}
