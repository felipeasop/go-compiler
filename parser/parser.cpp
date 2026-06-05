#include <iostream>
#include <string>
#include <vector>

#include "../lexer/Token.hpp"
#include "../lexer/TokenType.hpp"

// =====================================================================
// PARTE 2: A ÁRVORE SINTÁTICA ABSTRATA (AST - ABSTRACT SYNTAX TREE)
// =====================================================================
// PONTO CHAVE PARA OS ALUNOS:
// O Parser não gera apenas um "Sim/Não". Ele constrói uma árvore na memória!
// Nessa árvore, nós internos são operações (como "+" ou "atribuição")
// e as folhas são dados (como números ou variáveis).
// Usamos Programação Orientada a Objetos (polimorfismo) para representar isso.
// =====================================================================

// Classe Base Abstrata para todos os nós da nossa árvore
class ASTNode {
public:
    // Destrutor virtual: garante que ao deletar um nó pai, o C++ chame os
    // destrutores corretos dos nós filhos de forma limpa, evitando memory leaks.
    virtual ~ASTNode() = default;

    // Método virtual puro: cada tipo de nó implementará sua própria forma de
    // se desenhar na tela (Pretty Printing) com recuo estruturado (indent).
    virtual void print(int indent = 0) const = 0;
};

// Nó Raiz: Representa o programa completo, que é simplesmente uma lista de comandos/instruções
class ProgramNode : public ASTNode {
public:
    std::vector<ASTNode*> statements; // Vetor contendo os nós de cada comando do programa

    // Limpeza de Memória Recursiva: deleta cada comando armazenado no vetor
    ~ProgramNode() override {
        for (auto* stmt : statements) {
            delete stmt;
        }
    }

    void print(int indent = 0) const override {
        std::cout << std::string(indent, ' ') << "ProgramNode (Inicio do Programa)\n";
        for (auto* stmt : statements) {
            stmt->print(indent + 2); // Imprime os comandos filhos mais à direita
        }
    }
};

// Nó de Declaração de Variável (ex: "int x = 10;")
class VarDeclNode : public ASTNode {
public:
    std::string name;          // Nome da variável declarada
    ASTNode* initializer; // Nó da expressão de valor inicial (pode ser nulo, ex: "int x;")

    VarDeclNode(std::string n, ASTNode* init) : name(n), initializer(init) {}

    ~VarDeclNode() override {
        delete initializer; // Libera a expressão associada à inicialização
    }

    void print(int indent = 0) const override {
        std::cout << std::string(indent, ' ') << "VarDeclNode (Declaracao de Variavel: " << name << ")\n";
        if (initializer) {
            initializer->print(indent + 4);
        }
    }
};

// Nó de Atribuição de Valor (ex: "x = 20;")
class AssignNode : public ASTNode {
public:
    std::string name;    // Nome da variável que está recebendo o valor
    ASTNode* expr;  // Nó da expressão calculada que será guardada na variável

    AssignNode(std::string n, ASTNode* e) : name(n), expr(e) {}

    ~AssignNode() override {
        delete expr; // Libera a memória da expressão avaliada
    }

    void print(int indent = 0) const override {
        std::cout << std::string(indent, ' ') << "AssignNode (Atribuicao a variavel: " << name << ")\n";
        expr->print(indent + 4);
    }
};

// Nó de Impressão na Tela (ex: "print(soma);")
class PrintNode : public ASTNode {
public:
    ASTNode* expr;  // A expressão cujo resultado será impresso na tela

    PrintNode(ASTNode* e) : expr(e) {}

    ~PrintNode() override {
        delete expr;
    }

    void print(int indent = 0) const override {
        std::cout << std::string(indent, ' ') << "PrintNode (Comando Print)\n";
        expr->print(indent + 4);
    }
};

// Nó de Desvio Condicional (ex: "if (condicao) { bloco_then } else { bloco_else }")
class IfNode : public ASTNode {
public:
    ASTNode* condition;          // Expressão relacional da condição
    std::vector<ASTNode*> thenBranch; // Comandos a executar se a condição for verdadeira
    std::vector<ASTNode*> elseBranch; // Comandos a executar se for falsa (opcional)

    IfNode(ASTNode* cond, std::vector<ASTNode*> thenB, std::vector<ASTNode*> elseB)
        : condition(cond), thenBranch(thenB), elseBranch(elseB) {}

    ~IfNode() override {
        delete condition;
        for (auto* stmt : thenBranch) delete stmt;
        for (auto* stmt : elseBranch) delete stmt;
    }

    void print(int indent = 0) const override {
        std::cout << std::string(indent, ' ') << "IfNode (Condicional IF)\n";
        std::cout << std::string(indent + 2, ' ') << "Condicao:\n";
        condition->print(indent + 4);

        std::cout << std::string(indent + 2, ' ') << "Bloco 'Então':\n";
        for (auto* stmt : thenBranch) {
            stmt->print(indent + 4);
        }

        if (!elseBranch.empty()) {
            std::cout << std::string(indent + 2, ' ') << "Bloco 'Senão':\n";
            for (auto* stmt : elseBranch) {
                stmt->print(indent + 4);
            }
        }
    }
};

// Nó de Laço de Repetição (ex: "while (condicao) { corpo_do_laco }")
class WhileNode : public ASTNode {
public:
    ASTNode* condition;    // Condição lógica de permanência no laço
    std::vector<ASTNode*> body; // Instruções executadas repetidamente

    WhileNode(ASTNode* cond, std::vector<ASTNode*> b) : condition(cond), body(b) {}

    ~WhileNode() override {
        delete condition;
        for (auto* stmt : body) delete stmt;
    }

    void print(int indent = 0) const override {
        std::cout << std::string(indent, ' ') << "WhileNode (Laco de Repeticao WHILE)\n";
        std::cout << std::string(indent + 2, ' ') << "Condicao de entrada:\n";
        condition->print(indent + 4);

        std::cout << std::string(indent + 2, ' ') << "Corpo do laco:\n";
        for (auto* stmt : body) {
            stmt->print(indent + 4);
        }
    }
};

// Nó Binário Geral: Usado para qualquer operação que tenha lado ESQUERDO e DIREITO
// (ex: somas, subtrações, multiplicações, divisões e comparações lógicas).
class BinaryOpNode : public ASTNode {
public:
    TokenType op;     // O operador aplicado (ex: T_PLUS, T_EQ, T_GT)
    ASTNode* left;    // Operando do lado esquerdo
    ASTNode* right;   // Operando do lado direito

    BinaryOpNode(TokenType operation, ASTNode* l, ASTNode* r)
        : op(operation), left(l), right(r) {}

    ~BinaryOpNode() override {
        delete left;
        delete right;
    }

    void print(int indent = 0) const override {
        std::cout << std::string(indent, ' ') << "BinaryOpNode (Operacao Binaria: " << tokenTypeTostd::string(op) << ")\n";
        left->print(indent + 4);
        right->print(indent + 4);
    }
};

// Nó Literal Numérico (Nó Folha - guarda o valor bruto)
class NumberNode : public ASTNode {
public:
    int value;

    NumberNode(int v) : value(v) {}

    void print(int indent = 0) const override {
        std::cout << std::string(indent, ' ') << "NumberNode (Valor constante: " << value << ")\n";
    }
};

// Nó de Referência a Variável (Nó Folha - guarda o nome da variável buscada na memória)
class VariableNode : public ASTNode {
public:
    std::string name;

    VariableNode(std::string n) : name(n) {}

    void print(int indent = 0) const override {
        std::cout << std::string(indent, ' ') << "VariableNode (Busca variavel: " << name << ")\n";
    }
};


// =====================================================================
// PARTE 3: O ANALISADOR SINTÁTICO (PARSER DESCENDENTE RECURSIVO)
// =====================================================================
// PONTO CHAVE PARA OS ALUNOS:
// O Parser Descendente Recursivo analisa a lista de tokens do início ao fim.
// Cada regra da nossa gramática é convertida diretamente em uma função C++.
// Para saber qual regra seguir, o parser dá uma "espiada" (Lookahead) no
// próximo token usando o método peek().
// =====================================================================

class Parser {
private:
    std::vector<Token> tokens; // Lista estática de tokens gerada pela análise léxica
    size_t pos;           // Posição do ponteiro de leitura do Parser

    // Método de lookahead: Retorna o tipo do token atual sem avançar o ponteiro
    TokenType peek() {
        if (pos >= tokens.size()) return TokenType::T_EOF;
        return tokens[pos].type;
    }

    // Retorna a estrutura inteira do Token atual para obtermos valores de lexema e linha
    Token peekToken() {
        if (pos >= tokens.size()) return Token(TokenType::T_EOF, "", -1);
        return tokens[pos];
    }

    // Avança o ponteiro de leitura do parser em 1 unidade e retorna o token que acabou de ser pulado
    Token advance() {
        Token t = peekToken();
        if (t.type != TokenType::T_EOF) {
            pos++;
        }
        return t;
    }

    // O método mais importante do parser!
    // Ele valida se o token atual é EXATAMENTE o que a gramática espera.
    // Se for verdadeiro, ele o consome (avança); se não, acusa um erro estrutural.
    void match(TokenType expected) {
        if (peek() == expected) {
            advance(); // Avança para o próximo token
        } else {
            // Se falhar, interrompe com mensagem explicativa detalhada
            error("Esperava o token '" + tokenTypeTostd::string(expected) +
                  "' porem foi encontrado '" + peekToken().lexeme + "'");
        }
    }

    // Dispara uma exceção de erro sintático, indicando com precisão a linha do erro
    void error(const std::string& message) {
        Token t = peekToken();
        throw runtime_error("Erro Sintatico na linha " + to_std::string(t.line) + ": " + message);
    }

public:
    // O parser inicia apontando para a primeira posição (índice 0) dos tokens
    Parser(const std::vector<Token>& toks) : tokens(toks), pos(0) {}

    // Entrada mestre da Gramática:
    // Regra: Program -> Statement* (um programa é composto por zero ou mais instruções)
    ASTNode* parseProgram() {
        auto* prog = new ProgramNode();

        // Fica em loop consumindo comandos até achar o fim do arquivo (T_EOF)
        while (peek() != TokenType::T_EOF) {
            try {
                prog->statements.push_back(parseStatement());
            } catch (const std::exception& e) {
                // Muito didático: Se falhar em qualquer comando interno,
                // limpamos a árvore parcial criada até aqui para não deixar lixo na memória!
                delete prog;
                throw; // Repassa a exceção para a função principal tratar
            }
        }
        return prog;
    }

private:
    // Regra: Statement -> Declaration | Assignment | PrintStmt | IfStmt | WhileStmt
    // Usamos LOOKAHEAD (peek) para decidir qual caminho seguir!
    ASTNode* parseStatement() {
        switch (peek()) {
            case TokenType::T_INT:
                // Se começar com 'int', com certeza é uma declaração de variável!
                return parseDeclaration();
            case TokenType::T_ID:
                // Se começar com um nome de variável, com certeza é uma atribuição (ex: x = 5;)
                return parseAssignment();
            case TokenType::T_PRINT:
                // Se começar com 'print', segue para tratar a impressão
                return parsePrintStmt();
            case TokenType::T_IF:
                // Se começar com 'if', trata o desvio condicional
                return parseIfStmt();
            case TokenType::T_WHILE:
                // Se começar com 'while', trata o laço de repetição
                return parseWhileStmt();
            default:
                // Se não caiu em nenhuma dessas opções, o código foi escrito de forma inválida!
                error("Comando invalido ou nao reconhecido na linguagem");
                return nullptr;
        }
    }

    // Regra: Declaration -> "int" ID [ "=" Expression ] ";"
    ASTNode* parseDeclaration() {
        match(TokenType::T_INT); // Valida e consome a palavra "int"

        Token idTok = peekToken(); // Captura o token do identificador para ler o nome da variável
        match(TokenType::T_ID);  // Valida e consome o nome da variável

        ASTNode* initializer = nullptr; // Por padrão, a variável pode não ter valor inicial

        // Verifica se há uma inicialização com o caractere "=" (ex: int x = 10;)
        if (peek() == TokenType::T_ASSIGN) {
            match(TokenType::T_ASSIGN);      // Consome "="
            initializer = parseExpression(); // Processa a expressão após o igual
        }

        match(TokenType::T_SEMICOLON); // Toda declaração de variável termina obrigatoriamente com ";"

        return new VarDeclNode(idTok.lexeme, initializer);
    }

    // Regra: Assignment -> ID "=" Expression ";"
    ASTNode* parseAssignment() {
        Token idTok = peekToken(); // Guarda o token contendo o nome da variável destino
        match(TokenType::T_ID);  // Consome o ID

        match(TokenType::T_ASSIGN); // Garante que há um '=' logo em seguida

        ASTNode* expr = parseExpression(); // Parseia a expressão de valor atribuído

        match(TokenType::T_SEMICOLON); // Garante que a atribuição termina com ";"

        return new AssignNode(idTok.lexeme, expr);
    }

    // Regra: PrintStmt -> "print" "(" Expression ")" ";"
    ASTNode* parsePrintStmt() {
        match(TokenType::T_PRINT);  // Valida e consome "print"
        match(TokenType::T_LPAREN); // Valida e consome "("

        ASTNode* expr = parseExpression(); // Avalia o que está dentro do parêntese

        match(TokenType::T_RPAREN);    // Valida e consome ")"
        match(TokenType::T_SEMICOLON); // Garante o fechamento com ";"

        return new PrintNode(expr);
    }

    // Regra: IfStmt -> "if" "(" Expression ")" "{" Statement* "}" [ "else" "{" Statement* "}" ]
    ASTNode* parseIfStmt() {
        match(TokenType::T_IF);     // Valida "if"
        match(TokenType::T_LPAREN); // Valida "("

        ASTNode* cond = parseExpression(); // Processa a condição relacional interna

        match(TokenType::T_RPAREN); // Valida ")"

        match(TokenType::T_LBRACE); // Início do bloco then: "{"
        std::vector<ASTNode*> thenB;
        // Enquanto não achar o caractere de fechamento "}", lê os comandos do bloco interno
        while (peek() != TokenType::T_RBRACE && peek() != TokenType::T_EOF) {
            thenB.push_back(parseStatement());
        }
        match(TokenType::T_RBRACE); // Consome "}"

        std::vector<ASTNode*> elseB; // Bloco opcional
        // Se após fechar o bloco anterior o próximo token for 'else', tratamos o bloco alternativo
        if (peek() == TokenType::T_ELSE) {
            match(TokenType::T_ELSE);   // Consome "else"
            match(TokenType::T_LBRACE); // Início do bloco else: "{"
            while (peek() != TokenType::T_RBRACE && peek() != TokenType::T_EOF) {
                elseB.push_back(parseStatement());
            }
            match(TokenType::T_RBRACE); // Consome "}"
        }

        return new IfNode(cond, thenB, elseB);
    }

    // Regra: WhileStmt -> "while" "(" Expression ")" "{" Statement* "}"
    ASTNode* parseWhileStmt() {
        match(TokenType::T_WHILE);  // Consome "while"
        match(TokenType::T_LPAREN); // Consome "("

        ASTNode* cond = parseExpression(); // Processa a condição de permanência do laço

        match(TokenType::T_RPAREN); // Consome ")"

        match(TokenType::T_LBRACE); // Consome "{"
        std::vector<ASTNode*> body;
        while (peek() != TokenType::T_RBRACE && peek() != TokenType::T_EOF) {
            body.push_back(parseStatement());
        }
        match(TokenType::T_RBRACE); // Consome "}"

        return new WhileNode(cond, body);
    }

    // =====================================================================
    // CASCATA DE PRECEDÊNCIA (EXPRESSÕES)
    // =====================================================================
    // EXPLICAÇÃO PEDAGÓGICA PARA OS ALUNOS:
    // Por que não criamos uma regra única para expressões?
    // Porque operadores diferentes têm forças diferentes! A multiplicação (*)
    // deve rodar antes da soma (+).
    // Para resolver isso sem algoritmos complexos de ordenação, dividimos a
    // gramática em "camadas" aninhadas (Cascata de Precedência).
    //
    // Camada 1: Expression  -> Operadores Relacionais (==, <, >) - Precedência Baixa
    // Camada 2: SimpleExpr  -> Operadores Matemáticos (+, -)      - Precedência Média
    // Camada 3: Term        -> Operadores Matemáticos (*, /)      - Precedência Alta
    // Camada 4: Factor      -> Unidades básicas (número, ID, parênteses)
    //
    // Como o parser desce a cascata, ele sempre constrói os nós mais profundos
    // (de maior prioridade) primeiro!
    // =====================================================================

    // Camada 1: Expression -> SimpleExpr [ ( "==" | "<" | ">" ) SimpleExpr ]
    ASTNode* parseExpression() {
        ASTNode* left = parseSimpleExpr(); // Desce um nível para resolver a aritmética primeiro

        // Se houver um operador de comparação logo em seguida, processa-o
        if (peek() == TokenType::T_EQ || peek() == TokenType::T_LT || peek() == TokenType::T_GT) {
            TokenType op = peek(); // Guarda o tipo do operador relacional
            advance();             // Consome o operador
            ASTNode* right = parseSimpleExpr(); // Desce o nível para obter o lado direito
            left = new BinaryOpNode(op, left, right); // Agrupa os dois lados sob o nó operador
        }
        return left;
    }

    // Camada 2: SimpleExpr -> Term (( "+" | "-" ) Term)*
    // O "*" (asterisco) na gramática significa repetição (usamos laço 'while' em C++)
    ASTNode* parseSimpleExpr() {
        ASTNode* left = parseTerm(); // Desce para priorizar multiplicações/divisões

        // Enquanto o próximo caractere for '+' ou '-', continua agrupando da esquerda para a direita
        while (peek() == TokenType::T_PLUS || peek() == TokenType::T_MINUS) {
            TokenType op = peek();
            advance(); // Consome '+' ou '-'
            ASTNode* right = parseTerm(); // Obtém o próximo operando
            left = new BinaryOpNode(op, left, right); // Substitui a raiz parcial com a operação atual
        }
        return left;
    }

    // Camada 3: Term -> Factor (( "*" | "/" ) Factor)*
    ASTNode* parseTerm() {
        ASTNode* left = parseFactor(); // Desce para o nível dos dados básicos (fatores)

        // Enquanto o próximo token for '*' ou '/', aplica a operação binária forte
        while (peek() == TokenType::T_MULT || peek() == TokenType::T_DIV) {
            TokenType op = peek();
            advance(); // Consome '*' ou '/'
            ASTNode* right = parseFactor();
            left = new BinaryOpNode(op, left, right);
        }
        return left;
    }

    // Camada 4 (Base): Factor -> NUMBER | ID | "(" Expression ")"
    ASTNode* parseFactor() {
        // Caso 1: Se for um literal numérico bruto (ex: 42)
        if (peek() == TokenType::T_NUM) {
            Token t = peekToken();
            match(TokenType::T_NUM); // Valida e consome o número
            return new NumberNode(stoi(t.lexeme)); // Converte o texto em inteiro real
        }
        // Caso 2: Se for o nome de uma variável sendo referenciada (ex: soma)
        else if (peek() == TokenType::T_ID) {
            Token t = peekToken();
            match(TokenType::T_ID); // Valida e consome o identificador
            return new VariableNode(t.lexeme);
        }
        // Caso 3: Se houver parênteses, forçamos a precedência! (ex: (1 + 2) * 3)
        else if (peek() == TokenType::T_LPAREN) {
            match(TokenType::T_LPAREN);        // Consome "("
            ASTNode* expr = parseExpression(); // Reinicia toda a cascata lá do topo para o que está dentro!
            match(TokenType::T_RPAREN);        // Consome ")"
            return expr; // Retorna a expressão interna como o fator processado
        }
        else {
            // Se não for nenhum dos três casos válidos, há um erro de digitação na expressão
            error("Fator invalido na expressao (esperava numero, variavel ou '(')");
            return nullptr;
        }
    }
};


// =====================================================================
// PARTE 4: FUNÇÃO PRINCIPAL (ORQUESTRADORA DA COMPILAÇÃO)
// =====================================================================

int main() {
    // Código-fonte escrito na linguagem MicroC.
    // O programa declara variáveis, testa condições e roda laços matemáticos.
    std::string code = R"(
        int soma = 10 + 20;

        if (soma == 30) {
            print(soma);
        }

        while (soma > 0) {
            soma = soma - 1;
        }
    )";

    std::cout << "=== FASE 1: INICIANDO ANALISE LEXICA (SCANNER) ===" << std::endl;
    std::vector<Token> tokens;
    try {
        Scanner scanner(code);
        Token t = scanner.nextToken();
        // Consome tokens até achar o fim do arquivo
        while (t.type != TokenType::T_EOF) {
            tokens.push_back(t);
            t = scanner.nextToken();
        }
        tokens.push_back(t); // Adiciona o marcador de fim do arquivo (T_EOF) para o Parser saber onde parar

        // Exibe de forma visível a tabela de tokens identificados
        for (const auto& tok : tokens) {
            if (tok.type != TokenType::T_EOF) {
                std::cout << tokenTypeTostd::string(tok.type) << " -> \""
                     << tok.lexeme << "\" (linha " << tok.line << ")" << std::endl;
            }
        }
        std::cout << "Analise lexica realizada com sucesso!\n" << std::endl;

    } catch (const std::exception& e) {
        cerr << "Erro durante a fase de Scanner: " << e.what() << std::endl;
        return 1;
    }

    std::cout << "=== FASE 2: INICIANDO ANALISE SINTATICA (PARSER) ===" << std::endl;
    try {
        // Instancia o Parser com o vetor estruturado de tokens
        Parser parser(tokens);

        // Executa o início do parsing e recebe o nó raiz da AST
        ASTNode* programAST = parser.parseProgram();

        std::cout << "Sucesso! A sequencia de tokens forma um programa estruturalmente valido para o MicroC.\n" << std::endl;
        std::cout << "=== FASE 3: EXIBICAO DA ARVORE SINTATICA ABSTRATA (AST) ===" << std::endl;

        // Dispara o Pretty Printer a partir da raiz da árvore
        programAST->print();
        std::cout << "==========================================================" << std::endl;

        // Desalocação Recursiva da memória para manter o sistema limpo
        delete programAST;
        std::cout << "\nMemoria dinâmica da AST desalocada com seguranca (sem vazamento de memoria)." << std::endl;

    } catch (const std::exception& e) {
        // Se houver qualquer erro estrutural sintático, a falha será capturada e impressa detalhadamente aqui
        cerr << "FALHA NA COMPILACAO!" << std::endl;
        cerr << e.what() << std::endl;
        return 1;
    }

    return 0;
}

Exibindo 02_analise_sintatica.cpp…
