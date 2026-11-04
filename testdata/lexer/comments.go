package tests

import "fmt"

// Este é um comentário de linha simples
// Ele deve ser completamente ignorado pelo Lexer

func testComments() {
	var a int = 10 / 2 // A primeira barra é divisão, as outras são comentário

	/* Este é um comentário de bloco.
	   O Lexer deve pular todas essas linhas.
	*/

	var b float32 = 3.14

	fmt.Println(a)
	fmt.Print(b)
}
