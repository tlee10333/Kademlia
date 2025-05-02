//Main node class

package main

//Server Node
type Node struct {
	ID int
	KV map[string]string
}

func NewNode(id int) *Node {
	return &Node{
		ID: id,
		KV: make(map[string]string),
	}
}

func (node *Node) GetID() int {
	return node.ID
}

func (node *Node) FindKV(key string) (string, bool) {
	value, isFound := node.KV[key]
	return value, isFound
}

func (node *Node) InsertKV(key string, value string) bool {
	node.KV[key] = value
	_, isFound := node.FindKV(key)
	return isFound
}
func (node *Node) DeleteKV(key string) bool {
	delete(node.KV, key)
	_, isFound := node.FindKV(key)
	return isFound  
}