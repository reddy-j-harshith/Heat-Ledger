package main

import (
	"time"
)

type Message struct {
	Sender     string `json:"sender"`
	Message_Id int32  `json:"m_id"`
	Content    string `json:"content"`
}

type UTXO struct {
	txn_id string
	index  int32
	amount float64
	owner  string
}

type Transaction struct {
	tnx_id     string // SHA(SHA(from + to + amount + fee + timestamp))
	block_hash string // Parent block
	in_sz      int32
	out_sz     int32
	fee        float64
	inputs     []UTXO
	outputs    []UTXO
	timestamp  time.Time
}

type Block struct {
	block_hash    string // Primary Key
	block_height  int32
	previous_hash string
	nonce         int32
	difficulty    int32
	merkle_hash   string // Obtain the merkle root from the target node and send the list of transactions to build the local merkle tree
	timestamp     time.Time
}

type MerkleNode struct {
	value string // Concatenations of the left and right nodes
	left  *MerkleNode
	right *MerkleNode
	// Leaf nodes will contain the hash of the txid, and left / right are set in NIL
}
