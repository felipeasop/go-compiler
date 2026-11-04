package main

import "fmt"

// Fibonacci recursivo comparado com a versão iterativa.
func fib(n int) int {
	if n < 2 {
		return n
	}
	return fib(n-1) + fib(n-2)
}

func main() {
	var a int = 0
	var b int = 1
	for i := 0; i < 15; i++ {
		fmt.Print(a, " ")
		proximo := a + b
		a = b
		b = proximo
	}
	fmt.Println()
	fmt.Println("fib(20) =", fib(20))
}
