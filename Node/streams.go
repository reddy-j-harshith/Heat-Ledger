package main

import (
	"bufio"

	"github.com/libp2p/go-libp2p/core/network"
)

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

// Messaging Stream Handlers
func messageProtocol(stream network.Stream) {

	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	go readData(rw)
	// go writeData(rw)
}

// Download Request handlers
func downloadProtocol(stream network.Stream) {

}

func downloadBlock(stream network.Stream) {

}

func downloadTxn(stream network.Stream) {

}

// Reciever Stream handlers
func blockProtocol(stream network.Stream) {

}

func txnProtocol(stream network.Stream) {

}
