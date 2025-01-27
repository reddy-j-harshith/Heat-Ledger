package main

import (
	"bufio"

	"github.com/libp2p/go-libp2p/core/network"
)

func broadcastMessage(stream network.Stream) {

	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	go propagateMessage(rw, stream)

}

func broadcastTxn(stream network.Stream, txn Transaction) {

	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	go propagateTxn(rw, stream)

}

func broadcastBlock(stream network.Stream, block Block) {

	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	go propagateBlock(rw, stream)

}

func messageProtocol(stream network.Stream) {

	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	go readData(rw)
	// go writeData(rw)
}
