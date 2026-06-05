#include <string>
#include <stdexcept>
#include <cctype>

#include "Scanner.hpp"

Scanner::Scanner(std::string source)
    : input(source), pos(0), line(1) {

    // Cadastro das palavras reservadas da linguagem
    keywords["package"] = TokenType::T_PACKAGE;
    keywords["import"]  = TokenType::T_IMPORT;
    keywords["func"]    = TokenType::T_FUNC;
    keywords["var"]     = TokenType::T_VAR;
    keywords["int"]     = TokenType::T_INT;
    keywords["float"]   = TokenType::T_FLOAT;
    keywords["bool"]    = TokenType::T_BOOL;
    keywords["string"]  = TokenType::T_STRING;
    keywords["if"]      = TokenType::T_IF;
    keywords["else"]    = TokenType::T_ELSE;
    keywords["for"]     = TokenType::T_FOR;
    keywords["true"]    = TokenType::T_TRUE;
    keywords["false"]   = TokenType::T_FALSE;
}

char Scanner::peek() {
    if (pos >= input.length()) return '\0';
    return input[pos];
}

char Scanner::next() {
    char c = peek();
    if (c != '\0') pos++;
    return c;
}

void Scanner::skipWhitespace() {
    while (isspace(peek())) {
        if (next() == '\n') {
            line++;
        }
    }
}

void Scanner::skipComment() {
    while (peek() != '\n' && peek() != '\0') {
        next();
    }
}

void Scanner::skipBlockComment() {
    while (peek() != '\0') {
        char c = next();

        // Controla a contagem de linhas dentro do bloco
        if (c == '\n') {
            line++;
        }

        // Verifica se encontrou o fechamento "*/"
        if (c == '*' && peek() == '/') {
            next(); // consome o '/'
            return;
        }
    }

    // Chegou ao fim da entrada sem fechar o bloco
    throw std::runtime_error("Erro lexico: comentario de bloco nao fechado na linha: " + std::to_string(line));
}

Token Scanner::scanNumber(char start) {
    std::string buffer;
    buffer += start;

    // Enquanto não encontrar um ponto não é um float
    bool isFloat = false;

    // Continua enquanto encontrar outros dígitos ou ponto decimal
    while (isdigit(peek()) || peek() == '.') {
        // Se encontrar um ponto, marca que é um float
        if (peek() == '.') {
            // Se já tiver um ponto, é um float inválido
            if (isFloat) {
                throw std::runtime_error(
                    "Erro lexico: float invalido: " + buffer +
                    " na linha: " + std::to_string(line)
                );
            }
            isFloat = true;
        }
        buffer += next();
    }

    // Retorna um token float
    if (isFloat) return Token(TokenType::T_FLOAT_NUM, buffer, line);

    // Retorna um token numérico
    return Token(TokenType::T_NUM, buffer, line);
}

Token Scanner::scanIdentifier(char start) {
    std::string buffer;

    // Adiciona o primeiro caractere já lido
    buffer += start;

    // Continua lendo enquanto o padrão for válido para identificador
    while (isalnum(peek()) || peek() == '_') buffer += next();

    // Verifica se o texto lido é uma palavra reservada
    if (keywords.count(buffer)) return Token(keywords[buffer], buffer, line);

    // Caso contrário, trata como identificador comum
    return Token(TokenType::T_ID, buffer, line);
}

Token Scanner::scanString(char start) {
    std::string buffer;

    // Inclui o delimitador de abertura no lexema
    buffer += start;

    // Lê até encontrar o delimitador de fechamento ou fim da entrada
    // Usa start como delimitador
    while (peek() != start && peek() != '\0') {
        if (peek() == '\\') {
            buffer += next(); // consome a barra invertida

            // Valida o caractere de escape
            char escaped = peek();
            switch (escaped) {
                case 'n':
                case 't':
                case '\\':
                case '"':
                case '\'':
                case 'r':
                    buffer += next();
                    break;
                default:
                    throw std::runtime_error(
                        "Erro lexico: escape invalido '\\" +
                        std::string(1, escaped) +
                        "' na linha " + std::to_string(line)
                    );
            }
            continue;
        }

        // Controla linhas dentro de strings multilinha
        if (peek() == '\n') line++;
        buffer += next();
    }

    // Se chegou ao fim sem fechar a string, lança erro
    if (peek() == '\0') {
        throw std::runtime_error(
            "Erro lexico: string nao fechada na linha: " +
            std::to_string(line)
        );
    }

    // Consome e inclui o delimitador de fechamento
    buffer += next();

    return Token(TokenType::T_STRING_LITERAL, buffer, line);
}

Token Scanner::nextToken() {
    // Primeiro, ignora espaços em branco
    skipWhitespace();

    // Se chegou ao fim da entrada, retorna EOF com lexema visível
    if (pos >= input.length()) return Token(TokenType::T_EOF, "EOF", line);

    // Lê o próximo caractere
    char c = next();

    // Se começar com dígito, tenta formar um número
    if (isdigit(c)) return scanNumber(c);

    // Se começar com letra ou underscore, tenta formar identificador
    if (isalpha(c) || c == '_') return scanIdentifier(c);

    // Analisa símbolos e operadores
    switch (c) {
        case '+':
            return Token(TokenType::T_PLUS, "+", line);

        case '-':
            return Token(TokenType::T_MINUS, "-", line);

        case '*':
            return Token(TokenType::T_MULT, "*", line);

        case '/':

            // Se houver outro '/', então é comentário de linha
            if (peek() == '/') {
                next();             // consome o segundo '/'
                skipComment();      // ignora o restante da linha
                return nextToken(); // busca o próximo token válido
            }

            // Se houver '*', então é comentário de bloco /* ... */
            if (peek() == '*') {
                next();                // consome o '*'
                skipBlockComment();    // ignora até encontrar '*/'
                return nextToken();    // busca o próximo token válido
            }

            return Token(TokenType::T_DIV, "/", line);

        case '"':
            return scanString(c);

        case '=':

            // Verifica se é "==" (igualdade)
            if (peek() == '=') {
                next();
                return Token(TokenType::T_EQ, "==", line);
            }

            // Caso contrário, é "=" (atribuição)
            return Token(TokenType::T_ASSIGN, "=", line);

        case '<':
            return Token(TokenType::T_LT, "<", line);

        case '>':
            return Token(TokenType::T_GT, ">", line);

        case '(':
            return Token(TokenType::T_LPAREN, "(", line);

        case ')':
            return Token(TokenType::T_RPAREN, ")", line);

        case '{':
            return Token(TokenType::T_LBRACE, "{", line);

        case '}':
            return Token(TokenType::T_RBRACE, "}", line);

        case ';':
            return Token(TokenType::T_SEMICOLON, ";", line);

        case ':':
            if (peek() == '=') {
                next();
                return Token(TokenType::T_DECLARE_ASSIGN, ":=", line);
            }
            return Token(TokenType::T_COLON, ":", line);

        default:
            // Se encontrar um caractere que não pertence à linguagem,
            // lança erro léxico informando o símbolo e a linha.
            throw std::runtime_error("Erro Lexico: caractere invalido '" +
                std::string(1, c) + "' na linha " + std::to_string(line)
            );
    }
}
