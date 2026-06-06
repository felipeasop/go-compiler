package parser

import (
	"fmt"
	"strings"

	"modulariza/lexer"
)

// =====================================================================
// PARTE 2: A ÁRVORE SINTÁTICA ABSTRATA (AST - ABSTRACT SYNTAX TREE)
// =====================================================================
// O Parser não gera apenas um "Sim/Não". Ele constrói uma árvore na memória!
// Nessa árvore, nós internos são operações (como "+" ou "atribuição")
// e as folhas são dados (como números ou variáveis).
// Em Go, polimorfismo é expresso por interfaces — não por herança.
// =====================================================================

func pad(n int) string { return strings.Repeat(" ", n) }

// ASTNode é a interface implementada por todos os nós da árvore.
type ASTNode interface {
	Print(indent int)
}

// ── Nó Raiz: lista de comandos do programa ───────────────────────────

type ProgramNode struct {
	Statements []ASTNode
}

func (n *ProgramNode) Print(indent int) {
	fmt.Printf("%sProgramNode (Inicio do Programa)\n", pad(indent))
	for _, stmt := range n.Statements {
		stmt.Print(indent + 2)
	}
}

// ── VarDeclNode: "int x = 10;" ───────────────────────────────────────

type VarDeclNode struct {
	Name        string
	Initializer ASTNode // nil se não houver valor inicial
}

func (n *VarDeclNode) Print(indent int) {
	fmt.Printf("%sVarDeclNode (Declaracao de Variavel: %s)\n", pad(indent), n.Name)
	if n.Initializer != nil {
		n.Initializer.Print(indent + 4)
	}
}

// ── AssignNode: "x = 20;" ────────────────────────────────────────────

type AssignNode struct {
	Name string
	Expr ASTNode
}

func (n *AssignNode) Print(indent int) {
	fmt.Printf("%sAssignNode (Atribuicao a variavel: %s)\n", pad(indent), n.Name)
	n.Expr.Print(indent + 4)
}

// ── PrintNode: "print(soma);" ────────────────────────────────────────

type PrintNode struct {
	Expr ASTNode
}

func (n *PrintNode) Print(indent int) {
	fmt.Printf("%sPrintNode (Comando Print)\n", pad(indent))
	n.Expr.Print(indent + 4)
}

// ── IfNode: "if (cond) { ... } else { ... }" ─────────────────────────

type IfNode struct {
	Condition  ASTNode
	ThenBranch []ASTNode
	ElseBranch []ASTNode // vazio se não houver else
}

func (n *IfNode) Print(indent int) {
	fmt.Printf("%sIfNode (Condicional IF)\n", pad(indent))
	fmt.Printf("%sCondicao:\n", pad(indent+2))
	n.Condition.Print(indent + 4)

	fmt.Printf("%sBloco 'Então':\n", pad(indent+2))
	for _, stmt := range n.ThenBranch {
		stmt.Print(indent + 4)
	}

	if len(n.ElseBranch) > 0 {
		fmt.Printf("%sBloco 'Senão':\n", pad(indent+2))
		for _, stmt := range n.ElseBranch {
			stmt.Print(indent + 4)
		}
	}
}

// ── WhileNode: "while (cond) { ... }" ────────────────────────────────

type WhileNode struct {
	Condition ASTNode
	Body      []ASTNode
}

func (n *WhileNode) Print(indent int) {
	fmt.Printf("%sWhileNode (Laco de Repeticao WHILE)\n", pad(indent))
	fmt.Printf("%sCondicao de entrada:\n", pad(indent+2))
	n.Condition.Print(indent + 4)

	fmt.Printf("%sCorpo do laco:\n", pad(indent+2))
	for _, stmt := range n.Body {
		stmt.Print(indent + 4)
	}
}

// ── BinaryOpNode: left OP right ──────────────────────────────────────

type BinaryOpNode struct {
	Op    lexer.TokenType
	Left  ASTNode
	Right ASTNode
}

func (n *BinaryOpNode) Print(indent int) {
	fmt.Printf("%sBinaryOpNode (Operacao Binaria: %s)\n", pad(indent), n.Op)
	n.Left.Print(indent + 4)
	n.Right.Print(indent + 4)
}

// ── NumberNode: literal inteiro (folha) ──────────────────────────────

type NumberNode struct{ Value int }

func (n *NumberNode) Print(indent int) {
	fmt.Printf("%sNumberNode (Valor constante: %d)\n", pad(indent), n.Value)
}

// ── VariableNode: referência a variável (folha) ──────────────────────

type VariableNode struct{ Name string }

func (n *VariableNode) Print(indent int) {
	fmt.Printf("%sVariableNode (Busca variavel: %s)\n", pad(indent), n.Name)
}
