package main

func semRetorno(n int) int {
	if n > 0 {
		return 1
	}
	// falta return quando n <= 0
}

func tipoErrado() int {
	return "texto"
}

func vazia() {
	return 10
}

func main() {
	semRetorno(1)
	tipoErrado()
	vazia()
}
