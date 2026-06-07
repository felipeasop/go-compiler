# Go Compiler

Um compilador para a linguagem Go e implementado em Go, criado para estudo e aprendizado. Este projeto inclui um analisador léxico e um analisador sintático.

## 1. Analisador Léxico (Lexer)

O analisador léxico transforma código-fonte bruto em uma sequência de tokens estruturados, facilitando as etapas subsequentes de um compilador.

* **Escaneamento e Categorização de Tokens**: Identifica e classifica strings brutas em unidades lógicas com significado para o compilador (palavras reservadas, identificadores, tipos primitivos, operadores e símbolos de agrupamento).
* **Descarte de Espaços e Comentários**: Filtra e remove do fluxo de processamento os espaços em branco redundantes, comentários de linha (`//`) e comentários de bloco (`/* ... */`), deixando apenas o código útil para as próximas fases.
* **Processamento de Literais**: Reconhece e isola números inteiros, pontos flutuantes (`float`) e cadeias de texto (`string`), tratando internamente sequências de escape complexas como `\n` e `\t`.
* **Inserção Automática de Ponto e Vírgula (ASI)**: Replica a especificação nativa do Go, analisando o final das linhas e injetando pontos e vírgulas (`;`) invisíveis de forma automática para simplificar a escrita do código-fonte.
* **Rastreamento de Erros Léxicos**: Detecta instantaneamente caracteres inválidos (que não pertencem ao alfabeto da linguagem) ou strings que foram abertas e não foram fechadas, apontando a linha exata da ocorrência.
* **Visualização de Tokens**: Imprime uma tabela no terminal contendo o Tipo do Token, o Lexema exato e a Linha.

## 2. Analisador Sintático (Parser)

O analisador sintático consome a sequência de tokens gerada pelo Lexe, validando se a ordem dos tokens respeita as regras gramaticais da linguagem Go através de um algoritmo Descendente Recursivo.

* **Validação Gramatical**: Garante que estruturas como funções, laços e condicionais sigam a sintaxe correta.

* **Precedência de Operadores em 6 Camadas**: Resolve expressões matemáticas e lógicas complexas, garantindo que operações como * tenham prioridade sobre +.

* **Tratamento de Short Statements**: Processa construções exclusivas do Go, como inicializações de variáveis embutidas diretamente em escopos de `if` e `for` (ex: `if x := 0; x > 5`).

* **Recuperação de Erros (Panic Mode)**: Caso encontre um erro de sintaxe, o Parser não quebra; ele ativa barreiras de sincronização (como encontrar um `;` ou `}`) para reportar múltiplos erros em uma única execução.

## 3. Árvore Sintática Abstrata (AST)

A AST é uma representação estruturada da árvore sintática do código Go, que reflete o significado lógico do código fonte, usada para análise e transformação posterior.

* **Separação de Conceitos (Nodes):** Organiza o código de forma rígida entre Statements (instruções de fluxo como IfNode, ForNode, ReturnNode) e Expressions (estruturas que geram valores como BinaryOpNode, LiteralNode).

* **Visualização Textual Hierárquica:** Possui o método Print() integrado que exibe a árvore no terminal com recuos (indents) perfeitamente alinhados, facilitando a depuração da estrutura.

* **Serialização para JSON:** Implementa o método ToJSON(), permitindo exportar a árvore sintática completa para o formato JSON estruturado, ideal para ferramentas externas ou ferramentas de inspeção visual.

## Como Testar

Para testar o compilador, execute o seguinte comando:
```sh
go run . tests/nome_do_arquivo.go
```

É possível utilizar o compilador Go padrão para comparação.

```sh
go build tests/<nome_do_arquivo>.go
```
