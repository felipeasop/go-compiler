package main

import "fmt"

// Divisão por zero que só aparece em tempo de execução: o divisor é uma
// variável, então a análise semântica não pode detectar.
func divide(a int, b int) int {
	return a / b
}

func main() {
	for i := 3; i >= 0; i-- {
		fmt.Println(12, "/", i, "=", divide(12, i))
	}
}
