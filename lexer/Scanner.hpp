#pragma once

#include <string>
#include <unordered_map>

#include "TokenType.hpp"
#include "Token.hpp"

// Classe responsável por percorrer o código-fonte
// e transformar caracteres em tokens.
class Scanner {

private:

    std::string input;  // Código-fonte de entrada
    size_t pos;         // Posição atual de leitura
    int line;           // Linha atual da análise

    // Mapa de palavras reservadas:
    // associa texto (ex.: "if") ao tipo do token correspondente.
    std::unordered_map<std::string, TokenType> keywords;

public:

    Scanner(std::string source);

    // Retorna o caractere atual sem avançar na leitura.
    // Se chegar ao fim da entrada, retorna '\0'.
    char peek();

    // Retorna o caractere atual e avança para a próxima posição.
    char next();

    // Ignora espaços em branco, tabulações e quebras de linha.
    // Sempre que encontra '\n', incrementa o contador de linhas.
    void skipWhitespace();

    // Ignora comentários de uma linha iniciados por "//".
    // Continua lendo até o final da linha ou fim da entrada.
    void skipComment();

    // Ignora comentários de bloco iniciados por "/*" e fechados por "*/".
    // Incrementa o contador de linhas ao encontrar quebras de linha dentro
    // do bloco. Lança erro léxico se o bloco não for fechado.
    void skipBlockComment();

    // Lê um número inteiro ou float a partir do primeiro dígito já encontrado.
    Token scanNumber(char start);

    // Lê identificadores ou palavras reservadas.
    // Um identificador pode conter letras, números e underscore.
    Token scanIdentifier(char start);

    // Lê uma string literal delimitada pelo caractere recebido em start.
    // O lexema incluirá os delimitadores de abertura e fechamento.
    // Valida sequências de escape. Lança erro léxico se a string não for
    // fechada ou se encontrar um escape inválido.
    Token scanString(char start);

    // Método principal do scanner:
    // retorna o próximo token encontrado na entrada.
    Token nextToken();
};
