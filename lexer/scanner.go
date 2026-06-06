package lexer

import (
	"fmt"
	"unicode"
)

// Mapa de palavras reservadas:
// associa texto (ex.: "if") ao tipo do token correspondente.
var keywords = map[string]TokenType{
	// Cadastro das palavras reservadas da linguagem
	"package": T_PACKAGE,
	"import":  T_IMPORT,
	"func":    T_FUNC,
	"var":     T_VAR,
	"int":     T_INT,
	"float":   T_FLOAT,
	"bool":    T_BOOL,
	"string":  T_STRING,
	"if":      T_IF,
	"else":    T_ELSE,
	"for":     T_FOR,
	"true":    T_TRUE,
	"false":   T_FALSE,
}

// Struct responsável por percorrer o código-fonte
// e transformar caracteres em tokens.
type Scanner struct {
	input []rune // Código-fonte de entrada
	pos   int    // Posição atual de leitura
	line  int    // Linha atual da análise
}

// New inicializa um novo Scanner para o código-fonte fornecido.
func New(source string) *Scanner {
	return &Scanner{
		input: []rune(source),
		pos:   0,
		line:  1,
	}
}

// peek retorna o caractere atual sem avançar na leitura.
// Se chegar ao fim da entrada, retorna 0 (equivalente ao '\0' do C++).
func (s *Scanner) peek() rune {
	if s.pos >= len(s.input) {
		return 0
	}
	return s.input[s.pos]
}

// next retorna o caractere atual e avança para a próxima posição.
func (s *Scanner) next() rune {
	c := s.peek()
	if c != 0 {
		s.pos++
	}
	return c
}

// skipWhitespace ignora espaços em branco, tabulações e quebras de linha.
// Sempre que encontra '\n', incrementa o contador de linhas.
func (s *Scanner) skipWhitespace() {
	for unicode.IsSpace(s.peek()) {
		if s.next() == '\n' {
			s.line++
		}
	}
}

// skipComment ignora comentários de uma linha iniciados por "//".
// Continua lendo até o final da linha ou fim da entrada.
func (s *Scanner) skipComment() {
	for s.peek() != '\n' && s.peek() != 0 {
		s.next()
	}
}

// skipBlockComment ignora comentários de bloco iniciados por "/*" e fechados por "*/".
// Incrementa o contador de linhas ao encontrar quebras de linha dentro
// do bloco. Retorna erro léxico se o bloco não for fechado.
func (s *Scanner) skipBlockComment() error {
	for s.peek() != 0 {
		c := s.next()

		// Controla a contagem de linhas dentro do bloco
		if c == '\n' {
			s.line++
		}

		// Verifica se encontrou o fechamento "*/"
		if c == '*' && s.peek() == '/' {
			s.next() // consome o '/'
			return nil
		}
	}

	// Chegou ao fim da entrada sem fechar o bloco
	return fmt.Errorf("Erro lexico: comentario de bloco nao fechado na linha: %d", s.line)
}

// scanNumber lê um número inteiro ou float a partir do primeiro dígito já encontrado.
func (s *Scanner) scanNumber(start rune) (Token, error) {
	buf := []rune{start}

	// Enquanto não encontrar um ponto não é um float
	isFloat := false

	// Continua enquanto encontrar outros dígitos ou ponto decimal
	for unicode.IsDigit(s.peek()) || s.peek() == '.' {
		// Se encontrar um ponto, marca que é um float
		if s.peek() == '.' {
			// Se já tiver um ponto, é um float inválido
			if isFloat {
				return Token{}, fmt.Errorf(
					"Erro lexico: float invalido: %s na linha: %d",
					string(buf), s.line,
				)
			}
			isFloat = true
		}
		buf = append(buf, s.next())
	}

	// Retorna um token float
	if isFloat {
		return Token{T_FLOAT_NUM, string(buf), s.line}, nil
	}

	// Retorna um token numérico
	return Token{T_NUM, string(buf), s.line}, nil
}

// scanIdentifier lê identificadores ou palavras reservadas.
// Um identificador pode conter letras, números e underscore.
func (s *Scanner) scanIdentifier(start rune) Token {
	// Adiciona o primeiro caractere já lido
	buf := []rune{start}

	// Continua lendo enquanto o padrão for válido para identificador
	for isIdentRune(s.peek()) {
		buf = append(buf, s.next())
	}

	lexeme := string(buf)

	// Verifica se o texto lido é uma palavra reservada
	if tt, ok := keywords[lexeme]; ok {
		return Token{tt, lexeme, s.line}
	}

	// Caso contrário, trata como identificador comum
	return Token{T_ID, lexeme, s.line}
}

// scanString lê uma string literal delimitada pelo caractere recebido em start.
// O lexema incluirá os delimitadores de abertura e fechamento.
// Valida sequências de escape. Retorna erro léxico se a string não for
// fechada ou se encontrar um escape inválido.
func (s *Scanner) scanString(start rune) (Token, error) {
	// Inclui o delimitador de abertura no lexema
	buf := []rune{start}

	// Lê até encontrar o delimitador de fechamento ou fim da entrada
	// Usa start como delimitador
	for s.peek() != start && s.peek() != 0 {
		if s.peek() == '\\' {
			buf = append(buf, s.next()) // consome a barra invertida

			// Valida o caractere de escape
			escaped := s.peek()
			switch escaped {
			case 'n', 't', '\\', '"', '\'', 'r':
				buf = append(buf, s.next())
			default:
				return Token{}, fmt.Errorf(
					"Erro lexico: escape invalido '\\%c' na linha %d",
					escaped, s.line,
				)
			}
			continue
		}

		// Controla linhas dentro de strings multilinha
		if s.peek() == '\n' {
			s.line++
		}
		buf = append(buf, s.next())
	}

	// Se chegou ao fim sem fechar a string, retorna erro
	if s.peek() == 0 {
		return Token{}, fmt.Errorf(
			"Erro lexico: string nao fechada na linha: %d", s.line,
		)
	}

	// Consome e inclui o delimitador de fechamento
	buf = append(buf, s.next())

	return Token{T_STRING_LITERAL, string(buf), s.line}, nil
}

// NextToken é o método principal do scanner:
// retorna o próximo token encontrado na entrada.
func (s *Scanner) NextToken() (Token, error) {
	// Primeiro, ignora espaços em branco
	s.skipWhitespace()

	// Se chegou ao fim da entrada, retorna EOF com lexema visível
	if s.pos >= len(s.input) {
		return Token{T_EOF, "EOF", s.line}, nil
	}

	// Lê o próximo caractere
	c := s.next()

	// Se começar com dígito, tenta formar um número
	if unicode.IsDigit(c) {
		return s.scanNumber(c)
	}

	// Se começar com letra ou underscore, tenta formar identificador
	if unicode.IsLetter(c) || c == '_' {
		return s.scanIdentifier(c), nil
	}

	// Analisa símbolos e operadores
	switch c {
	case '+':
		return Token{T_PLUS, "+", s.line}, nil

	case '-':
		return Token{T_MINUS, "-", s.line}, nil

	case '*':
		return Token{T_MULT, "*", s.line}, nil

	case '/':
		// Se houver outro '/', então é comentário de linha
		if s.peek() == '/' {
			s.next()             // consome o segundo '/'
			s.skipComment()      // ignora o restante da linha
			return s.NextToken() // busca o próximo token válido
		}

		// Se houver '*', então é comentário de bloco /* ... */
		if s.peek() == '*' {
			s.next()                                     // consome o '*'
			if err := s.skipBlockComment(); err != nil { // ignora até encontrar '*/'
				return Token{}, err
			}
			return s.NextToken() // busca o próximo token válido
		}

		return Token{T_DIV, "/", s.line}, nil

	case '"':
		return s.scanString(c)

	case '=':
		// Verifica se é "==" (igualdade)
		if s.peek() == '=' {
			s.next()
			return Token{T_EQ, "==", s.line}, nil
		}

		// Caso contrário, é "=" (atribuição)
		return Token{T_ASSIGN, "=", s.line}, nil

	case '<':
		return Token{T_LT, "<", s.line}, nil

	case '>':
		return Token{T_GT, ">", s.line}, nil

	case '(':
		return Token{T_LPAREN, "(", s.line}, nil

	case ')':
		return Token{T_RPAREN, ")", s.line}, nil

	case '{':
		return Token{T_LBRACE, "{", s.line}, nil

	case '}':
		return Token{T_RBRACE, "}", s.line}, nil

	case ';':
		return Token{T_SEMICOLON, ";", s.line}, nil

	case ':':
		if s.peek() == '=' {
			s.next()
			return Token{T_DECLARE_ASSIGN, ":=", s.line}, nil
		}
		return Token{T_COLON, ":", s.line}, nil
	}

	// Se encontrar um caractere que não pertence à linguagem,
	// retorna erro léxico informando o símbolo e a linha.
	return Token{}, fmt.Errorf(
		"Erro Lexico: caractere invalido '%c' na linha %d", c, s.line,
	)
}

// Tokenize percorre toda a entrada e retorna todos os tokens,
// incluindo o token EOF final. Interrompe e retorna erro no primeiro
// problema léxico encontrado.
func (s *Scanner) Tokenize() ([]Token, error) {
	var tokens []Token
	for {
		tok, err := s.NextToken()
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, tok)
		if tok.Type == T_EOF {
			break
		}
	}
	return tokens, nil
}

// isIdentRune reporta se r é válido no interior de um identificador.
func isIdentRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}
