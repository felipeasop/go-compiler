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
    T_INT,
    T_IF,
    T_ELSE,
    T_WHILE,
    T_PRINT,

    // Identificadores e números
    T_ID,
    T_NUM,

    // Operadores de atribuição e comparação
    T_ASSIGN,
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
        keywords["int"] = TokenType::T_INT;
        keywords["if"] = TokenType::T_IF;
        keywords["else"] = TokenType::T_ELSE;
        keywords["while"] = TokenType::T_WHILE;
        keywords["print"] = TokenType::T_PRINT;
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

    // Lê um número inteiro a partir do primeiro dígito já encontrado.
    Token scanNumber(char start) {

        string buffer;

        // Coloca o primeiro dígito no buffer
        buffer += start;

        // Continua enquanto encontrar outros dígitos
        while (isdigit(peek())) {
            buffer += next();
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
                next();           // consome o segundo '/'
                skipComment();    // ignora o restante da linha
                return nextToken(); // busca o próximo token válido
            }

            return Token(TokenType::T_DIV, "/", line);

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

    case TokenType::T_INT:
        return "T_INT";

    case TokenType::T_IF:
        return "T_IF";

    case TokenType::T_ELSE:
        return "T_ELSE";

    case TokenType::T_WHILE:
        return "T_WHILE";

    case TokenType::T_PRINT:
        return "T_PRINT";

    case TokenType::T_ID:
        return "T_ID";

    case TokenType::T_NUM:
        return "T_NUM";

    case TokenType::T_ASSIGN:
        return "T_ASSIGN";

    case TokenType::T_EQ:
        return "T_EQ";

    case TokenType::T_PLUS:
        return "T_PLUS";

    case TokenType::T_MINUS:
        return "T_MINUS";

    case TokenType::T_MULT:
        return "T_MULT";

    case TokenType::T_DIV:
        return "T_DIV";

    case TokenType::T_LT:
        return "T_LT";

    case TokenType::T_GT:
        return "T_GT";

    case TokenType::T_LPAREN:
        return "T_LPAREN";

    case TokenType::T_RPAREN:
        return "T_RPAREN";

    case TokenType::T_LBRACE:
        return "T_LBRACE";

    case TokenType::T_RBRACE:
        return "T_RBRACE";

    case TokenType::T_SEMICOLON:
        return "T_SEMICOLON";

    case TokenType::T_EOF:
        return "T_EOF";

    default:
        return "UNKNOWN";
    }
}

int main() {

    // Código-fonte de exemplo que será analisado.
    // A string raw (R"( ... )") permite escrever o texto em múltiplas linhas.
    string code = R"(

int soma = 10 + 20;

if (soma == 30) {
print(soma);
}

// comentario ignorado

)";

    // Cria o scanner com o código de entrada
    Scanner scanner(code);

    try {

        // Lê o primeiro token
        Token token = scanner.nextToken();

        // Continua analisando até encontrar o fim da entrada
        while (token.type != TokenType::T_EOF) {

            // Exibe o tipo do token, o lexema e a linha correspondente
            cout
                << tokenTypeToString(token.type)
                << " -> "
                << token.lexeme
                << " (linha "
                << token.line
                << ")"
                << endl;

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
