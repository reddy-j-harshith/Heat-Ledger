package main

import (
	"crypto/sha256"
	"fmt"
	"math"
	"time"
)

// Transaction definition
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

// Input and Output structure
type input struct {
	Txn_id    string `json:"txn_id"`
	Index     int32  `json:"index"`
	Signature string `json:"sign"`
}

type output struct {
	Pubkey string  `json:"pubkey"`
	Value  float64 `json:"amount"`
}

// MerkleNode definition
type MerkleNode struct {
	Value string
	Left  *MerkleNode
	Right *MerkleNode
}

// In-memory storage
var Merkle_Roots map[string]*MerkleNode = map[string]*MerkleNode{}
var Transactions map[string]Transaction = map[string]Transaction{}

// Dummy function to build the Merkle Tree
func buildMerkle(txnList []Transaction) string {
	if len(txnList) == 0 {
		return ""
	}

	// Calculate the depth of the tree

	merkleDepth := math.Ceil(math.Log2(float64(len(txnList))))

	// Create the Merkle tree
	MerkleRoot := merkleFunc(txnList, 0, 0, int(merkleDepth))

	// Save the Merkle root to the database (simulation here)
	Merkle_Roots[MerkleRoot.Value] = MerkleRoot

	return MerkleRoot.Value
}

// Merkle Tree Construction Function
func merkleFunc(txnList []Transaction, treeIdx int, level int, depth int) *MerkleNode {

	len := int(len(txnList))

	node := &MerkleNode{}

	if len == 0 {
		return node
	}

	// If this is not the last level
	if level < depth {

		node.Left = merkleFunc(txnList, (2*treeIdx)+1, level+1, depth)
		node.Right = merkleFunc(txnList, (2*treeIdx)+2, level+1, depth)
		leftData := node.Left.Value
		rightData := node.Right.Value
		hash_data := sha256.Sum256([]byte(leftData + rightData))
		node.Value = fmt.Sprintf("%x", hash_data)
		return node
	}

	// Get the leaf node from transactions
	txnIdx := treeIdx - (1 << depth) + 1
	var hash_data [32]byte

	if txnIdx >= len {
		// Padding the remaining transactions
		hash_data = sha256.Sum256([]byte(txnList[len-1].Tnx_id))
	} else {
		// If this is a leaf node
		hash_data = sha256.Sum256([]byte(txnList[txnIdx].Tnx_id))
	}

	node.Value = fmt.Sprintf("%x", hash_data)
	return node
}

// Function to create dummy transactions
func createDummyTransactions() []Transaction {
	transactions := []Transaction{
		{Tnx_id: "txn_001", Block_hash: "block_001", In_sz: 2, Out_sz: 3, Fee: 0.1, Timestamp: time.Now()},
		{Tnx_id: "txn_002", Block_hash: "block_002", In_sz: 1, Out_sz: 2, Fee: 0.05, Timestamp: time.Now()},
		{Tnx_id: "txn_003", Block_hash: "block_003", In_sz: 3, Out_sz: 2, Fee: 0.15, Timestamp: time.Now()},
		{Tnx_id: "txn_004", Block_hash: "block_004", In_sz: 2, Out_sz: 4, Fee: 0.2, Timestamp: time.Now()},
	}

	return transactions
}

func main() {
	// Create dummy transactions
	txnList := createDummyTransactions()

	// Build the Merkle tree and get the Merkle Root
	merkleRoot := buildMerkle(txnList)
	fmt.Printf("Merkle Root: %s\n", merkleRoot)

	// Print the saved Merkle Roots
	fmt.Println("Stored Merkle Roots:")
	for root, node := range Merkle_Roots {
		fmt.Printf("Root: %s, Node Value: %s\n", root, node.Value)
	}
}
