package ast

import (
	"fmt"
	"strings"
)

// =====================================================================
// AST — ÁRVORE SINTÁTICA ABSTRATA
// =====================================================================

type ASTNode interface {
	Print(indent int)
	ToJSON(indent int) string
}

func pad(n int) string { return strings.Repeat(" ", n) }
func jpd(n int) string { return strings.Repeat("  ", n) }

// ─── ProgramNode ───────────────────────────────────────────────────
type ProgramNode struct {
	PackageName string
	Imports     []string
	Functions   []ASTNode
	Globals     []ASTNode
}

func (n *ProgramNode) Print(indent int) {
	fmt.Printf("%sProgramNode (package %s)\n", pad(indent), n.PackageName)
	for _, imp := range n.Imports {
		fmt.Printf("%sImport: %s\n", pad(indent+2), imp)
	}
	for _, g := range n.Globals {
		g.Print(indent + 2)
	}
	for _, f := range n.Functions {
		f.Print(indent + 2)
	}
}

func (n *ProgramNode) ToJSON(indent int) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s{\n%s\"tipo\": \"ProgramNode\",\n%s\"package\": %q,\n", jpd(indent), jpd(indent+1), jpd(indent+1), n.PackageName))
	sb.WriteString(fmt.Sprintf("%s\"imports\": [", jpd(indent+1)))
	for i, imp := range n.Imports {
		sb.WriteString(fmt.Sprintf("%q", imp))
		if i < len(n.Imports)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteString("],\n")
	sb.WriteString(fmt.Sprintf("%s\"globals\": [\n", jpd(indent+1)))
	for i, g := range n.Globals {
		sb.WriteString(g.ToJSON(indent + 2))
		if i < len(n.Globals)-1 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}
	sb.WriteString(fmt.Sprintf("%s],\n", jpd(indent+1)))
	sb.WriteString(fmt.Sprintf("%s\"functions\": [\n", jpd(indent+1)))
	for i, f := range n.Functions {
		sb.WriteString(f.ToJSON(indent + 2))
		if i < len(n.Functions)-1 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}
	sb.WriteString(fmt.Sprintf("%s]\n%s}", jpd(indent+1), jpd(indent)))
	return sb.String()
}
