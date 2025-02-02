package main

import (
	"bufio"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"os"
	"time"

	"github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
	"github.com/multiformats/go-multiaddr"
)

var logger = log.Logger("rendezvous")

func readData(rw *bufio.ReadWriter) {
	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			fmt.Println("User Went Offline", err)
			return
		}

		if str == "" || str == "\n" {
			continue
		}

		fmt.Printf("\x1b[32m%s\x1b[0m> ", str)
	}
}

func main() {
	log.SetAllLoggers(log.LevelWarn)
	log.SetLogLevel("rendezvous", "info")
	var err error
	// Parsing the flags
	config, _ = ParseFlags()

	fmt.Println("Enter the private key:")
	reader := bufio.NewReader(os.Stdin)
	privKeyString, _ := reader.ReadString('\n')
	privKeyString = strings.TrimSpace(privKeyString)
	privKeyBytes, _ := hex.DecodeString(privKeyString)

	privKey, _ := crypto.UnmarshalPrivateKey(privKeyBytes)

	// Creating the current node
	User, err = libp2p.New(
		libp2p.Identity(privKey),
		libp2p.ListenAddrs(
			[]multiaddr.Multiaddr(config.ListenAddresses)...,
		),
	)
	if err != nil {
		panic(err)
	}

	// Display the public key
	pubKey := privKey.GetPublic()
	pubKeyBytes, _ := pubKey.Raw()
	fmt.Println("Public Key (Hex):", hex.EncodeToString(pubKeyBytes))

	logger.Info("Node created with the ID: ", User.ID().String())

	// Set the handlers for the node
	SetNodeHandlers()

	// Extract Bootstrap peers
	ctx := context.Background()
	bootstrapPeers := make([]peer.AddrInfo, len(config.BootstrapPeers))

	// Add the Bootstrap peers to the peer slice
	for i, addr := range config.BootstrapPeers {
		peerInfo, _ := peer.AddrInfoFromP2pAddr(addr)
		bootstrapPeers[i] = *peerInfo
	}

	// Create a local dht with custom buket size
	kademliaDHT, err = dht.New(
		ctx, User,
		dht.BootstrapPeers(bootstrapPeers...),
		dht.ProtocolPrefix("/custom-dht"),
		dht.BucketSize(5),
	)

	if err != nil {
		panic(err)
	}

	// Clean-up scheduled
	defer kademliaDHT.Close()

	logger.Debug("Bootstrapping the node's DHT")
	if err = kademliaDHT.Bootstrap(ctx); err != nil {
		panic(err)
	}

	// Wait for the Bootstrapping to finish
	time.Sleep(2 * time.Second)

	// Announce your arrival
	logger.Debug("Announcing your arrival")
	routingDiscovery := drouting.NewRoutingDiscovery(kademliaDHT)
	dutil.Advertise(ctx, routingDiscovery, config.RendezvousString)
	logger.Debug("Successfully Announced")

	logger.Debug("Starting the Peer discovery")

	go func() {
		for {
			// Create a peer channel for all the available users
			peerChan, err := routingDiscovery.FindPeers(ctx, config.RendezvousString)
			if err != nil {
				logger.Error("Error finding peers:", err)
				time.Sleep(2 * time.Second)
				continue
			}

			for peer := range peerChan {
				if peer.ID == User.ID() {
					continue
				}

				if _, exists := peerSet[peer.ID.String()]; exists {
					continue
				}

				if err := User.Connect(ctx, peer); err != nil {
					logger.Warn("Failed to connect to peer:", err)
				} else {
					logger.Info("Connected to peer:", peer.ID.String())
					peerArray = append(peerArray, peer)
					peerSet[peer.ID.String()] = peer
				}
			}

			// The local DHT updates for every two seconds
			time.Sleep(2 * time.Second)
		}
	}()

	// Initialize and poplutate the databases
	startUp()

	// For the user communication
	reader = bufio.NewReader(os.Stdin)
	var userStream network.Stream = nil

	for {
		// Mode Selection: Direct Message or Gossip Mode
		print("> Select Mode (1: Send Transaction, 2: Gossip Mode, 3: Exit)\n> ")
		mode, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading the input")
			continue
		}

		mode = strings.TrimSpace(mode)

		// Exit the program if selected
		if mode == "3" {
			fmt.Println("Exiting...")
			break
		}

		// Handle Direct Message Mode
		if mode == "1" {

			// Direct Message logic (same as before)
			for {
				if userStream == nil {

					// Taking the input to select user index
					print("> Enter User Node ID\n>")
					input, err := reader.ReadString('\n')
					if err != nil {
						fmt.Println("Error reading the input")
						continue
					}

					input = strings.TrimSpace(input)

					logger.Info("Connecting to user")

					// Establish the Stream
					targetUser, exists := peerSet[input]

					if !exists {
						println("User not found!")
						continue
					}
					newStream, err := User.NewStream(ctx, targetUser.ID, protocol.ID(config.ProtocolID+"/message"))
					if err != nil {
						println("Error occurred creating a stream!\n")
						continue
					}

					// Set the current stream
					userStream = newStream

					logger.Info("Connected to: ", targetUser.ID.String())
				}

				println("> Enter UTXO hash (type 'Cancel' to go back to mode selection)")

				// Taking the input to send a transaction

				// Entering number of inputs, Outputs, Fee
				println("> Enter Number of Inputs")
				read, _ := reader.ReadString('\n')
				read = strings.TrimSpace(read)
				inputs, _ := strconv.ParseInt(read, 10, 8)

				read, _ = reader.ReadString('\n')
				read = strings.TrimSpace(read)
				outputs, _ := strconv.ParseInt(read, 10, 64)

				read, _ = reader.ReadString('\n')
				read = strings.TrimSpace(read)
				fee, _ := strconv.ParseFloat(read, 64)

				utxos := make([]string, 0)
				nodeID := make([]string, 0)
				amounts := make([]float64, 0)

				for i := 0; i < int(inputs); i++ {
					println("> Enter UTXO hash")
					utxo, _ := reader.ReadString('\n')
					utxo = strings.TrimSpace(utxo)
					utxos = append(utxos, utxo)
				}

				for i := 0; i < int(outputs); i++ {
					println("> Enter Node ID")
					id, _ := reader.ReadString('\n')
					id = strings.TrimSpace(id)
					nodeID = append(nodeID, id)

					println("> Enter Amount")
					amt, _ := reader.ReadString('\n')
					amt = strings.TrimSpace(amt)
					amount, _ := strconv.ParseFloat(amt, 64)
					amounts = append(amounts, amount)
				}

				sendFunds(utxos, nodeID, amounts, fee)
			}
		}

		// Dislpay all the transactions in the Mempool
		if mode == "2" {
			displayMempool()
		}

		// Display the Blockchain
		if mode == "3" {
			displayBlockchain()
		}

		// Sync the Blockchain
		if mode == "4" {
			// Pick a random peer to sync the blockchain
			peerMutex.RLock()
			peers := peerArray
			if len(peers) == 0 {
				fmt.Println("No peers available for syncing")
				continue
			}
			randomPeer := peers[rand.Intn(len(peers))]
			fmt.Println("Syncing with peer:", randomPeer.ID.String())
			// Add your blockchain syncing logic here
			peerMutex.RUnlock()

			stream, err := User.NewStream(ctx, randomPeer.ID, protocol.ID(config.ProtocolID+"/download/blockchain"))
			if err != nil {
				fmt.Println("Failed to create stream with peer:", randomPeer.ID)
				continue
			}

			// Create a buffered writer for the stream
			rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

			line, err := rw.ReadString('\n')
			if line == "" || line == "\n" {
				continue
			}

			var blockchain []Block

			err = json.Unmarshal([]byte(line), &blockchain)

			if err != nil {
				fmt.Println("Failed to parse JSON:", err)
				continue
			}

			makeBlockchain(blockchain)
		}

		// Handle Gossip Mode
		if mode == "5" {
			// Taking the input to send a message in gossip mode
			println("> Enter Message for Gossip (type 'Cancel' to go back to mode selection)")

			sendData, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("Error reading from stdin:", err)
				continue
			}

			sendData = strings.TrimSpace(sendData)

			if sendData == "Cancel" {
				break
			}

			// Increment global m_id for Gossip messages
			m_id++

			// Create a new Message
			message := Message{
				Sender:     User.ID().String(),
				Message_Id: m_id,
				Content:    sendData,
			}

			// Send the message to all connected peers (Gossip)
			for _, peer := range peerArray {
				// Create a new stream for the peer
				stream, err := User.NewStream(ctx, peer.ID, protocol.ID(config.ProtocolID+"/gossip"))
				if err != nil {
					fmt.Println("Failed to create stream with peer:", peer.ID)
					continue
				}

				// Create a buffered writer for the stream
				rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

				// Serialize the message to JSON
				data, err := json.Marshal(message)
				if err != nil {
					fmt.Println("Failed to serialize message:", err)
					stream.Close()
					continue
				}

				// Send the serialized message
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
			}
		}
	}
}
