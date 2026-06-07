package lexer

import "strconv"

// Token representa a classe léxica de um lexema.
type Token int

const (
	// ── Especiais ────────────────────────────────────────────────────
	ILLEGAL Token = iota
	EOF
	COMMENT

	literal_beg
	// ── Literais ─────────────────────────────────────────────────────
	IDENT  // identificadores e palavras não reservadas (true, false, int...)
	INT    // 123
	FLOAT  // 3.14
	IMAG   // 3i
	CHAR   // 'a'
	STRING // "abc" ou `abc`
	literal_end

	operator_beg
	// ── Operadores aritméticos ────────────────────────────────────────
	ADD // +
	SUB // -
	MUL // *
	QUO // /
	REM // %

	// ── Operadores bit a bit ──────────────────────────────────────────
	AND // &
	OR  // |
	XOR // ^
	SHL // <<
	SHR // >>

	// ── Operadores de atribuição composta ─────────────────────────────
	ADD_ASSIGN // +=
	SUB_ASSIGN // -=
	MUL_ASSIGN // *=
	QUO_ASSIGN // /=
	REM_ASSIGN // %=

	AND_ASSIGN // &=
	OR_ASSIGN  // |=
	XOR_ASSIGN // ^=
	SHL_ASSIGN // <<=
	SHR_ASSIGN // >>=

	// ── Operadores lógicos ────────────────────────────────────────────
	LAND // &&
	LOR  // ||

	// ── Incremento / Decremento ───────────────────────────────────────
	INC // ++
	DEC // --

	// ── Comparação e atribuição ───────────────────────────────────────
	EQL    // ==
	LSS    // <
	GTR    // >
	ASSIGN // =
	NOT    // !

	NEQ      // !=
	LEQ      // <=
	GEQ      // >=
	DEFINE   // :=
	ELLIPSIS // ...

	// ── Agrupamento ───────────────────────────────────────────────────
	LPAREN // (
	LBRACK // [
	LBRACE // {
	COMMA  // ,
	PERIOD // .

	RPAREN    // )
	RBRACK    // ]
	RBRACE    // }
	SEMICOLON // ;
	COLON     // :
	operator_end

	keyword_beg
	// ── Palavras reservadas ───────────────────────────────────────────
	BREAK
	CASE
	CONST
	CONTINUE
	DEFAULT
	DEFER
	ELSE
	FOR
	FUNC
	GO
	IF
	IMPORT
	INTERFACE
	MAP
	PACKAGE
	RANGE
	RETURN
	SELECT
	STRUCT
	SWITCH
	TYPE
	VAR
	keyword_end

	additional_beg
	TILDE // ~
	additional_end
)

var tokens = [...]string{
	ILLEGAL: "ILLEGAL",
	EOF:     "EOF",
	COMMENT: "COMMENT",

	IDENT:  "IDENT",
	INT:    "INT",
	FLOAT:  "FLOAT",
	IMAG:   "IMAG",
	CHAR:   "CHAR",
	STRING: "STRING",

	ADD: "+",
	SUB: "-",
	MUL: "*",
	QUO: "/",
	REM: "%",

	AND: "&",
	OR:  "|",
	XOR: "^",
	SHL: "<<",
	SHR: ">>",

	ADD_ASSIGN: "+=",
	SUB_ASSIGN: "-=",
	MUL_ASSIGN: "*=",
	QUO_ASSIGN: "/=",
	REM_ASSIGN: "%=",

	AND_ASSIGN: "&=",
	OR_ASSIGN:  "|=",
	XOR_ASSIGN: "^=",
	SHL_ASSIGN: "<<=",
	SHR_ASSIGN: ">>=",

	LAND: "&&",
	LOR:  "||",
	INC:  "++",
	DEC:  "--",

	EQL:      "==",
	LSS:      "<",
	GTR:      ">",
	ASSIGN:   "=",
	NOT:      "!",
	NEQ:      "!=",
	LEQ:      "<=",
	GEQ:      ">=",
	DEFINE:   ":=",
	ELLIPSIS: "...",

	LPAREN:    "(",
	LBRACK:    "[",
	LBRACE:    "{",
	COMMA:     ",",
	PERIOD:    ".",
	RPAREN:    ")",
	RBRACK:    "]",
	RBRACE:    "}",
	SEMICOLON: ";",
	COLON:     ":",

	BREAK:     "break",
	CASE:      "case",
	CONST:     "const",
	CONTINUE:  "continue",
	DEFAULT:   "default",
	DEFER:     "defer",
	ELSE:      "else",
	FOR:       "for",
	FUNC:      "func",
	GO:        "go",
	IF:        "if",
	IMPORT:    "import",
	INTERFACE: "interface",
	MAP:       "map",
	PACKAGE:   "package",
	RANGE:     "range",
	RETURN:    "return",
	SELECT:    "select",
	STRUCT:    "struct",
	SWITCH:    "switch",
	TYPE:      "type",
	VAR:       "var",

	TILDE: "~",
}

// MarshalJSON serializa o Token como string legível no JSON
// em vez do número inteiro do iota.
func (tok Token) MarshalJSON() ([]byte, error) {
	return []byte(`"` + tok.String() + `"`), nil
}

func (tok Token) String() string {
	s := ""
	if 0 <= tok && int(tok) < len(tokens) {
		s = tokens[tok]
	}
	if s == "" {
		s = "token(" + strconv.Itoa(int(tok)) + ")"
	}
	return s
}

var keywords map[string]Token

func init() {
	keywords = make(map[string]Token, keyword_end-(keyword_beg+1))
	for i := keyword_beg + 1; i < keyword_end; i++ {
		keywords[tokens[i]] = i
	}
}

// Lookup mapeia um identificador para palavra reservada ou IDENT.
func Lookup(ident string) Token {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}

// TokenData é a estrutura consumida pelo Parser e serializada em JSON.
type TokenData struct {
	Type   Token  `json:"tipo"`
	Lexeme string `json:"lexema"`
	Line   int    `json:"linha"`
}
