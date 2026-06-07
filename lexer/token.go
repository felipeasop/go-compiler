package lexer

// Struct que representa um token encontrado na análise léxica.
type Token struct {
	Type   TokenType // Tipo do token
	Lexeme string    // Texto exato encontrado na entrada
	Line   int       // Linha em que o token foi encontrado
}
