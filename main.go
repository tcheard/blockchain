package main

import (
	"log"
	"os"
)

func addBlock(bc *Blockchain, data string) {
	err := bc.AddBlock(data)
	if err != nil {
		log.Fatal("Failed to add block")
		os.Exit(1)
	}
}

func main() {
	cli := CLI{}
	cli.Run()
}
