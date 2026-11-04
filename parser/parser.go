package parser

import (
	"fmt"

	"go-compiler/ast"
	"go-compiler/lexer"
)

// =====================================================================
// PARSER DESCENDENTE RECURSIVO
// =====================================================================
// Gramática (EBNF completa em docs/gramatica.md):
//   Program      → "package" IDENT ";" { Import } { FuncDecl | VarDecl }
//   FuncDecl     → "func" IDENT "(" [ ParamList ] ")" [ Type ] Block
//   ParamList    → ParamGroup { "," ParamGroup }
//   ParamGroup   → IDENT { "," IDENT } Type
//   Block        → "{" { Statement } "}"
//   Statement    → VarDecl | SimpleStmt ";" | PrintStmt | IfStmt | ForStmt
//                | ReturnStmt | "break" | "continue"
//   SimpleStmt   → IDENT ":=" Expr | IDENT AssignOp Expr | IDENT ("++"|"--")
//                | CallExpr
//   Expr         → OrExpr
//   OrExpr       → AndExpr ( "||" AndExpr )*
//   AndExpr      → RelationalExpr ( "&&" RelationalExpr )*
//   RelExpr      → SimpleExpr ( ( "==" | "!=" | "<" | ">" | "<=" | ">=" ) SimpleExpr )*
//   SimpleExpr   → Term ( ( "+" | "-" ) Term )*
//   Term         → Unary ( ( "*" | "/" | "%" ) Unary )*
//   Unary        → [ "!" | "-" ] Unary | Factor
//   Factor       → LITERAL | IDENT | CallExpr | "(" Expr ")"
//   CallExpr     → IDENT "(" [ Expr { "," Expr } ] ")"
// =====================================================================

type Parser struct {
	tokens []lexer.TokenData
	pos    int
	errors []error
}

func NewParser(tokens []lexer.TokenData) *Parser {
	return &Parser{tokens: tokens}
}

func (p *Parser) Errors() []error { return p.errors }

// ─── helpers ───────────────────────────────────────────────────────

func (p *Parser) peek() lexer.Token {
	if p.pos >= len(p.tokens) {
		return lexer.EOF
	}
	return p.tokens[p.pos].Type
}

func (p *Parser) peekToken() lexer.TokenData {
	if p.pos >= len(p.tokens) {
		return lexer.TokenData{Type: lexer.EOF}
	}
	return p.tokens[p.pos]
}

func (p *Parser) peekLexeme() string {
	return p.peekToken().Lexeme
}

// lookahead retorna o tipo do token na posição pos+offset sem avançar.
func (p *Parser) lookahead(offset int) lexer.Token {
	i := p.pos + offset
	if i >= len(p.tokens) {
		return lexer.EOF
	}
	return p.tokens[i].Type
}

func (p *Parser) advance() lexer.TokenData {
	tok := p.peekToken()
	if tok.Type != lexer.EOF {
		p.pos++
	}
	return tok
}

// errorf registra um erro sintático na linha do token atual.
func (p *Parser) errorf(format string, args ...interface{}) error {
	tok := p.peekToken()
	err := fmt.Errorf("erro sintatico linha %d: %s", tok.Line, fmt.Sprintf(format, args...))
	p.errors = append(p.errors, err)
	return err
}

// match consome o token se for do tipo esperado; registra erro caso contrário.
func (p *Parser) match(expected lexer.Token) (lexer.TokenData, error) {
	if p.peek() == expected {
		return p.advance(), nil
	}
	tok := p.peekToken()
	return lexer.TokenData{}, p.errorf("esperava %s, encontrou %s (%q)", expected, tok.Type, tok.Lexeme)
}

// syncError registra erro e aplica Panic Mode: avança até ponto seguro.
func (p *Parser) syncError(msg string) {
	p.errorf("%s (encontrou %q)", msg, p.peekLexeme())
	for p.peek() != lexer.EOF &&
		p.peek() != lexer.SEMICOLON &&
		p.peek() != lexer.RBRACE {
		p.advance()
	}
	if p.peek() == lexer.SEMICOLON {
		p.advance()
	}
}

// skipSemicolons descarta ponto e vírgula soltos (inseridos automaticamente).
func (p *Parser) skipSemicolons() {
	for p.peek() == lexer.SEMICOLON {
		p.advance()
	}
}

// endStatement exige o fim de uma instrução: ';' (explícito ou inserido
// pelo lexer) ou um '}' que fecha o bloco na mesma linha.
func (p *Parser) endStatement() {
	switch p.peek() {
	case lexer.SEMICOLON:
		p.advance()
	case lexer.RBRACE, lexer.EOF: // EOF inesperado já é reportado por quem esperava '}'
	default:
		p.syncError("fim de instrucao esperado (faltou quebra de linha ou ';')")
	}
}

// ─── Program ───────────────────────────────────────────────────────

func (p *Parser) ParseProgram() *ast.ProgramNode {
	prog := &ast.ProgramNode{}

	p.skipSemicolons()

	// package <nome>
	if p.peek() == lexer.PACKAGE {
		p.advance()
		if id, err := p.match(lexer.IDENT); err == nil {
			prog.PackageName = id.Lexeme
		}
		p.match(lexer.SEMICOLON)
	}

	// import "pkg"  ou  import ( "pkg1" "pkg2" )
	for p.peek() == lexer.IMPORT {
		p.advance()
		if p.peek() == lexer.LPAREN {
			p.advance()
			p.skipSemicolons()
			for p.peek() == lexer.STRING {
				imp := p.advance()
				prog.Imports = append(prog.Imports, imp.Lexeme)
				p.match(lexer.SEMICOLON)
			}
			p.match(lexer.RPAREN)
		} else {
			if imp, err := p.match(lexer.STRING); err == nil {
				prog.Imports = append(prog.Imports, imp.Lexeme)
			}
		}
		p.match(lexer.SEMICOLON)
	}

	// declarações de nível de pacote: func e var globais
	for p.peek() != lexer.EOF {
		p.skipSemicolons()
		if p.peek() == lexer.EOF {
			break
		}

		saved := p.pos
		switch p.peek() {
		case lexer.FUNC:
			if f := p.parseFunc(); f != nil {
				prog.Functions = append(prog.Functions, f)
			}
		case lexer.VAR:
			if v := p.parseVarDecl(); v != nil {
				prog.Globals = append(prog.Globals, v)
			}
		case lexer.TYPE, lexer.STRUCT, lexer.CONST, lexer.INTERFACE:
			p.syncError(fmt.Sprintf("'%s' nao suportado nesta versao do compilador", p.peekLexeme()))
		default:
			p.syncError("esperava 'func' ou 'var' no nivel do pacote")
		}
		// salvaguarda anti-loop infinito
		if p.pos == saved {
			p.advance()
		}
	}

	return prog
}

// ─── FuncDecl ──────────────────────────────────────────────────────

func (p *Parser) parseFunc() ast.ASTNode {
	p.advance() // consome 'func'

	id, err := p.match(lexer.IDENT)
	if err != nil {
		p.syncError("nome de funcao esperado")
		return nil
	}
	fn := &ast.FuncNode{Name: id.Lexeme, Line: id.Line}

	if _, err := p.match(lexer.LPAREN); err != nil {
		p.syncError("'(' esperado apos o nome da funcao")
		return nil
	}
	fn.Params = p.parseParams()
	if _, err := p.match(lexer.RPAREN); err != nil {
		p.syncError("')' esperado ao fim dos parametros")
		return nil
	}

	// tipo de retorno opcional (um único tipo nomeado)
	if p.peek() == lexer.IDENT {
		fn.ReturnType = p.advance().Lexeme
	}

	if _, err := p.match(lexer.LBRACE); err != nil {
		p.syncError("'{' esperado no inicio do corpo da funcao")
		return nil
	}
	fn.Body = p.parseBlock()
	p.match(lexer.RBRACE)
	p.endStatement()

	return fn
}

// parseParams reconhece as duas formas do Go: (a int, b int) e (a, b int).
func (p *Parser) parseParams() []ast.Param {
	var params []ast.Param
	var pending []lexer.TokenData // nomes que ainda aguardam o tipo

	for p.peek() == lexer.IDENT {
		name := p.advance()
		pending = append(pending, name)

		if p.peek() == lexer.COMMA {
			p.advance()
			continue
		}
		if p.peek() == lexer.RPAREN {
			break // reportado abaixo: parâmetro sem tipo
		}
		typ, err := p.match(lexer.IDENT)
		if err != nil {
			return params
		}
		for _, n := range pending {
			params = append(params, ast.Param{Name: n.Lexeme, TypeName: typ.Lexeme, Line: n.Line})
		}
		pending = nil
		if p.peek() == lexer.COMMA {
			p.advance()
		}
	}
	if len(pending) > 0 {
		p.errorf("parametro '%s' sem tipo", pending[len(pending)-1].Lexeme)
	}
	return params
}

// ─── Block ─────────────────────────────────────────────────────────

func (p *Parser) parseBlock() []ast.ASTNode {
	var stmts []ast.ASTNode
	for p.peek() != lexer.RBRACE && p.peek() != lexer.EOF {
		saved := p.pos
		if s := p.parseStatement(); s != nil {
			stmts = append(stmts, s)
		}
		if p.pos == saved {
			p.advance() // salvaguarda anti-loop
		}
	}
	return stmts
}

// ─── Statement ─────────────────────────────────────────────────────

func (p *Parser) parseStatement() ast.ASTNode {
	p.skipSemicolons()

	if p.peek() == lexer.RBRACE || p.peek() == lexer.EOF {
		return nil
	}

	switch p.peek() {

	case lexer.VAR:
		return p.parseVarDecl()

	case lexer.IDENT:
		// fmt.Println(...) / fmt.Print(...) / fmt.Printf(...)
		if p.peekLexeme() == "fmt" && p.lookahead(1) == lexer.PERIOD {
			return p.parsePrintStmt()
		}
		if stmt := p.parseSimpleStmt(); stmt != nil {
			if _, isCall := stmt.(*ast.CallNode); !isCall && !isAssignLike(stmt) {
				p.errorf("expressao avaliada mas nao utilizada")
			}
			p.endStatement()
			return stmt
		}
		p.syncError("instrucao invalida com identificador")
		return nil

	case lexer.IF:
		return p.parseIf()

	case lexer.FOR:
		return p.parseFor()

	case lexer.RETURN:
		return p.parseReturn()

	case lexer.BREAK:
		tok := p.advance()
		p.endStatement()
		return &ast.BreakNode{Line: tok.Line}

	case lexer.CONTINUE:
		tok := p.advance()
		p.endStatement()
		return &ast.ContinueNode{Line: tok.Line}

	case lexer.TYPE, lexer.STRUCT, lexer.MAP, lexer.CONST,
		lexer.SELECT, lexer.SWITCH, lexer.DEFER, lexer.RANGE,
		lexer.INTERFACE, lexer.GO:
		p.syncError(fmt.Sprintf("'%s' nao suportado nesta versao do compilador", p.peekLexeme()))
		return nil

	default:
		p.syncError("instrucao invalida")
		return nil
	}
}

func isAssignLike(n ast.ASTNode) bool {
	switch n.(type) {
	case *ast.AssignNode, *ast.VarDeclNode:
		return true
	}
	return false
}

// ─── VarDecl ───────────────────────────────────────────────────────

func (p *Parser) parseVarDecl() ast.ASTNode {
	p.advance() // consome 'var'

	id, err := p.match(lexer.IDENT)
	if err != nil {
		p.syncError("nome de variavel esperado apos 'var'")
		return nil
	}
	decl := &ast.VarDeclNode{Name: id.Lexeme, Line: id.Line}

	if p.peek() == lexer.IDENT {
		decl.TypeName = p.advance().Lexeme
	}

	if p.peek() == lexer.ASSIGN {
		p.advance()
		if p.peek() == lexer.SEMICOLON || p.peek() == lexer.RBRACE {
			p.syncError("expressao esperada apos '='")
			return nil
		}
		decl.Initializer = p.parseExpr()
	}

	if decl.TypeName == "" && decl.Initializer == nil {
		p.syncError(fmt.Sprintf("declaracao de '%s' precisa de tipo ou valor inicial", decl.Name))
		return nil
	}

	p.endStatement()
	return decl
}

// ─── SimpleStmt ────────────────────────────────────────────────────
// Usado em instruções comuns e nas cláusulas init/post de if e for.
// Não consome o ';' final.

// compoundOps mapeia operadores de atribuição composta para o operador binário.
var compoundOps = map[lexer.Token]string{
	lexer.ADD_ASSIGN: "+",
	lexer.SUB_ASSIGN: "-",
	lexer.MUL_ASSIGN: "*",
	lexer.QUO_ASSIGN: "/",
	lexer.REM_ASSIGN: "%",
}

func (p *Parser) parseSimpleStmt() ast.ASTNode {
	if p.peek() == lexer.IDENT {
		switch next := p.lookahead(1); {
		case next == lexer.DEFINE: // x := expr
			id := p.advance()
			p.advance()
			expr := p.parseExpr()
			return &ast.VarDeclNode{Name: id.Lexeme, Initializer: expr, Line: id.Line}

		case next == lexer.ASSIGN: // x = expr
			id := p.advance()
			p.advance()
			expr := p.parseExpr()
			return &ast.AssignNode{Name: id.Lexeme, Expr: expr, Line: id.Line}

		case compoundOps[next] != "": // x += expr  →  x = x + expr
			id := p.advance()
			op := compoundOps[p.advance().Type]
			expr := p.parseExpr()
			return &ast.AssignNode{
				Name: id.Lexeme,
				Expr: &ast.BinaryOpNode{OpStr: op,
					Left:  &ast.VariableNode{Name: id.Lexeme, Line: id.Line},
					Right: expr, Line: id.Line},
				Line: id.Line,
			}

		case next == lexer.INC || next == lexer.DEC: // x++  →  x = x + 1
			id := p.advance()
			op := p.advance()
			opStr := "+"
			if op.Type == lexer.DEC {
				opStr = "-"
			}
			return &ast.AssignNode{
				Name: id.Lexeme,
				Expr: &ast.BinaryOpNode{OpStr: opStr,
					Left:  &ast.VariableNode{Name: id.Lexeme, Line: id.Line},
					Right: &ast.LiteralNode{Value: "1", Line: id.Line},
					Line:  id.Line},
				Line: id.Line,
			}
		}
	}
	return p.parseExpr()
}

// ─── PrintStmt ─────────────────────────────────────────────────────

func (p *Parser) parsePrintStmt() ast.ASTNode {
	start := p.advance()  // consome 'fmt'
	p.match(lexer.PERIOD) // consome '.'
	fn, err := p.match(lexer.IDENT)
	if err != nil {
		p.syncError("nome de funcao do pacote fmt esperado")
		return nil
	}
	switch fn.Lexeme {
	case "Println", "Print", "Printf":
	default:
		p.syncError(fmt.Sprintf("fmt.%s nao suportado (use Println, Print ou Printf)", fn.Lexeme))
		return nil
	}
	if _, err := p.match(lexer.LPAREN); err != nil {
		p.syncError("'(' esperado")
		return nil
	}
	args := p.parseArgs()
	if _, err := p.match(lexer.RPAREN); err != nil {
		p.syncError("')' esperado")
		return nil
	}
	p.endStatement()
	return &ast.PrintCallNode{Func: fn.Lexeme, Args: args, Line: start.Line}
}

// parseArgs lê uma lista de expressões separadas por vírgula (pode ser vazia).
func (p *Parser) parseArgs() []ast.ASTNode {
	var args []ast.ASTNode
	if p.peek() == lexer.RPAREN {
		return args
	}
	args = append(args, p.parseExpr())
	for p.peek() == lexer.COMMA {
		p.advance()
		args = append(args, p.parseExpr())
	}
	return args
}

// ─── IfStmt ────────────────────────────────────────────────────────

func (p *Parser) parseIf() ast.ASTNode {
	ifTok := p.advance() // consome 'if'

	var init ast.ASTNode
	var cond ast.ASTNode

	// Tenta ler a primeira parte (pode ser a inicialização ou já a condição)
	firstPt := p.parseSimpleStmt()

	if p.peek() == lexer.SEMICOLON {
		p.advance() // consome o ';'
		init = firstPt
		cond = p.parseExpr()
	} else {
		// Se não houver ';', a condição é lida diretamente após o 'if'
		cond = firstPt
		if isAssignLike(cond) {
			p.errorf("atribuicao usada como condicao do if")
		}
	}

	if _, err := p.match(lexer.LBRACE); err != nil {
		p.syncError("'{' esperado apos a condicao do if")
		return nil
	}
	thenBranch := p.parseBlock()
	p.match(lexer.RBRACE)

	var elseBranch []ast.ASTNode
	if p.peek() == lexer.ELSE {
		p.advance() // consome 'else'
		if p.peek() == lexer.IF {
			// else if
			if nested := p.parseIf(); nested != nil {
				elseBranch = append(elseBranch, nested)
			}
			return &ast.IfNode{Init: init, Condition: cond, ThenBranch: thenBranch, ElseBranch: elseBranch, Line: ifTok.Line}
		}
		if _, err := p.match(lexer.LBRACE); err != nil {
			p.syncError("'{' esperado apos 'else'")
			return nil
		}
		elseBranch = p.parseBlock()
		p.match(lexer.RBRACE)
	}
	p.endStatement()

	return &ast.IfNode{Init: init, Condition: cond, ThenBranch: thenBranch, ElseBranch: elseBranch, Line: ifTok.Line}
}

// ─── ForStmt ───────────────────────────────────────────────────────

func (p *Parser) parseFor() ast.ASTNode {
	forTok := p.advance() // consome 'for'
	node := &ast.ForNode{Line: forTok.Line}

	if p.peek() != lexer.LBRACE {
		// Diferencia "for cond {" de "for init; cond; post {" procurando
		// um ';' antes do '{' do corpo.
		isClassic := false
		for i := p.pos; i < len(p.tokens); i++ {
			tt := p.tokens[i].Type
			if tt == lexer.LBRACE || tt == lexer.EOF {
				break
			}
			if tt == lexer.SEMICOLON {
				isClassic = true
				break
			}
		}

		if isClassic {
			if p.peek() != lexer.SEMICOLON {
				node.Init = p.parseSimpleStmt()
			}
			p.match(lexer.SEMICOLON)
			if p.peek() != lexer.SEMICOLON {
				node.Condition = p.parseExpr()
			}
			p.match(lexer.SEMICOLON)
			if p.peek() != lexer.LBRACE {
				node.Post = p.parseSimpleStmt()
			}
		} else {
			node.Condition = p.parseExpr()
		}
	}

	if _, err := p.match(lexer.LBRACE); err != nil {
		p.syncError("'{' esperado no corpo do for")
		return nil
	}
	node.Body = p.parseBlock()
	p.match(lexer.RBRACE)
	p.endStatement()

	return node
}

// ─── ReturnStmt ────────────────────────────────────────────────────

func (p *Parser) parseReturn() ast.ASTNode {
	tok := p.advance()

	if p.peek() == lexer.SEMICOLON || p.peek() == lexer.RBRACE || p.peek() == lexer.EOF {
		p.endStatement()
		return &ast.ReturnNode{Line: tok.Line}
	}

	val := p.parseExpr()
	p.endStatement()
	return &ast.ReturnNode{Value: val, Line: tok.Line}
}

// =====================================================================
// CASCATA DE PRECEDÊNCIA (LOGICAL -> RELATIONAL -> MATH)
// =====================================================================

// Camada 1: Operadores Lógicos. Como em Go, && tem precedência maior
// que ||: a || b && c  ≡  a || (b && c).
func (p *Parser) parseExpr() ast.ASTNode {
	return p.parseOrExpr()
}

func (p *Parser) parseOrExpr() ast.ASTNode {
	left := p.parseAndExpr()
	for p.peek() == lexer.LOR {
		op := p.advance()
		right := p.parseAndExpr()
		left = &ast.BinaryOpNode{OpStr: op.Lexeme, Left: left, Right: right, Line: op.Line}
	}
	return left
}

func (p *Parser) parseAndExpr() ast.ASTNode {
	left := p.parseRelationalExpr()
	for p.peek() == lexer.LAND {
		op := p.advance()
		right := p.parseRelationalExpr()
		left = &ast.BinaryOpNode{OpStr: op.Lexeme, Left: left, Right: right, Line: op.Line}
	}
	return left
}

// Camada 2: Operadores Relacionais (==, !=, <, >, <=, >=)
// Associam à esquerda, como em Go: 1 < 2 == true  ≡  (1 < 2) == true.
func (p *Parser) parseRelationalExpr() ast.ASTNode {
	left := p.parseSimpleExpr()
	for isRelational(p.peek()) {
		op := p.advance()
		right := p.parseSimpleExpr()
		left = &ast.BinaryOpNode{OpStr: op.Lexeme, Left: left, Right: right, Line: op.Line}
	}
	return left
}

func isRelational(t lexer.Token) bool {
	switch t {
	case lexer.EQL, lexer.NEQ, lexer.LSS, lexer.GTR, lexer.LEQ, lexer.GEQ:
		return true
	}
	return false
}

// Camada 3: Soma e Subtração
func (p *Parser) parseSimpleExpr() ast.ASTNode {
	left := p.parseTerm()
	for p.peek() == lexer.ADD || p.peek() == lexer.SUB {
		op := p.advance()
		right := p.parseTerm()
		left = &ast.BinaryOpNode{OpStr: op.Lexeme, Left: left, Right: right, Line: op.Line}
	}
	return left
}

// Camada 4: Multiplicação, Divisão e Resto
func (p *Parser) parseTerm() ast.ASTNode {
	left := p.parseUnary()
	for p.peek() == lexer.MUL || p.peek() == lexer.QUO || p.peek() == lexer.REM {
		op := p.advance()
		right := p.parseUnary()
		left = &ast.BinaryOpNode{OpStr: op.Lexeme, Left: left, Right: right, Line: op.Line}
	}
	return left
}

// Camada 5: Unários (Negação Lógica e Numérica)
func (p *Parser) parseUnary() ast.ASTNode {
	switch p.peek() {
	case lexer.NOT, lexer.SUB:
		op := p.advance()
		return &ast.UnaryOpNode{OpStr: op.Lexeme, Operand: p.parseUnary(), Line: op.Line}
	}
	return p.parseFactor()
}

// Camada 6: Literais, Variáveis, Chamadas e Parênteses
func (p *Parser) parseFactor() ast.ASTNode {
	tok := p.peekToken()

	switch p.peek() {
	case lexer.INT, lexer.FLOAT, lexer.STRING, lexer.IMAG, lexer.CHAR:
		p.advance()
		return &ast.LiteralNode{Value: tok.Lexeme, Line: tok.Line}

	case lexer.IDENT:
		p.advance()
		if tok.Lexeme == "true" || tok.Lexeme == "false" {
			return &ast.LiteralNode{Value: tok.Lexeme, Line: tok.Line}
		}
		if p.peek() == lexer.LPAREN {
			p.advance()
			args := p.parseArgs()
			p.match(lexer.RPAREN)
			return &ast.CallNode{Name: tok.Lexeme, Args: args, Line: tok.Line}
		}
		return &ast.VariableNode{Name: tok.Lexeme, Line: tok.Line}

	case lexer.LPAREN:
		p.advance()
		expr := p.parseExpr()
		p.match(lexer.RPAREN)
		return expr

	default:
		p.syncError(fmt.Sprintf("fator invalido: %q", tok.Lexeme))
		return &ast.LiteralNode{Value: "0", Line: tok.Line}
	}
}
