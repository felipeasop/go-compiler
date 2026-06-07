package parser

import (
	"fmt"
	"strconv"

	"go-compiler/lexer"
)

// =====================================================================
// PARSER DESCENDENTE RECURSIVO
// =====================================================================
// Recebe a lista de tokens do Scanner e constrói a AST.
// Cada método corresponde a uma regra da gramática.
//
// Gramática base (linguagem Go compilada):
//   Program    → Statement*
//   Statement  → VarDecl | Assignment | PrintStmt | IfStmt | ForStmt
//   VarDecl    → "var" ID Type [ "=" Expr ] ";"
//              | ID ":=" Expr ";"
//   Assignment → ID "=" Expr ";"
//   PrintStmt  → "fmt" "." "Println" "(" Expr ")" ";"
//   IfStmt     → "if" Expr "{" Statement* "}" [ "else" ( IfStmt | "{" Statement* "}" ) ]
//   ForStmt    → "for" Expr "{" Statement* "}"
//   Expr       → SimpleExpr [ ( "==" | "<" | ">" | "<=" | ">=" ) SimpleExpr ]
//   SimpleExpr → Term ( ( "+" | "-" ) Term )*
//   Term       → Factor ( ( "*" | "/" ) Factor )*
//   Factor     → NUMBER | FLOAT_NUM | STRING_LITERAL | ID | "(" Expr ")"
// =====================================================================

type Parser struct {
	tokens []lexer.Token
	pos    int
	errors []error
}

func NewParser(tokens []lexer.Token) *Parser {
	return &Parser{tokens: tokens}
}

func (p *Parser) Errors() []error { return p.errors }

// ─── helpers ───────────────────────────────────────────────────────

// peek retorna o tipo do token atual sem consumi-lo (lookahead).
func (p *Parser) peek() lexer.TokenType {
	if p.pos >= len(p.tokens) {
		return lexer.T_EOF
	}
	return p.tokens[p.pos].Type
}

// peekToken retorna o token atual completo.
func (p *Parser) peekToken() lexer.Token {
	if p.pos >= len(p.tokens) {
		return lexer.Token{Type: lexer.T_EOF}
	}
	return p.tokens[p.pos]
}

// advance consome o token atual e retorna ele.
func (p *Parser) advance() lexer.Token {
	tok := p.peekToken()
	if tok.Type != lexer.T_EOF {
		p.pos++
	}
	return tok
}

// match valida que o token atual é do tipo esperado e o consome.
// Se não for, registra o erro e retorna token vazio (Panic Mode).
func (p *Parser) match(expected lexer.TokenType) (lexer.Token, error) {
	if p.peek() == expected {
		return p.advance(), nil
	}
	err := fmt.Errorf("erro sintatico linha %d: esperava %s, encontrou %s (%q)",
		p.peekToken().Line, expected, p.peek(), p.peekToken().Lexeme)
	p.errors = append(p.errors, err)
	return lexer.Token{}, err
}

// sync avança tokens até um ponto seguro após um erro (Panic Mode).
func (p *Parser) sync(msg string) {
	tok := p.peekToken()
	err := fmt.Errorf("erro sintatico linha %d: %s (encontrou %q)", tok.Line, msg, tok.Lexeme)
	p.errors = append(p.errors, err)
	for p.peek() != lexer.T_EOF && p.peek() != lexer.T_SEMICOLON && p.peek() != lexer.T_RBRACE {
		p.advance()
	}
	if p.peek() == lexer.T_SEMICOLON {
		p.advance()
	}
}

// ─── Program ───────────────────────────────────────────────────────

func (p *Parser) ParseProgram() *ProgramNode {
	prog := &ProgramNode{}
	for p.peek() != lexer.T_EOF {
		if s := p.parseStatement(); s != nil {
			prog.Statements = append(prog.Statements, s)
		}
	}
	return prog
}

// ─── Statement ─────────────────────────────────────────────────────

func (p *Parser) parseStatement() ASTNode {
	switch p.peek() {

	case lexer.T_VAR:
		// var x int = expr;   ou   var x int;
		return p.parseVarDecl()

	case lexer.T_ID:
		// Distingue  x := expr   de   x = expr
		// e também   fmt.Println(...)
		if p.pos+1 < len(p.tokens) && p.tokens[p.pos+1].Type == lexer.T_DECLARE_ASSIGN {
			return p.parseShortVarDecl()
		}
		// fmt.Println(expr) — "fmt" é lexer.T_ID
		if p.peekToken().Lexeme == "fmt" {
			return p.parsePrintStmt()
		}
		return p.parseAssignment()

	case lexer.T_IF:
		return p.parseIfStmt()

	case lexer.T_FOR:
		return p.parseForStmt()

	default:
		p.sync("instrucao invalida")
		return nil
	}
}

// ─── VarDecl ───────────────────────────────────────────────────────
// var x int = expr;   ou   var x int;

func (p *Parser) parseVarDecl() ASTNode {
	p.advance() // consome "var"

	idTok, err := p.match(lexer.T_ID)
	if err != nil {
		return nil
	}

	// tipo: int | float | bool | string
	if p.peek() != lexer.T_INT && p.peek() != lexer.T_FLOAT && p.peek() != lexer.T_BOOL && p.peek() != lexer.T_STRING {
		p.sync("tipo esperado apos nome da variavel")
		return nil
	}
	p.advance() // consome o tipo

	var init ASTNode
	if p.peek() == lexer.T_ASSIGN {
		p.advance() // consome "="
		init = p.parseExpr()
	}

	p.match(lexer.T_SEMICOLON)
	return &VarDeclNode{Name: idTok.Lexeme, Initializer: init}
}

// ─── ShortVarDecl ──────────────────────────────────────────────────
// x := expr;

func (p *Parser) parseShortVarDecl() ASTNode {
	idTok := p.advance() // consome ID
	p.advance()          // consome :=
	expr := p.parseExpr()
	p.match(lexer.T_SEMICOLON)
	return &VarDeclNode{Name: idTok.Lexeme, Initializer: expr}
}

// ─── Assignment ────────────────────────────────────────────────────
// x = expr;

func (p *Parser) parseAssignment() ASTNode {
	idTok := p.advance() // consome ID
	if _, err := p.match(lexer.T_ASSIGN); err != nil {
		return nil
	}
	expr := p.parseExpr()
	p.match(lexer.T_SEMICOLON)
	return &AssignNode{Name: idTok.Lexeme, Expr: expr}
}

// ─── PrintStmt ─────────────────────────────────────────────────────
// fmt.Println(expr);
func (p *Parser) parsePrintStmt() ASTNode {
	p.match(lexer.T_ID)  // Consome "fmt"
	p.match(lexer.T_DOT) // Consome "." (Agora tratado corretamente!)
	p.match(lexer.T_ID)  // Consome "Println"

	if _, err := p.match(lexer.T_LPAREN); err != nil {
		return nil
	}
	expr := p.parseExpr()
	if _, err := p.match(lexer.T_RPAREN); err != nil {
		return nil
	}
	p.match(lexer.T_SEMICOLON)

	return &PrintCallNode{Expr: expr}
}

// ─── IfStmt ────────────────────────────────────────────────────────
// if (cond) { ... }  else if (cond) { ... }  else { ... }
// Em Go a condição não usa parênteses, mas o professor usa — aceitamos os dois.

func (p *Parser) parseIfStmt() ASTNode {
	p.advance() // consome "if"

	// parênteses opcionais ao redor da condição
	hasParen := p.peek() == lexer.T_LPAREN
	if hasParen {
		p.advance()
	}
	cond := p.parseExpr()
	if hasParen {
		p.match(lexer.T_RPAREN)
	}

	if _, err := p.match(lexer.T_LBRACE); err != nil {
		return nil
	}
	thenBranch := p.parseBlock()
	p.match(lexer.T_RBRACE)

	node := &IfNode{Condition: cond, ThenBranch: thenBranch}

	for p.peek() == lexer.T_ELSE {
		p.advance() // consome "else"
		if p.peek() == lexer.T_IF {
			// else if
			p.advance() // consome "if"
			hasParen2 := p.peek() == lexer.T_LPAREN
			if hasParen2 {
				p.advance()
			}
			eic := p.parseExpr()
			if hasParen2 {
				p.match(lexer.T_RPAREN)
			}
			p.match(lexer.T_LBRACE)
			eib := p.parseBlock()
			p.match(lexer.T_RBRACE)
			node.ElseIfBlocks = append(node.ElseIfBlocks, ElseIfBlock{Condition: eic, Body: eib})
		} else {
			// else simples
			p.match(lexer.T_LBRACE)
			node.ElseBranch = p.parseBlock()
			p.match(lexer.T_RBRACE)
			break
		}
	}

	return node
}

// ─── ForStmt ───────────────────────────────────────────────────────
// Go usa "for" onde outras linguagens usam "while".
// Forma: for (cond) { ... }   — parênteses opcionais (professor usa).

func (p *Parser) parseForStmt() ASTNode {
	p.advance() // consome "for"

	hasParen := p.peek() == lexer.T_LPAREN
	if hasParen {
		p.advance()
	}
	cond := p.parseExpr()
	if hasParen {
		p.match(lexer.T_RPAREN)
	}

	p.match(lexer.T_LBRACE)
	body := p.parseBlock()
	p.match(lexer.T_RBRACE)

	return &ForNode{Condition: cond, Body: body}
}

// ─── Block ─────────────────────────────────────────────────────────

func (p *Parser) parseBlock() []ASTNode {
	var stmts []ASTNode
	for p.peek() != lexer.T_RBRACE && p.peek() != lexer.T_EOF {
		if s := p.parseStatement(); s != nil {
			stmts = append(stmts, s)
		}
	}
	return stmts
}

// =====================================================================
// CASCATA DE PRECEDÊNCIA
// =====================================================================
// Camada 1: parseExpr       → operadores relacionais  (menor precedência)
// Camada 2: parseSimpleExpr → + e -
// Camada 3: parseTerm       → * e /
// Camada 4: parseFactor     → valores literais e variáveis (maior precedência)
// =====================================================================

func (p *Parser) parseExpr() ASTNode {
	left := p.parseSimpleExpr()
	switch p.peek() {
	case lexer.T_EQ, lexer.T_LT, lexer.T_GT, lexer.T_LE, lexer.T_GE:
		op := p.peek()
		p.advance()
		right := p.parseSimpleExpr()
		return &BinaryOpNode{Op: op, Left: left, Right: right}
	}
	return left
}

func (p *Parser) parseSimpleExpr() ASTNode {
	left := p.parseTerm()
	for p.peek() == lexer.T_PLUS || p.peek() == lexer.T_MINUS {
		op := p.peek()
		p.advance()
		right := p.parseTerm()
		left = &BinaryOpNode{Op: op, Left: left, Right: right}
	}
	return left
}

func (p *Parser) parseTerm() ASTNode {
	left := p.parseFactor()
	for p.peek() == lexer.T_MULT || p.peek() == lexer.T_DIV {
		op := p.peek()
		p.advance()
		right := p.parseFactor()
		left = &BinaryOpNode{Op: op, Left: left, Right: right}
	}
	return left
}

func (p *Parser) parseFactor() ASTNode {
	tok := p.peekToken()
	switch p.peek() {
	case lexer.T_NUM:
		p.advance()
		v, _ := strconv.Atoi(tok.Lexeme)
		return &NumberNode{Value: v}
	case lexer.T_FLOAT_NUM:
		p.advance()
		v, _ := strconv.ParseFloat(tok.Lexeme, 64)
		return &FloatNode{Value: v}
	case lexer.T_STRING_LITERAL:
		p.advance()
		return &StringNode{Value: tok.Lexeme}
	case lexer.T_TRUE:
		p.advance()
		return &VariableNode{Name: "true"}
	case lexer.T_FALSE:
		p.advance()
		return &VariableNode{Name: "false"}
	case lexer.T_ID:
		p.advance()
		return &VariableNode{Name: tok.Lexeme}
	case lexer.T_LPAREN:
		p.advance()
		expr := p.parseExpr()
		p.match(lexer.T_RPAREN)
		return expr
	default:
		p.sync(fmt.Sprintf("fator invalido: %q", tok.Lexeme))
		return &NumberNode{Value: 0}
	}
}
