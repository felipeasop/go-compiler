package main

import "fmt"

// && e || com avaliação em curto-circuito: a função do lado direito
// só é chamada quando necessário.
var chamadas int = 0

func verdadeiro() bool {
	chamadas++
	return true
}

func main() {
	if false && verdadeiro() {
		fmt.Println("não deveria imprimir")
	}
	if true || verdadeiro() {
		fmt.Println("ok 1")
	}
	if verdadeiro() && verdadeiro() {
		fmt.Println("ok 2")
	}
	fmt.Println("chamadas:", chamadas)

	a := 5
	b := 0
	if b != 0 && a/b > 1 {
		fmt.Println("não deveria imprimir")
	}
	fmt.Println("sem divisão por zero")
}
