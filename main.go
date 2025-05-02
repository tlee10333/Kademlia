//We define K nodes to keep track, the # of bits for our hashing

package main

import (
	"fmt"
	"math/rand"
)

func RandomNodeKey() int {
	return rand.Intn(16) // 0 to 15 inclusive
}

func main() {
	nodeA := NewNode(1)
	fmt.Println(nodeA.GetID())
	nodeA.InsertKV("E", "YO")
	nodeA.InsertKV("F", "OPYO")

	fmt.Println(nodeA.FindKV("E"))
	nodeA.DeleteKV("E")
	fmt.Println(nodeA.FindKV("E"))

}
