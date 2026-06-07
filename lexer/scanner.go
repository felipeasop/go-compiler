package lexer

import (
	"fmt"
	"unicode"
)

type Scanner struct {
	source   []rune
	pos      int
	line     int
	lastType Token
}

func NewScanner(source string) *Scanner {
	return &Scanner{source: []rune(source), pos: 0, line: 1, lastType: EOF}
}

func (s *Scanner) peek() rune {
	if s.pos >= len(s.source) {
		return 0
	}
	return s.source[s.pos]
}

// peekAt retorna o caractere na posição pos+offset sem avançar.
func (s *Scanner) peekAt(offset int) rune {
	i := s.pos + offset
	if i >= len(s.source) {
		return 0
	}
	return s.source[i]
}

func (s *Scanner) next() rune {
	c := s.peek()
	if c != 0 {
		s.pos++
	}
	return c
}

// skipWhitespace ignora espaços, tabs e \r — mas NÃO \n.
// \n é tratado em NextToken para inserção automática de ';'.
func (s *Scanner) skipWhitespace() {
	for s.peek() == ' ' || s.peek() == '\t' || s.peek() == '\r' {
		s.next()
	}
}

func (s *Scanner) skipLineComment() {
	for s.peek() != '\n' && s.peek() != 0 {
		s.next()
	}
}

func (s *Scanner) skipBlockComment() error {
	for s.peek() != 0 {
		c := s.next()
		if c == '\n' {
			s.line++
		}
		if c == '*' && s.peek() == '/' {
			s.next()
			return nil
		}
	}
	return fmt.Errorf("comentario de bloco nao fechado na linha %d", s.line)
}

// tokenRequiresSemicolon implementa a regra de inserção automática de ';' do Go.
// Ref: https://go.dev/ref/spec#Semicolons
func tokenRequiresSemicolon(t Token) bool {
	switch t {
	case IDENT, INT, FLOAT, IMAG, CHAR, STRING,
		BREAK, CONTINUE, RETURN,
		INC, DEC, RPAREN, RBRACK, RBRACE:
		return true
	}
	return false
}

func (s *Scanner) NextToken() (TokenData, error) {
	s.skipWhitespace()

	// Inserção automática de ';' ao encontrar \n
	if s.peek() == '\n' {
		s.next()
		s.line++
		if tokenRequiresSemicolon(s.lastType) {
			s.lastType = SEMICOLON
			return TokenData{SEMICOLON, ";", s.line - 1}, nil
		}
		return s.NextToken()
	}

	// EOF — emite ';' pendente se necessário
	if s.pos >= len(s.source) {
		if tokenRequiresSemicolon(s.lastType) {
			s.lastType = SEMICOLON
			return TokenData{SEMICOLON, ";", s.line}, nil
		}
		return TokenData{EOF, "EOF", s.line}, nil
	}

	c := s.next()
	startLine := s.line

	// ── Números ──────────────────────────────────────────────────────
	if unicode.IsDigit(c) {
		buf := []rune{c}
		isFloat := false
		isImag := false
		for unicode.IsDigit(s.peek()) || s.peek() == '.' || s.peek() == 'i' {
			if s.peek() == '.' {
				isFloat = true
			}
			if s.peek() == 'i' {
				isImag = true
				buf = append(buf, s.next())
				break
			}
			buf = append(buf, s.next())
		}
		if isImag {
			return s.emit(IMAG, string(buf), startLine)
		}
		if isFloat {
			return s.emit(FLOAT, string(buf), startLine)
		}
		return s.emit(INT, string(buf), startLine)
	}

	// ── Identificadores e palavras reservadas ─────────────────────────
	if unicode.IsLetter(c) || c == '_' {
		buf := []rune{c}
		for unicode.IsLetter(s.peek()) || unicode.IsDigit(s.peek()) || s.peek() == '_' {
			buf = append(buf, s.next())
		}
		lex := string(buf)
		return s.emit(Lookup(lex), lex, startLine)
	}

	switch c {

	// ── Strings ───────────────────────────────────────────────────────
	case '"':
		buf := []rune{c}
		for s.peek() != '"' && s.peek() != 0 {
			if s.peek() == '\\' {
				buf = append(buf, s.next()) // consome '\'
				escaped := s.peek()
				switch escaped {
				case 'n', 't', 'r', '\\', '"', '\'':
					buf = append(buf, s.next())
				default:
					return TokenData{}, fmt.Errorf("escape invalido '\\%c' linha %d", escaped, startLine)
				}
				continue
			}
			if s.peek() == '\n' {
				return TokenData{}, fmt.Errorf("string nao fechada linha %d", startLine)
			}
			buf = append(buf, s.next())
		}
		if s.peek() == 0 {
			return TokenData{}, fmt.Errorf("string nao fechada linha %d", startLine)
		}
		buf = append(buf, s.next()) // consome '"' de fechamento
		return s.emit(STRING, string(buf), startLine)

	// ── Raw string literal (backtick) ───────────────────────────────
	// Em Go, `...` é uma raw string: sem escapes, \n é literal.
	case '`':
		buf := []rune{c}
		for s.peek() != '`' && s.peek() != 0 {
			if s.peek() == '\n' {
				s.line++
			}
			buf = append(buf, s.next())
		}
		if s.peek() == 0 {
			return TokenData{}, fmt.Errorf("raw string nao fechada linha %d", startLine)
		}
		buf = append(buf, s.next()) // consome '`' de fechamento
		return s.emit(STRING, string(buf), startLine)

	// ── Char literal ─────────────────────────────────────────────────
	case '\'':
		buf := []rune{c}
		for s.peek() != '\'' && s.peek() != 0 {
			if s.peek() == '\\' {
				buf = append(buf, s.next())
			}
			buf = append(buf, s.next())
		}
		if s.peek() == 0 {
			return TokenData{}, fmt.Errorf("char literal nao fechado linha %d", startLine)
		}
		buf = append(buf, s.next())
		return s.emit(CHAR, string(buf), startLine)

	// ── Operadores aritméticos ────────────────────────────────────────
	case '+':
		if s.peek() == '+' {
			s.next()
			return s.emit(INC, "++", startLine)
		}
		if s.peek() == '=' {
			s.next()
			return s.emit(ADD_ASSIGN, "+=", startLine)
		}
		return s.emit(ADD, "+", startLine)

	case '-':
		if s.peek() == '-' {
			s.next()
			return s.emit(DEC, "--", startLine)
		}
		if s.peek() == '=' {
			s.next()
			return s.emit(SUB_ASSIGN, "-=", startLine)
		}
		return s.emit(SUB, "-", startLine)

	case '*':
		if s.peek() == '=' {
			s.next()
			return s.emit(MUL_ASSIGN, "*=", startLine)
		}
		return s.emit(MUL, "*", startLine)

	case '/':
		if s.peek() == '/' {
			s.next()
			s.skipLineComment()
			return s.NextToken()
		}
		if s.peek() == '*' {
			s.next()
			if err := s.skipBlockComment(); err != nil {
				return TokenData{}, err
			}
			return s.NextToken()
		}
		if s.peek() == '=' {
			s.next()
			return s.emit(QUO_ASSIGN, "/=", startLine)
		}
		return s.emit(QUO, "/", startLine)

	case '%':
		if s.peek() == '=' {
			s.next()
			return s.emit(REM_ASSIGN, "%=", startLine)
		}
		return s.emit(REM, "%", startLine)

	// ── Operadores bit a bit ──────────────────────────────────────────
	case '&':
		if s.peek() == '&' {
			s.next()
			return s.emit(LAND, "&&", startLine)
		}
		if s.peek() == '=' {
			s.next()
			return s.emit(AND_ASSIGN, "&=", startLine)
		}
		return s.emit(AND, "&", startLine)

	case '|':
		if s.peek() == '|' {
			s.next()
			return s.emit(LOR, "||", startLine)
		}
		if s.peek() == '=' {
			s.next()
			return s.emit(OR_ASSIGN, "|=", startLine)
		}
		return s.emit(OR, "|", startLine)

	case '^':
		if s.peek() == '=' {
			s.next()
			return s.emit(XOR_ASSIGN, "^=", startLine)
		}
		return s.emit(XOR, "^", startLine)

	case '~':
		return s.emit(TILDE, "~", startLine)

	// ── Shifts ────────────────────────────────────────────────────────
	case '<':
		if s.peek() == '<' {
			s.next()
			if s.peek() == '=' {
				s.next()
				return s.emit(SHL_ASSIGN, "<<=", startLine)
			}
			return s.emit(SHL, "<<", startLine)
		}
		if s.peek() == '=' {
			s.next()
			return s.emit(LEQ, "<=", startLine)
		}
		return s.emit(LSS, "<", startLine)

	case '>':
		if s.peek() == '>' {
			s.next()
			if s.peek() == '=' {
				s.next()
				return s.emit(SHR_ASSIGN, ">>=", startLine)
			}
			return s.emit(SHR, ">>", startLine)
		}
		if s.peek() == '=' {
			s.next()
			return s.emit(GEQ, ">=", startLine)
		}
		return s.emit(GTR, ">", startLine)

	// ── Comparação / atribuição ───────────────────────────────────────
	case '=':
		if s.peek() == '=' {
			s.next()
			return s.emit(EQL, "==", startLine)
		}
		return s.emit(ASSIGN, "=", startLine)

	case '!':
		if s.peek() == '=' {
			s.next()
			return s.emit(NEQ, "!=", startLine)
		}
		return s.emit(NOT, "!", startLine)

	case ':':
		if s.peek() == '=' {
			s.next()
			return s.emit(DEFINE, ":=", startLine)
		}
		return s.emit(COLON, ":", startLine)

	// ── Ellipsis ──────────────────────────────────────────────────────
	case '.':
		if s.peek() == '.' && s.peekAt(1) == '.' {
			s.next()
			s.next()
			return s.emit(ELLIPSIS, "...", startLine)
		}
		return s.emit(PERIOD, ".", startLine)

	// ── Agrupamento ───────────────────────────────────────────────────
	case '(':
		return s.emit(LPAREN, "(", startLine)
	case ')':
		return s.emit(RPAREN, ")", startLine)
	case '[':
		return s.emit(LBRACK, "[", startLine)
	case ']':
		return s.emit(RBRACK, "]", startLine)
	case '{':
		return s.emit(LBRACE, "{", startLine)
	case '}':
		return s.emit(RBRACE, "}", startLine)

	// ── Delimitadores ─────────────────────────────────────────────────
	case ';':
		return s.emit(SEMICOLON, ";", startLine)
	case ',':
		return s.emit(COMMA, ",", startLine)
	}

	return TokenData{}, fmt.Errorf("caractere invalido '%c' linha %d", c, startLine)
}

// emit atualiza lastType e retorna o TokenData — ponto único de saída.
func (s *Scanner) emit(t Token, lex string, line int) (TokenData, error) {
	s.lastType = t
	return TokenData{t, lex, line}, nil
}

func (s *Scanner) Tokenize() ([]TokenData, error) {
	var tokens []TokenData
	for {
		tok, err := s.NextToken()
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, tok)
		if tok.Type == EOF {
			break
		}
	}
	return tokens, nil
}
