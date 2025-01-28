package main

import (
	"sync"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

// Global variables
var (
	// Functional variables
	User        host.Host                             // Current User Node
	peerSet     map[peer.ID]bool = map[peer.ID]bool{} // Set of neighbors
	peerMutex   sync.RWMutex                          // Mutex for the message database
	peerArray   []peer.AddrInfo  = []peer.AddrInfo{}  // Array of neighbors
	kademliaDHT *dht.IpfsDHT                          // Local DHT

	// Message database
	m_id     int32                       = 1
	least    map[string]int32            = map[string]int32{}
	database map[string]map[int32]string = map[string]map[int32]string{}

	// Blockchain database
	Mempool      []Transaction          = []Transaction{}
	UTXO_SET     map[string]UTXO        = map[string]UTXO{}
	Blockchain   map[string]Block       = map[string]Block{}
	Merkle_Roots map[string]*MerkleNode = map[string]*MerkleNode{}
	Transactions map[string]Transaction = map[string]Transaction{}
)
