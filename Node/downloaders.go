package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/libp2p/go-libp2p/core/network"
)

// Send the blockchain when the full node starts up
func downloadBlockchain(rw *bufio.ReadWriter, strm network.Stream) {

	// Read the latest block of the remote peer
	data, _ := rw.ReadString('\n')

	remote_block := Block{}

	json.Unmarshal([]byte(data), &remote_block)

	// Check if the remote peer has a longer blockchain
	if len(Blockchain) < int(remote_block.Block_height) {
		rw.WriteString("Shorter. Try from others\n")
		rw.Flush()
		return
	}

	// Get the block height of the remote peer
	remote_height := remote_block.Block_height

	// Send the length of the blockchain
	_, err := rw.WriteString(strconv.Itoa(len(Blockchain)) + "\n")
	rw.Flush()
	if err != nil {
		fmt.Println("Failed to send blockchain length:", err)
		return
	}

	currentBlock := Blockchain[Latest_Block]
	// Send the blockchain from the remote peer
	for i := int(remote_height); i < len(Blockchain); i++ {
		// Convert to JSON
		blockJSON, _ := json.Marshal(currentBlock)
		// Send the block
		rw.WriteString(string(blockJSON) + "\n")
		rw.Flush()
		// Move to the next Block
		currentBlock = Blockchain[currentBlock.Block_hash]
	}

	fmt.Println("Blockchain sent to", strm.Conn().RemotePeer())
}

// Send the Mempool
func downloadMempool(rw *bufio.ReadWriter, strm network.Stream) {

	// Send the length of the Mempool
	_, err := rw.WriteString(strconv.Itoa(len(Mempool)) + "\n")
	rw.Flush()
	if err != nil {
		fmt.Println("Failed to send Mempool length:", err)
		return
	}

	// Send the Mempool
	for _, txn := range Mempool {
		// Convert to JSON
		txnJSON, _ := json.Marshal(txn)
		// Send the transaction
		rw.WriteString(string(txnJSON) + "\n")
		rw.Flush()
	}

	fmt.Println("Mempool sent to", strm.Conn().RemotePeer())
}

// Send a block
func downloadBlock(rw *bufio.ReadWriter, strm network.Stream) {

}

// Send a transaction
func downloadTransaction(rw *bufio.ReadWriter, strm network.Stream) {

}
