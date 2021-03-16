package gee

import (
	"fmt"
	"strings"
)

type node struct {
	pattern  string//全路由
	part     string//分割的部分路由
	children []*node//存放子节点
	isWild   bool
}
//输出封装
func (n *node)String() string {
	return fmt.Sprintf("node{pattern=%s, part=%s, isWild=%t}", n.pattern, n.part, n.isWild)
}

//前缀树插入
func (n *node)insert(pattern string,parts []string,height int)  {
	fmt.Println("yb test",pattern,parts,height)
	if len(parts)==height {
		n.pattern=pattern
		return
	}
	part:=parts[height]
	fmt.Println("yb test",part)
	child:=n.matchChild(part)
	if child==nil{
		child=&node{part: part, isWild: part[0] == ':' || part[0] == '*'}
		n.children=append(n.children,child)
	}
	child.insert(pattern,parts,height+1)
}

func (n *node) search(parts []string, height int) *node {
	if len(parts) == height || strings.HasPrefix(n.part, "*") {
		if n.pattern == "" {
			return nil
		}
		return n
	}

	part := parts[height]
	children := n.matchChildren(part)

	for _, child := range children {
		result := child.search(parts, height+1)
		if result != nil {
			return result
		}
	}

	return nil
}

func (n *node) travel(list *[]*node) {
	if n.pattern != "" {
		*list = append(*list, n)
	}
	for _, child := range n.children {
		child.travel(list)
	}
}

func (n *node) matchChild(part string) *node {
	for _, child := range n.children {
		fmt.Println("yb test1",child.pattern,child.part,child.children)
		if child.part == part || child.isWild {
			return child
		}
	}
	return nil
}

func (n *node) matchChildren(part string) []*node {
	nodes := make([]*node, 0)
	for _, child := range n.children {
		if child.part == part || child.isWild {
			nodes = append(nodes, child)
		}
	}
	return nodes
}