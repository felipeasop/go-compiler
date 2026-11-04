package tests

import "fmt"

func testLiterals() {
	var integer int = 42
	floatNum := 123.456

	// Strings com escape
	var greeting string = "Olá, mundo!\nBem-vindo ao compilador."
	var tabbed string = "Coluna1\tColuna2\tColuna3"
	var quotes string = "Ele disse: \"Isso é um teste\""

	// Números colados em operadores
	var math int = 10 + 20*30

	fmt.Println(integer)
	fmt.Println(floatNum)
	fmt.Println(greeting)
	fmt.Println(tabbed)
	fmt.Println(quotes)
	fmt.Println(math)
}
