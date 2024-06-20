package main

import (
	"fmt"
	"log"
	"os"
)

func (cli *CLI) send(from, to string, amount int, nodeID string, mineNow bool) {
	if !ValidateAddress(from) {
		log.Println("ERROR: Sender address is not valid")
		os.Exit(1)
	}
	if !ValidateAddress(to) {
		log.Println("ERROR: Recipient address is not valid")
		os.Exit(1)
	}
	if from == to {

		log.Println("ERROR: Sender and Recipient cannot be same")
		os.Exit(1)
	}
	bc := NewBlockchain(nodeID)
	UTXOSet := UTXOSet{bc}
	defer bc.db.Close()

	wallets, err := NewWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}
	wallet := wallets.GetWallet(from)

	tx := NewUTXOTransaction(&wallet, to, amount, &UTXOSet)

	if mineNow {

		cbTx := NewCoinbaseTX(from, "")
		txs := []*Transaction{cbTx, tx}

		newBlock := bc.MineBlock(txs)
		UTXOSet.Update(newBlock)

	} else {

		sendTx(knownNodes[0], tx)
	}

	fmt.Println("Success!")
}
