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
	host.SetStreamHandler(protocol.ID(config.ProtocolID+"/braodcast/block"), broadcastBlock)

	// Recieve handlers
	host.SetStreamHandler(protocol.ID(config.ProtocolID+"/recieve/block"), blockProtocol)
	host.SetStreamHandler(protocol.ID(config.ProtocolID+"/recieve/transaction"), txnProtocol)

	// Download handlers
	host.SetStreamHandler(protocol.ID(config.ProtocolID+"/download/blockchain"), downloadProtocol)
	host.SetStreamHandler(protocol.ID(config.ProtocolID+"/download/block"), downloadBlock)
	host.SetStreamHandler(protocol.ID(config.ProtocolID+"/download/transaction"), downloadTxn)
}
