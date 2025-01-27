package main

import (
	"sync"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

// Local database
var (
	// Set the user
	User host.Host

	// Local DHT
	kademliaDHT *dht.IpfsDHT

	// Mutex for the message database
	peerMutex sync.RWMutex

	// Maintain a set of neighbors
	peerArray []peer.AddrInfo  = []peer.AddrInfo{}
	peerSet   map[peer.ID]bool = map[peer.ID]bool{}

	// Message database
	m_id     int32                       = 1
	least    map[string]int32            = map[string]int32{}
	database map[string]map[int32]string = map[string]map[int32]string{}

	// Blockchain database
	Blockchain   map[string]Block       = map[string]Block{}
	Merkle_Roots map[string]MerkleNode  = map[string]MerkleNode{}
	Transactions map[string]Transaction = map[string]Transaction{}
)
