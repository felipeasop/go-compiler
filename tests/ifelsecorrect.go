package testdata

import "fmt"

// CASO: else e else if na MESMA linha que o }
// Este é o único jeito válido em Go.
// O scanner insere ";" após "}" quando há quebra de linha —
// então "} \n else" vira "} ; else" e quebra.
// "} else {" na mesma linha NÃO gera ";" e funciona corretamente.

func ifElseCorrect() {
	var x int = 10

	if x > 5 {
		fmt.Println(x)
	} else {
		fmt.Println(x)
	}

	var y int = 10

	if y > 10 {
		fmt.Println(y)
	} else if y == 10 {
		fmt.Println(y)
	} else {
		fmt.Println(y)
	}
}
