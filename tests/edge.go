package testdata

import "fmt"

func edge3() {
	// O Go permite declarar variáveis DENTRO do cabeçalho do if.
	// O escopo da variável 'status' só existe dentro deste if.
	if status := 200; status == 200 {
		fmt.Println(status)
	}

	// Funciona até mesmo com o "else if"
	var max int = 500
	if a := 10; a > max {
		fmt.Println(a)
	} else if b := 1000; b > max {
		fmt.Println(b)
	}
}
