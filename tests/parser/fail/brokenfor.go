package testdata

import "fmt"

func brokenFor() {
	var i int = 0

	// Colocou a condição dentro de chaves e esqueceu as chaves do corpo
	for {
		i < 10
	}
	i = i + 1

	fmt.Println(i)
}
