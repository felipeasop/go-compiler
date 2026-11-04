package main

import "fmt"

// Precedência e associatividade de operadores, conferidas contra o Go.
func main() {
	fmt.Println(2+3*4, (2+3)*4, 10-3-2, 100/10/2, 17%5*2)
	fmt.Println(-2*-3, !false && false, true || false && false)

	a := true
	b := false
	c := false
	fmt.Println(a || b && c, (a || b) && c, !a || !b)
	fmt.Println(1 < 2 == true, 3+4 > 2*3)
}
