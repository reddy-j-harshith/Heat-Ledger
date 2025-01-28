package main

import (
	"crypto/sha256"
	"fmt"
	"strconv"
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
	value  float64
	pubkey string
}

type input struct {
	txn_id    string
	index     int32
	signature string
}

type output struct {
	value  float64
	pubkey string
}

type Transaction struct {
	tnx_id     string // SHA(SHA(from + to + amount + fee + timestamp))
	block_hash string // Parent block
	in_sz      int32
	out_sz     int32
	fee        float64
	inputs     []input
	outputs    []output
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
	// Concatenations of the left and right nodes
	value string
	left  *MerkleNode
	right *MerkleNode
}

// Leaf nodes will contain the hash of the txid, and left / right are set in NIL

// Hash of the inputs, outputs and the timestamp
func (txn Transaction) generateTxn() {
	var data string

	for _, input := range txn.inputs {
		data += input.txn_id + strconv.Itoa(int(input.index))
	}

	for _, output := range txn.outputs {
		data += fmt.Sprintf("%.8f", output.value) + output.pubkey
	}

	data += fmt.Sprintf("%.8f", txn.fee)
	data += txn.timestamp.String()

	hash := sha256.Sum256([]byte(data))

	txn.tnx_id = fmt.Sprintf("%x", hash)
}
