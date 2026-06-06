package lexer

// Enumeração que representa todos os tipos de tokens
// reconhecidos pelo analisador léxico.
type TokenType int

const (
	// Palavras reservadas
	// Estrutura do programa
	T_PACKAGE TokenType = iota
	T_IMPORT
	T_FUNC
	T_VAR

	// Tipos primitivos
	T_INT
	T_FLOAT
	T_BOOL
	T_STRING

	// Condicionais e loops
	T_IF
	T_ELSE
	T_FOR

	// Valores booleanos
	T_TRUE
	T_FALSE

	// Comandos built-in
	T_PRINT
	T_WHILE

	// Identificadores e números
	T_ID
	T_NUM
	T_FLOAT_NUM
	T_STRING_LITERAL

	// Operadores de atribuição e comparação
	T_ASSIGN
	T_DECLARE_ASSIGN
	T_EQ

	// Operadores aritméticos
	T_PLUS
	T_MINUS
	T_MULT
	T_DIV

	// Operadores relacionais
	T_LT
	T_GT
	T_LE
	T_GE

	// Símbolos de agrupamento
	T_LPAREN
	T_RPAREN

	T_LBRACE
	T_RBRACE

	// Delimitador de instrução
	T_SEMICOLON
	T_COLON

	// Fim do arquivo/entrada
	T_EOF
)

// String converte o enum TokenType em texto.
// Isso facilita a exibição dos tokens no terminal.
// Em Go, implementar String() satisfaz a interface fmt.Stringer,
// substituindo a função tokenTypeToString() do Utils.cpp/hpp.
// Isso facilita a exibição dos tokens no terminal.
func (t TokenType) String() string {
	switch t {
	case T_PACKAGE:
		return "T_PACKAGE"
	case T_IMPORT:
		return "T_IMPORT"
	case T_FUNC:
		return "T_FUNC"
	case T_VAR:
		return "T_VAR"
	case T_INT:
		return "T_INT"
	case T_FLOAT:
		return "T_FLOAT"
	case T_BOOL:
		return "T_BOOL"
	case T_STRING:
		return "T_STRING"
	case T_IF:
		return "T_IF"
	case T_ELSE:
		return "T_ELSE"
	case T_FOR:
		return "T_FOR"
	case T_TRUE:
		return "T_TRUE"
	case T_FALSE:
		return "T_FALSE"
	case T_ID:
		return "T_ID"
	case T_NUM:
		return "T_NUM"
	case T_FLOAT_NUM:
		return "T_FLOAT_NUM"
	case T_STRING_LITERAL:
		return "T_STRING_LITERAL"
	case T_ASSIGN:
		return "T_ASSIGN"
	case T_DECLARE_ASSIGN:
		return "T_DECLARE_ASSIGN"
	case T_EQ:
		return "T_EQ"
	case T_PLUS:
		return "T_PLUS"
	case T_MINUS:
		return "T_MINUS"
	case T_MULT:
		return "T_MULT"
	case T_DIV:
		return "T_DIV"
	case T_LT:
		return "T_LT"
	case T_GT:
		return "T_GT"
	case T_LE:
		return "T_LE"
	case T_GE:
		return "T_GE"
	case T_LPAREN:
		return "T_LPAREN"
	case T_RPAREN:
		return "T_RPAREN"
	case T_LBRACE:
		return "T_LBRACE"
	case T_RBRACE:
		return "T_RBRACE"
	case T_SEMICOLON:
		return "T_SEMICOLON"
	case T_COLON:
		return "T_COLON"
	case T_EOF:
		return "T_EOF"
	default:
		return "UNKNOWN"
	}
}
