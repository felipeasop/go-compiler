# Go Compiler

Compilador para um subconjunto da linguagem Go, escrito em Go. Desenvolvido como Trabalho
Prático Final da disciplina de Compiladores (Ciência da Computação, UENP, 2026/1), com a troca de
MicroC por Go aprovada pelo professor.

- **Professor:** Luciano Nascimento
- **Equipe:** Carlos Eduardo Costa Nardon e Felipe Augusto Santos Oliveira

O compilador cobre todas as etapas, do código-fonte à execução:

```
fonte → léxico → sintático (AST) → semântico → código de três endereços → otimização → interpretador
```

Todo programa aceito também é Go válido. Os testes comparam a saída do nosso compilador com a
do compilador Go oficial.

## Requisitos

- **Go 1.26** (sem dependências externas), **ou**
- **Nix** com flakes: `nix develop` abre um shell com Go e ferramentas.

## Como usar

```sh
go run . testdata/programs/fatorial.go          # mostra todas as etapas e executa
go run . -run testdata/programs/fatorial.go     # só executa
go run . -ir -opt testdata/programs/otimizacao.go   # código intermediário antes/depois
go run . -compare testdata/programs/series.go   # executa sem e com otimização e compara
```

| Opção | Efeito |
|---|---|
| (nenhuma) | mostra todas as etapas e executa |
| `-tokens` | tokens (análise léxica) |
| `-ast` | árvore sintática abstrata |
| `-sem` | tabela de símbolos e relatório semântico |
| `-ir` | código intermediário (três endereços) |
| `-opt` | código otimizado + resumo das otimizações |
| `-explain` | cada transformação feita pelo otimizador |
| `-run` | executa o código otimizado |
| `-O0` | executa o código **não** otimizado |
| `-compare` | executa as duas versões e compara as saídas |
| `-out DIR` | onde gravar os artefatos (padrão `outputs/`) |
| `-max-steps N` | limite de instruções executadas |

Os artefatos de cada etapa ficam em `outputs/`: `tokens.json`, `ast.json`,
`tabela_simbolos.json`, `relatorio_semantico.txt`, `codigo_intermediario.txt`,
`codigo_otimizado.txt` e `otimizacao.txt`.

**Códigos de saída:** `0` sucesso · `1` erro de compilação · `2` erro de execução.

### Com Nix

```sh
nix develop                                       # shell de desenvolvimento
nix run . -- testdata/programs/fatorial.go        # compila e roda o compilador
nix build                                         # gera ./result/bin/go-compiler
nix flake check                                   # build + todos os testes
```

> Flakes só enxergam arquivos rastreados pelo git: arquivos novos precisam de `git add` (ou use
> `path:.` no lugar de `.`).

## Testes

```sh
go test ./...                    # 111 casos
go test ./... -coverpkg=./...    # cobertura (~80%)
```

- `testdata/programs/`: 10 programas completos, cada um com a saída esperada (`.esperado`)
  gerada pelo `go run` oficial. São executados sem e com otimização, e as duas saídas têm de
  ser iguais à do Go.
- `testdata/{lexer,parser,semantic,runtime}/`: casos válidos (`pass`) e inválidos (`fail`) de
  cada etapa.
- `compiler/checks_test.go`: uma regra sintática/semântica por caso, com trechos curtos.
- Testes unitários do lexer e de cada técnica do otimizador.

Detalhes dos casos e resultados em [`docs/documentacao.md`](docs/documentacao.md#8-testes).

## Estrutura

```
lexer/       análise léxica (tokens, inserção automática de ';')
parser/      análise sintática descendente recursiva com modo pânico
ast/         nós da árvore sintática, impressão e JSON
semantic/    tabela de símbolos, escopos, tipos e verificações
ir/          código de três endereços: representação, gerador e avaliação
optimizer/   12 técnicas de otimização + análise de vivacidade
vm/          interpretador do código de três endereços
compiler/    pipeline que liga as etapas (+ testes de integração)
docs/        documentação técnica, gramática, roteiro de demonstração, changelog
testdata/    programas de teste
```

## Linguagem suportada

- Tipos `int`, `float64` (e `float32`), `string`, `bool`.
- `var x T = e`, `var x = e`, `var x T`, `x := e`, `x = e`, `x += e` (e `-= *= /= %=`),
  `x++`, `x--`, `_ = e`.
- Funções com parâmetros e retorno, recursão, variáveis globais.
- `if` / `else if` / `else` (com inicialização: `if x := f(); x > 0`).
- `for` nas três formas (clássico, condição, infinito), `break`, `continue`.
- Operadores aritméticos, relacionais e lógicos (com curto-circuito), conversões `int(x)` e
  `float64(x)`.
- `fmt.Println`, `fmt.Print` e `fmt.Printf`.

Gramática completa em [`docs/gramatica.md`](docs/gramatica.md).

## Documentação

Toda a documentação está em Markdown, na pasta `docs/`:

- [Documentação técnica](docs/documentacao.md): arquitetura, cada etapa, otimização, testes,
  decisões e limitações (relatório técnico).
- [Gramática](docs/gramatica.md): EBNF da linguagem suportada.
- [Roteiro da apresentação](docs/demonstracao.md): demonstração prática passo a passo, com os
  comandos de cada etapa e um exemplo de erro de cada tipo.
- [Registro de desenvolvimento](docs/CHANGELOG.md): evolução do projeto.

### Entregáveis do enunciado

| Entregável | Onde |
|---|---|
| Código-fonte completo do compilador | `lexer/` `parser/` `ast/` `semantic/` `ir/` `optimizer/` `vm/` `compiler/` `main.go` |
| Gramática da linguagem | [`docs/gramatica.md`](docs/gramatica.md) |
| Documentação da arquitetura e das etapas | [`docs/documentacao.md`](docs/documentacao.md) |
| Instruções de compilação, configuração e execução | este README (seções *Requisitos* e *Como usar*) e `flake.nix` |
| Programas de teste | [`testdata/`](testdata) |
| Relatório técnico com as decisões de implementação | [`docs/documentacao.md` §9](docs/documentacao.md#9-decisões-e-dificuldades-de-desenvolvimento) |
| Apresentação e demonstração prática | [`docs/demonstracao.md`](docs/demonstracao.md) |
