package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
)

func startUp() {

}

// Mining of a block
func mineBlock() {

}

// Sending UTXOs
func sendFunds() {

	// Mention the amount to be sent
	// Mention the public key of the receiver
	// Sign the transaction

}

// Validation of a block by checking the previous hash, and all the transactions
func validateBlock() {

}

// Validate whether the recieved blockchain copy has some inconsistencies
func validateBlockchain() {

}

// Download the blockchain when the full node starts up
func downloadBlockchain() {

}

// Compress the merkle tree
func compressMerkle() {

}

// Creation of a merkle tree and storage
func buildMerkle(txnList []Transaction) string {
	if len(txnList) == 0 {
		return ""
	}

	// Calculate the depth of the tree
	merkleDepth := math.Ceil(math.Log2(float64(len(txnList))))
	// Create the merkle tree
	MerkleRoot := merkleFunc(txnList, 0, 0, int(merkleDepth))

	// Save the merkle root to the database
	MerkleMutex.RLock()
	Merkle_Roots[MerkleRoot.Value] = MerkleRoot
	MerkleMutex.RUnlock()

	return MerkleRoot.Value
}

func merkleFunc(txnList []Transaction, treeIdx int, level int, depth int) *MerkleNode {

	len := int(len(txnList))

	node := &MerkleNode{}

	if len == 0 {
		return node
	}

	// If this not the last level
	if level < depth {

		node.Left = merkleFunc(txnList, (2*treeIdx)+1, level+1, depth)
		node.Right = merkleFunc(txnList, (2*treeIdx)+2, level+1, depth)
		leftData := node.Left.Value
		rightData := node.Right.Value
		hash_data := sha256.Sum256([]byte(leftData + rightData))
		node.Value = fmt.Sprintf("%x", hash_data)
		return node
	}

	txnIdx := treeIdx - (1 << uint(depth)) + 1
	var hash_data [32]byte

	if txnIdx >= len {
		// Padding the remaining transactions
		hash_data = sha256.Sum256([]byte(txnList[len-1].Txn_id))
	} else {
		// If this is a leaf node
		hash_data = sha256.Sum256([]byte(txnList[txnIdx].Txn_id))
	}

	node.Value = fmt.Sprintf("%x", hash_data)
	return node
}

// Remove the confirmed transactions from the mempool
func removeTxns(block Block) {
	if len(block.Transactions) == 0 {
		return
	}

	for _, txn := range block.Transactions {
		delete(Mempool, txn.Txn_id)
	}
}

// Create UTXO
func handleUTXO(txn Transaction) {
	for _, input := range txn.Inputs {
		// Remove the UTXO from the UTXO set
		utxoHashInput := sha256.Sum256([]byte(fmt.Sprintf("%s:%d", input.Txn_id, input.Index)))
		delete(UTXO_SET, hex.EncodeToString(utxoHashInput[:]))
	}

	for idx, output := range txn.Outputs {
		// Add the new UTXO into the UTXO set
		utxoHashOutput := sha256.Sum256([]byte(fmt.Sprintf("%s:%d", txn.Txn_id, idx)))
		UTXO_SET[hex.EncodeToString(utxoHashOutput[:])] = UTXO{txn.Txn_id, int32(idx), output.Value, output.Pubkey}
	}
}
