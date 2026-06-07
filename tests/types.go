package testdata

import "fmt"

// CASO: todos os tipos suportados com declaração var + tipo explícito
func typesVarDecl() {
	var a int = 42
	var b float32 = 3.14
	var c string = "compiladores"
	var d bool = true

	fmt.Println(a)
	fmt.Println(b)
	fmt.Println(c)
	fmt.Println(d)
}

// CASO: declaração curta := (infere o tipo pelo valor)
func typesShortDecl() {
	x := 100
	y := 2.71
	z := "hello"
	w := false

	fmt.Println(x)
	fmt.Println(y)
	fmt.Println(z)
	fmt.Println(w)
}

// CASO: reatribuição após declaração
func typesReassign() {
	var n int = 1
	n = n + 1
	n = n * 2
	fmt.Println(n)
}
