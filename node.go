//Main node class

package main

import "net"

//Server Node
type Node struct {
	ID           int
	ADDR         net.UDPAddr
	KV           map[int]string
	RoutingTable RoutingTable
}

func NewNode(id int, addr net.UDPAddr) *Node {
	return &Node{
		ID:           id,
		ADDR:         addr,
		KV:           make(map[int]string),
		RoutingTable: *NewRoutingTable(id, bitLength, K),
	}
}

func (node *Node) GetID() int {
	return node.ID
}

func (node *Node) FindKV(key int) (string, bool) {
	value, isFound := node.KV[key]
	return value, isFound
}

func (node *Node) InsertKV(key int, value string) bool {
	node.KV[key] = value
	_, isFound := node.FindKV(key)
	return isFound
}
func (node *Node) DeleteKV(key int) bool {
	delete(node.KV, key)
	_, isFound := node.FindKV(key)
	return isFound
}

func Xor(a int, b int) int {
	return a ^ b
}
