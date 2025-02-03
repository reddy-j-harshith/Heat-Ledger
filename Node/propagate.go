package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"
)

func propagateMessage(rw *bufio.ReadWriter, strm network.Stream) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic in gossipExecute:", r)
		}
	}()

	for {
		line, err := rw.ReadString('\n')
		if err != nil {
			return
		}

		if line == "" || line == "\n" {
			continue
		}

		var message Message
		err = json.Unmarshal([]byte(line), &message)
		if err != nil {
			fmt.Println("Failed to parse JSON:", err)
			continue
		}

		peerMutex.Lock()
		if small, exist := least[message.Sender]; !exist || message.Message_Id > small {
			least[message.Sender] = message.Message_Id
		} else {
			peerMutex.Unlock()
			continue
		}
		peerMutex.Unlock()

		// Database Access with Mutex
		peerMutex.Lock()
		if _, exists := database[message.Sender]; !exists {
			database[message.Sender] = make(map[int32]string)
			if _, exists := database[message.Sender][message.Message_Id]; !exists {
				database[message.Sender][message.Message_Id] = message.Content
			} else {
				peerMutex.Unlock()
				continue
			}
		}
		peerMutex.Unlock()

		fmt.Printf("\x1b[32m> Message Sent by: %s\n> Message: %s\n> Sent from %s\x1b[0m\n", message.Sender, message.Content, strm.Conn().RemotePeer())

		peerMutex.RLock()
		peers := peerArray
		peerMutex.RUnlock()

		for _, peer := range peers {
			if peer.ID == strm.Conn().RemotePeer() || peer.ID.String() == message.Sender {
				continue
			}

			stream, err := User.NewStream(context.Background(), peer.ID, protocol.ID("/chat/1.0.0/gossip"))
			if err != nil {
				continue
			}

			rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

			data, err := json.Marshal(message)
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

			err = rw.Flush()
			if err != nil {
				fmt.Println("Failed to flush data to peer:", peer.ID, err)
				stream.Close()
				continue
			}

			stream.Close()
		}
	}
}

// Propagate the transaction to the network - Mempool
func propagateTxn(rw *bufio.ReadWriter, strm network.Stream) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic in propagateTxn:", r)
		}
	}()

	for {
		line, err := rw.ReadString('\n')
		if err != nil {
			return
		}

		if line == "" || line == "\n" {
			continue
		}

		var transaction Transaction
		err = json.Unmarshal([]byte(line), &transaction)
		if err != nil {
			fmt.Println("Failed to parse JSON:", err)
			continue
		}

		// Add the transaction to the mempool
		MempoolMutex.Lock()
		if _, exists := Mempool[transaction.Txn_id]; !exists {
			Mempool[transaction.Txn_id] = transaction
		} else {
			MempoolMutex.Unlock()
			continue
		}
		MempoolMutex.Unlock()

		fmt.Printf("\x1b[32m> New Txn Added to Mempool\n> Txn_id: %s\n> Sent from %s\x1b[0m\n", transaction.Txn_id, strm.Conn().RemotePeer())

		peerMutex.RLock()
		peers := peerArray
		peerMutex.RUnlock()

		for _, peer := range peers {
			if peer.ID == strm.Conn().RemotePeer() {
				continue
			}

			stream, err := User.NewStream(context.Background(), peer.ID, protocol.ID("/blockchain/1.0.0/broadcast/transaction"))
			if err != nil {
				continue
			}

			rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

			data, err := json.Marshal(transaction)
			if err != nil {
				fmt.Println("Failed to serialize transaction:", err)
				stream.Close()
				continue
			}

			_, err = rw.WriteString(string(data) + "\n")
			if err != nil {
				fmt.Println("Failed to send transaction to peer:", peer.ID, err)
				stream.Close()
				continue
			}

			err = rw.Flush()
			if err != nil {
				fmt.Println("Failed to flush data to peer:", peer.ID, err)
				stream.Close()
				continue
			}

			stream.Close()
		}
	}
}

// Validate the block and propagate it to the network
func propagateBlock(rw *bufio.ReadWriter, strm network.Stream) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic in gossipExecute:", r)
		}
	}()

	for {
		line, err := rw.ReadString('\n')
		if err != nil {
			return
		}

		if line == "" || line == "\n" {
			continue
		}

		var blockDTO BlockDTO
		err = json.Unmarshal([]byte(line), &blockDTO)
		if err != nil {
			fmt.Println("Failed to parse JSON:", err)
			continue
		}

		// Add the block to the blockchain database
		BlockMutex.RLock()
		_, exists := Blockchain[blockDTO.Block_hash]
		BlockMutex.RUnlock()

		if !exists {
			// Make the Block
			block := Block{
				Block_hash:    blockDTO.Block_hash,
				Block_height:  blockDTO.Block_height,
				Previous_hash: blockDTO.Previous_hash,
				Nonce:         blockDTO.Nonce,
				Difficulty:    blockDTO.Difficulty,
				Merkle_hash:   blockDTO.Merkle_hash,
				Timestamp:     blockDTO.Timestamp,
				Transactions:  []Transaction{},
			}

			// Add Transactions to the Block
			for _, txn := range blockDTO.Transactions {
				transaction := Mempool[txn]
				block.Transactions = append(block.Transactions, transaction)
			}

			// Validate the Block
			err := validateBlock(block)
			if err != nil {
				fmt.Println("Block Validation Failed: Stopping propagation", err)
				return
			}

			// Stop mining since a new block is confirmed
			if miningCancel != nil {
				miningCancel()
			}

			// Remove Transactions from Mempool & Update UTXO
			go removeFromMempool(block)
			go buildMerkle(block.Transactions)
			Latest_Block = block.Block_hash

			// Add the Block to the database
			BlockMutex.Lock()
			Blockchain[block.Block_hash] = block
			BlockMutex.Unlock()
		} else {
			continue
		}

		fmt.Printf("\x1b[32m> New Block Added to Blockchain\n> Block_hash: %s\n> Sent from %s\x1b[0m\n", blockDTO.Block_hash, strm.Conn().RemotePeer())

		peerMutex.RLock()
		peers := peerArray
		peerMutex.RUnlock()

		for _, peer := range peers {
			if peer.ID == strm.Conn().RemotePeer() {
				continue
			}

			stream, err := User.NewStream(context.Background(), peer.ID, protocol.ID(config.ProtocolID+"/broadcast/block"))
			if err != nil {
				continue
			}

			rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

			data, err := json.Marshal(blockDTO)
			if err != nil {
				fmt.Println("Failed to serialize transaction:", err)
				stream.Close()
				continue
			}

			_, err = rw.WriteString(string(data) + "\n")
			if err != nil {
				fmt.Println("Failed to send transaction to peer:", peer.ID, err)
				stream.Close()
				continue
			}

			err = rw.Flush()
			if err != nil {
				fmt.Println("Failed to flush data to peer:", peer.ID, err)
				stream.Close()
				continue
			}

			stream.Close()
		}
	}
}
