package ast

import (
	"fmt"
	"strings"
)

// ─── BinaryOpNode ──────────────────────────────────────────────────
type BinaryOpNode struct {
	OpStr string
	Left  ASTNode
	Right ASTNode
	Line  int
}

func (n *BinaryOpNode) Print(indent int) {
	fmt.Printf("%sBinaryOpNode (%s)\n", pad(indent), n.OpStr)
	n.Left.Print(indent + 4)
	n.Right.Print(indent + 4)
}

func (n *BinaryOpNode) ToJSON(indent int) string {
	return fmt.Sprintf("%s{\n%s\"tipo\": \"BinaryOpNode\",\n%s\"op\": %q,\n%s\"esq\":\n%s,\n%s\"dir\":\n%s\n%s}",
		jpd(indent), jpd(indent+1), jpd(indent+1), n.OpStr,
		jpd(indent+1), n.Left.ToJSON(indent+2),
		jpd(indent+1), n.Right.ToJSON(indent+2),
		jpd(indent))
}

// ─── UnaryOpNode ───────────────────────────────────────────────────
// Cobre: !expr, -expr
type UnaryOpNode struct {
	OpStr   string
	Operand ASTNode
	Line    int
}

func (n *UnaryOpNode) Print(indent int) {
	fmt.Printf("%sUnaryOpNode (%s)\n", pad(indent), n.OpStr)
	n.Operand.Print(indent + 4)
}

func (n *UnaryOpNode) ToJSON(indent int) string {
	return fmt.Sprintf("%s{\n%s\"tipo\": \"UnaryOpNode\",\n%s\"op\": %q,\n%s\"operando\":\n%s\n%s}",
		jpd(indent), jpd(indent+1), jpd(indent+1), n.OpStr,
		jpd(indent+1), n.Operand.ToJSON(indent+2), jpd(indent))
}

// ─── LiteralNode ───────────────────────────────────────────────────
// Cobre INT, FLOAT, STRING, IMAG, CHAR, true, false
type LiteralNode struct {
	Value string
	Line  int
}

func (n *LiteralNode) Print(indent int) {
	fmt.Printf("%sLiteralNode (%s)\n", pad(indent), n.Value)
}
func (n *LiteralNode) ToJSON(indent int) string {
	return fmt.Sprintf("%s{\"tipo\": \"LiteralNode\", \"valor\": %q}", jpd(indent), n.Value)
}

// ─── VariableNode ──────────────────────────────────────────────────
type VariableNode struct {
	Name string
	Line int // usado pelo Semantic Analyzer no relatório de erros
}

func (n *VariableNode) Print(indent int) {
	fmt.Printf("%sVariableNode (%s)\n", pad(indent), n.Name)
}
func (n *VariableNode) ToJSON(indent int) string {
	return fmt.Sprintf("%s{\"tipo\": \"VariableNode\", \"nome\": %q}", jpd(indent), n.Name)
}

// ─── CallNode ──────────────────────────────────────────────────────
// Chamada de função: nome(arg1, arg2, ...). Também representa as
// conversões de tipo int(x) e float64(x), que em Go têm a mesma sintaxe.
type CallNode struct {
	Name string
	Args []ASTNode
	Line int
}

func (n *CallNode) Print(indent int) {
	fmt.Printf("%sCallNode (%s)\n", pad(indent), n.Name)
	for _, a := range n.Args {
		a.Print(indent + 4)
	}
}

func (n *CallNode) ToJSON(indent int) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s{\n%s\"tipo\": \"CallNode\",\n%s\"nome\": %q,\n%s\"args\": [\n",
		jpd(indent), jpd(indent+1), jpd(indent+1), n.Name, jpd(indent+1)))
	writeJSONList(&sb, n.Args, indent+2)
	sb.WriteString(fmt.Sprintf("%s]\n%s}", jpd(indent+1), jpd(indent)))
	return sb.String()
}
