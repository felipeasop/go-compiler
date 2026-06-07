package main

import (
	"fmt"
	"os"
	"strings"

	"go-compiler/lexer"
)

func main() {
	code := `
package main

import "fmt"

func main() {
	var x int = 10 // ou x := 10

	fmt.Println(x)

	var y int = 10

	if y > 5 {
		fmt.Println(y)
	}

	var x int = 5

	for z > 0 {
		fmt.Println(z)
		z = z - 1 // ou x--
	}
}
`
	scanner := lexer.NewScanner(code)

	// Cabeçalho da tabela formatado com espaços em branco (equivalente ao std::setw e std::left)
	fmt.Printf("%-20s%-20s%-10s\n", "TIPO DE TOKEN", "LEXEMA", "LINHA")
	fmt.Println(strings.Repeat("-", 50))

	for {
		// Lê o token e verifica se ocorreu algum erro léxico (substitui o try/catch do C++)
		token, err := scanner.NextToken()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Erro: %v\n", err)
			os.Exit(1)
		}

		// Imprime o token atual
		fmt.Printf("%-20s%-20s%-10d\n", token.Type, token.Lexeme, token.Line)

		// Verifica se chegou ao fim
		if token.Type == lexer.T_EOF {
			break
		}
	}

	fmt.Println(strings.Repeat("-", 50))
	fmt.Println("Fim da analise lexica.")
}
