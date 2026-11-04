# Registro de desenvolvimento

Evolução do projeto, da mais recente para a mais antiga. Os detalhes de cada decisão estão em
[`documentacao.md`](documentacao.md#9-decisões-e-dificuldades-de-desenvolvimento).

## 2026-11-03: Back-end, otimização e integração

**Front-end (correções necessárias para gerar código)**
- AST: `VarDeclNode.TypeName`, `FuncNode.Params/ReturnType`, novo `CallNode` com argumentos,
  `PrintCallNode` com vários argumentos (`Println`, `Print`, `Printf`), linha em todos os nós.
- `ForNode` com `Init`/`Post` separados do corpo. Antes, `continue` pulava o `post`.
- Parser: parâmetros `(a int, b int)` e `(a, b int)`, tipo de retorno, chamadas como expressão
  e como instrução, atribuições compostas (`+= -= *= /= %=`), identificador vazio `_`.
- Parser: corrigida a precedência `&&` > `||` e a associatividade dos relacionais, para ficar
  igual ao Go.
- Lexer: não para mais no primeiro erro; reporta todos.

**Análise semântica**
- Assinaturas de função registradas numa primeira passada (permite recursão e qualquer ordem).
- Novas verificações: chamadas (existência, nº e tipo de argumentos), retorno (tipo, ausência,
  `return` em `void`), condição booleana, operadores por tipo, `break`/`continue` fora de laço,
  tipo desconhecido, `func main` sem parâmetros/retorno.
- Constantes não tipadas: `var f float64 = 1` e `f * 2` reescritos para `float64`.
- Tabela de símbolos com categoria (variável/parâmetro/função) e assinatura; impressa no terminal.

**Novas etapas**
- `ir/`: gerador de código de três endereços (quádruplas), com curto-circuito e renomeação de
  variáveis com *shadowing*.
- `optimizer/`: 12 técnicas (dobramento e propagação de constantes, propagação de cópias,
  simplificação algébrica, redução de força, subexpressões comuns, desvios constantes, código
  inalcançável, otimização de desvios, rótulos não usados, atribuições mortas com *liveness*,
  coalescência de cópias), com relatório de cada transformação.
- `vm/`: interpretador do código de três endereços com recursão, globais e erros de execução
  com linha.
- `compiler/`: pipeline que encadeia as etapas.
- `main.go`: opções `-tokens -ast -sem -ir -opt -run -O0 -compare -explain`.

**Testes e infraestrutura**
- `tests/` → `testdata/`, para que `go test ./...` ignore os programas de exemplo.
- 10 programas completos com saída esperada gerada pelo Go oficial. Novos casos de erro léxico,
  semântico e de execução. Testes unitários do lexer e do otimizador.
- `flake.nix` (`nix develop`, `nix build`, `nix run`, `nix flake check`).
- Documentação: `docs/documentacao.md`, `docs/gramatica.md`, `docs/demonstracao.md`.

## 2026-08-03: Análise semântica
- Analisador semântico (padrão Visitor) com tabela de símbolos em pilha de escopos.
- Verificações: variável não declarada, declaração duplicada, uso antes da inicialização,
  incompatibilidade de tipos, variável não usada, divisão por zero literal, código morto.
- Exportação da tabela de símbolos e do relatório semântico.

## 2026-06-05 a 2026-06-07: Análise sintática e AST
- Projeto reescrito em Go; parser descendente recursivo com modo pânico.
- AST separada do parser (`ast/`), com impressão hierárquica e exportação JSON.
- Precedência de operadores em camadas; funções; `if` com inicialização; `for` nas três formas.
- Palavras reservadas do compilador Go oficial; inserção automática de `;` no lexer.
- Programas de teste organizados em `pass`/`fail` por etapa.

## 2026-05-23 a 2026-05-25: Análise léxica
- Primeira versão do analisador léxico: identificadores, palavras reservadas, inteiros, floats,
  strings com escapes, raw strings e comentários de linha e de bloco.
- Tabela de tokens no terminal; README com instruções.
