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

		peerMutex.RLock()
		small, exist := least[message.Sender]
		peerMutex.RUnlock()

		if !exist || message.Message_Id > small {
			peerMutex.Lock()
			least[message.Sender] = message.Message_Id
			peerMutex.Unlock()
		} else {
			// Skip as the message might be very old or already reached
			continue
		}

		// Database Access with Mutex
		peerMutex.RLock()
		_, exists := database[message.Sender]
		peerMutex.RUnlock()

		if !exists {
			peerMutex.Lock()
			database[message.Sender] = make(map[int32]string)
			peerMutex.Unlock()
		}

		peerMutex.RLock()
		_, exists = database[message.Sender][message.Message_Id]
		peerMutex.RUnlock()

		if exists {
			continue
		}

		peerMutex.Lock()
		database[message.Sender][message.Message_Id] = message.Content
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

func propagateTxn(rw *bufio.ReadWriter, strm network.Stream) {

}

func propagateBlock(rw *bufio.ReadWriter, strm network.Stream) {

}
