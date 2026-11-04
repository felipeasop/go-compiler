package main

import "fmt"

func soma(a int, b int) int {
	return a + b
}

func avisa(msg string) {
	fmt.Println(msg)
}

func main() {
	x := soma(1)            // número de argumentos errado
	y := soma(1, "dois")    // tipo de argumento errado
	z := subtrai(5, 3)      // função não declarada
	w := avisa("oi") + 1    // função sem retorno usada em expressão
	soma = 3                // atribuição a uma função
	fmt.Println(x, y, z, w)
}
