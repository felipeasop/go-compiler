package testdata

import "fmt"

// CASO: return dentro de função
func funcWithReturn() {
	var x int = 10
	if x > 5 {
		return
	}
	fmt.Println(x)
}

// CASO: return com valor
func funcWithReturnValue() {
	var x int = 42
	fmt.Println(x)
	return
}

// CASO: break dentro de if dentro de for
func breakInsideIf() {
	var i int = 0
	for i < 100 {
		if i == 5 {
			break
		}
		i = i + 1
	}
	fmt.Println(i)
}

// CASO: continue dentro de if dentro de for
func continueInsideIf() {
	var i int = 0
	for i < 10 {
		i = i + 1
		if i == 5 {
			continue
		}
		fmt.Println(i)
	}
}
