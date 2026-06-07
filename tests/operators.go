package testdata

import "fmt"

func relationalOps() {
	var a int = 5
	var b int = 10

	if a == b {
		fmt.Println(a)
	}
	if a != b {
		fmt.Println(b)
	}
	if a < b {
		fmt.Println(a)
	}
	if a <= b {
		fmt.Println(a)
	}
	if b > a {
		fmt.Println(b)
	}
	if b >= a {
		fmt.Println(b)
	}
}

// CASO: precedência de operadores
// (2 + 3) * 4 deve ser 20, não 2 + 12
func operatorPrecedence() {
	var a int = 2 + 3*4
	var b int = (2 + 3) * 4
	fmt.Println(a)
	fmt.Println(b)
}

// CASO: expressão com múltiplos termos e associatividade à esquerda
// 10 - 3 - 2 deve ser (10 - 3) - 2 = 5, não 10 - (3 - 2) = 9
func leftAssociativity() {
	var x int = 10 - 3 - 2
	fmt.Println(x)
}
