package main

import "fmt"

// Soma de série, contagem regressiva e potência com laços aninhados.
func potencia(base int, exp int) int {
	r := 1
	for exp > 0 {
		if exp%2 == 1 {
			r = r * base
		}
		base = base * base
		exp = exp / 2
	}
	return r
}

func main() {
	soma := 0.0
	for i := 1; i <= 10; i++ {
		soma += 1.0 / float64(i)
	}
	fmt.Printf("H(10) = %.4f\n", soma)

	for i := 0; i <= 10; i++ {
		fmt.Print(potencia(2, i), " ")
	}
	fmt.Println()

	linha := ""
	for i := 1; i <= 4; i++ {
		linha = ""
		for j := 0; j < i; j++ {
			linha += "*"
		}
		fmt.Println(linha)
	}
}
