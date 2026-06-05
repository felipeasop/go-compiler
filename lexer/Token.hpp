#pragma once

#include "TokenType.hpp"
#include <string>

// Estrutura que representa um token encontrado na análise léxica.
struct Token {

    TokenType type;   // Tipo do token
    std::string lexeme;    // Texto exato encontrado na entrada
    int line;         // Linha em que o token foi encontrado

    // Construtor para inicializar um token
    Token(TokenType t, std::string l, int ln)
        : type(t), lexeme(l), line(ln) {
    }
};
