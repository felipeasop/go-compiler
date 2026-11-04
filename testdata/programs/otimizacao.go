package main

import "fmt"

// Programa feito para exercitar o otimizador.

// Em calcula, a e b só são conhecidos em tempo de execução: aqui
// aparecem subexpressões comuns, simplificação algébrica, redução de
// força e propagação de cópias.
func calcula(a int, b int) int {
	c := (a + b) * (a + b) // (a + b) é calculado uma vez só
	d := b*1 + 0           // vira apenas b
	e := d * 2             // vira d + d
	f := c                 // cópia: usos de f passam a usar c
	return f + e
}

func main() {
	// Exemplo do enunciado: x = 10 * 2 + 5  →  x = 25
	x := 10*2 + 5

	// Propagação de constantes através de variáveis
	largura := 4
	altura := 5
	area := largura * altura

	// Código morto: valor calculado e nunca usado
	naoUsada := area * 1000
	naoUsada = 0
	_ = naoUsada

	// Condição constante: o else nunca executa
	if largura < altura {
		fmt.Println("largura menor")
	} else {
		fmt.Println("nunca impresso")
	}

	fmt.Println(x, area)
	for i := 1; i <= 3; i++ {
		fmt.Println(calcula(i, x))
	}
}
