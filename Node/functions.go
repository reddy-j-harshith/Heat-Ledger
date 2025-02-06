package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
)

// Creation of a merkle tree and storage
func buildMerkle(txnList []Transaction) *MerkleNode {
	if len(txnList) == 0 {
		return &MerkleNode{}
	}

	// Calculate the depth of the tree
	merkleDepth := math.Ceil(math.Log2(float64(len(txnList))))
	// Create the merkle tree
	MerkleRoot := merkleFunc(txnList, 0, 0, int(merkleDepth))

	return MerkleRoot
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

// Sending UTXOs - Create the transaction and add it to the mempool
func sendFunds(utxo []string, nodeID []string, amount []float64, fee float64) (string, error) {

	inputs := make([]Input, len(utxo))
	outputs := make([]Output, len(nodeID))

	// Create inputs of the transaction
	inputSum := 0.0
	for idx, utxoHash := range utxo {

		UTXOMutex.RLock()
		utxoInput := UTXO_SET[utxoHash]
		UTXOMutex.RUnlock()

		inputs[idx] = Input{
			Txn_id: utxoInput.Txn_id,
			Index:  int32(utxoInput.Value),
		}
		inputSum += utxoInput.Value
	}

	// Create outputs of the transaction
	outputSum := 0.0
	for idx, key := range nodeID {
		outputs[idx] = Output{
			Pubkey: key,
			Value:  amount[idx],
		}
		outputSum += amount[idx]
	}

	if outputSum+fee > inputSum {
		return "", fmt.Errorf("output sum and fee is greater than the input")
	}

	// Create another output for the change
	if inputSum-outputSum-fee > 0 {
		outputs = append(outputs, Output{
			Pubkey: User.ID().String(),
			Value:  inputSum - outputSum - fee,
		})
	}

	// Create the transaction
	transaction := &Transaction{
		In_sz:     int32(len(inputs)),
		Out_sz:    int32(len(outputs)),
		Fee:       fee,
		Inputs:    inputs,
		Outputs:   outputs,
		Timestamp: time.Now(),
	}

	transaction.generateTxn()

	// Obtain the peers available
	peerMutex.RLock()
	peers := peerArray
	peerMutex.RUnlock()

	success := false

	// Broadcast the transactions to the peers
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
		return "", fmt.Errorf("failed to communicate with any peers")
	}

	// Add the transaction to the mempool
	MempoolMutex.Lock()
	if _, exists := Mempool[transaction.Txn_id]; !exists {
		Mempool[transaction.Txn_id] = *transaction
	}
	MempoolMutex.Unlock()

	return transaction.Txn_id, nil
}

// Remove the confirmed transactions from the mempool - typically called after mining a block
func removeFromMempool(block Block) {
	if len(block.Transactions) == 0 {
		return
	}

	for _, txn := range block.Transactions {
		// Add the transaction to Transaction database
		TransMutex.Lock()
		Transactions[txn.Txn_id] = txn
		TransMutex.Unlock()

		// Handle the UTXOs, creating and destorying the UTXOs
		go handleUTXO(&txn)

		// Remove the transaction from the mempool
		MempoolMutex.Lock()
		delete(Mempool, txn.Txn_id)
		MempoolMutex.Unlock()
	}
}

// Add and Remove UTXOs
func handleUTXO(txn *Transaction) {

	// Destory the UTXOs
	for _, input := range txn.Inputs {
		utxoHashInput := sha256.Sum256([]byte(fmt.Sprintf("%s:%d", input.Txn_id, input.Index)))

		UTXOMutex.Lock()
		delete(UTXO_SET, hex.EncodeToString(utxoHashInput[:]))
		UTXOMutex.Unlock()
	}

	// Create the UTXOS
	for idx, output := range txn.Outputs {
		// Add the new UTXO into the UTXO set
		utxoHashOutput := sha256.Sum256([]byte(fmt.Sprintf("%s:%d", txn.Txn_id, idx)))

		UTXOMutex.Lock()
		UTXO_SET[hex.EncodeToString(utxoHashOutput[:])] = UTXO{txn.Txn_id, int32(idx), output.Value, output.Pubkey}
		UTXOMutex.Unlock()
	}
}

// Display Mempool
func displayMempool() {
	MempoolMutex.RLock()
	fmt.Println("Mempool:")
	for _, txn := range Mempool {
		fmt.Println(txn)
	}
	MempoolMutex.RUnlock()
}

// Display the blockchain upto three blocks
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

// Update the Mempool
func updateMempool(peer peer.AddrInfo) error {

	stream, err := User.NewStream(context.Background(), peer.ID, protocol.ID(config.ProtocolID+"/download/mempool"))
	if err != nil {
		return fmt.Errorf("failed to create stream with peer: %s", peer.String())
	}

	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	// Recieve the length of the mempool
	mempoolLength, _ := rw.ReadString('\n')
	mempoolLength = strings.TrimSpace(mempoolLength)
	mempoolLengthInt, _ := strconv.Atoi(mempoolLength)

	for i := 0; i < mempoolLengthInt; i++ {
		// Read the transaction from the peer
		txnData, _ := rw.ReadString('\n')

		// Save them to the mempool
		var txn Transaction
		err = json.Unmarshal([]byte(txnData), &txn)
		if err != nil {
			fmt.Println("Failed to parse transaction data:", err)
			continue
		}

		// Add the transaction to the mempool
		MempoolMutex.Lock()
		Mempool[txn.Txn_id] = txn
		MempoolMutex.Unlock()
	}

	return nil
}

// Sync blockchain
func syncBlockchain(randomPeer peer.AddrInfo) error {
	// Update the Mempool
	go updateMempool(randomPeer)

	// Communicate with remote peer to obtain the new blocks
	blocks, err := getBlockchain(randomPeer)
	if err != nil {
		fmt.Println("Failed to update blockchain:", err)
		return err
	}

	// Create the blockchain with the received blocks
	err = createBlockchain(blocks)
	if err != nil {
		fmt.Println("Failed to create blockchain:", err)
		return err
	}
	return nil
}

// Update the Blockchain
func getBlockchain(peer peer.AddrInfo) ([]Block, error) {

	stream, err := User.NewStream(context.Background(), peer.ID, protocol.ID(config.ProtocolID+"/download/blockchain"))

	if err != nil {
		fmt.Println("Failed to create stream with peer:", peer.String())
		return nil, err
	}

	// Create a buffered reader writer for the stream
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	// Send the latest block height
	_, err = rw.WriteString(Latest_Block + "\n")
	if err != nil {
		fmt.Println("Failed to send latest block height to peer:",
			stream.Conn().RemotePeer().String(),
			err,
		)
		stream.Close()
		return nil, err
	}

	height_gap, _ := rw.ReadString('\n')

	if height_gap == "Shorter. Try from others\n" {
		fmt.Println("Peer has a shorter blockchain. Try from others")
		stream.Close()
		return nil, fmt.Errorf("peer has a shorter blockchain")
	}

	height_gap = strings.TrimSpace(height_gap)
	height_gap_int, _ := strconv.Atoi(height_gap)

	blocks := make([]Block, height_gap_int)

	// Receive the blockchain from the peer
	for i := 0; i < height_gap_int; i++ {
		// Read the block from the peer
		blockData, _ := rw.ReadString('\n')

		// Save them to the blockchain
		var block Block
		err = json.Unmarshal([]byte(blockData), &block)
		if err != nil {
			fmt.Println("Failed to parse block data:", err)
			continue
		}

		// Add it to the blocks slice
		blocks = append(blocks, block)
	}

	// Close the stream
	stream.Close()

	return blocks, nil
}

// Add the new blocks to the existing blockchain
func createBlockchain(blockchain []Block) error {

	// Note: We assume that the []Block is sorted in the order of the blockchain
	// Last one is the earliest block

	// Populate the blockchain database
	for _, block := range blockchain {
		// Ignore the block making for genesis block
		if block.Block_height == 0 {
			continue
		}

		BlockMutex.Lock()
		Blockchain[block.Block_hash] = block
		BlockMutex.Unlock()

		// Handle the UTXOs, creating and destorying the UTXOs for the new blocks
		for _, txn := range block.Transactions {
			go handleUTXO(&txn)
			TransMutex.Lock()
			Transactions[txn.Txn_id] = txn
			TransMutex.Unlock()
		}

		// No need to assign the result as the merkle root is already sent from the target node
		MerkleRoot := buildMerkle(block.Transactions)

		// Save the merkle root to the database
		MerkleMutex.RLock()
		Merkle_Roots[MerkleRoot.Value] = MerkleRoot
		MerkleMutex.RUnlock()
	}

	// Set the genesis block if applicable
	if blockchain[len(blockchain)-1].Block_height == 0 {
		Genesis_Block = blockchain[0].Block_hash
	}

	// Set the latest block
	Latest_Block = blockchain[len(blockchain)-1].Block_hash
	return nil
}

func createBlock(transaction []string, coinbaseFee float64) (Block, error) {
	current_block := Blockchain[Latest_Block]

	transactions := make([]Transaction, len(transaction)+1)

	// Make the coinbase transaction
	transactions[0] = Transaction{
		In_sz:  0,
		Out_sz: 1,
		Fee:    coinbaseFee,
		Inputs: []Input{},
		Outputs: []Output{
			{
				Pubkey: User.ID().String(),
				Value:  coinbaseFee,
			},
		},
		Timestamp: time.Now(),
	}

	// Generate the transaction fee for the miner
	transactions[0].generateTxn()

	// The include the rest of the transactions
	for idx, txn := range transaction {
		MempoolMutex.RLock()
		transactions[idx+1] = Mempool[txn]
		MempoolMutex.RUnlock()
	}

	// Create a new block
	newBlock := Block{
		Block_height:  current_block.Block_height + 1,
		Previous_hash: current_block.Block_hash,
		Transactions:  transactions,
		Timestamp:     time.Now(),
	}

	// Add the block hash to all the transactions
	for idx := range transactions {
		transactions[idx].Block_hash = newBlock.Block_hash
	}

	// Create the merkle root
	newBlock.Merkle_hash = buildMerkle(newBlock.Transactions).Value

	return newBlock, nil
}

// Validate the transaction by checking the UTXO set
func validateTransaction(txn Transaction) error {

	// Check the availablity in the UTXO Set
	inputSum := 0.0
	for _, input := range txn.Inputs {
		utxoHash := sha256.Sum256([]byte(fmt.Sprintf("%s:%d", input.Txn_id, input.Index)))

		UTXOMutex.RLock()
		utxo, exists := UTXO_SET[hex.EncodeToString(utxoHash[:])]
		UTXOMutex.RUnlock()

		if !exists {
			return fmt.Errorf("input does not exist in the UTXO set")
		}

		inputSum += utxo.Value
	}

	// Check the negative sums
	outputSum := 0.0
	for _, output := range txn.Outputs {
		if output.Value < 0 {
			return fmt.Errorf("output value is negative")
		}
		outputSum += output.Value
	}

	// Validate the fee
	if outputSum+txn.Fee > inputSum {
		return fmt.Errorf("output sum and fee is greater than the input")
	}

	return nil
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

	// Validate the transactions
	for _, txn := range block.Transactions {
		err := validateTransaction(txn)
		if err != nil {
			return err
		}
	}

	return nil
}

func startMining(block Block) {
	if miningCancel != nil {
		miningCancel() // Stop previous mining
	}

	miningCtx, miningCancel = context.WithCancel(context.Background())
	go mineBlock(miningCtx, block)
}

// Mining of a block
func mineBlock(ctx context.Context, block Block) error {

	// Check if the block is valid
	err := validateBlock(block)
	if err != nil {
		return fmt.Errorf("block is invalid")
	}

	// Mine the block using generateBlockHash
	nonce := int32(0)
	foundValidBlock := false
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Mining aborted due to new block arrival")
			return fmt.Errorf("mining interrupted")
		default:
			block.Nonce = nonce
			block.generateBlockHash()

			// Check if the hash is valid
			if block.Block_hash[:block.Difficulty] == "0000" {
				foundValidBlock = true
				break // This ensures the loop exits
			}
			nonce++
		}
		if foundValidBlock {
			break // Break the for loop
		}
	}

	// Check if the current block height is the same as the latest block height
	BlockMutex.RLock()
	currentBlock := Blockchain[Latest_Block]
	BlockMutex.RUnlock()

	if currentBlock.Block_height != block.Block_height-1 {
		return fmt.Errorf("block at the height is already mined!! Please try again")
	}

	// Add the block to the blockchain
	BlockMutex.Lock()
	Blockchain[block.Block_hash] = block
	BlockMutex.Unlock()

	// Remove the transactions from the mempool
	removeFromMempool(block)

	// Update the latest block
	Latest_Block = block.Block_hash

	// Broadcast the block to the peers

	// Create BlockDTO
	blockDTO := BlockDTO{
		Block_hash:    block.Block_hash,
		Block_height:  block.Block_height,
		Previous_hash: block.Previous_hash,
		Nonce:         block.Nonce,
		Difficulty:    block.Difficulty,
		Merkle_hash:   block.Merkle_hash,
		Timestamp:     block.Timestamp,
		Transactions:  []string{},
	}

	// Add the transactions to the BlockDTO
	for _, txn := range block.Transactions {
		blockDTO.Transactions = append(blockDTO.Transactions, txn.Txn_id)
	}

	// Propagate the block to all the connected peers
	for _, peer := range peerArray {
		// Create a new stream for the peer
		stream, err := User.NewStream(context.Background(), peer.ID, protocol.ID(config.ProtocolID+"/propagate"))
		if err != nil {
			fmt.Println("Failed to create stream with peer:", peer.ID)
			continue
		}

		// Create a buffered writer for the stream
		rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

		// Serialize the block to JSON
		data, err := json.Marshal(blockDTO)
		if err != nil {
			fmt.Println("Failed to serialize block:", err)
			stream.Close()
			continue
		}

		// Send the serialized block
		_, err = rw.WriteString(string(data) + "\n")
		if err != nil {
			fmt.Println("Failed to send block to peer:", peer.ID, err)
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
	}

	return nil
}

func createGenesis() {
	// Create the genesis block if it doesn't exist
	if Genesis_Block == "" {
		genesisBlock := Block{
			Block_height:  0,
			Previous_hash: "",
			Transactions:  []Transaction{},
			Timestamp:     time.Unix(0, 0),
		}
		genesisBlock.generateBlockHash()
		Blockchain[genesisBlock.Block_hash] = genesisBlock
		Genesis_Block = genesisBlock.Block_hash
		Latest_Block = genesisBlock.Block_hash
	}
}

// TODO
func startUp() error {
	// Create the genesis block
	createGenesis()

	// Pick a random peer to sync the blockchain
	peerMutex.RLock()
	peers := peerArray
	if len(peers) == 0 {
		return fmt.Errorf("no peers available")
	}
	randomPeer := peers[rand.Intn(len(peers))]
	peerMutex.RUnlock()
	fmt.Println("Syncing with peer:", randomPeer.ID.String())

	err := syncBlockchain(randomPeer)
	if err != nil {
		fmt.Println("Failed to sync blockchain:", err)
		return err
	}

	// Download the blockchain
	err = syncBlockchain(randomPeer)
	if err != nil {
		fmt.Println("Failed to download blockchain:", err)
		return err
	}

	// Download the mempool
	err = updateMempool(randomPeer)
	if err != nil {
		fmt.Println("Failed to download mempool:", err)
		return err
	}

	return nil
}

// Validate whether the recieved blockchain copy has some inconsistencies
func validateBlockchain() error {
	BlockMutex.RLock()
	defer BlockMutex.RUnlock()

	// Start from the latest block and traverse backwards
	block := Blockchain[Latest_Block]
	for block.Block_hash != Genesis_Block {
		// Validate the block
		err := validateBlock(block)
		if err != nil {
			return fmt.Errorf("block %s is invalid: %v", block.Block_hash, err)
		}

		// Move to the previous block
		block = Blockchain[block.Previous_hash]
	}

	// Validate the genesis block
	err := validateBlock(Blockchain[Genesis_Block])
	if err != nil {
		return fmt.Errorf("genesis block is invalid: %v", err)
	}

	return nil
}
