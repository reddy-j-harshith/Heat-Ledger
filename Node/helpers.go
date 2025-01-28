package main

import (
	"crypto/sha256"
	"fmt"
	"math"
)

// Creation of a merkle tree and storage
func buildMerkle(txnList []Transaction) string {

	merkleDepth := math.Ceil(math.Log2(float64(len(txnList))))
	// Create the merkle tree
	MerkleRoot := merkleFunc(txnList, 0, 0, int(merkleDepth))

	// Save the merkle root to the database
	MerkleMutex.RLock()
	Merkle_Roots[MerkleRoot.Value] = MerkleRoot
	MerkleMutex.RUnlock()

	return MerkleRoot.Value
}

func merkleFunc(txnList []Transaction, idx int, level int, depth int) *MerkleNode {

	len := int(len(txnList))

	node := &MerkleNode{
		Value: "",
	}

	if len == 0 || idx >= len {
		return node
	}

	// If this not the last level
	if level < depth {

		node.Left = merkleFunc(txnList, (2*idx)+1, level+1, depth)
		node.Right = merkleFunc(txnList, (2*idx)+2, level+1, depth)
		leftData := node.Left.Value
		rightData := node.Right.Value
		hash_data := sha256.Sum256([]byte(leftData + rightData))
		node.Value = fmt.Sprintf("%x", hash_data)
		return node
	}

	txnIdx := idx - (1 << uint(depth)) - 1
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

// Mining of a block
func mineBlock() {

}

// Sending UTXOs
func sendFunds() {

}

// Validation of a block by checking the previous hash, and all the transactions
func validateBlock() {

}

// Download the blockchain when the full node starts up
func downloadBlockchain() {

}

// Remove the confirmed transactions from the mempool
func removeTxns(block Block) {

}

// Compress the merkle tree
func compressMerkle() {

}

// Validate whether the recieved blockchain copy has some inconsistencies
func validateBlockchain() {

}

// Creation of UTXO entries
func storeUTXO() {

}
