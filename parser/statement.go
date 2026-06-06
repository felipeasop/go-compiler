package parser

import (
	"fmt"

	"modulariza/lexer"
)

// =====================================================================
// PARTE 3: O ANALISADOR SINTÁTICO (PARSER DESCENDENTE RECURSIVO)
// =====================================================================
// Cada regra da gramática vira diretamente uma função em Go.
// O parser dá uma "espiada" (Lookahead) no próximo token via peek()
// para decidir qual caminho seguir — sem backtracking.
// =====================================================================

type Parser struct {
	tokens []lexer.Token
	pos    int
}

func NewParser(tokens []lexer.Token) *Parser {
	return &Parser{tokens: tokens}
}

// ── helpers de lookahead ─────────────────────────────────────────────

func (p *Parser) peek() lexer.TokenType {
	if p.pos >= len(p.tokens) {
		return lexer.T_EOF
	}
	return p.tokens[p.pos].Type
}

func (p *Parser) peekToken() lexer.Token {
	if p.pos >= len(p.tokens) {
		return lexer.Token{Type: lexer.T_EOF, Lexeme: "", Line: -1}
	}
	return p.tokens[p.pos]
}

// advance avança o ponteiro e retorna o token consumido.
func (p *Parser) advance() lexer.Token {
	t := p.peekToken()
	if t.Type != lexer.T_EOF {
		p.pos++
	}
	return t
}

// match valida que o token atual é o esperado e o consome; retorna erro caso contrário.
func (p *Parser) match(expected lexer.TokenType) error {
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
	for p.peek() != lexer.T_EOF {
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
	case lexer.T_INT:
		return p.parseDeclaration()
	case lexer.T_ID:
		return p.parseAssignment()
	case lexer.T_PRINT:
		return p.parsePrintStmt()
	case lexer.T_IF:
		return p.parseIfStmt()
	case lexer.T_WHILE:
		return p.parseWhileStmt()
	default:
		return nil, p.syntaxError(
			fmt.Sprintf("comando inválido ou não reconhecido: '%s'", p.peekToken().Lexeme),
		)
	}
}

// Regra: Declaration -> "int" ID [ "=" Expression ] ";"
func (p *Parser) parseDeclaration() (ASTNode, error) {
	if err := p.match(lexer.T_INT); err != nil {
		return nil, err
	}
	idTok := p.peekToken()
	if err := p.match(lexer.T_ID); err != nil {
		return nil, err
	}

	var initializer ASTNode
	if p.peek() == lexer.T_ASSIGN {
		p.advance()
		var err error
		initializer, err = p.parseExpression()
		if err != nil {
			return nil, err
		}
	}

	if err := p.match(lexer.T_SEMICOLON); err != nil {
		return nil, err
	}
	return &VarDeclNode{Name: idTok.Lexeme, Initializer: initializer}, nil
}

// Regra: Assignment -> ID "=" Expression ";"
func (p *Parser) parseAssignment() (ASTNode, error) {
	idTok := p.peekToken()
	if err := p.match(lexer.T_ID); err != nil {
		return nil, err
	}
	if err := p.match(lexer.T_ASSIGN); err != nil {
		return nil, err
	}
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if err := p.match(lexer.T_SEMICOLON); err != nil {
		return nil, err
	}
	return &AssignNode{Name: idTok.Lexeme, Expr: expr}, nil
}

// Regra: PrintStmt -> "print" "(" Expression ")" ";"
func (p *Parser) parsePrintStmt() (ASTNode, error) {
	if err := p.match(lexer.T_PRINT); err != nil {
		return nil, err
	}
	if err := p.match(lexer.T_LPAREN); err != nil {
		return nil, err
	}
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if err := p.match(lexer.T_RPAREN); err != nil {
		return nil, err
	}
	if err := p.match(lexer.T_SEMICOLON); err != nil {
		return nil, err
	}
	return &PrintNode{Expr: expr}, nil
}

// Regra: IfStmt -> "if" "(" Expression ")" "{" Statement* "}" [ "else" "{" Statement* "}" ]
func (p *Parser) parseIfStmt() (ASTNode, error) {
	if err := p.match(lexer.T_IF); err != nil {
		return nil, err
	}
	if err := p.match(lexer.T_LPAREN); err != nil {
		return nil, err
	}
	cond, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if err := p.match(lexer.T_RPAREN); err != nil {
		return nil, err
	}

	thenBranch, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	var elseBranch []ASTNode
	if p.peek() == lexer.T_ELSE {
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
	if err := p.match(lexer.T_WHILE); err != nil {
		return nil, err
	}
	if err := p.match(lexer.T_LPAREN); err != nil {
		return nil, err
	}
	cond, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if err := p.match(lexer.T_RPAREN); err != nil {
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
	if err := p.match(lexer.T_LBRACE); err != nil {
		return nil, err
	}
	var stmts []ASTNode
	for p.peek() != lexer.T_RBRACE && p.peek() != lexer.T_EOF {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		stmts = append(stmts, stmt)
	}
	if err := p.match(lexer.T_RBRACE); err != nil {
		return nil, err
	}
	return stmts, nil
}
