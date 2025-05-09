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

const bitLength = 4 // Align with node.go's routing table creation
const K = 2         // Align with node.go's routing table creation

// Message represents a Kademlia RPC-style message
type Message struct {
	Type    string      `json:"type"` // ping, store, find_node, find_value, etc
	From    int         `json:"from"` // Sender node ID
	To      int         `json:"to"`   // Recipient ID
	IP      net.UDPAddr `json: "IP"`  //IP of original (so from & IP can be different servers)
	Key     int         `json:"key,omitempty"`
	Value   string      `json:"value,omitempty"`
	Closest []NodeInfo  `json:"closest,omitempty"` // Used for find_node responses
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

	// Routine superloop
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

	// Update the routing table with the sender's information
	node.RoutingTable.InsertNode(msg.From, msg.IP)

	var response Message

	switch msg.Type {
	case "ping":
		// RPC 1: PING - just respond to confirm we're alive
		response = Message{
			Type: "node_alive",
			IP:   node.ADDR,
			From: node.ID,
			To:   msg.From,
		}

	case "store":
		// RPC 2: STORE - store the key-value pair locally
		fmt.Printf("Storing key %d with value %s\n", msg.Key, msg.Value)
		node.InsertKV(msg.Key, msg.Value)
		response = Message{
			Type: "store_ack",
			IP:   node.ADDR,
			From: node.ID,
			To:   msg.From,
		}

	case "find_node":
		// RPC 3: FIND_NODE - return k closest nodes to the target ID
		fmt.Printf("Looking for closest nodes to %d\n", msg.Key)
		closest := node.RoutingTable.FindClosestNodes(msg.Key)
		response = Message{
			Type:    "node_response",
			From:    node.ID,
			To:      msg.From,
			IP:      node.ADDR,
			Closest: closest,
		}

	case "find_value":
		// RPC 4: FIND_VALUE - return value if we have it, otherwise act like FIND_NODE
		if value, found := node.FindKV(msg.Key); found {
			fmt.Printf("Found value for key %d: %s\n", msg.Key, value)
			response = Message{
				Type:  "found_value",
				From:  node.ID,
				To:    msg.From,
				IP:    node.ADDR,
				Key:   msg.Key,
				Value: value}
		} else {
			fmt.Printf("Value not found for key %d, returning closest nodes\n", msg.Key)

			closest := node.RoutingTable.FindClosestNodes(msg.Key)
			response = Message{
				Type:    "node_response",
				From:    node.ID,
				To:      msg.From,
				IP:      msg.IP, //Preserve original server IP
				Key:     msg.Key,
				Closest: closest,
			}

		}

	default:
		response = Message{
			Type: "error",
			From: node.ID,
			To:   msg.From,
			IP:   node.ADDR}
	}

	//SEND RESPONSE BACK
	respBytes, _ := json.Marshal(response)
	conn.WriteToUDP(respBytes, addr)
}

// Sending messages
func sendMessage(addr net.UDPAddr, msg Message, node *Node) (*Message, error) {

	conn, err := net.DialUDP("udp", nil, &addr) //When we actually send messages, we open up a random port and send it
	if err != nil {
		fmt.Println("Could not dial remote UDP:", err)
		return nil, err
	}
	defer conn.Close()

	msgBytes, _ := json.Marshal(msg)
	_, err = conn.Write(msgBytes)
	if err != nil {
		fmt.Println("Error sending message:", err)
		return nil, err
	}

	// Wait for response
	buffer := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	n, _, err := conn.ReadFromUDP(buffer)
	if err != nil {
		//If we don't get a response
		fmt.Println("Error receiving response:", err)
		//CHeck if this was a node we used to have in our routing table
		isFound, _, _ := node.RoutingTable.FindNode(msg.To)
		//If it was in our routing table but the server is now down, remove it
		if isFound {
			node.RoutingTable.DeleteNode(msg.To)
			fmt.Printf("REMOVING SERVER %d FROM RTABLE DUE TO NO RESPONSE \n", msg.To)
		}
		return nil, err
	}

	var response Message
	err = json.Unmarshal(buffer[:n], &response)
	if err != nil {
		fmt.Println("Error parsing response:", err)
		return nil, err
	}

	fmt.Printf("Response type: %s\n", response.Type)

	node.RoutingTable.InsertNode(response.From, response.IP)

	return &response, nil
}

// Helper Functions
func Atoi(s string) int {
	num, err := strconv.Atoi(s)
	if err != nil {
		fmt.Println("Cannot Convert String to Int")
		return -1
	}
	return num
}

// Our Hash Function
func HashToIntNBits(input string, n int) int {
	if n <= 0 || n > 31 { // conservative for 32-bit systems
		panic("HashToIntNBits: n must be between 1 and 31")
	}

	fullHash := sha1.Sum([]byte(input)) // [20]byte = 160 bits

	//THIS CONVERTS OUR HASH TO AN INT. ONLY FOR DEMO PURPOSES.
	// Convert hash to big.Int
	hashInt := new(big.Int).SetBytes(fullHash[:])

	// Shift right to extract the top `n` bits
	shifted := new(big.Int).Rsh(hashInt, uint(160-n))
	fmt.Printf("Key: %s is now %d\n", input, shifted)

	// Convert to int safely
	return int(shifted.Int64())
}

// NodeIDGenerator generates a node ID based on the port number
func NodeIDGenerator(number int, n int) int {
	// Extract the last `n` bits by performing a modulo operation
	lastBits := number % (1 << n) // (1 << n) is equivalent to 2^n
	return lastBits
}

// find_value - tries to find a value across the network
func (node *Node) FindValue(addr net.UDPAddr, key int) string {
	// First check if we have the value locally
	if value, found := node.FindKV(key); found {
		return value
	}

	// First find the K closest nodes to the key
	closestNodes := node.IterativeFindNode(addr, key)

	//Make sure closestNodes != 0
	if len(closestNodes) != 0 {
		//In case len(closestNodes) < K, we avoid a out of index error
		limit := K
		if len(closestNodes) < K {
			limit = len(closestNodes)
		}

		// Query nodes of up to K nodes (you can change this to something else too)
		for _, targetNode := range closestNodes[0:limit] {
			// Send FIND_VALUE message
			msg := Message{
				Type: "find_value",
				From: node.ID,
				To:   targetNode.ID,
				IP:   addr,
				Key:  key,
			}

			response, err := sendMessage(targetNode.Addr, msg, node)
			if err == nil && response.Type == "found_value" {
				fmt.Printf("Found value for key %d at node %d: %s\n", key, targetNode.ID, response.Value)
				return response.Value
			} else {
				fmt.Printf("Not found in server: %d \n", response.From)
			}

		}

	}

	//We didn't find the key
	return ""

}

// Iterative FindNode - finds the K closest nodes to a given key
func (node *Node) IterativeFindNode(addr net.UDPAddr, targetID int) []NodeInfo {
	// Start with the closest nodes we know
	closestNodes := node.RoutingTable.FindClosestNodes(targetID)

	// Track nodes we've already queried
	queried := make(map[int]bool)

	// Create a list to track all nodes we've found
	//Don't forget to include the current server node in allFoundNodes

	var allFoundNodes []NodeInfo
	for _, nodeInfo := range closestNodes {
		if nodeInfo.ID != node.ID { // Don't include ourselves
			allFoundNodes = append(allFoundNodes, nodeInfo)
		}
	}

	for len(allFoundNodes) > 0 {
		// Take the closest unqueried node
		node.RoutingTable.SortByDistance(allFoundNodes, targetID)
		var targetNode NodeInfo
		foundUnqueried := false

		for i, candidate := range allFoundNodes {
			if !queried[candidate.ID] {
				targetNode = candidate
				foundUnqueried = true
				// Remove this node from allFoundNodes, we'll add it back if it responds
				allFoundNodes = append(allFoundNodes[:i], allFoundNodes[i+1:]...)
				break
			}
		}

		if !foundUnqueried {
			break // All nodes have been queried
		}

		queried[targetNode.ID] = true

		// Send FIND_NODE message
		msg := Message{
			Type: "find_node",
			From: node.ID,
			To:   targetNode.ID,
			IP:   addr,
			Key:  targetID,
		}

		response, err := sendMessage(targetNode.Addr, msg, node)
		if err != nil {
			fmt.Printf("Error contacting node %d: %v\n", targetNode.ID, err)
			continue
		}

		// Add the node back to our list since it responded
		allFoundNodes = append(allFoundNodes, targetNode)

		if response.Type == "node_response" {
			// Add new nodes to our list if we haven't seen them yet
			for _, newNode := range response.Closest {
				found := false
				for _, existingNode := range allFoundNodes {
					if existingNode.ID == newNode.ID {
						found = true
						break
					}
				}

				if !found && newNode.ID != node.ID {
					allFoundNodes = append(allFoundNodes, newNode)
				}
			}
		}
	}

	// Return the K closest nodes from all we've found
	node.RoutingTable.SortByDistance(allFoundNodes, targetID)

	if len(allFoundNodes) > K {
		return allFoundNodes[:K]
	}
	return allFoundNodes
}

// Our store_dht logic
func (node *Node) storeDHT(addr net.UDPAddr, key int, value string) {

	// First find the K closest nodes to the key
	closestNodes := node.IterativeFindNode(addr, key)
	fmt.Println(closestNodes)

	//Don't forget to sort with the current server node as well
	currentNode := NewNodeInfo(node.GetID(), node.GetADDR())
	closestNodes = append(closestNodes, *currentNode)

	//Complete Sorted List of Nodes
	node.RoutingTable.SortByDistance(closestNodes, key)
	if len(closestNodes) != 0 {
		//In case len(closestNodes) < K, we avoid a out of index error
		limit := K
		if len(closestNodes) < K {
			limit = len(closestNodes)
		}
		// Store the key-value pair at each of these nodes
		for _, targetNode := range closestNodes[0:limit] {
			msg := Message{
				Type:  "store",
				From:  node.ID,
				IP:    addr,
				To:    targetNode.ID,
				Key:   key,
				Value: value,
			}

			_, err := sendMessage(targetNode.Addr, msg, node)
			if err != nil {
				fmt.Printf("Failed to store at node %d: %v\n", targetNode.ID, err)
			} else {
				fmt.Printf("Successfully stored key %d at node %d\n", key, targetNode.ID)
			}
		}

	}

}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <port>")
		os.Exit(1)
	}

	port := Atoi(os.Args[1])
	nodeID := NodeIDGenerator(port, bitLength)
	addr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP("127.0.0.1"),
	}

	node := NewNode(nodeID, addr)
	fmt.Printf("SERVER NODE ID: %d AT LOCALHOST ADDR PORT:%d \n", node.GetID(), node.ADDR.Port)

	// Run the server in a goroutine
	go StartServer(node, port)

	// Give it a moment to start
	time.Sleep(1 * time.Second)

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\nKademlia Node Commands:")
	fmt.Println("  ping <port> - Send ping to a node")
	fmt.Println("  store <port> <key> <value> - Store key-value pair at specific node")
	fmt.Println("  store_dht <key> <value> - Store key-value pair in DHT")
	fmt.Println("  find_value <key> - Find value for key in DHT")
	fmt.Println("  find_node <nodeID> - Find nodes closest to nodeID")
	fmt.Println("  print_rtable - Print Server Node Routing Table")
	fmt.Println("  quit - Exit the program")

	// SuperLoop for SENDING Messages
	for {
		fmt.Print("\nEnter command: ")
		line, _ := reader.ReadString('\n')
		parts := strings.Fields(strings.TrimSpace(line))

		if len(parts) == 0 {
			continue
		}

		cmd := parts[0]

		switch cmd {
		case "ping":
			if len(parts) < 2 {
				fmt.Println("Usage: ping <port>")
				continue
			}
			toPort := Atoi(parts[1])
			key := NodeIDGenerator(toPort, bitLength)

			addr := net.UDPAddr{
				Port: toPort,
				IP:   net.ParseIP("127.0.0.1"),
			}

			msg := Message{
				Type: "ping",
				IP:   node.ADDR,
				From: node.ID,
				To:   key,
			}
			sendMessage(addr, msg, node)

		case "store":
			if len(parts) < 4 {
				fmt.Println("Usage: store <port> <key> <value>")
				continue
			}
			toPort := Atoi(parts[1])
			key := NodeIDGenerator(toPort, bitLength)
			value := parts[3]

			msg := Message{
				Type:  "store",
				From:  node.ID,
				To:    key,
				IP:    node.ADDR,
				Key:   key,
				Value: value,
			}
			addr := net.UDPAddr{
				Port: toPort,
				IP:   net.ParseIP("127.0.0.1"),
			}

			sendMessage(addr, msg, node)

		case "store_dht":
			if len(parts) < 3 {
				fmt.Println("Usage: store_dht <key> <value>")
				continue
			}
			key := HashToIntNBits(parts[1], bitLength)
			value := parts[2]

			node.storeDHT(addr, key, value)

		case "find_value":
			if len(parts) < 2 {
				fmt.Println("Usage: find_value <key>")
				continue
			}
			key := HashToIntNBits(parts[1], bitLength)

			// First check locally
			value, isFound := node.FindKV(key)
			if isFound {
				fmt.Printf("Key %s (ID: %d) is in local node %d with value: %s\n", parts[1], key, node.ID, value)
			} else {
				fmt.Printf("Key %s (ID: %d) not found locally, searching DHT...\n", parts[1], key)
				value = node.FindValue(addr, key)
				if value != "" {
					fmt.Printf("Found key %s (ID: %d) in DHT with value: %s\n", parts[1], key, value)
				} else {
					fmt.Printf("Key %s (ID: %d) not found in DHT\n", parts[1], key)
				}
			}

		case "find_node":
			if len(parts) < 2 {
				fmt.Println("Usage: find_node <nodeID>")
				continue
			}
			targetID := Atoi(parts[1])

			fmt.Printf("Finding nodes closest to %d...\n", targetID)
			closestNodes := node.IterativeFindNode(node.ADDR, targetID)

			fmt.Printf("Found %d nodes:\n", len(closestNodes))
			for i, nodeInfo := range closestNodes {
				fmt.Printf("%d. Node ID: %d, Address: %s:%d\n",
					i+1, nodeInfo.ID, nodeInfo.Addr.IP, nodeInfo.Addr.Port)
			}

		case "print_rtable":
			node.RoutingTable.PrintRoutingTableSummary()
		case "quit":
			fmt.Println("Exiting...")
			return

		default:
			fmt.Println("Unknown command. Available commands: ping, store, store_dht, find_value, find_node, quit")
		}

		fmt.Println("") //So messages aren't so clumped together
	}
}
