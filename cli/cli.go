package cli

import (
	"flag"
	"fmt"
	"os"
)

// CLI provides a handler for the basic CLI
type CLI struct {
}

// Run runs the CLI
func (cli *CLI) Run() {
	cli.validateArgs()

	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	createBlockchainAddress := createBlockchainCmd.String("address", "", "Address")

	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)

	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	getBalanceAddress := getBalanceCmd.String("address", "", "Address")

	listAddressesCmd := flag.NewFlagSet("listaddresses", flag.ExitOnError)

	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)

	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	sendFrom := sendCmd.String("from", "", "Sender Address")
	sendTo := sendCmd.String("to", "", "Receiver Address")
	sendAmount := sendCmd.Int("amount", 0, "Amount being sent")

	versionCmd := flag.NewFlagSet("version", flag.ExitOnError)

	switch os.Args[1] {
	case "createblockchain":
		if err := createBlockchainCmd.Parse(os.Args[2:]); err != nil {
			fmt.Printf("Failed to parse createblockchain arguments")
			os.Exit(1)
		}
	case "createwallet":
		if err := createWalletCmd.Parse(os.Args[2:]); err != nil {
			fmt.Printf("Failed to parse getwallet arguments")
			os.Exit(1)
		}
	case "getbalance":
		if err := getBalanceCmd.Parse(os.Args[2:]); err != nil {
			fmt.Printf("Failed to parse getbalance arguments")
			os.Exit(1)
		}
	case "listaddresses":
		if err := listAddressesCmd.Parse(os.Args[2:]); err != nil {
			fmt.Printf("Failed to parse listaddresses arguments")
			os.Exit(1)
		}
	case "printchain":
		if err := printChainCmd.Parse(os.Args[2:]); err != nil {
			fmt.Printf("Failed to parse printchain arguments")
			os.Exit(1)
		}
	case "send":
		if err := sendCmd.Parse(os.Args[2:]); err != nil {
			fmt.Printf("Failed to parse send arguments")
			os.Exit(1)
		}
	case "version":
		if err := versionCmd.Parse(os.Args[2:]); err != nil {
			fmt.Printf("Failed to parse version arguments")
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

	if createWalletCmd.Parsed() {
		cli.createWallet()
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			os.Exit(1)
		}

		cli.getBalance(*getBalanceAddress)
	}

	if listAddressesCmd.Parsed() {
		cli.listAddresses()
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount == 0 {
			sendCmd.Usage()
			os.Exit(1)
		}

		cli.send(*sendFrom, *sendTo, *sendAmount)
	}

	if versionCmd.Parsed() {
		cli.version()
	}
}

func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  createblockchain -address ADDRESS - Create a blockchain and send genesis block reward to ADDRESS")
	fmt.Println("  createwallet - create a new wallet")
	fmt.Println("  getbalance -address ADDRESS - Get balance of ADDRESS")
	fmt.Println("  listaddresses - get a list of all created wallet addresses")
	fmt.Println("  printchain - Print all the blocks of the blockchain")
	fmt.Println("  send -from FROM -to TO -amount AMOUNT - Send AMOUNT of coins from FROM address to TO")
	fmt.Println("  version - Print version info")
}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}
