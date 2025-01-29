package main

import (
	"crypto/sha256"
	"fmt"
	"math"
	"strconv"
	"sync"
	"time"
)

type input struct {
	Txn_id string
	Index  int
}

type output struct {
	Value  float64
	Pubkey string
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

type MerkleNode struct {
	Value string
	Left  *MerkleNode
	Right *MerkleNode
}

var (
	Merkle_Roots map[string]*MerkleNode = map[string]*MerkleNode{}
	MerkleMutex  sync.RWMutex
)

func (txn *Transaction) generateTxn() {
	var data string

	for _, input := range txn.Inputs {
		data += input.Txn_id + strconv.Itoa(input.Index)
	}

	for _, output := range txn.Outputs {
		data += fmt.Sprintf("%.8f", output.Value) + output.Pubkey
	}

	data += fmt.Sprintf("%.8f", txn.Fee)
	data += txn.Timestamp.String()

	hash := sha256.Sum256([]byte(data))
	txn.Tnx_id = fmt.Sprintf("%x", hash)
}

func merkleFunc(txnList []Transaction, treeIdx int, level int, depth int) *MerkleNode {
	len := len(txnList)
	node := &MerkleNode{}
	if len == 0 {
		return node
	}

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
		hash_data = sha256.Sum256([]byte(txnList[len-1].Tnx_id))
	} else {
		hash_data = sha256.Sum256([]byte(txnList[txnIdx].Tnx_id))
	}

	node.Value = fmt.Sprintf("%x", hash_data)
	return node
}

func buildMerkle(txnList []Transaction) string {
	if len(txnList) == 0 {
		return ""
	}

	merkleDepth := math.Ceil(math.Log2(float64(len(txnList))))
	MerkleRoot := merkleFunc(txnList, 0, 0, int(merkleDepth))

	MerkleMutex.RLock()
	Merkle_Roots[MerkleRoot.Value] = MerkleRoot
	MerkleMutex.RUnlock()

	return MerkleRoot.Value
}

func displayMerkleTree(node *MerkleNode, level int) {
	if node == nil {
		return
	}
	fmt.Printf("Level %d: %s\n", level, node.Value)
	displayMerkleTree(node.Left, level+1)
	displayMerkleTree(node.Right, level+1)
}

func main() {
	txns := []Transaction{
		{Inputs: []input{{"prev_txn_1", 0}}, Outputs: []output{{10.0, "pubkey1"}}, Fee: 0.01, Timestamp: time.Now()},
		{Inputs: []input{{"prev_txn_2", 1}}, Outputs: []output{{5.5, "pubkey2"}}, Fee: 0.02, Timestamp: time.Now()},
		{Inputs: []input{{"prev_txn_3", 2}}, Outputs: []output{{8.2, "pubkey3"}}, Fee: 0.03, Timestamp: time.Now()},
		{Inputs: []input{{"prev_txn_4", 3}}, Outputs: []output{{6.7, "pubkey4"}}, Fee: 0.04, Timestamp: time.Now()},
		{Inputs: []input{{"prev_txn_5", 4}}, Outputs: []output{{9.3, "pubkey5"}}, Fee: 0.05, Timestamp: time.Now()},
	}

	for i := range txns {
		txns[i].generateTxn()
	}

	rootHash := buildMerkle(txns)

	fmt.Println("Merkle Root:", rootHash)
	fmt.Println("\nDisplaying Merkle Tree:")
	displayMerkleTree(Merkle_Roots[rootHash], 0)

	thirdTxnHash := txns[2].Tnx_id
	fmt.Println("\nThird Transaction Hash:", thirdTxnHash)
	fmt.Println("\nHash of the third txn:", fmt.Sprintf("%x", sha256.Sum256([]byte(thirdTxnHash))))
	valid := false
	for _, node := range Merkle_Roots {
		if node.Value == rootHash {
			valid = true
			break
		}
	}
	fmt.Println("\nMerkle Root Validity:", valid)
}
