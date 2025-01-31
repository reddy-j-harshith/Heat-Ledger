package main

import (
	"github.com/libp2p/go-libp2p/core/protocol"
)

func SetNodeHandlers() {
	// Message handlers
	User.SetStreamHandler(protocol.ID(config.ProtocolID+"/message"), messageProtocol)
	User.SetStreamHandler(protocol.ID(config.ProtocolID+"/gossip"), broadcastMessage)

	// Propagation handlers
	User.SetStreamHandler(protocol.ID(config.ProtocolID+"/broadcast/transaction"), broadcastTxn)
	User.SetStreamHandler(protocol.ID(config.ProtocolID+"/broadcast/block"), broadcastBlock)

	// Download handlers
	User.SetStreamHandler(protocol.ID(config.ProtocolID+"/download/blockchain"), exportBlockchain)
	User.SetStreamHandler(protocol.ID(config.ProtocolID+"/download/block"), exportBlock)
	User.SetStreamHandler(protocol.ID(config.ProtocolID+"/download/transaction"), exportTransaction)
}
