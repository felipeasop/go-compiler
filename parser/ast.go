package lexer

import (
	"fmt"
	"strings"

	"go-compiler/lexer"
)

// =====================================================================
// AST — ÁRVORE SINTÁTICA ABSTRATA
// =====================================================================
// Em Go usamos interfaces no lugar de classes abstratas do C++.
// Qualquer tipo que implemente Print() e ToJSON() é um ASTNode.
// =====================================================================

type ASTNode interface {
	Print(indent int)
	ToJSON(indent int) string
}

func pad(n int) string { return strings.Repeat(" ", n) }
func jpd(n int) string { return strings.Repeat("  ", n) }

// ─── ProgramNode ───────────────────────────────────────────────────
// Raiz da árvore. Corresponde a: Program → Statement*
type ProgramNode struct {
	Statements []ASTNode
}

func (n *ProgramNode) Print(indent int) {
	fmt.Printf("%sProgramNode\n", pad(indent))
	for _, s := range n.Statements {
		s.Print(indent + 2)
	}
}
func (n *ProgramNode) ToJSON(indent int) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s{\n%s\"tipo\": \"ProgramNode\",\n%s\"statements\": [\n", jpd(indent), jpd(indent+1), jpd(indent+1)))
	for i, s := range n.Statements {
		sb.WriteString(s.ToJSON(indent + 2))
		if i < len(n.Statements)-1 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}
	sb.WriteString(fmt.Sprintf("%s]\n%s}", jpd(indent+1), jpd(indent)))
	return sb.String()
}

// ─── VarDeclNode ───────────────────────────────────────────────────
// int x = expr;   ou   int x;
// Corresponde a: Declaration → "int" ID [ "=" Expr ] ";"
type VarDeclNode struct {
	Name        string
	Initializer ASTNode // nil se não houver inicialização
}

func (n *VarDeclNode) Print(indent int) {
	fmt.Printf("%sVarDeclNode (int %s)\n", pad(indent), n.Name)
	if n.Initializer != nil {
		n.Initializer.Print(indent + 4)
	}
}
func (n *VarDeclNode) ToJSON(indent int) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s{\n%s\"tipo\": \"VarDeclNode\",\n%s\"nome\": %q", jpd(indent), jpd(indent+1), jpd(indent+1), n.Name))
	if n.Initializer != nil {
		sb.WriteString(fmt.Sprintf(",\n%s\"inicializador\":\n", jpd(indent+1)))
		sb.WriteString(n.Initializer.ToJSON(indent + 2))
	}
	sb.WriteString(fmt.Sprintf("\n%s}", jpd(indent)))
	return sb.String()
}

// ─── AssignNode ────────────────────────────────────────────────────
// x = expr;
// Corresponde a: Assignment → ID "=" Expr ";"
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
// Em Go a chamada de impressão é fmt.Println — o scanner emite
// "fmt" como T_ID, "." não é um token, então parseamos como
// chamada de função genérica reconhecida pelo nome "fmt.Println".
// Corresponde a: PrintStmt → "fmt" "." "Println" "(" Expr ")" ";"
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
// if (cond) { ... } else { ... }
// Melhoria: suporte a else if em cadeia.
type ElseIfBlock struct {
	Condition ASTNode
	Body      []ASTNode
}

type IfNode struct {
	Condition    ASTNode
	ThenBranch   []ASTNode
	ElseIfBlocks []ElseIfBlock
	ElseBranch   []ASTNode
}

func (n *IfNode) Print(indent int) {
	fmt.Printf("%sIfNode\n", pad(indent))
	fmt.Printf("%sCondicao:\n", pad(indent+2))
	n.Condition.Print(indent + 4)
	fmt.Printf("%sThen:\n", pad(indent+2))
	for _, s := range n.ThenBranch {
		s.Print(indent + 4)
	}
	for _, eib := range n.ElseIfBlocks {
		fmt.Printf("%sElseIf:\n", pad(indent+2))
		eib.Condition.Print(indent + 4)
		for _, s := range eib.Body {
			s.Print(indent + 4)
		}
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
// Go usa "for" onde C/Java usam "while".
// Forma: for (cond) { ... }   →  for cond { ... }
// Corresponde a: ForStmt → "for" "(" Expr ")" "{" Statement* "}"
type ForNode struct {
	Condition ASTNode
	Body      []ASTNode
}

func (n *ForNode) Print(indent int) {
	fmt.Printf("%sForNode\n", pad(indent))
	fmt.Printf("%sCondicao:\n", pad(indent+2))
	n.Condition.Print(indent + 4)
	fmt.Printf("%sCorpo:\n", pad(indent+2))
	for _, s := range n.Body {
		s.Print(indent + 4)
	}
}
func (n *ForNode) ToJSON(indent int) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s{\n%s\"tipo\": \"ForNode\",\n%s\"condicao\":\n%s,\n%s\"corpo\": [\n",
		jpd(indent), jpd(indent+1), jpd(indent+1), n.Condition.ToJSON(indent+2), jpd(indent+1)))
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

// ─── BinaryOpNode ──────────────────────────────────────────────────
// Qualquer operação binária: +, -, *, /, ==, <, >, <=, >=
type BinaryOpNode struct {
	Op    lexer.TokenType
	Left  ASTNode
	Right ASTNode
}

func (n *BinaryOpNode) Print(indent int) {
	fmt.Printf("%sBinaryOpNode (%s)\n", pad(indent), n.Op)
	n.Left.Print(indent + 4)
	n.Right.Print(indent + 4)
}
func (n *BinaryOpNode) ToJSON(indent int) string {
	return fmt.Sprintf("%s{\n%s\"tipo\": \"BinaryOpNode\",\n%s\"op\": %q,\n%s\"esq\":\n%s,\n%s\"dir\":\n%s\n%s}",
		jpd(indent), jpd(indent+1), jpd(indent+1), n.Op.String(),
		jpd(indent+1), n.Left.ToJSON(indent+2),
		jpd(indent+1), n.Right.ToJSON(indent+2),
		jpd(indent))
}

// ─── NumberNode ────────────────────────────────────────────────────
type NumberNode struct{ Value int }

func (n *NumberNode) Print(indent int) {
	fmt.Printf("%sNumberNode (%d)\n", pad(indent), n.Value)
}
func (n *NumberNode) ToJSON(indent int) string {
	return fmt.Sprintf("%s{\"tipo\": \"NumberNode\", \"valor\": %d}", jpd(indent), n.Value)
}

// ─── FloatNode ─────────────────────────────────────────────────────
// Melhoria: suporte a float
type FloatNode struct{ Value float64 }

func (n *FloatNode) Print(indent int) {
	fmt.Printf("%sFloatNode (%g)\n", pad(indent), n.Value)
}
func (n *FloatNode) ToJSON(indent int) string {
	return fmt.Sprintf("%s{\"tipo\": \"FloatNode\", \"valor\": %g}", jpd(indent), n.Value)
}

// ─── StringNode ────────────────────────────────────────────────────
// Melhoria: suporte a string literal
type StringNode struct{ Value string }

func (n *StringNode) Print(indent int) {
	fmt.Printf("%sStringNode (%s)\n", pad(indent), n.Value)
}
func (n *StringNode) ToJSON(indent int) string {
	return fmt.Sprintf("%s{\"tipo\": \"StringNode\", \"valor\": %q}", jpd(indent), n.Value)
}

// ─── VariableNode ──────────────────────────────────────────────────
type VariableNode struct{ Name string }

func (n *VariableNode) Print(indent int) {
	fmt.Printf("%sVariableNode (%s)\n", pad(indent), n.Name)
}
func (n *VariableNode) ToJSON(indent int) string {
	return fmt.Sprintf("%s{\"tipo\": \"VariableNode\", \"nome\": %q}", jpd(indent), n.Name)
}
