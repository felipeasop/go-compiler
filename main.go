package main

import (
	"fmt"
	"os"
)

func main() {
	codeTeste1 := `package main

import "fmt"

func main() {
    var x int = 10
    var y int = 20
    var soma int = x + y

    z := 5
    resultado := soma + z

    var pi float = 3.14
    area := pi * 2

    var ativo bool = true
    inativo := false

    var nome string = "joao"

    if (soma == 30) {
        var dobro int = soma + soma
    } else {
        var metade int = soma - 5
    }

    if (x < y) {
        diff := y - x
    }
}
`

	codeTeste2 := `package main

import "fmt"

func main() {
    var nome string = "maria silva"
    var vazia string = ""
    saudacao := "ola, mundo!"
    escapada := "ele disse \"oi\" pra mim"

    /* comentario de bloco
       com multiplas linhas
       deve ser ignorado */

    // declaracao curta de numericos
    contador := 0
    taxa := 1.5

    for (contador < 5) {
        contador = contador + 1
    }

    /* outro bloco antes de instrucao */
    for (taxa < 3.0) {
        taxa = taxa + 0.5
    }

    var limite int = 100
    acumulado := 0

    for (acumulado < limite) {
        acumulado = acumulado + 10
    }
}
`

	codeTeste3 := `package main

import "fmt"

func main() {
    var valor_inicial int = 0
    var preco_total float = 99.99
    var nome_completo string = "ana souza"
    _contador := 1

    var a int = 10 + 2
    var b int = 10 - 3
    var c int = a * b
    var d int = c / 4

    if (a == 12) {
        resultado := a + b
    }

    if (b < a) {
        diff := a - b
    }

    if (c > d) {
        var grande int = c
    }

    var i int = 0
    for (i < 5) {
        i = i + 1
        var parcial float = preco_total * i
        if (parcial > 200) {
            var aviso string = "limite atingido"
        } else {
            var ok string = "dentro do limite"
        }
    }

    if (valor_inicial == 0) {
        if (_contador > 0) {
            _contador = _contador + 1
        } else {
            _contador = 0
        }
    }
}
`
	// Selecione o teste desejado trocando a variável abaixo:
	code := codeTeste1
	_ = codeTeste2
	_ = codeTeste3

	s := scanner.New(code)

	// Cabeçalho da tabela
	fmt.Printf("%-20s %-20s %-10s\n", "TIPO DE TOKEN", "LEXEMA", "LINHA")
	fmt.Println(repeat('-', 50))

	for {
		tok, err := s.NextToken()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Erro: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("%-20s %-20s %-10d\n", tok.Type, tok.Lexeme, tok.Line)

		if tok.Type == scanner.T_EOF {
			break
		}
	}

	fmt.Println(repeat('-', 50))
	fmt.Println("Fim da analise lexica.")
}

func repeat(ch rune, n int) string {
	buf := make([]rune, n)
	for i := range buf {
		buf[i] = ch
	}
	return string(buf)
}
