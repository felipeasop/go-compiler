package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"go-compiler/lexer"
	"go-compiler/parser"
)

func main() {
	// Código de teste (Instruções válidas para a gramática do nosso Parser)
	code :=
		`var x int = 10;
    fmt.Println(x);

    var y int = 10;
    if y > 5 {
        fmt.Println(y);
    }

    var z int = 5;
    for z > 0 {
        fmt.Println(z);
        z = z - 1;
    }`

	// Criar a pasta 'outputs' se ela não existir
	outputDir := "outputs"
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		os.Mkdir(outputDir, 0755)
	}

	// ==========================================
	// ANÁLISE LÉXICA
	// ==========================================
	fmt.Println("=== INICIANDO ANÁLISE LÉXICA ===")

	scanner := lexer.NewScanner(code)

	tokens, err := scanner.Tokenize()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro léxico fatal: %v\n", err)
		os.Exit(1)
	}

	// Imprime a Tabela de Tokens
	fmt.Printf("%-20s%-20s%-10s\n", "TIPO DE TOKEN", "LEXEMA", "LINHA")
	fmt.Println(strings.Repeat("-", 50))
	for _, token := range tokens {
		fmt.Printf("%-20s%-20s%-10d\n", token.Type.String(), token.Lexeme, token.Line)
	}
	fmt.Println(strings.Repeat("-", 50))

	// Exportação JSON dos Tokens para a pasta outputs/
	tokensJSON, err := json.MarshalIndent(tokens, "", "  ")
	if err == nil {
		err = os.WriteFile(outputDir+"/tokens.json", tokensJSON, 0644)
		if err == nil {
			fmt.Println("Arquivo 'outputs/tokens.json' gerado com sucesso!")
		} else {
			fmt.Printf("Erro ao salvar tokens.json: %v\n", err)
		}
	}

	// ==========================================
	// ANÁLISE SINTÁTICA (PARSER)
	// ==========================================
	fmt.Println("\n=== INICIANDO ANÁLISE SINTÁTICA ===")

	p := parser.NewParser(tokens)
	ast := p.ParseProgram()

	// Verifica se houveram erros sintáticos
	erros := p.Errors()
	if len(erros) > 0 {
		fmt.Println("Erros sintáticos encontrados:")
		for _, erro := range erros {
			fmt.Println("  -", erro)
		}
	} else {
		fmt.Println("Análise sintática concluída sem erros.")

		fmt.Println("\n=== FASE 3: EXIBICAO DA ARVORE SINTATICA ABSTRATA (AST) ===")
		if ast != nil {
			ast.Print(0) // Chama o print bonitão que acabamos de montar
		}
		fmt.Println(strings.Repeat("=", 58))
	}

	// Exportação JSON da AST para a pasta outputs/
	if ast != nil {
		astJSON := ast.ToJSON(0)
		err = os.WriteFile(outputDir+"/ast.json", []byte(astJSON), 0644)
		if err == nil {
			fmt.Println("Arquivo 'outputs/ast.json' gerado com sucesso!")
		} else {
			fmt.Printf("Erro ao salvar ast.json: %v\n", err)
		}
	} else {
		fmt.Println("Erro: AST não foi gerada devido a erros sintáticos críticos.")
	}
}
