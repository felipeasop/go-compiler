package main

import "fmt"

func main() {
	n := 10
	if n {                     // condição não booleana
		fmt.Println(n)
	}
	s := "abc" * 2             // operador inválido para string
	ok := n && true            // && exige bool
	var f float64 = n          // int não converte implicitamente
	fmt.Println(s, ok, f)
	break                      // fora de laço
}
