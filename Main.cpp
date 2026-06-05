#include <iostream>
#include <string>
#include <iomanip>

#include "./lexer/Utils.hpp"
#include "./lexer/Scanner.hpp"

int main() {

    // ----------------------------------------------------------------
    // TESTE 1: tipos primitivos, aritmetica, comparacao e if/else
    // ----------------------------------------------------------------
    std::string code_teste1 =
R"(package main

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
)";

    // ----------------------------------------------------------------
    // TESTE 2: strings, :=, for e comentarios de linha e bloco
    // ----------------------------------------------------------------
    std::string code_teste2 =
R"(package main

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
)";

    // ----------------------------------------------------------------
    // TESTE 3: underscores, aninhamento, todos os operadores
    // ----------------------------------------------------------------
    std::string code_teste3 =
R"(package main

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
)";

    std::string code = code_teste1;
    // std::string code = code_teste2;
    // std::string code = code_teste3;

    // Cria o scanner com o código de entrada
    Scanner scanner(code);

    try {
        // Cabeçalho da tabela
        std::cout  << std::left
                    << std::setw(20) << "TIPO DE TOKEN"
                    << std::setw(20) << "LEXEMA"
                    << std::setw(10) << "LINHA" << std::endl;
        std::cout << std::string(50, '-') << std::endl;

        // Lê o primeiro token
        Token token = scanner.nextToken();

        // Continua analisando até encontrar o fim da entrada
        while (token.type != TokenType::T_EOF) {
            std::cout << std::left
                        << std::setw(20) << tokenTypeToString(token.type)
                        << std::setw(20) << token.lexeme
                        << std::setw(10) << token.line << std::endl;

            // Busca o próximo token
            token = scanner.nextToken();
        }

        // Exibe o token EOF final
        std::cout << std::left
                    << std::setw(20) << tokenTypeToString(token.type)
                    << std::setw(20) << token.lexeme
                    << std::setw(10) << token.line << std::endl;

        std::cout << std::string(50, '-') << std::endl;
        std::cout << "Fim da analise lexica." << std::endl;

    } catch (std::exception& e) {
        // Caso ocorra erro léxico, a mensagem será exibida aqui
        std::cerr << "Erro: " << e.what() << std::endl;
    }

    return 0;
}
