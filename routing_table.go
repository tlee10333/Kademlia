// Handles Figuring out which node to search for next
// Is the binary tree
package main

import (
	"fmt"
	"math/bits"
	"net"
	"time"
)

type NodeInfo struct {
	Addr      net.UDPAddr
	TimeStamp int64
}

func NewNodeInfo(address net.UDPAddr) *NodeInfo {
	return &NodeInfo{
		Addr:      address,
		TimeStamp: time.Now().Unix(),
	}

}

// Map since k = 20 so fast lookup for larger memory used is fine.
type Bucket map[int]NodeInfo

type RoutingTable struct {
	SelfID  int
	Buckets []Bucket
	K       int
}

func NewRoutingTable(selfID int, IDBits int, k int) *RoutingTable {
	buckets := make([]Bucket, IDBits)
	for i := range buckets {
		buckets[i] = make(Bucket)
	}

	return &RoutingTable{
		SelfID:  selfID,
		Buckets: buckets,
		K:       k,
	}
}

// Either find the node, or the bucket it should be in
func (rt *RoutingTable) FindNode(nodeID int) (bool, int, Bucket) {
	distance := Xor(rt.SelfID, nodeID)

	//Finds highest set bit (highest 1 in binary from right to left)
	prefixIndex := bits.Len(uint(distance)) - 1 //math/bits only accepts uint
	bucket := rt.Buckets[prefixIndex]
	_, isFound := bucket[nodeID]
	return isFound, prefixIndex, bucket

}

func (rt *RoutingTable) PrintBucketList(bucket Bucket) {

	for nodeID, nodeInfo := range bucket {
		fmt.Printf("NodeID: %d, Addr: %s, Time: %d\n", nodeID, nodeInfo.Addr, nodeInfo.TimeStamp)
	}

}

func (rt *RoutingTable) InsertNode(nodeID int) {
	if nodeID == rt.SelfID {
		return // Don't insert self
	}

	isFound, prefixIndex, bucket := rt.FindNode(nodeID)

	if !isFound {
		if len(bucket) < rt.K {
			placeholder := net.UDPAddr{
				Port: 1800,
				IP:   net.ParseIP("127.0.0.1"),
			}
			bucket[nodeID] = *NewNodeInfo(placeholder) // Replace with actual address
			fmt.Printf("Successfully added node %d to bucket %d\n", nodeID, prefixIndex)
		} else {
			fmt.Printf("Bucket %d is full, dropping node %d\n", prefixIndex, nodeID)
		}
	} else {
		fmt.Printf("Node %d already in bucket %d\n", nodeID, prefixIndex)
	}
}

func (rt *RoutingTable) DeleteNode(nodeID int) {
	if nodeID == rt.SelfID {
		return // Don't insert self
	}

	isFound, _, bucket := rt.FindNode(nodeID)
	if isFound {
		delete(bucket, nodeID)

	}

}
