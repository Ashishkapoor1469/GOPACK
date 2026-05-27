package graph

import (
	"fmt"
	"strings"
)

type DepNode struct {
	Name         string
	Version      string
	Dependencies []*DepNode
}

// BuildTree builds a mock/simulated dependency tree based on installed/manifest packages for rendering
func BuildTree(directDeps map[string]string) *DepNode {
	root := &DepNode{
		Name:    "my-app",
		Version: "1.0.0",
	}

	// We seed a mock tree for rendering so that it demonstrates circular and conflict detection:
	// react -> scheduler
	// next -> react, react-dom -> react
	// circular: a -> b -> a
	for k, v := range directDeps {
		node := &DepNode{Name: k, Version: strings.TrimLeft(v, "^~")}
		root.Dependencies = append(root.Dependencies, node)

		// Seed some transitive dependencies
		if k == "next" {
			node.Dependencies = append(node.Dependencies, &DepNode{Name: "react", Version: "18.3.1"})
			node.Dependencies = append(node.Dependencies, &DepNode{Name: "react-dom", Version: "18.3.1"})
		} else if k == "react" {
			node.Dependencies = append(node.Dependencies, &DepNode{Name: "scheduler", Version: "0.23.0"})
		} else if k == "lodash" {
			// Seed a circular dependency b -> a -> b for lodash just to showcase the ↺ warning
			circle1 := &DepNode{Name: "lodash-utils", Version: "1.0.0"}
			circle2 := &DepNode{Name: "lodash", Version: "4.17.21"} // points back
			circle1.Dependencies = append(circle1.Dependencies, circle2)
			node.Dependencies = append(node.Dependencies, circle1)
		}
	}

	return root
}

// RenderTree renders the dependency tree as an ASCII/Unicode string
func RenderTree(node *DepNode, depth int, prefixes []string, visited map[string]bool) string {
	var sb strings.Builder

	if depth == 0 {
		sb.WriteString(fmt.Sprintf("%s@%s\n", node.Name, node.Version))
	}

	if visited == nil {
		visited = make(map[string]bool)
	}

	// Detect circular dependency
	nodeKey := fmt.Sprintf("%s@%s", node.Name, node.Version)
	isCircular := visited[nodeKey]

	visited[nodeKey] = true
	defer func() { visited[nodeKey] = false }()

	numChildren := len(node.Dependencies)
	for i, child := range node.Dependencies {
		isLast := i == numChildren-1

		// Prepare tree drawing prefix
		var connector string
		if isLast {
			connector = "└── "
		} else {
			connector = "├── "
		}

		prefixStr := strings.Join(prefixes, "")
		sb.WriteString(prefixStr)
		sb.WriteString(connector)

		// Check for conflicts (mock detection: if name is react and version is different from the direct dep version)
		warning := ""
		if child.Name == "react" && child.Version != "18.3.1" && depth > 0 {
			warning = " \033[33m⚠ duplicate version conflict (using multiple versions)\033[0m"
		}

		if isCircular {
			sb.WriteString(fmt.Sprintf("\033[31m%s@%s ↺ (circular dependency)\033[0m\n", child.Name, child.Version))
			continue
		}

		sb.WriteString(fmt.Sprintf("%s@%s%s\n", child.Name, child.Version, warning))

		// Recurse
		var nextPrefixes []string
		nextPrefixes = append(nextPrefixes, prefixes...)
		if isLast {
			nextPrefixes = append(nextPrefixes, "    ")
		} else {
			nextPrefixes = append(nextPrefixes, "│   ")
		}

		sb.WriteString(RenderTree(child, depth+1, nextPrefixes, visited))
	}

	return sb.String()
}

// RenderDOT exports the graph as a Graphviz DOT script
func RenderDOT(node *DepNode) string {
	var sb strings.Builder
	sb.WriteString("digraph G {\n")
	sb.WriteString("  rankdir=LR;\n")
	sb.WriteString("  node [shape=box, fontname=\"Courier\"];\n")

	var declare func(n *DepNode, visited map[string]bool)
	declare = func(n *DepNode, visited map[string]bool) {
		key := fmt.Sprintf("%s_%s", strings.ReplaceAll(n.Name, "-", "_"), strings.ReplaceAll(n.Version, ".", "_"))
		key = strings.ReplaceAll(key, "@", "")
		key = strings.ReplaceAll(key, "/", "_")

		if visited[key] {
			return
		}
		visited[key] = true

		sb.WriteString(fmt.Sprintf("  %s [label=\"%s@%s\"];\n", key, n.Name, n.Version))

		for _, c := range n.Dependencies {
			ckey := fmt.Sprintf("%s_%s", strings.ReplaceAll(c.Name, "-", "_"), strings.ReplaceAll(c.Version, ".", "_"))
			ckey = strings.ReplaceAll(ckey, "@", "")
			ckey = strings.ReplaceAll(ckey, "/", "_")

			sb.WriteString(fmt.Sprintf("  %s -> %s;\n", key, ckey))
			declare(c, visited)
		}
	}

	declare(node, make(map[string]bool))
	sb.WriteString("}\n")
	return sb.String()
}
