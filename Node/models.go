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
	Txn_id    string `json:"txn_id"`
	Index     int32  `json:"index"`
	Signature string `json:"sign"`
}

type output struct {
	Pubkey string  `json:"pubkey"`
	Value  float64 `json:"amount"`
}

type Transaction struct {
	Tnx_id     string    `json:"txn_id"`
	Block_hash string    `json:"block_hash"`
	In_sz      int32     `json:"in_sz"`
	Out_sz     int32     `json:"out_sz"`
	Fee        float64   `json:"fee"`
	Inputs     []input   `json:"inputs"`
	Outputs    []output  `json:"outputs"`
	Timestamp  time.Time `json:"timestamp"`
}

type Block struct {
	Block_hash    string        `json:"block_hash"`
	Block_height  int32         `json:"block_height"`
	Previous_hash string        `json:"previous_hash"`
	Nonce         int32         `json:"nonce"`
	Difficulty    int32         `json:"difficulty"`
	Merkle_hash   string        `json:"merkle_hash"` // Obtain the merkle root from the target node and send the list of transactions to build the local merkle tree
	Timestamp     time.Time     `json:"timestamp"`
	Transactions  []Transaction `json:"transactions"`
}

type MerkleNode struct {
	// Concatenations of the left and right nodes
	Value string
	Left  *MerkleNode
	Right *MerkleNode
}

// Leaf nodes will contain the hash of the txid, and left / right are set in NIL

// Hash of the inputs, outputs and the timestamp
func (txn *Transaction) generateTxn() {
	var data string

	for _, input := range txn.Inputs {
		data += input.Txn_id + strconv.Itoa(int(input.Index))
	}

	for _, output := range txn.Outputs {
		data += fmt.Sprintf("%.8f", output.Value) + output.Pubkey
	}

	data += fmt.Sprintf("%.8f", txn.Fee)
	data += txn.Timestamp.String()

	hash := sha256.Sum256([]byte(data))

	txn.Tnx_id = fmt.Sprintf("%x", hash)
}
