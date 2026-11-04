package main

import "fmt"

// Fatorial recursivo e iterativo: chamadas de função, recursão e laço.
func fatorialRec(n int) int {
	if n <= 1 {
		return 1
	}
	return n * fatorialRec(n-1)
}

func fatorialIter(n int) int {
	resultado := 1
	for i := 2; i <= n; i++ {
		resultado *= i
	}
	return resultado
}

func main() {
	for n := 0; n <= 10; n++ {
		fmt.Println(n, fatorialRec(n), fatorialIter(n))
	}
}
