package main

import (
	"context"
	"sync"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

// Global variables
var (
	// Functional variables
	User        host.Host                                             // Current User Node
	config      Config                                                // Configuration
	peerMutex   sync.RWMutex                                          // Mutex for the message database
	kademliaDHT *dht.IpfsDHT                                          // Local DHT
	peerArray   []peer.AddrInfo          = []peer.AddrInfo{}          // Array of neighbors
	peerSet     map[string]peer.AddrInfo = map[string]peer.AddrInfo{} // Set of neighbors

	// Message database
	m_id     int32                       = 1
	least    map[string]int32            = map[string]int32{}
	database map[string]map[int32]string = map[string]map[int32]string{}

	// Blockchain database
	Genesis_Block string                 = ""
	Latest_Block  string                 = ""
	Mempool       map[string]Transaction = map[string]Transaction{}
	UTXO_SET      map[string]UTXO        = map[string]UTXO{}
	Blockchain    map[string]Block       = map[string]Block{}
	Merkle_Roots  map[string]*MerkleNode = map[string]*MerkleNode{}
	Transactions  map[string]Transaction = map[string]Transaction{}

	// Mutex for the respective Databases
	MempoolMutex sync.RWMutex // Mutex for the mempool
	UTXOMutex    sync.RWMutex // Mutex for the UTXO set
	BlockMutex   sync.RWMutex // Mutex for the blockchain
	MerkleMutex  sync.RWMutex // Mutex for the Merkle roots
	TransMutex   sync.RWMutex // Mutex for the transactions

	miningCtx    context.Context
	miningCancel context.CancelFunc
)
