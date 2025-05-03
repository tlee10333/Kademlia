package main

import (
	"bufio"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

const bitLength = 4

// Message represents a Kademlia RPC-style message
type Message struct {
	Type  string `json:"type"` // ping, store, find_node, find_value
	From  int    `json:"from"` // Sender node ID
	Key   int    `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

// StartServer initializes a UDP server and listens for incoming messages
func StartServer(node *Node, port int) {
	addr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP("127.0.0.1"),
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		fmt.Printf("Failed to bind to port %d: %v\n", port, err)
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Printf("Node %d listening on port %d...\n", node.ID, port)

	buffer := make([]byte, 1024)

	//routine superloop
	for {
		n, remoteAddr, err := conn.ReadFromUDP(buffer)

		if err != nil {
			fmt.Println("Error reading from UDP:", err)
			continue
		}

		go handleMessage(conn, node, buffer[:n], remoteAddr)
	}

}

func handleMessage(conn *net.UDPConn, node *Node, data []byte, addr *net.UDPAddr) {
	var msg Message
	err := json.Unmarshal(data, &msg)
	if err != nil {
		fmt.Println("Invalid message:", err)
		return
	}

	fmt.Printf("Received %s from %d\n", msg.Type, msg.From)

	var response Message

	switch msg.Type {
	case "ping":
		response = Message{Type: "pong", From: node.ID}

	case "store":
		fmt.Println(msg.Key, msg.Value)
		node.InsertKV(msg.Key, msg.Value)
		response = Message{Type: "store_ack", From: node.ID}

	case "find_value":
		if value, found := node.FindKV(msg.Key); found {
			response = Message{Type: "found_value", From: node.ID, Key: msg.Key, Value: value}
		} else {
			response = Message{Type: "not_found", From: node.ID}
		}

	case "find_node":
		// This would normally return a list of closest nodes. Simplified here.
		response = Message{Type: "node_info", From: node.ID}

	default:
		response = Message{Type: "error", From: node.ID}
	}

	respBytes, _ := json.Marshal(response)
	conn.WriteToUDP(respBytes, addr)
}

func sendMessage(addr net.UDPAddr, msg Message) {

	// addr := net.UDPAddr{
	// 	Port: toPort,
	// 	IP:   net.ParseIP("127.0.0.1"),
	// }

	conn, err := net.DialUDP("udp", nil, &addr)

	if err != nil {
		fmt.Println("Could not dial remote UDP:", err)
		return
	}
	defer conn.Close()

	msgBytes, _ := json.Marshal(msg)
	conn.Write(msgBytes)
	//We successfully write to the port, then something happens

	buffer := make([]byte, 1024)
	n, _, _ := conn.ReadFromUDP(buffer)

	fmt.Printf("Response: %s\n", string(buffer[:n]))
}

//******************Helper Functions*********************

func Atoi(s string) int {
	num, err := strconv.Atoi(s)

	if err != nil {
		fmt.Println("Cannot Convert String to Int")
		return -1
	}

	return num
}

func HashToIntNBits(input string, n int) int {
	if n <= 0 || n > 31 { // conservative for 32-bit systems
		panic("HashToIntNBits: n must be between 1 and 31")
	}

	fullHash := sha1.Sum([]byte(input)) // [20]byte = 160 bits

	// Convert hash to big.Int
	hashInt := new(big.Int).SetBytes(fullHash[:])

	// Shift right to extract the top `n` bits
	shifted := new(big.Int).Rsh(hashInt, uint(160-n))
	fmt.Printf("Key: %s is  now %d \n", input, shifted)

	// Convert to int safely
	return int(shifted.Int64())

}

// arbitrary NodeID generator (based off of port number)
func NodeIDGenerator(number int, n int) int {
	// Extract the last `n` bits by performing a modulo operation
	lastBits := number % (1 << n) // (1 << n) is equivalent to 2^n

	return lastBits
}

// // FindValue locates which bucket/key should be queried and sends a FIND_VALUE message
// func (node *Node) FindValue(key string) {
// 	// Step 1: Hash key to int ID

// 	// Step 2: Compute distance to SelfID
// 	keyID := Atoi(key)
// 	if keyID == -1 {
// 		fmt.Println("FAILED KEY")
// 		return
// 	}

// 	_, bucketIndex, bucket := node.RoutingTable.FindNode(keyID) //Only care about the bucket

// 	if bucketIndex < 0 || bucketIndex >= len(node.RoutingTable.Buckets) {
// 		fmt.Println("No valid bucket found for this key.")
// 		return
// 	}

// 	// Step 4: Get bucket and iterate through nodes
// 	if len(bucket) == 0 {
// 		fmt.Printf("Bucket %d is empty. No nodes to contact.\n", bucketIndex)
// 		return
// 	}

// 	// Step 5: Send FIND_VALUE to all nodes in the selected bucket
// 	for _, nodeInfo := range bucket {
// 		msg := Message{
// 			Type:  "find_value",
// 			From:  node.ID,
// 			Key:   key,
// 			Value: "",
// 		}

// 		sendMessage(nodeInfo.Addr, msg)

// 	}
// }

func main() {

	port, _ := strconv.Atoi(os.Args[1])
	nodeID := NodeIDGenerator(port, bitLength)
	node := NewNode(nodeID)
	fmt.Printf("SERVER NODE ID: %d\n", node.GetID())

	// Run the server in a goroutine
	go StartServer(node, port)

	// Give it a moment to start
	select {
	case <-time.After(1 * time.Second):
	}

	reader := bufio.NewReader(os.Stdin)

	//SuperLoop for Sending Messages
	for {
		fmt.Println("SEND A MESSAGE")
		line, _ := reader.ReadString('\n')
		parts := strings.Fields(strings.TrimSpace(line))

		cmd := parts[0]
		value := ""
		key := 0

		switch cmd {
		case "ping":
			toPort := Atoi(parts[1])

			addr := net.UDPAddr{
				Port: toPort,
				IP:   net.ParseIP("127.0.0.1"),
			}
			msg := Message{
				Type:  cmd,
				From:  node.ID,
				Key:   key,
				Value: value,
			}
			sendMessage(addr, msg)
		case "store":
			toPort := Atoi(parts[1])
			key = HashToIntNBits(parts[2], bitLength)
			value = parts[3]

			msg := Message{
				Type:  cmd,
				From:  node.ID,
				Key:   key,
				Value: value,
			}
			addr := net.UDPAddr{
				Port: toPort,
				IP:   net.ParseIP("127.0.0.1"),
			}

			sendMessage(addr, msg)

		case "find_value":
			key = HashToIntNBits(parts[2], bitLength)

			//try to find key in our's first
			value, isFound := node.FindKV(key)
			if isFound {
				fmt.Printf("Key %s is in Server Node %d with value: %s \n", parts[2], node.ID, value)
			} else {
				//Find everyone in the closest bucket and ask them if they have it

			}

		}

	}

}
