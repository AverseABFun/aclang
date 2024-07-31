package syntaxtree

import (
	"fmt"
	"math"
	"strings"

	"github.com/averseabfun/logger"
)

type SyntaxType string

const (
	TYPE_COMMAND          SyntaxType = "SYNTAX_TYPE_COMMAND"
	TYPE_GROUPING         SyntaxType = "SYNTAX_TYPE_GROUPING"
	TYPE_GROUPING_COMMAND SyntaxType = "SYNTAX_TYPE_GROUPING_COMMAND"
	TYPE_VALUE            SyntaxType = "SYNTAX_TYPE_VALUE"
)

type SyntaxData struct {
	Type      SyntaxType
	Depth     uint
	Keyword   string
	Arguments []string
}

type SyntaxNode struct {
	Data          SyntaxData
	Id            uint64
	Children      map[uint64]*SyntaxNode
	MaxChild      uint64
	SmallestChild uint64
	Parent        *SyntaxNode
}

func (node *SyntaxNode) AddChild(child SyntaxNode) {
	logger.Logf(logger.LogDebug, "Adding child %d to %d", child.Id, node.Id)
	node.Children[child.Id] = &child
	child.Parent = node
	if child.Id > node.MaxChild {
		node.MaxChild = child.Id
	}
	if child.Id < node.SmallestChild {
		node.SmallestChild = child.Id
	}
}

func (node SyntaxNode) String() string {
	var childrenString = ""
	logger.Logf(logger.LogDebug, "SmallestChild: %d, MaxChild: %d", node.SmallestChild, node.MaxChild)
	for id := node.SmallestChild; id <= node.MaxChild; id++ {
		var val, ok = node.Children[id]
		if !ok {
			logger.Logf(logger.LogWarning, "Cannot find node with ID %d", id)
			continue
		}
		childrenString += val.String()
		childrenString += ", "
	}
	var numTabs = node.Data.Depth
	var tabs = strings.Repeat("\t", int(numTabs))
	return fmt.Sprintf("%d %s(%s): {\n%s%s}", node.Id, node.Data.Keyword, node.Data.Type, tabs, childrenString)
}

type SyntaxTree struct {
	RootNode SyntaxNode
	maxID    uint64
}

func (tree *SyntaxTree) String() string {
	return tree.RootNode.String()
}

func (tree *SyntaxTree) CreateNode() SyntaxNode {
	tree.maxID++
	var newNode = SyntaxNode{Id: tree.maxID, Children: make(map[uint64]*SyntaxNode), Data: SyntaxData{}, Parent: &tree.RootNode, MaxChild: 0, SmallestChild: math.MaxUint64}
	return newNode
}

func CreateTree() SyntaxTree {
	var tree = SyntaxTree{maxID: math.MaxUint64}
	tree.RootNode = tree.CreateNode()
	return tree
}
