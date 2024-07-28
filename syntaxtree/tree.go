package syntaxtree

import "math"

type Node[T any] struct {
	Data       T
	properties map[string]any
	Id         uint64
}

type Tree[T any] struct {
	nodes       map[uint64]Node[T]
	maxID       uint64
	connections map[uint64]uint64
}

func (tree Tree[T]) AddNode(node Node[T]) {
	if _, ok := tree.nodes[node.Id]; ok {
		panic("Attempted to add node with same ID as other node")
	}
	tree.nodes[node.Id] = node
}

func (tree *Tree[T]) CreateNode() Node[T] {
	if tree.maxID == math.MaxUint64 {
		panic("tree.maxID is equal to math.MaxUint64")
	}
	var newNode = Node[T]{Id: tree.maxID}
	tree.maxID++
	return newNode
}
