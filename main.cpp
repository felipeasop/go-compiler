#include <iostream>
#include <string>
#include <unordered_map>
#include <stdexcept>
#include <cctype>

using namespace std;

// Enumeração que representa todos os tipos de tokens
// reconhecidos pelo analisador léxico.
enum class TokenType {
    // Palavras reservadas
    // Estrutura do programa
    T_PACKAGE,
    T_IMPORT,
    T_FUNC,
    T_MAIN,
    T_VAR,

    // Tipos
    T_INT,
    T_FLOAT,
    T_BOOL,
    T_STRING,

    // Condicionais e loops
    T_IF,
    T_ELSE,
    T_FOR,

    // Identificadores e números
    T_ID,
    T_NUM,
    T_FLOAT_NUM,
    T_STRING_CONTENT,

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

// Estrutura que representa um token encontrado na análise léxica.
struct Token {

    TokenType type;   // Tipo do token
    string lexeme;    // Texto exato encontrado na entrada
    int line;         // Linha em que o token foi encontrado

    // Construtor para inicializar um token
    Token(TokenType t, string l, int ln)
        : type(t), lexeme(l), line(ln) {
    }
};

// Classe responsável por percorrer o código-fonte
// e transformar caracteres em tokens.
class Scanner {

private:

    string input;     // Código-fonte de entrada
    size_t pos;       // Posição atual de leitura
    int line;         // Linha atual da análise

    // Mapa de palavras reservadas:
    // associa texto (ex.: "if") ao tipo do token correspondente.
    unordered_map<string, TokenType> keywords;

public:

    // Construtor do scanner.
    // Recebe o código-fonte e inicializa os dados internos.
    Scanner(string source)
        : input(source), pos(0), line(1) {

        // Cadastro das palavras reservadas da linguagem
        keywords["package"] = TokenType::T_PACKAGE;
        keywords["import"]  = TokenType::T_IMPORT;
        keywords["func"]    = TokenType::T_FUNC;
        keywords["main"]    = TokenType::T_MAIN;
        keywords["var"]     = TokenType::T_VAR;
        keywords["int"]     = TokenType::T_INT;
        keywords["float"]   = TokenType::T_FLOAT;
        keywords["bool"]    = TokenType::T_BOOL;
        keywords["string"]  = TokenType::T_STRING;
        keywords["if"]      = TokenType::T_IF;
        keywords["else"]    = TokenType::T_ELSE;
        keywords["for"]     = TokenType::T_FOR;
    }

    // Retorna o caractere atual sem avançar na leitura.
    // Se chegar ao fim da entrada, retorna '\0'.
    char peek() {

        if (pos >= input.length()) {
            return '\0';
        }

        return input[pos];
    }

    // Retorna o caractere atual e avança para a próxima posição.
    char next() {

        char c = peek();

        if (c != '\0') {
            pos++;
        }

        return c;
    }

    // Ignora espaços em branco, tabulações e quebras de linha.
    // Sempre que encontra '\n', incrementa o contador de linhas.
    void skipWhitespace() {

        while (isspace(peek())) {

            if (next() == '\n') {
                line++;
            }
        }
    }

    // Ignora comentários de uma linha iniciados por "//".
    // Continua lendo até o final da linha ou fim da entrada.
    void skipComment() {

        while (peek() != '\n' && peek() != '\0') {
            next();
        }
    }

    // Ignora comentários de bloco iniciados por "/*" e fechados por "*/".
    // Incrementa o contador de linhas ao encontrar quebras de linha dentro
    // do bloco. Lança erro léxico se o bloco não for fechado.
    void skipBlockComment() {

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
        throw runtime_error(
            "Erro lexico: comentario de bloco nao fechado na linha: " +
            to_string(line)
        );
    }

    // Lê um número inteiro ou float a partir do primeiro dígito já encontrado.
    Token scanNumber(char start) {

        string buffer;

        // Coloca o primeiro dígito no buffer
        buffer += start;
        bool isFloat = false;

        // Continua enquanto encontrar outros dígitos ou ponto decimal
        while (isdigit(peek()) || peek() == '.') {
            // Se encontrar um ponto, marca que é um float
            if (peek() == '.') {
                // Se já tiver um ponto, é um float inválido
                if (isFloat) {
                    throw runtime_error(
                        "Erro lexico: float invalido: " + buffer +
                        " na linha: " + to_string(line)
                    );
                }
                isFloat = true;
            }
            buffer += next();
        }

        // Retorna um token float
        if (isFloat) {
            return Token(TokenType::T_FLOAT_NUM, buffer, line);
        }

        // Retorna um token numérico
        return Token(TokenType::T_NUM, buffer, line);
    }

    // Lê identificadores ou palavras reservadas.
    // Um identificador pode conter letras, números e underscore.
    Token scanIdentifier(char start) {

        string buffer;

        // Adiciona o primeiro caractere já lido
        buffer += start;

        // Continua lendo enquanto o padrão for válido para identificador
        while (isalnum(peek()) || peek() == '_') {
            buffer += next();
        }

        // Verifica se o texto lido é uma palavra reservada
        if (keywords.count(buffer)) {
            return Token(keywords[buffer], buffer, line);
        }

        // Caso contrário, trata como identificador comum
        return Token(TokenType::T_ID, buffer, line);
    }

    // Lê uma string literal delimitada por aspas duplas.
    // O lexema incluirá as aspas de abertura e fechamento.
    // Lança erro léxico se a string não for fechada.
    Token scanString(char start) {

        string buffer;

        // Inclui a aspa de abertura no lexema
        buffer += start;

        // Lê até encontrar a aspa de fechamento ou fim da entrada
        while (peek() != '"' && peek() != '\0') {
            // Suporte a escape de aspas: \"
            if (peek() == '\\') {
                buffer += next(); // consome a barra invertida
                if (peek() != '\0') {
                    buffer += next(); // consome o caractere escapado
                }
                continue;
            }
            // Controla linhas dentro de strings multilinha
            if (peek() == '\n') {
                line++;
            }
            buffer += next();
        }

        // Se chegou ao fim sem fechar a string, lança erro
        if (peek() == '\0') {
            throw runtime_error(
                "Erro lexico: string nao fechada na linha: " +
                to_string(line)
            );
        }

        // Consome e inclui a aspa de fechamento
        buffer += next();

        return Token(TokenType::T_STRING_CONTENT, buffer, line);
    }

    // Método principal do scanner:
    // retorna o próximo token encontrado na entrada.
    Token nextToken() {

        // Primeiro, ignora espaços em branco
        skipWhitespace();

        // Se chegou ao fim da entrada, retorna EOF
        if (pos >= input.length()) {
            return Token(TokenType::T_EOF, "", line);
        }

        // Lê o próximo caractere
        char c = next();

        // Se começar com dígito, tenta formar um número
        if (isdigit(c)) {
            return scanNumber(c);
        }

        // Se começar com letra ou underscore, tenta formar identificador
        if (isalpha(c) || c == '_') {
            return scanIdentifier(c);
        }

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
            throw runtime_error(
                "Erro Lexico: caractere invalido '" +
                string(1, c) +
                "' na linha " +
                to_string(line)
            );
        }
    }
};

// Função auxiliar para converter o enum TokenType em texto.
// Isso facilita a exibição dos tokens no terminal.
string tokenTypeToString(TokenType type) {

    switch(type) {

    case TokenType::T_PACKAGE:        return "T_PACKAGE";
    case TokenType::T_IMPORT:         return "T_IMPORT";
    case TokenType::T_FUNC:           return "T_FUNC";
    case TokenType::T_MAIN:           return "T_MAIN";
    case TokenType::T_VAR:            return "T_VAR";
    case TokenType::T_INT:            return "T_INT";
    case TokenType::T_FLOAT:          return "T_FLOAT";
    case TokenType::T_BOOL:           return "T_BOOL";
    case TokenType::T_STRING:         return "T_STRING";
    case TokenType::T_IF:             return "T_IF";
    case TokenType::T_ELSE:           return "T_ELSE";
    case TokenType::T_FOR:            return "T_FOR";
    case TokenType::T_ID:             return "T_ID";
    case TokenType::T_NUM:            return "T_NUM";
    case TokenType::T_FLOAT_NUM:      return "T_FLOAT_NUM";
    case TokenType::T_STRING_CONTENT: return "T_STRING_CONTENT";
    case TokenType::T_ASSIGN:         return "T_ASSIGN";
    case TokenType::T_DECLARE_ASSIGN: return "T_DECLARE_ASSIGN";
    case TokenType::T_EQ:             return "T_EQ";
    case TokenType::T_PLUS:           return "T_PLUS";
    case TokenType::T_MINUS:          return "T_MINUS";
    case TokenType::T_MULT:           return "T_MULT";
    case TokenType::T_DIV:            return "T_DIV";
    case TokenType::T_LT:             return "T_LT";
    case TokenType::T_GT:             return "T_GT";
    case TokenType::T_LPAREN:         return "T_LPAREN";
    case TokenType::T_RPAREN:         return "T_RPAREN";
    case TokenType::T_LBRACE:         return "T_LBRACE";
    case TokenType::T_RBRACE:         return "T_RBRACE";
    case TokenType::T_SEMICOLON:      return "T_SEMICOLON";
    case TokenType::T_COLON:          return "T_COLON";
    case TokenType::T_EOF:            return "T_EOF";
    default:                          return "UNKNOWN";
    }
}

int main() {

    // ----------------------------------------------------------------
    // TESTE 1: tipos primitivos, aritmetica, comparacao e if/else
    // ----------------------------------------------------------------
    string code_teste1 = R"(
package main

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
    string code_teste2 = R"(
package main

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
    string code_teste3 = R"(
package main

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

    // string code = code_teste1;
    string code = code_teste2;
    // string code = code_teste3;

    // Cria o scanner com o código de entrada
    Scanner scanner(code);

    try {

        // Lê o primeiro token
        Token token = scanner.nextToken();

        // Continua analisando até encontrar o fim da entrada
        while (token.type != TokenType::T_EOF) {

            // Exibe o tipo do token, o lexema e a linha correspondente
            cout << tokenTypeToString(token.type) << " -> " << token.lexeme << " (linha " << token.line<< ")" << endl;

            // Busca o próximo token
            token = scanner.nextToken();
        }

        cout << endl;
        cout << "Fim da analise lexica." << endl;

    }
    catch (exception& e) {

        // Caso ocorra erro léxico, a mensagem será exibida aqui
        cerr << e.what() << endl;
    }

    return 0;
}
