package tests

import "fmt"

// CASO: for com condição apenas (equivalente ao while de outras linguagens)
func forWhileStyle() {
	var i int = 0
	for i < 5 {
		fmt.Println(i)
		i = i + 1
	}
}

// CASO: for com break — interrompe o laço antes da condição falhar
func forWithBreak() {
	var i int = 0
	for i < 10 {
		if i == 3 {
			break
		}
		fmt.Println(i)
		i = i + 1
	}
}

// CASO: for com continue — pula para a próxima iteração
func forWithContinue() {
	var i int = 0
	for i < 5 {
		i = i + 1
		if i == 3 {
			continue
		}
		fmt.Println(i)
	}
}

// CASO: for aninhado — dois laços um dentro do outro
func forNested() {
	var i int = 0
	for i < 3 {
		var j int = 0
		for j < 3 {
			fmt.Println(j)
			j = j + 1
		}
		i = i + 1
	}
}
