package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
)

// CLI provides a handler for the basic CLI
type CLI struct {
}

func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  createblockchain -address ADDRESS - create a new blockchain")
	fmt.Println("  getbalance -address ADDRESS - get an address' balance")
	fmt.Println("  printchain - print all the blocks of the blockchain")
	fmt.Println("  version - print version info")
}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

func (cli *CLI) createBlockchain(address string) {
	bc, err := CreateBlockchain(address)
	if err != nil {
		fmt.Printf("Failed to create blockchain: %v\n", err)
		os.Exit(1)
	}
	bc.db.Close()

	fmt.Println("Done!")
}

func (cli *CLI) getBalance(address string) {
	bc, err := NewBlockchain()
	if err != nil {
		fmt.Printf("Failed to retrieve blockchain: %v\n", err)
		os.Exit(1)
	}
	defer bc.db.Close()

	balance := 0
	UTXOs, err := bc.FindUnspentTransactionOutputs(address)
	if err != nil {
		fmt.Printf("Failed to find unspent transaction outputs: %v\n", err)
	}

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance for '%s': %d\n", address, balance)
}

func (cli *CLI) printChain(bc *Blockchain) {
	bci := bc.Iterator()

	for {
		block, err := bci.Next()
		if err != nil {
			fmt.Printf("Failed to retrieve next block: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Hash: %x\n", block.Hash)
		pow := NewProofOfWork(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

func printVersion() {
	fmt.Printf("blockchain %s (Git SHA: %s, Go Version: %s)\n", Version, GitSHA, runtime.Version())
}

// Run runs the CLI
func (cli *CLI) Run() {
	cli.validateArgs()

	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	versionCmd := flag.NewFlagSet("version", flag.ExitOnError)

	createBlockchainAddress := createBlockchainCmd.String("address", "", "Address")
	getBalanceAddress := getBalanceCmd.String("address", "", "Address")

	switch os.Args[1] {
	case "createblockchain":
		if err := createBlockchainCmd.Parse(os.Args[2:]); err != nil {
			fmt.Println("Failed to parse createblockchain arguments")
			os.Exit(1)
		}
	case "getbalance":
		if err := getBalanceCmd.Parse(os.Args[2:]); err != nil {
			fmt.Println("Failed to parse getbalance arguments")
			os.Exit(1)
		}
	case "printchain":
		if err := printChainCmd.Parse(os.Args[2:]); err != nil {
			fmt.Println("Failed to parse printchain arguments")
			os.Exit(1)
		}
	case "version":
		if err := versionCmd.Parse(os.Args[2:]); err != nil {
			fmt.Println("Failed to parse version arguments")
			os.Exit(1)
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			os.Exit(1)
		}

		cli.createBlockchain(*createBlockchainAddress)
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			os.Exit(1)
		}

		cli.getBalance(*getBalanceAddress)
	}

	if printChainCmd.Parsed() {
		bc, err := NewBlockchain()
		if err != nil {
			log.Fatal("Failed to get blockchain", err)
			os.Exit(1)
		}
		defer bc.db.Close()

		cli.printChain(bc)
	}

	if versionCmd.Parsed() {
		printVersion()
	}
}
