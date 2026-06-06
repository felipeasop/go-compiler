package parser

import (
	"fmt"
	"strconv"
	"strings"
)

// =====================================================================
// PARTE 2: A ÁRVORE SINTÁTICA ABSTRATA (AST - ABSTRACT SYNTAX TREE)
// =====================================================================
// O Parser não gera apenas um "Sim/Não". Ele constrói uma árvore na memória!
// Nessa árvore, nós internos são operações (como "+" ou "atribuição")
// e as folhas são dados (como números ou variáveis).
// Em Go, polimorfismo é expresso por interfaces — não por herança.
// =====================================================================

// ASTNode é a interface base para todos os nós da árvore.
// Substitui a classe abstrata com método virtual puro do C++.
// Qualquer struct que implemente Print(int) satisfaz a interface automaticamente.
type ASTNode interface {
	Print(indent int)
}

func pad(n int) string { return strings.Repeat(" ", n) }

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
	Op    scanner.TokenType
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

// =====================================================================
// PARTE 3: O ANALISADOR SINTÁTICO (PARSER DESCENDENTE RECURSIVO)
// =====================================================================
// Cada regra da gramática vira diretamente uma função em Go.
// O parser dá uma "espiada" (Lookahead) no próximo token via peek()
// para decidir qual caminho seguir — sem backtracking.
// =====================================================================

type Parser struct {
	tokens []scanner.Token
	pos    int
}

func NewParser(tokens []scanner.Token) *Parser {
	return &Parser{tokens: tokens}
}

// ── helpers de lookahead ─────────────────────────────────────────────

func (p *Parser) peek() scanner.TokenType {
	if p.pos >= len(p.tokens) {
		return scanner.T_EOF
	}
	return p.tokens[p.pos].Type
}

func (p *Parser) peekToken() scanner.Token {
	if p.pos >= len(p.tokens) {
		return scanner.Token{Type: scanner.T_EOF, Lexeme: "", Line: -1}
	}
	return p.tokens[p.pos]
}

// advance avança o ponteiro e retorna o token consumido.
func (p *Parser) advance() scanner.Token {
	t := p.peekToken()
	if t.Type != scanner.T_EOF {
		p.pos++
	}
	return t
}

// match valida que o token atual é o esperado e o consome; retorna erro caso contrário.
func (p *Parser) match(expected scanner.TokenType) error {
	if p.peek() == expected {
		p.advance()
		return nil
	}
	t := p.peekToken()
	return fmt.Errorf("erro sintático na linha %d: esperava '%s' porém foi encontrado '%s'",
		t.Line, expected, t.Lexeme)
}

func (p *Parser) syntaxError(msg string) error {
	t := p.peekToken()
	return fmt.Errorf("erro sintático na linha %d: %s", t.Line, msg)
}

// ── regras da gramática ──────────────────────────────────────────────

// Regra: Program -> Statement*
func (p *Parser) ParseProgram() (*ProgramNode, error) {
	prog := &ProgramNode{}
	for p.peek() != scanner.T_EOF {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		prog.Statements = append(prog.Statements, stmt)
	}
	return prog, nil
}

// Regra: Statement -> Declaration | Assignment | PrintStmt | IfStmt | WhileStmt
// Usa LOOKAHEAD (peek) para decidir qual caminho seguir.
func (p *Parser) parseStatement() (ASTNode, error) {
	switch p.peek() {
	case scanner.T_INT:
		return p.parseDeclaration()
	case scanner.T_ID:
		return p.parseAssignment()
	case scanner.T_PRINT:
		return p.parsePrintStmt()
	case scanner.T_IF:
		return p.parseIfStmt()
	case scanner.T_WHILE:
		return p.parseWhileStmt()
	default:
		return nil, p.syntaxError(
			fmt.Sprintf("comando inválido ou não reconhecido: '%s'", p.peekToken().Lexeme),
		)
	}
}

// Regra: Declaration -> "int" ID [ "=" Expression ] ";"
func (p *Parser) parseDeclaration() (ASTNode, error) {
	if err := p.match(scanner.T_INT); err != nil {
		return nil, err
	}
	idTok := p.peekToken()
	if err := p.match(scanner.T_ID); err != nil {
		return nil, err
	}

	var initializer ASTNode
	if p.peek() == scanner.T_ASSIGN {
		p.advance()
		var err error
		initializer, err = p.parseExpression()
		if err != nil {
			return nil, err
		}
	}

	if err := p.match(scanner.T_SEMICOLON); err != nil {
		return nil, err
	}
	return &VarDeclNode{Name: idTok.Lexeme, Initializer: initializer}, nil
}

// Regra: Assignment -> ID "=" Expression ";"
func (p *Parser) parseAssignment() (ASTNode, error) {
	idTok := p.peekToken()
	if err := p.match(scanner.T_ID); err != nil {
		return nil, err
	}
	if err := p.match(scanner.T_ASSIGN); err != nil {
		return nil, err
	}
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if err := p.match(scanner.T_SEMICOLON); err != nil {
		return nil, err
	}
	return &AssignNode{Name: idTok.Lexeme, Expr: expr}, nil
}

// Regra: PrintStmt -> "print" "(" Expression ")" ";"
func (p *Parser) parsePrintStmt() (ASTNode, error) {
	if err := p.match(scanner.T_PRINT); err != nil {
		return nil, err
	}
	if err := p.match(scanner.T_LPAREN); err != nil {
		return nil, err
	}
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if err := p.match(scanner.T_RPAREN); err != nil {
		return nil, err
	}
	if err := p.match(scanner.T_SEMICOLON); err != nil {
		return nil, err
	}
	return &PrintNode{Expr: expr}, nil
}

// Regra: IfStmt -> "if" "(" Expression ")" "{" Statement* "}" [ "else" "{" Statement* "}" ]
func (p *Parser) parseIfStmt() (ASTNode, error) {
	if err := p.match(scanner.T_IF); err != nil {
		return nil, err
	}
	if err := p.match(scanner.T_LPAREN); err != nil {
		return nil, err
	}
	cond, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if err := p.match(scanner.T_RPAREN); err != nil {
		return nil, err
	}

	thenBranch, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	var elseBranch []ASTNode
	if p.peek() == scanner.T_ELSE {
		p.advance()
		elseBranch, err = p.parseBlock()
		if err != nil {
			return nil, err
		}
	}

	return &IfNode{Condition: cond, ThenBranch: thenBranch, ElseBranch: elseBranch}, nil
}

// Regra: WhileStmt -> "while" "(" Expression ")" "{" Statement* "}"
func (p *Parser) parseWhileStmt() (ASTNode, error) {
	if err := p.match(scanner.T_WHILE); err != nil {
		return nil, err
	}
	if err := p.match(scanner.T_LPAREN); err != nil {
		return nil, err
	}
	cond, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if err := p.match(scanner.T_RPAREN); err != nil {
		return nil, err
	}
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	return &WhileNode{Condition: cond, Body: body}, nil
}

// parseBlock lê "{" Statement* "}" e retorna a lista de nós.
func (p *Parser) parseBlock() ([]ASTNode, error) {
	if err := p.match(scanner.T_LBRACE); err != nil {
		return nil, err
	}
	var stmts []ASTNode
	for p.peek() != scanner.T_RBRACE && p.peek() != scanner.T_EOF {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		stmts = append(stmts, stmt)
	}
	if err := p.match(scanner.T_RBRACE); err != nil {
		return nil, err
	}
	return stmts, nil
}

// =====================================================================
// CASCATA DE PRECEDÊNCIA (EXPRESSÕES)
// =====================================================================
// Camada 1: Expression  -> relacionais  (==, <, >)  — precedência mais baixa
// Camada 2: SimpleExpr  -> aditivos     (+, -)
// Camada 3: Term        -> multiplicativos (*, /)
// Camada 4: Factor      -> unidades básicas (número, variável, parênteses)
//
// O parser desce a cascata, construindo os nós de maior prioridade primeiro.
// =====================================================================

// Camada 1: Expression -> SimpleExpr [ ( "==" | "<" | ">" ) SimpleExpr ]
func (p *Parser) parseExpression() (ASTNode, error) {
	left, err := p.parseSimpleExpr()
	if err != nil {
		return nil, err
	}
	op := p.peek()
	if op == scanner.T_EQ || op == scanner.T_LT || op == scanner.T_GT {
		p.advance()
		right, err := p.parseSimpleExpr()
		if err != nil {
			return nil, err
		}
		left = &BinaryOpNode{Op: op, Left: left, Right: right}
	}
	return left, nil
}

// Camada 2: SimpleExpr -> Term ( ( "+" | "-" ) Term )*
func (p *Parser) parseSimpleExpr() (ASTNode, error) {
	left, err := p.parseTerm()
	if err != nil {
		return nil, err
	}
	for p.peek() == scanner.T_PLUS || p.peek() == scanner.T_MINUS {
		op := p.peek()
		p.advance()
		right, err := p.parseTerm()
		if err != nil {
			return nil, err
		}
		left = &BinaryOpNode{Op: op, Left: left, Right: right}
	}
	return left, nil
}

// Camada 3: Term -> Factor ( ( "*" | "/" ) Factor )*
func (p *Parser) parseTerm() (ASTNode, error) {
	left, err := p.parseFactor()
	if err != nil {
		return nil, err
	}
	for p.peek() == scanner.T_MULT || p.peek() == scanner.T_DIV {
		op := p.peek()
		p.advance()
		right, err := p.parseFactor()
		if err != nil {
			return nil, err
		}
		left = &BinaryOpNode{Op: op, Left: left, Right: right}
	}
	return left, nil
}

// Camada 4 (Base): Factor -> NUMBER | ID | "(" Expression ")"
func (p *Parser) parseFactor() (ASTNode, error) {
	switch p.peek() {
	case scanner.T_NUM:
		t := p.advance()
		val, err := strconv.Atoi(t.Lexeme)
		if err != nil {
			return nil, fmt.Errorf("erro sintático na linha %d: número inválido '%s'", t.Line, t.Lexeme)
		}
		return &NumberNode{Value: val}, nil

	case scanner.T_ID:
		t := p.advance()
		return &VariableNode{Name: t.Lexeme}, nil

	case scanner.T_LPAREN:
		p.advance() // consome "("
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if err := p.match(scanner.T_RPAREN); err != nil {
			return nil, err
		}
		return expr, nil

	default:
		return nil, p.syntaxError(
			fmt.Sprintf("fator inválido na expressão (esperava número, variável ou '('): encontrado '%s'",
				p.peekToken().Lexeme),
		)
	}
}
