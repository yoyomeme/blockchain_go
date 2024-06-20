package main

import (
	"fmt"
	"log"
	"os"
)

func (cli *CLI) createBlockchain(address, nodeID string) {
	if !ValidateAddress(address) {
		log.Println("ERROR: Address is not valid")
		os.Exit(1)
	}
	bc := CreateBlockchain(address, nodeID)
	defer bc.db.Close()

	UTXOSet := UTXOSet{bc}
	UTXOSet.Reindex()

	fmt.Println("Done!")
}
