package lexer

import "fmt"

// Struct que representa um token encontrado na análise léxica.
type Token struct {
	Type   TokenType `json:"tipo"`   // Tipo do token
	Lexeme string    `json:"lexema"` // Texto exato encontrado na entrada
	Line   int       `json:"linha"`  // Linha em que o token foi encontrado
}

func (t TokenType) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", t.String())), nil
}
