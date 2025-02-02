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
		return
	}

	// Get the block height of the remote peer
	remote_height := remote_block.Block_height

	BlockMutex.RLock()

	// Send the length of the blockchain
	_, err := rw.WriteString(strconv.Itoa(len(Blockchain)) + "\n")
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

	defer BlockMutex.RUnlock()

}

// Send a block
func downloadBlock(rw *bufio.ReadWriter, strm network.Stream) {

}

// Send a transaction
func downloadTransaction(rw *bufio.ReadWriter, strm network.Stream) {

}
