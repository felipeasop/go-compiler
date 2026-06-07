# Go Compiler

Um compilador para a linguagem Go e implementado em Go, criado para estudo e aprendizado. Este projeto inclui um analisador léxico e um analisador sintático.

## 1. Analisador Léxico (Lexer)

O analisador léxico transforma código-fonte bruto em uma sequência de tokens estruturados, facilitando as etapas subsequentes de um compilador.

* **Reconhecimento de Tokens**: Identifica palavras reservadas, tipos primitivos, operadores aritméticos/relacionais e símbolos de agrupamento.
* **Tratamento de Comentários**: Suporte a comentários de linha (`//`) e de bloco (`/* ... */`).
* **Suporte a Literais**: Suporta inteiros, números de ponto flutuante (`float`) e strings com suporte a sequências de escape (`\n`, `\t`, etc.).
* **Pontuação**: Suporte a sintaxe sem ponto e vírgula, com inferência automática de delimitadores, assim como o Go.
* **Relatório de Erros**: Mensagens detalhadas de erro léxico com indicação da linha.
* **Saída Formatada**: Exibe os tokens encontrados em uma tabela organizada no terminal.

## 2. Analisador Sintático (Parser)

O analisador sintático transforma a sequência de tokens gerada pelo analisador léxico em uma árvore sintática, utilizando um parser descendente recursivo que valida a gramática do código.

* **Construção de Árvore Sintática**: Constrói uma árvore sintática a partir dos tokens reconhecidos.
* **Precedência Matemática e Lógica Baseada em 6 Camadas**: Resolve expressões respeitando a ordem de operadores
* **Tratamento de Escopo e Short Statements**: Capaz de parsear estruturas complexas e exclusivas do Go, como inicializações de variáveis embutidas diretamente nos laços repetitivos e condicionais 
* **Recuperação de Erros (Panic Mode)**: Implementa sincronização através de barreiras para evitar travamentos, reportando múltiplos erros sintáticos e a linha da ocorrência sem interromper a execução do Parser.

## Como Testar

Para testar o compilador, execute o seguinte comando:
```sh
go run . tests/nome_do_arquivo.go
```

É possível utilizar o compilador Go padrão para comparação.

```sh
go build tests/<nome_do_arquivo>.go
```
