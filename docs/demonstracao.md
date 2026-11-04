# Roteiro da demonstração

Roteiro da apresentação e da demonstração prática (itens 7 e 12 do enunciado). Cada passo mostra uma
etapa. Todos os comandos rodam a partir da raiz do projeto, dentro de `nix develop` ou com Go
1.26 instalado.

## 0. Preparação

```sh
nix develop          # ou: Go 1.26 instalado
go test ./...        # mostrar que os 111 testes passam
```

## 1. Programa de entrada

Abrir `testdata/programs/fatorial.go`: funções, recursão, laço e impressão.

## 2. Análise léxica: reconhecimento dos tokens

```sh
go run . -tokens testdata/programs/fatorial.go
```
Pontos a mostrar: os `;` inseridos automaticamente no fim das linhas e o número da linha de
cada token.

## 3. Análise sintática: AST

```sh
go run . -ast testdata/programs/fatorial.go
```
Mostrar `FuncNode` com assinatura, `ForNode` com `Init`/`Condicao`/`Post`, `CallNode` e a
precedência em `BinaryOpNode`. O JSON fica em `outputs/ast.json`.

## 4. Análise semântica: tabela de símbolos e regras

```sh
go run . -sem testdata/programs/escopo.go
```
Mostrar as três variáveis `x` em escopos diferentes (*shadowing*), as funções com assinatura e
as globais no escopo 0.

## 5. Código intermediário

```sh
go run . -ir testdata/programs/fatorial.go
```
Explicar o `for` traduzido em rótulos e desvios, e a convenção `param` + `call`.

## 6. Otimização: antes e depois

```sh
go run . -ir -opt testdata/programs/otimizacao.go
go run . -opt -explain testdata/programs/otimizacao.go     # cada transformação
```
- `main`: `x := 10*2 + 5` vira a constante `25`, como no exemplo do enunciado. O `if` com
  condição constante perde o `else`.
- `calcula`: subexpressão comum `(a + b)`, `b*1 + 0` → `b`, `d * 2` → `b + b`.

## 7. Execução e preservação do comportamento

```sh
go run . -run testdata/programs/fatorial.go
go run . -compare testdata/programs/series.go
```
O `-compare` executa as duas versões e mostra: saídas idênticas e menos instruções executadas
(709 → 578).

Opcional: comparar com o Go oficial: `go run testdata/programs/fatorial.go`.

## 8. Tratamento de erros (uma etapa de cada vez)

```sh
go run . testdata/lexer/fail/invalidos.go        # léxico: 4 erros reportados de uma vez
go run . testdata/parser/fail/brokenifelse.go    # sintático: else na linha seguinte
go run . -sem testdata/semantic/fail/funcoes.go  # semântico: argumentos, função inexistente...
go run . -run testdata/runtime/divzero.go        # execução: divisão por zero com linha
```

## Opções do compilador

| Opção | Efeito |
|---|---|
| (nenhuma) | mostra todas as etapas e executa |
| `-tokens` `-ast` `-sem` `-ir` `-opt` | mostra só as etapas escolhidas |
| `-run` | executa (código otimizado) |
| `-O0` | executa o código **não** otimizado |
| `-compare` | executa as duas versões e compara |
| `-explain` | lista cada transformação do otimizador |
| `-out DIR` | diretório dos artefatos (padrão `outputs/`) |
| `-max-steps N` | limite de instruções executadas |
