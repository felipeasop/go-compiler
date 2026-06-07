package parser

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

// ─── FuncNode ──────────────────────────────────────────────────────
type FuncNode struct {
	Name string
	Body []ASTNode
}

func (n *FuncNode) Print(indent int) {
	fmt.Printf("%sFuncNode (func %s)\n", pad(indent), n.Name)
	for _, s := range n.Body {
		s.Print(indent + 2)
	}
}

func (n *FuncNode) ToJSON(indent int) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s{\n%s\"tipo\": \"FuncNode\",\n%s\"nome\": %q,\n%s\"corpo\": [\n", jpd(indent), jpd(indent+1), jpd(indent+1), n.Name, jpd(indent+1)))
	for i, s := range n.Body {
		sb.WriteString(s.ToJSON(indent + 2))
		if i < len(n.Body)-1 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}
	sb.WriteString(fmt.Sprintf("%s]\n%s}", jpd(indent+1), jpd(indent)))
	return sb.String()
}

// ─── VarDeclNode ───────────────────────────────────────────────────
type VarDeclNode struct {
	Name        string
	Initializer ASTNode
}

func (n *VarDeclNode) Print(indent int) {
	fmt.Printf("%sVarDeclNode (%s)\n", pad(indent), n.Name)
	if n.Initializer != nil {
		n.Initializer.Print(indent + 4)
	}
}

func (n *VarDeclNode) ToJSON(indent int) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s{\n%s\"tipo\": \"VarDeclNode\",\n%s\"nome\": %q", jpd(indent), jpd(indent+1), jpd(indent+1), n.Name))
	if n.Initializer != nil {
		sb.WriteString(fmt.Sprintf(",\n%s\"inicializador\":\n%s", jpd(indent+1), n.Initializer.ToJSON(indent+2)))
	}
	sb.WriteString(fmt.Sprintf("\n%s}", jpd(indent)))
	return sb.String()
}

// ─── AssignNode ────────────────────────────────────────────────────
type AssignNode struct {
	Name string
	Expr ASTNode
}

func (n *AssignNode) Print(indent int) {
	fmt.Printf("%sAssignNode (%s)\n", pad(indent), n.Name)
	n.Expr.Print(indent + 4)
}

func (n *AssignNode) ToJSON(indent int) string {
	return fmt.Sprintf("%s{\n%s\"tipo\": \"AssignNode\",\n%s\"nome\": %q,\n%s\"expr\":\n%s\n%s}",
		jpd(indent), jpd(indent+1), jpd(indent+1), n.Name, jpd(indent+1), n.Expr.ToJSON(indent+2), jpd(indent))
}

// ─── PrintCallNode ─────────────────────────────────────────────────
// fmt.Println(expr)
type PrintCallNode struct {
	Expr ASTNode
}

func (n *PrintCallNode) Print(indent int) {
	fmt.Printf("%sPrintCallNode (fmt.Println)\n", pad(indent))
	n.Expr.Print(indent + 4)
}

func (n *PrintCallNode) ToJSON(indent int) string {
	return fmt.Sprintf("%s{\n%s\"tipo\": \"PrintCallNode\",\n%s\"expr\":\n%s\n%s}",
		jpd(indent), jpd(indent+1), jpd(indent+1), n.Expr.ToJSON(indent+2), jpd(indent))
}

// ─── IfNode ────────────────────────────────────────────────────────
// else if é representado colocando um IfNode dentro de ElseBranch.
type IfNode struct {
	Init       ASTNode
	Condition  ASTNode
	ThenBranch []ASTNode
	ElseBranch []ASTNode
}

func (n *IfNode) Print(indent int) {
	fmt.Printf("%sIfNode\n", pad(indent))
	fmt.Printf("%sCondicao:\n", pad(indent+2))
	n.Condition.Print(indent + 4)
	fmt.Printf("%sThen:\n", pad(indent+2))
	for _, s := range n.ThenBranch {
		s.Print(indent + 4)
	}
	if len(n.ElseBranch) > 0 {
		fmt.Printf("%sElse:\n", pad(indent+2))
		for _, s := range n.ElseBranch {
			s.Print(indent + 4)
		}
	}
}

func (n *IfNode) ToJSON(indent int) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s{\n%s\"tipo\": \"IfNode\",\n%s\"condicao\":\n%s,\n%s\"then\": [\n",
		jpd(indent), jpd(indent+1), jpd(indent+1), n.Condition.ToJSON(indent+2), jpd(indent+1)))
	for i, s := range n.ThenBranch {
		sb.WriteString(s.ToJSON(indent + 2))
		if i < len(n.ThenBranch)-1 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}
	sb.WriteString(fmt.Sprintf("%s]", jpd(indent+1)))
	if len(n.ElseBranch) > 0 {
		sb.WriteString(fmt.Sprintf(",\n%s\"else\": [\n", jpd(indent+1)))
		for i, s := range n.ElseBranch {
			sb.WriteString(s.ToJSON(indent + 2))
			if i < len(n.ElseBranch)-1 {
				sb.WriteString(",")
			}
			sb.WriteString("\n")
		}
		sb.WriteString(fmt.Sprintf("%s]", jpd(indent+1)))
	}
	sb.WriteString(fmt.Sprintf("\n%s}", jpd(indent)))
	return sb.String()
}

// ─── ForNode ───────────────────────────────────────────────────────
type ForNode struct {
	Condition ASTNode
	Body      []ASTNode
}

func (n *ForNode) Print(indent int) {
	fmt.Printf("%sForNode\n", pad(indent))
	fmt.Printf("%sCondicao:\n", pad(indent+2))
	if n.Condition != nil {
		n.Condition.Print(indent + 4)
	} else {
		fmt.Printf("%s(loop infinito)\n", pad(indent+4))
	}
	fmt.Printf("%sCorpo:\n", pad(indent+2))
	for _, s := range n.Body {
		s.Print(indent + 4)
	}
}

func (n *ForNode) ToJSON(indent int) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s{\n%s\"tipo\": \"ForNode\",\n", jpd(indent), jpd(indent+1)))
	if n.Condition != nil {
		sb.WriteString(fmt.Sprintf("%s\"condicao\":\n%s,\n", jpd(indent+1), n.Condition.ToJSON(indent+2)))
	} else {
		sb.WriteString(fmt.Sprintf("%s\"condicao\": null,\n", jpd(indent+1)))
	}
	sb.WriteString(fmt.Sprintf("%s\"corpo\": [\n", jpd(indent+1)))
	for i, s := range n.Body {
		sb.WriteString(s.ToJSON(indent + 2))
		if i < len(n.Body)-1 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}
	sb.WriteString(fmt.Sprintf("%s]\n%s}", jpd(indent+1), jpd(indent)))
	return sb.String()
}

// ─── ReturnNode ────────────────────────────────────────────────────
// return [expr]
type ReturnNode struct {
	Value ASTNode // nil para return sem valor
}

func (n *ReturnNode) Print(indent int) {
	fmt.Printf("%sReturnNode\n", pad(indent))
	if n.Value != nil {
		n.Value.Print(indent + 4)
	}
}

func (n *ReturnNode) ToJSON(indent int) string {
	if n.Value == nil {
		return fmt.Sprintf("%s{\"tipo\": \"ReturnNode\", \"valor\": null}", jpd(indent))
	}
	return fmt.Sprintf("%s{\n%s\"tipo\": \"ReturnNode\",\n%s\"valor\":\n%s\n%s}",
		jpd(indent), jpd(indent+1), jpd(indent+1), n.Value.ToJSON(indent+2), jpd(indent))
}

// ─── BreakNode ─────────────────────────────────────────────────────
type BreakNode struct{}

func (n *BreakNode) Print(indent int) {
	fmt.Printf("%sBreakNode\n", pad(indent))
}
func (n *BreakNode) ToJSON(indent int) string {
	return fmt.Sprintf("%s{\"tipo\": \"BreakNode\"}", jpd(indent))
}

// ─── ContinueNode ──────────────────────────────────────────────────
type ContinueNode struct{}

func (n *ContinueNode) Print(indent int) {
	fmt.Printf("%sContinueNode\n", pad(indent))
}
func (n *ContinueNode) ToJSON(indent int) string {
	return fmt.Sprintf("%s{\"tipo\": \"ContinueNode\"}", jpd(indent))
}

// ─── BinaryOpNode ──────────────────────────────────────────────────
type BinaryOpNode struct {
	OpStr string
	Left  ASTNode
	Right ASTNode
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
}

func (n *VariableNode) Print(indent int) {
	fmt.Printf("%sVariableNode (%s)\n", pad(indent), n.Name)
}
func (n *VariableNode) ToJSON(indent int) string {
	return fmt.Sprintf("%s{\"tipo\": \"VariableNode\", \"nome\": %q}", jpd(indent), n.Name)
}
