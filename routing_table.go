// Handles Figuring out which node to search for next
// Is the binary tree
package main

import (
	"fmt"
	"math/bits"
	"net"
	"sort"
	"time"
)

type NodeInfo struct {
	ID        int
	Addr      net.UDPAddr
	TimeStamp int64
}

func NewNodeInfo(id int, address net.UDPAddr) *NodeInfo {
	return &NodeInfo{
		ID:        id,
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

func (rt *RoutingTable) InsertNode(nodeID int, NodeAddr net.UDPAddr) {

	if nodeID == rt.SelfID {
		return // Don't insert self
	}

	isFound, prefixIndex, bucket := rt.FindNode(nodeID)

	if !isFound {
		if len(bucket) < rt.K {
			fmt.Printf("ADDING PORT: %d \n", NodeAddr.Port)

			bucket[nodeID] = *NewNodeInfo(nodeID, NodeAddr) // Replace with actual address
			distance := Xor(rt.SelfID, nodeID)

			//PRINT STATEMENTS FOR DEMO/CONCEPTUAL COMPREHENSION
			fmt.Printf("Node ID %d distance to self Node %d is XOR of %04b and %04b which is %04b \n", nodeID, rt.SelfID, nodeID, rt.SelfID, distance)
			fmt.Printf("Successfully added node %d to bucket %d\n", nodeID, prefixIndex)
			fmt.Println("")
			rt.PrintRoutingTableSummary()
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

//Need to verify these

// FindClosestNodes returns the k nodes closest to the target ID
func (rt *RoutingTable) FindClosestNodes(targetID int) []NodeInfo {
	count := rt.K

	// Collect all nodes from all buckets
	var allNodes []NodeInfo
	for _, bucket := range rt.Buckets {
		for _, node := range bucket {
			allNodes = append(allNodes, node)
		}
	}



	// Sort nodes by XOR distance to target
	rt.SortByDistance(allNodes, targetID)

	// Return the k closest nodes
	if len(allNodes) > count {
		return allNodes[:count]
	}
	return allNodes
}

// SortByDistance sorts nodes by their XOR distance to the target ID
func (rt *RoutingTable) SortByDistance(nodes []NodeInfo, targetID int) {
	sort.Slice(nodes, func(i, j int) bool {
		distI := Xor(nodes[i].ID, targetID)
		distJ := Xor(nodes[j].ID, targetID)
		return distI < distJ
	})

}

// PrintRoutingTableSummary prints a summary of the routing table
func (rt *RoutingTable) PrintRoutingTableSummary() {
	fmt.Printf("Routing Table for Node %d:\n", rt.SelfID)
	totalNodes := 0

	for i, bucket := range rt.Buckets {
		nodeCount := len(bucket)
		if nodeCount > 0 {
			fmt.Printf("BUCKET %d \n", i)
			rt.PrintBucketList(bucket)
			totalNodes += nodeCount
		}
	}
	fmt.Printf("Total nodes: %d\n", totalNodes)
}
