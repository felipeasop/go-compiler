package main

import "fmt"

// Laços aninhados, break, continue, resto da divisão e bool.
func ehPrimo(n int) bool {
	if n < 2 {
		return false
	}
	for d := 2; d*d <= n; d++ {
		if n%d == 0 {
			return false
		}
	}
	return true
}

func main() {
	contador := 0
	n := 0
	for {
		n++
		if !ehPrimo(n) {
			continue
		}
		fmt.Print(n, " ")
		contador++
		if contador == 15 {
			break
		}
	}
	fmt.Println()
	fmt.Println("total:", contador)
}
