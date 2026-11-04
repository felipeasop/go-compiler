package main

import "fmt"

// Algoritmo de Euclides (MDC) e MMC: vários parâmetros e laço while.
func mdc(a, b int) int {
	for b != 0 {
		resto := a % b
		a = b
		b = resto
	}
	return a
}

func mmc(a int, b int) int {
	return a / mdc(a, b) * b
}

func main() {
	fmt.Println(mdc(48, 18), mmc(4, 6))
	fmt.Println(mdc(1071, 462), mmc(21, 6))
}
