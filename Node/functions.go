package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/libp2p/go-libp2p/core/protocol"
)

// Sending UTXOs
func sendFunds(utxo []string, nodeID []string, amount []float64, fee float64) error {

	inputs := make([]Input, len(utxo))
	outputs := make([]Output, len(nodeID))

	for idx, utxoHash := range utxo {

		UTXOMutex.RLock()
		utxoInput := UTXO_SET[utxoHash]
		UTXOMutex.RUnlock()

		inputs[idx] = Input{
			Txn_id: utxoInput.Txn_id,
			Index:  int32(utxoInput.Value),
		}
	}

	for idx, key := range nodeID {
		outputs[idx] = Output{
			Pubkey: key,
			Value:  amount[idx],
		}
	}

	transaction := &Transaction{
		In_sz:     int32(len(inputs)),
		Out_sz:    int32(len(outputs)),
		Fee:       fee,
		Inputs:    inputs,
		Outputs:   outputs,
		Timestamp: time.Now(),
	}

	transaction.generateTxn()

	peerMutex.RLock()
	peers := peerArray
	peerMutex.RUnlock()

	success := false

	for _, peer := range peers {
		stream, err := User.NewStream(
			context.Background(),
			peer.ID,
			protocol.ID(config.ProtocolID+"/broadcast/transaction"),
		)

		if err != nil {
			fmt.Println("Failed to create a stream:", err)
			continue
		}

		rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

		data, err := json.Marshal(transaction)
		if err != nil {
			fmt.Println("Failed to serialize message:", err)
			stream.Close()
			continue
		}

		_, err = rw.WriteString(string(data) + "\n")
		if err != nil {
			fmt.Println("Failed to send message to peer:", peer.ID, err)
			stream.Close()
			continue
		}

		// Flush the buffer to ensure data is sent
		err = rw.Flush()
		if err != nil {
			fmt.Println("Failed to flush data to peer:", peer.ID, err)
			stream.Close()
			continue
		}
		stream.Close()
		success = true
	}

	if !success {
		return fmt.Errorf("failed to communicate with any peers")
	}

	// Add the transaction to the mempool
	MempoolMutex.Lock()
	if _, exists := Mempool[transaction.Txn_id]; !exists {
		Mempool[transaction.Txn_id] = *transaction
	}
	MempoolMutex.Unlock()

	return nil
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
		handleUTXO(txn)

		MempoolMutex.Lock()
		delete(Mempool, txn.Txn_id)
		MempoolMutex.Unlock()
	}
}

// Add and Remove UTXOs
func handleUTXO(txn Transaction) {
	for _, input := range txn.Inputs {
		// Remove the UTXO from the UTXO set
		utxoHashInput := sha256.Sum256([]byte(fmt.Sprintf("%s:%d", input.Txn_id, input.Index)))

		UTXOMutex.Lock()
		delete(UTXO_SET, hex.EncodeToString(utxoHashInput[:]))
		UTXOMutex.Unlock()
	}

	for idx, output := range txn.Outputs {
		// Add the new UTXO into the UTXO set
		utxoHashOutput := sha256.Sum256([]byte(fmt.Sprintf("%s:%d", txn.Txn_id, idx)))

		UTXOMutex.Lock()
		UTXO_SET[hex.EncodeToString(utxoHashOutput[:])] = UTXO{txn.Txn_id, int32(idx), output.Value, output.Pubkey}
		UTXOMutex.Unlock()
	}
}
func displayMempool() {
	MempoolMutex.RLock()
	fmt.Println("Mempool:")
	for _, txn := range Mempool {
		fmt.Println(txn)
	}
	MempoolMutex.RUnlock()
}

// TODO
func startUp() {

	// Create the genesis block
	// Download the UTXO set
	// Download the blockchain
	// Download the mempool
	// Download the transactions
	// Create the merkle roots

}

func makeBlockchain(blockchain []Block) {
	BlockMutex.Lock()
	for _, block := range blockchain {
		Blockchain[block.Block_hash] = block

		// Create the UTXOs
		for _, txn := range block.Transactions {
			handleUTXO(txn)
		}

	}
	BlockMutex.Unlock()

}

func displayBlockchain() {
	block := Latest_Block
	BlockMutex.RLock()
	i := 0
	for block != Genesis_Block {
		if i == 3 {
			break
		}
		fmt.Println(Blockchain[block])
		block = Blockchain[block].Previous_hash
		i++
	}
	BlockMutex.RUnlock()
}

// Mining of a block
func mineBlock() {

}

// Validate whether the recieved blockchain copy has some inconsistencies
func validateBlockchain() {

}

// Validation of a block by checking the previous hash, and all the transactions
func validateBlock(block Block) error {
	// Check if the previous hash is correct
	BlockMutex.RLock()
	previousBlock, exists := Blockchain[block.Previous_hash]
	BlockMutex.RUnlock()

	if previousBlock.Block_hash != block.Previous_hash ||
		!exists ||
		block.Block_height != previousBlock.Block_height+1 {
		return fmt.Errorf("previous hash does not match")
	}

	// TODO: Validate the transactions

	return nil
}
