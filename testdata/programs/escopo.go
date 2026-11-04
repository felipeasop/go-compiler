package main

import "fmt"

// Variáveis globais, shadowing em blocos aninhados e if com inicialização.
var contador int = 100
var mensagem = "global"

func incrementa() {
	contador++
}

func main() {
	fmt.Println(mensagem, contador)
	incrementa()
	incrementa()
	fmt.Println(contador)

	x := 1
	if x := 2; x > 1 {
		fmt.Println("if:", x)
		x := 3
		fmt.Println("bloco:", x)
	}
	fmt.Println("fora:", x)

	for i := 0; i < 2; i++ {
		x := i * 10
		fmt.Println("laço:", x)
	}

	var mensagem string = "local"
	fmt.Println(mensagem)
}
