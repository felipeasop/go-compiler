package main

import "fmt"

// Vários erros léxicos: o lexer reporta todos de uma vez.
func main() {
	x := 10 @ 2
	nome := "texto sem fechar
	y := 3 $ 1
	fmt.Println("escape \q inválido")
}
