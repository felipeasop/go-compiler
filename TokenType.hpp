#pragma once

// Enumeração que representa todos os tipos de tokens
// reconhecidos pelo analisador léxico.
enum class TokenType {
    // Palavras reservadas
    // Estrutura do programa
    T_PACKAGE,
    T_IMPORT,
    T_FUNC,
    T_VAR,

    // Tipos primitivos
    T_INT,
    T_FLOAT,
    T_BOOL,
    T_STRING,

    // Condicionais e loops
    T_IF,
    T_ELSE,
    T_FOR,

    // Valores booleanos
    T_TRUE,
    T_FALSE,

    // Identificadores e números
    T_ID,
    T_NUM,
    T_FLOAT_NUM,
    T_STRING_LITERAL,

    // Operadores de atribuição e comparação
    T_ASSIGN,
    T_DECLARE_ASSIGN,
    T_EQ,

    // Operadores aritméticos
    T_PLUS,
    T_MINUS,
    T_MULT,
    T_DIV,

    // Operadores relacionais
    T_LT,
    T_GT,

    // Símbolos de agrupamento
    T_LPAREN,
    T_RPAREN,

    T_LBRACE,
    T_RBRACE,

    // Delimitador de instrução
    T_SEMICOLON,
    T_COLON,

    // Fim do arquivo/entrada
    T_EOF
};
