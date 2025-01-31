package main

import (
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/protocol"
)

func SetNodeHandlers(host host.Host, config Config) {
	// Message handlers
	host.SetStreamHandler(protocol.ID(config.ProtocolID+"/message"), messageProtocol)
	host.SetStreamHandler(protocol.ID(config.ProtocolID+"/gossip"), broadcastMessage)

	// Propagation handlers
	host.SetStreamHandler(protocol.ID(config.ProtocolID+"/broadcast/transaction"), broadcastTxn)
	host.SetStreamHandler(protocol.ID(config.ProtocolID+"/broadcast/block"), broadcastBlock)

	// Download handlers
	host.SetStreamHandler(protocol.ID(config.ProtocolID+"/download/blockchain"), exportBlockchain)
	host.SetStreamHandler(protocol.ID(config.ProtocolID+"/download/block"), exportBlock)
	host.SetStreamHandler(protocol.ID(config.ProtocolID+"/download/transaction"), exportTransaction)
}
