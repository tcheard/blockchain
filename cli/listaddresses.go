package cli

import (
	"fmt"

	"github.com/tcheard/blockchain/pkg/blockchain"
)

func (cli *CLI) listAddresses() {
	wallets, err := blockchain.NewWallets()
	if err != nil {
		fmt.Printf("Failed to retrieve wallets: %v\n", err)
	}

	addresses := wallets.GetAddresses()
	for _, address := range addresses {
		fmt.Println(address)
	}
}
