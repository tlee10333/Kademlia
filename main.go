package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// Message represents a Kademlia RPC-style message
type Message struct {
	Type  string `json:"type"` // ping, store, find_node, find_value
	From  int    `json:"from"` // Sender node ID
	Key   string `json:"key,omitempty"`
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

func sendMessage(toPort int, msg Message) {

	addr := net.UDPAddr{
		Port: toPort,
		IP:   net.ParseIP("127.0.0.1"),
	}

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

func main() {

	port, _ := strconv.Atoi(os.Args[1])
	node := NewNode(port)

	// Run the server in a goroutine
	go StartServer(node, port)

	// Give it a moment to start
	select {
	case <-time.After(1 * time.Second):
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("SEND A MESSAGE")
		line, _ := reader.ReadString('\n')
		parts := strings.Fields(strings.TrimSpace(line))
		if len(parts) < 3 {
			fmt.Println("Invalid command")
			continue
		}

		cmd, targetAddr, key := parts[0], parts[1], parts[2]
		toPort, _ := strconv.Atoi(targetAddr)

		value := ""
		if len(parts) == 4 {
			value = parts[3]
		}

		msg := Message{
			Type:  cmd,
			From:  node.ID,
			Key:   key,
			Value: value,
		}

		sendMessage(toPort, msg)

	}

}
