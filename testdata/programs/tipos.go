package main

import "fmt"

// Tipos int, float64, string e bool; conversões e Printf.
func media(soma int, n int) float64 {
	return float64(soma) / float64(n)
}

func saudacao(nome string) string {
	return "Olá, " + nome + "!"
}

func main() {
	var altura float64 = 1.75
	var peso float64 = 70
	imc := peso / (altura * altura)
	fmt.Printf("IMC: %.2f\n", imc)

	fmt.Println(media(10, 4))
	f := 9.99
	fmt.Println(int(f), int(-f), 7/2, 7.0/2)

	var nome string = "MicroGo"
	fmt.Println(saudacao(nome), len2(nome))

	ativo := true
	fmt.Println(ativo && !false, ativo == false, "abc" < "abd")
	fmt.Println(-5%3, -7/2)
}

func len2(s string) int {
	if s == "" {
		return 0
	}
	return 1
}
