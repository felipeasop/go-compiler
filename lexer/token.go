package lexer

import "fmt"

// Struct que representa um token encontrado na análise léxica.
type Token struct {
	Type   TokenType // Tipo do token
	Lexeme string    // Texto exato encontrado na entrada
	Line   int       // Linha em que o token foi encontrado
}

// String formata o token de forma legível para depuração.
func (t Token) String() string {
	return fmt.Sprintf("Token{Type: %s, Lexeme: %q, Line: %d}", t.Type, t.Lexeme, t.Line)
}
