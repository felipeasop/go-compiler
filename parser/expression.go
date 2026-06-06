package parser

import (
	"fmt"
	"strconv"

	"modulariza/lexer"
)

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
	if op == lexer.T_EQ || op == lexer.T_LT || op == lexer.T_GT {
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
	for p.peek() == lexer.T_PLUS || p.peek() == lexer.T_MINUS {
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
	for p.peek() == lexer.T_MULT || p.peek() == lexer.T_DIV {
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
	case lexer.T_NUM:
		t := p.advance()
		val, err := strconv.Atoi(t.Lexeme)
		if err != nil {
			return nil, fmt.Errorf("erro sintático na linha %d: número inválido '%s'", t.Line, t.Lexeme)
		}
		return &NumberNode{Value: val}, nil

	case lexer.T_ID:
		t := p.advance()
		return &VariableNode{Name: t.Lexeme}, nil

	case lexer.T_LPAREN:
		p.advance() // consome "("
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if err := p.match(lexer.T_RPAREN); err != nil {
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
