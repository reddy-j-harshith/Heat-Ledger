package main

import (
	"bufio"

	"github.com/libp2p/go-libp2p/core/network"
)

// Messaging Stream Handlers
func messageProtocol(stream network.Stream) {
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	go readData(rw)
}

// Propagation handlers
func broadcastMessage(stream network.Stream) {
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	go propagateMessage(rw, stream)
}

func broadcastTxn(stream network.Stream) {
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	go propagateTxn(rw, stream)
}

func broadcastBlock(stream network.Stream) {
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	go propagateBlock(rw, stream)
}

// Export Request handlers
func exportBlockchain(stream network.Stream) {
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	downloadBlockchain(rw, stream)
}

func exportMempool(stream network.Stream) {
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	downloadMempool(rw, stream)
}

func exportBlock(stream network.Stream) {
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	go downloadBlock(rw, stream)
}

func exportTransaction(stream network.Stream) {
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	go downloadTransaction(rw, stream)
}
