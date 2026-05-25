# Analisador Léxico (Go-Lexer)

Um analisador léxico eficiente escrito em C++ para uma linguagem inspirada na sintaxe de Go. Este projeto transforma código-fonte bruto em uma sequência de tokens estruturados, facilitando as etapas subsequentes de um compilador.

## Funcionalidades
* **Reconhecimento de Tokens**: Identifica palavras reservadas, tipos primitivos, operadores aritméticos/relacionais e símbolos de agrupamento.
* **Tratamento de Comentários**: Suporte a comentários de linha (`//`) e de bloco (`/* ... */`).
* **Suporte a Literais**: Suporta inteiros, números de ponto flutuante (`float`) e strings com suporte a sequências de escape (`\n`, `\t`, etc.).
* **Relatório de Erros**: Mensagens detalhadas de erro léxico com indicação da linha.
* **Saída Formatada**: Exibe os tokens encontrados em uma tabela organizada no terminal.

## Como Compilar e Rodar

Certifique-se de estar na pasta raiz do projeto. Você pode compilar todos os arquivos `.cpp` de uma vez:

```bash
# Compilação
g++ *.cpp -o main

# Execução
./main
```

## Como Testar

O arquivo `Main.cpp` contém três conjuntos de testes pré-configurados. Para alternar entre eles, edite o final da função `main()` em `Main.cpp`:
```cpp
// Descomente apenas um por vez:
std::string code = code_teste1;
// std::string code = code_teste2;
// std::string code = code_teste3;
```

## Saída Esperada

O Analisador imprime os resultados em uma tabela:

| Token Type | Lexema | Linha |
| T_PACKAGE |	package |	1 |
| T_ID |	main |	1 |
T_IMPORT | import |3
T_STRING_LITERAL | "fmt" | 3 |
T_FUNC | func | 5 |
T_ID | main | 5 |
| ... |	... |	... |
