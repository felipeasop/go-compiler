package tests

import "fmt"

func testSemicolons() {
	// Aqui o lexer deve injetar um ';' invisível no final da linha do x
	var x int = 100
	var y int = 200

	// O lexer NÃO deve injetar ';' depois do '+', porque a expressão não acabou
	var z int = x +
		y

	// Deve injetar ';' após a chamada da função
	fmt.Println(z)
}
