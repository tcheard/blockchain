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
	fmt.Println("  addblock -data BLOCK_DATA - add a block to the blockchain")
	fmt.Println("  printchain - print all the blocks of the blockchain")
	fmt.Println("  version - print version info")
}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

func (cli *CLI) addBlock(bc *Blockchain, data string) {
	err := bc.AddBlock(data)
	if err != nil {
		fmt.Println("Failed to add block", err)
		os.Exit(1)
	}

	fmt.Println("Success!")
}

func (cli *CLI) printChain(bc *Blockchain) {
	bci := bc.Iterator()

	for {
		block, err := bci.Next()
		if err != nil {
			fmt.Println("Failed to retrieve next block", err)
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

func printVersion() {
	fmt.Printf("blockchain %s (Git SHA: %s, Go Version: %s)\n", Version, GitSHA, runtime.Version())
}

// Run runs the CLI
func (cli *CLI) Run() {
	cli.validateArgs()

	addBlockCmd := flag.NewFlagSet("addblock", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	versionCmd := flag.NewFlagSet("version", flag.ExitOnError)

	addBlockData := addBlockCmd.String("data", "", "Block data")

	switch os.Args[1] {
	case "addblock":
		if err := addBlockCmd.Parse(os.Args[2:]); err != nil {
			fmt.Println("Failed to parse addblock arguments")
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

	if addBlockCmd.Parsed() {
		if *addBlockData == "" {
			addBlockCmd.Usage()
			os.Exit(1)
		}

		bc, err := NewBlockchain()
		if err != nil {
			log.Fatal("Failed to get blockchain", err)
			os.Exit(1)
		}
		defer bc.db.Close()

		cli.addBlock(bc, *addBlockData)
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