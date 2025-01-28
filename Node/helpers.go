package main

import (
	"crypto/sha256"
	"fmt"
)

// Creation of a merkle tree and storage - <= 8 transactions
func buildMerkle(txnList []Transaction) string {

	MerkleRoot := merkleFunc(txnList, 0)

	// Save the node
	Merkle_Roots[MerkleRoot.value] = MerkleRoot

	return MerkleRoot.value
}

func merkleFunc(txnList []Transaction, idx int) *MerkleNode {
	len := len(txnList)
	if idx >= len {
		return nil
	}

	data := txnList[idx].tnx_id

	node := &MerkleNode{
		left:  nil,
		right: nil,
		value: fmt.Sprintf("%x", sha256.Sum256([]byte(data))),
	}

	if (2*idx)+1 >= len || (2*idx)+2 >= len {
		return node
	}

	node.left = merkleFunc(txnList, (2*idx)+1)
	node.right = merkleFunc(txnList, (2*idx)+2)

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
