# Documentação Técnica — Compilador Go (subconjunto) escrito em Go

**Disciplina:** Compiladores, Ciência da Computação, UENP, 2026/1
**Professor:** Luciano Nascimento
**Equipe:** Carlos Eduardo Costa Nardon e Felipe Augusto Santos Oliveira

> A linguagem-fonte é um subconjunto de **Go**, no lugar da MicroC sugerida no enunciado. A troca
> foi aprovada pelo professor. Todo programa aceito pelo compilador também é um programa Go
> válido, o que permitiu usar o compilador Go oficial como referência nos testes: a saída do
> nosso compilador é comparada com a saída do `go run`.

## Sumário

1. [Arquitetura do compilador](#1-arquitetura-do-compilador)
2. [Análise léxica](#2-análise-léxica)
3. [Análise sintática](#3-análise-sintática)
4. [Análise semântica](#4-análise-semântica)
5. [Código intermediário](#5-código-intermediário)
6. [Otimização](#6-otimização)
7. [Back-end: interpretador](#7-back-end-interpretador)
8. [Testes](#8-testes)
9. [Decisões e dificuldades](#9-decisões-e-dificuldades-de-desenvolvimento)
10. [Conclusão](#10-conclusão)

---

## 1. Arquitetura do compilador

### 1.1 Visão geral

O compilador é escrito em Go puro (sem bibliotecas externas e sem geradores de analisadores),
com um pacote por etapa:

| Pacote        | Etapa                      | Entrada → Saída                                    |
|---------------|----------------------------|----------------------------------------------------|
| `lexer/`      | Análise léxica             | texto → `[]TokenData` + erros                      |
| `parser/`     | Análise sintática          | tokens → AST (`*ast.ProgramNode`) + erros          |
| `ast/`        | Árvore sintática abstrata  | definição dos nós, impressão e exportação JSON     |
| `semantic/`   | Análise semântica          | AST → relatório (erros/avisos) + tabela de símbolos|
| `ir/`         | Código intermediário       | AST validada → código de três endereços            |
| `optimizer/`  | Otimização                 | código de três endereços → código otimizado        |
| `vm/`         | Back-end (interpretador)   | código de três endereços → execução                |
| `compiler/`   | Integração                 | encadeia as etapas (usado pelo `main` e pelos testes) |
| `main.go`     | Interface de linha de comando | opções, impressão e gravação dos artefatos      |

### 1.2 Fluxo de dados

```
 arquivo.go
     │
     ▼
 ┌─────────┐  []TokenData   ┌─────────┐  AST   ┌───────────┐ AST + tabela ┌──────────┐
 │  Lexer  │ ─────────────▶ │ Parser  │ ─────▶ │ Semântico │ ───────────▶ │ Gerador  │
 └─────────┘                └─────────┘        └───────────┘              │   TAC    │
  erros léxicos            erros sintáticos    erros semânticos           └──────────┘
                                                                               │ ir.Program
                                                                               ▼
                        ┌──────────────┐  ir.Program otimizado  ┌──────────────────────┐
     saída do programa ◀│ Interpretador│ ◀───────────────────── │      Otimizador      │
                        └──────────────┘                        └──────────────────────┘
```

Cada etapa só roda se a anterior terminou sem erros (`compiler.Compile`). Avisos semânticos não
interrompem a compilação. Os artefatos de cada etapa vão para `outputs/`:

| Arquivo                        | Conteúdo                                  |
|--------------------------------|-------------------------------------------|
| `tokens.json`                  | lista de tokens (tipo, lexema, linha)     |
| `ast.json`                     | árvore sintática                          |
| `tabela_simbolos.json`         | tabela de símbolos completa               |
| `relatorio_semantico.txt`      | erros e avisos semânticos                 |
| `codigo_intermediario.txt`     | código de três endereços                  |
| `codigo_otimizado.txt`         | código de três endereços otimizado        |
| `otimizacao.txt`               | resumo e lista de cada transformação      |

---

## 2. Análise léxica

**Arquivos:** `lexer/token.go`, `lexer/scanner.go`

### 2.1 Estratégia

Scanner escrito à mão. Funciona como um autômato finito implementado com `switch` sobre o
caractere atual e *lookahead* de um caractere (`peek`) para os operadores de dois caracteres
(`:=`, `==`, `+=`, `&&`, `++`…). O texto é convertido para `[]rune`, então identificadores e
strings com acentos funcionam.

### 2.2 Tokens

| Classe            | Exemplos                                                           |
|-------------------|--------------------------------------------------------------------|
| Identificador     | `x`, `soma`, `fatorialRec`, `_`                                    |
| Literais          | `INT` 42 · `FLOAT` 3.14 · `STRING` "a\tb" / \`raw\` · `CHAR` 'a'   |
| Palavras reservadas | as 25 do Go (`func`, `var`, `if`, `else`, `for`, `return`…)      |
| Operadores        | `+ - * / %`, `+= -= *= /= %=`, `++ --`, `== != < <= > >=`, `&& \|\| !`, `= :=` |
| Delimitadores     | `( ) { } [ ] , ; . :`                                              |

A lista completa e a gramática estão em [`gramatica.md`](gramatica.md). Como no compilador
oficial, os nomes de token seguem o pacote `go/token`.

### 2.3 Regras

- Espaços, tabulações e comentários (`//` e `/* */`) são descartados. O comentário de bloco
  continua contando as linhas.
- **Inserção automática de `;`:** quando há quebra de linha e o último token foi identificador,
  literal, `return`, `break`, `continue`, `++`, `--`, `)`, `]` ou `}`, o lexer emite um `;`. É
  essa regra que permite escrever Go sem `;`, e é ela que faz `}\nelse` ser erro (como no Go).
- Sequências de escape aceitas: `\n \t \r \\ \" \'`.

### 2.4 Tratamento de erros

O lexer **não para no primeiro erro**: registra o erro, descarta o caractere inválido e continua,
para reportar todos os erros de uma vez.

| Erro                          | Exemplo        | Mensagem                                              |
|-------------------------------|----------------|-------------------------------------------------------|
| caractere inválido            | `x := 10 @ 2`  | `erro lexico linha 7: caractere invalido '@'`         |
| string não fechada            | `"texto`       | `erro lexico linha 8: string nao fechada`             |
| escape inválido               | `"\q"`         | `erro lexico linha 10: sequencia de escape invalida '\q'` |
| comentário de bloco não fechado | `/* ...`     | `erro lexico linha 4: comentario de bloco nao fechado` |

---

## 3. Análise sintática

**Arquivos:** `parser/parser.go`, `ast/*.go`

### 3.1 Estratégia

Parser **descendente recursivo** escrito à mão, com uma função por não-terminal da gramática
(`parseFunc`, `parseIf`, `parseFor`, `parseExpr`…). A gramática é LL(1) quase toda; nos dois pontos
em que não é, o parser olha mais à frente:

- `x := …`, `x = …`, `x += …`, `x++` e `f(…)` começam todos com `IDENT`. A decisão usa o
  segundo token (`lookahead(1)`).
- `for cond {` × `for init; cond; post {`: o parser procura um `;` antes do `{`.

A precedência de operadores fica codificada na própria gramática, com uma camada por nível
(`parseOrExpr` → `parseAndExpr` → `parseRelationalExpr` → `parseSimpleExpr` → `parseTerm` →
`parseUnary` → `parseFactor`). As repetições `{ op operando }` dão associatividade à esquerda.

### 3.2 AST

O parser produz uma AST, e não uma árvore de derivação. Os nós ficam separados em instruções
(`FuncNode`, `VarDeclNode`, `AssignNode`, `PrintCallNode`, `IfNode`, `ForNode`, `ReturnNode`,
`BreakNode`, `ContinueNode`) e expressões (`BinaryOpNode`, `UnaryOpNode`, `LiteralNode`,
`VariableNode`, `CallNode`). Todo nó guarda a **linha** do código-fonte, usada nas mensagens de
erro.

Algumas construções são reescritas já no parser (*desugaring*), para que as etapas seguintes
tratem menos casos:

| Código-fonte | AST equivalente  |
|--------------|------------------|
| `x++`        | `x = x + 1`      |
| `x += e`     | `x = x + e`      |
| `x := e`     | `var x = e`      |

### 3.3 Exemplo

```go
func soma(a, b int) int {
	return a + b
}
```
```
FuncNode (func soma(a int, b int) int)
  ReturnNode
      BinaryOpNode (+)
          VariableNode (a)
          VariableNode (b)
```

### 3.4 Tratamento de erros: modo pânico

Ao encontrar um erro, o parser registra a mensagem com a linha e **sincroniza**: descarta tokens
até um ponto seguro (`;` ou `}`) e continua, para reportar vários erros numa só execução. Toda
repetição tem uma salvaguarda contra laço infinito: se nenhum token foi consumido, o parser
avança um token.

```
erro sintatico linha 21: instrucao invalida (encontrou "else")
erro sintatico linha 10: expressao avaliada mas nao utilizada
erro sintatico linha 13: esperava }, encontrou EOF ("EOF")
```

---

## 4. Análise semântica

**Arquivos:** `semantic/analyzer.go`, `semantic/symboltable.go`, `semantic/errors.go`

### 4.1 Tabela de símbolos

A tabela é uma **pilha de escopos**. Cada escopo é um `map[string]*Symbol`:

```go
type Symbol struct {
    Name        string     // identificador
    Kind        SymbolKind // variável | parâmetro | função
    Type        SymbolType // int | float64 | string | bool | void
    Signature   string     // funções: "func(a int, b int) int"
    Scope       int        // nível (0 = global)
    DeclLine    int
    Initialized bool
    Used        bool
}
```

- `EnterScope`/`ExitScope` empilham e desempilham escopos.
- `Declare` só verifica duplicidade **no escopo atual**, o que permite *shadowing*.
- `Lookup` procura do escopo mais interno para o global.
- As funções ficam numa tabela à parte, preenchida numa **primeira passada** sobre o programa.
  Assim uma função pode chamar outra declarada mais abaixo, ou a si mesma (recursão).

Exemplo (`testdata/programs/escopo.go`):

```
NOME           CATEGORIA   TIPO                         ESCOPO  LINHA
incrementa     função      func()                       0       9
main           função      func()                       0       13
contador       variável    int                          0       6
mensagem       variável    string                       0       7
x              variável    int                          1       19
x              variável    int                          2       20
x              variável    int                          3       22
...
```

### 4.2 Regras de escopo

| Construção                         | Escopo criado                                         |
|------------------------------------|-------------------------------------------------------|
| programa                           | escopo 0: variáveis globais                           |
| `func f(params)`                   | escopo da função: parâmetros e variáveis do corpo     |
| `if init; cond { } else { }`       | escopo do `if` (contém `init`) + um escopo para cada ramo |
| `for init; cond; post { }`         | escopo do `for` (contém `init`) + escopo do corpo     |

Uma declaração só é visível depois dela. Em `x := x + 1` o `x` da direita é o de fora, porque o
inicializador é analisado antes da declaração.

### 4.3 Sistema de tipos

Tipos: `int`, `float64` (`float32` é aceito como sinônimo), `string` e `bool`, mais `void` para
funções sem retorno. A tipagem é **estática e forte**, sem conversões implícitas entre variáveis,
como em Go:

| Operador              | Operandos aceitos                       | Resultado |
|-----------------------|-----------------------------------------|-----------|
| `+`                   | int, float64, string (mesmo tipo dos dois lados) | o mesmo tipo |
| `- * /`               | int, float64 (mesmo tipo)               | o mesmo tipo |
| `%`                   | int                                     | int       |
| `== !=`               | qualquer tipo, igual dos dois lados     | bool      |
| `< <= > >=`           | int, float64, string                    | bool      |
| `&& \|\| !`           | bool                                    | bool      |
| `-` unário            | int, float64                            | o mesmo tipo |
| `int(x)`, `float64(x)`| int, float64                            | o tipo da conversão |

**Constantes não tipadas.** Em Go, `var f float64 = 1` e `f * 2` são válidos, porque um literal
inteiro é uma constante "não tipada". O analisador reproduz isso: quando uma expressão formada
só por literais inteiros é usada onde se espera `float64`, ele **reescreve a AST** (`2` → `2.0`,
ou `10/4` → `float64(10/4)`). Assim o código intermediário já sai com o tipo certo. Uma
*variável* `int` atribuída a `float64` continua sendo erro.

### 4.4 Verificações realizadas

**Erros**, que interrompem a compilação:

| Verificação | Exemplo |
|---|---|
| variável não declarada | `y = x + 1` → `variável 'y' não declarada` |
| declaração duplicada no mesmo escopo | `variável 'x' já declarada neste escopo` |
| função não declarada / duplicada | `função 'subtrai' não declarada` |
| tipo desconhecido | `var x inteiro` → `tipo desconhecido 'inteiro'` |
| tipos incompatíveis em atribuição/declaração | `x = "texto"` (x int) |
| operador inválido para os tipos | `"abc" * 2` → `operador '*' inválido entre 'string' e 'int'` |
| condição de `if`/`for` não booleana | `if n { … }` |
| nº de argumentos | `soma(1)` → `espera 2 argumento(s), mas recebeu 1` |
| tipo de argumento | `soma(1, "dois")` |
| função `void` usada como valor | `x := avisa("oi") + 1` |
| tipo de retorno | `return "texto"` em `func f() int` |
| `return` com valor em função `void` (e vice-versa) | |
| função com retorno que pode terminar sem `return` | análise de "instrução terminal" (`return`, `if/else` com os dois ramos terminando, `for {}` sem `break`) |
| `break`/`continue` fora de laço | |
| atribuição a função / função usada como variável | `soma = 3` |
| divisão por zero com divisor literal | `x / 0` |
| `func main` com parâmetros ou retorno | |

**Avisos**, que não interrompem: variável declarada e não usada, variável usada antes de receber
valor (vale o valor zero do tipo, como em Go) e código inalcançável depois de
`return`/`break`/`continue`.

---

## 5. Código intermediário

**Arquivos:** `ir/ir.go`, `ir/generator.go`, `ir/values.go`

### 5.1 Representação escolhida

**Código de três endereços (TAC) em quádruplas.** Escolhemos o TAC por três motivos:

- é linear, então as otimizações locais e de fluxo de dados (blocos básicos, *liveness*) ficam
  simples;
- é próximo de um assembly, então interpretar é direto;
- é a representação mostrada no exemplo do próprio enunciado.

```go
type Instr struct {
    Op       Op     // tipo da instrução
    Dst      string // destino
    Arg1     string // operando 1
    Arg2     string // operando 2
    Operator string // + - * / % == < && ! int float64 ...
    Label    string // rótulo / nome da função / tipo de print
    N        int    // nº de parâmetros (call/print)
    Line     int    // linha no fonte (erros de execução)
}
```

| Instrução                   | Significado                                   |
|-----------------------------|-----------------------------------------------|
| `x = y`                     | cópia                                         |
| `x = y op z`                | operação binária                              |
| `x = -y`, `x = !y`, `x = float64(y)` | operação unária / conversão          |
| `L1:`                       | rótulo                                        |
| `goto L1`                   | desvio incondicional                          |
| `ifFalse x goto L1` / `if x goto L1` | desvio condicional                   |
| `param x`                   | empilha argumento                             |
| `t = call f, n` / `call f, n` | chamada com n argumentos                    |
| `return` / `return x`       | retorno                                       |
| `print Println, n`          | `fmt.Println` com n argumentos                |

**Operandos:** constantes (`25`, `2.5`, `"texto"`, `true`), variáveis locais (`x`), globais
(`@contador`) e temporários (`t1`, `t2`…). Quando há *shadowing* dentro de uma função, a variável
interna é renomeada (`x`, `x.2`, `x.3`), porque o TAC não tem blocos aninhados.

### 5.2 Esquemas de tradução (exemplos reais)

**Expressão e declaração:** `x := 10*2 + 5`
```
t1 = 10 * 2
t2 = t1 + 5
x = t2
```

**Função recursiva com `if`:** `testdata/programs/fatorial.go`
```go
func fatorialRec(n int) int {
	if n <= 1 {
		return 1
	}
	return n * fatorialRec(n-1)
}
```
```
func fatorialRec(n):
    t1 = n <= 1
    ifFalse t1 goto L1
    return 1
L1:
    t2 = n - 1
    param t2
    t3 = call fatorialRec, 1
    t4 = n * t3
    return t4
```

**Laço `for` clássico:** o `post` fica num rótulo próprio (L2), que é o destino do `continue`.
```go
for i := 2; i <= n; i++ { resultado *= i }
```
```
    i = 2
L1:
    t1 = i <= n
    ifFalse t1 goto L3      ← L3 é o destino do 'break'
    t2 = resultado * i
    resultado = t2
L2:                         ← destino do 'continue'
    t3 = i + 1
    i = t3
    goto L1
L3:
```

**Curto-circuito:** em `b != 0 && a/b > 1`, a divisão só é avaliada se `b != 0`.
```
    t9 = b != 0
    t8 = t9
    ifFalse t8 goto L7
    t10 = a / b
    t11 = t10 > 1
    t8 = t11
L7:
```

**Variáveis globais:** a inicialização delas vai para a pseudo-função `init.globais`, que o
interpretador executa antes de `main`.

---

## 6. Otimização

**Arquivos:** `optimizer/optimizer.go`, `optimizer/local.go`, `optimizer/flow.go`

### 6.1 Técnicas implementadas

O otimizador recebe uma **cópia** do programa, para que o original continue disponível para a
comparação. Ele aplica os passes abaixo repetidamente, até que nenhum altere o código (ponto
fixo):

| # | Técnica | Antes | Depois |
|---|---------|-------|--------|
| 1 | Dobramento de constantes | `t1 = 10 * 2` | `t1 = 20` |
| 2 | Propagação de constantes | `x = 4; t = x * 5` | `t = 4 * 5` |
| 3 | Propagação de cópias | `f = c; t = f + e` | `t = c + e` |
| 4 | Simplificação algébrica | `x * 1`, `x + 0`, `x / 1`, `x * 0`, `s + ""` | `x`, `x`, `x`, `0`, `s` |
| 5 | Redução de força | `t = d * 2` | `t = d + d` |
| 6 | Eliminação de subexpressões comuns | `t1 = a + b; t2 = a + b` | `t1 = a + b; t2 = t1` |
| 7 | Desvios com condição constante | `ifFalse true goto L1` | (removido) |
| 8 | Eliminação de código inalcançável | instruções depois de `goto`/`return` | (removidas) |
| 9 | Otimização de desvios | `goto L1` seguido de `L1:`; `goto L1` onde `L1: goto L2` | removido; `goto L2` |
| 10 | Remoção de rótulos não utilizados | `L2:` sem desvio para ele | (removido) |
| 11 | Eliminação de código morto (atribuições mortas) | `naoUsada = area * 1000` | (removida) |
| 12 | Coalescência de cópias | `t3 = i + 1; i = t3` | `i = i + 1` |

**Passe local (1–7):** uma varredura por bloco básico, com três tabelas: constantes
conhecidas, cópias e expressões já calculadas. Quando uma variável é redefinida, tudo o que
depende dela é invalidado.

**Atribuições mortas (11):** usa **análise de vivacidade** (*liveness*) global. O código é
dividido em blocos básicos, monta-se o grafo de fluxo de controle e as equações de fluxo de dados
são resolvidas para trás até estabilizar:

```
out[B] = ∪ in[S]   para todo sucessor S de B
in[B]  = use[B] ∪ (out[B] − def[B])
```

Uma atribuição sem efeito colateral cujo destino não está vivo logo depois dela é removida.

### 6.2 Justificativa das técnicas

- Dobramento e propagação de constantes são as técnicas cobradas no exemplo do enunciado e as que
  mais reduzem o código gerado por expressões constantes.
- O gerador cria muitos temporários e cópias (`t2 = …; x = t2`). Propagação de cópias,
  coalescência e eliminação de código morto limpam esse "ruído" típico de um gerador simples.
- Subexpressões comuns e redução de força melhoram código com valores que só se conhecem em
  execução (parâmetros de função).
- As otimizações de desvio limpam o resultado das anteriores: depois de dobrar uma condição,
  sobram ramos e rótulos inúteis.

### 6.3 Exemplo antes/depois (`testdata/programs/otimizacao.go`)

```go
func calcula(a int, b int) int {
	c := (a + b) * (a + b)
	d := b*1 + 0
	e := d * 2
	f := c
	return f + e
}
```

| Antes (12 instruções) | Depois (5 instruções) |
|---|---|
| <pre>t1 = a + b<br>t2 = a + b<br>t3 = t1 * t2<br>c = t3<br>t4 = b * 1<br>t5 = t4 + 0<br>d = t5<br>t6 = d * 2<br>e = t6<br>f = c<br>t7 = f + e<br>return t7</pre> | <pre>t1 = a + b<br>t3 = t1 * t1<br>t6 = b + b<br>t7 = t3 + t6<br>return t7</pre> |

Transformações: `t2 = a + b` vira `t2 = t1` (subexpressão comum); `b*1+0` vira `b`
(simplificação algébrica); `d * 2` vira `b + b` (propagação + redução de força); `f` é substituído
por `c` (propagação de cópias); as cópias e os temporários que sobraram são eliminados (código
morto).

No `main` do mesmo programa, todo o cálculo com constantes (`x := 10*2 + 5`, `area := largura *
altura`, a condição `largura < altura`) é resolvido em tempo de compilação. O ramo `else`, que
nunca executa, é removido. Use `go run . -opt -explain testdata/programs/otimizacao.go` para ver
cada transformação.

### 6.4 Por que a otimização preserva o comportamento

1. **Mesma aritmética:** o dobramento de constantes usa `ir.EvalBinary`/`ir.EvalUnary`, as
   **mesmas funções** que o interpretador usa para executar. Uma constante calculada na compilação
   tem sempre o valor que teria em execução.
2. **Erros de execução preservados:** uma divisão inteira com divisor zero ou desconhecido nunca é
   dobrada nem removida, então o erro continua acontecendo no mesmo ponto.
3. **Propagação só no bloco básico:** as tabelas são esvaziadas em todo rótulo (onde caminhos se
   juntam) e em todo desvio. Assim um valor nunca é propagado para dentro de um laço onde a
   variável muda.
4. **Globais e chamadas:** depois de um `call`, tudo o que se sabia sobre variáveis globais é
   descartado, porque a função chamada pode alterá-las. Uma chamada nunca é removida, só o seu
   resultado, se não for usado.
5. **Código morto:** só se remove uma atribuição cujo valor não é lido em **nenhum** caminho,
   segundo a análise de vivacidade.
6. **Verificação empírica:** `go run . -compare arquivo.go` executa as duas versões e compara as
   saídas. A suíte de testes faz isso para todos os programas e ainda compara as duas saídas com
   a do Go oficial (seção 8).

| Programa | Instruções no código | Instruções executadas |
|---|---|---|
| curtocircuito.go | 44 → 19 | 36 → 21 |
| escopo.go | 42 → 33 | 57 → 45 |
| fatorial.go | 33 → 30 | 865 → 764 |
| fibonacci.go | 33 → 31 | 142 480 → 142 450 |
| mdc.go | 33 → 32 | 120 → 109 |
| otimizacao.go | 44 → 22 | 88 → 52 |
| precedencia.go | 66 → 18 | 57 → 18 |
| primos.go | 38 → 35 | 1 395 → 1 292 |
| series.go | 60 → 50 | 709 → 578 |
| tipos.go | 62 → 39 | 61 → 38 |

Em todos os casos a saída otimizada é **idêntica** à não otimizada e à do Go oficial.

---

## 7. Back-end: interpretador

**Arquivo:** `vm/vm.go`

### 7.1 Arquitetura-alvo e estratégia

O back-end é um **interpretador do código de três endereços** (uma máquina virtual). A
"arquitetura-alvo" é, portanto, a própria máquina abstrata do TAC. Escolhemos interpretar em vez
de gerar assembly porque:

- é portável (roda igual em Linux, macOS e Windows);
- a execução pode ser comparada instrução a instrução entre o código otimizado e o não otimizado,
  inclusive pelo número de instruções executadas;
- os erros de execução apontam a linha do código-fonte.

### 7.2 Formato do código executado

O mesmo `ir.Program` impresso em `codigo_otimizado.txt`. Antes de executar, cada função passa por
um pré-processamento: os rótulos viram índices de instrução (`map[label]int`) e as constantes são
convertidas uma vez só e guardadas em cache.

### 7.3 Processo de execução

1. Executa `init.globais` (inicialização das variáveis globais).
2. Chama `main`.
3. Cada chamada cria um **registro de ativação** com um `map` de variáveis locais e temporários.
   Os parâmetros recebem os valores acumulados pelas instruções `param`. A recursão usa a pilha
   do próprio Go, limitada a 10 000 níveis ("estouro de pilha").
4. O laço principal busca a instrução em `pc`, executa e avança. Os desvios alteram `pc`.
5. `print` usa o próprio pacote `fmt` do Go com os valores nativos (`int`, `float64`, `string`,
   `bool`), por isso a formatação é idêntica à de um programa Go compilado.

**Erros de execução:**
```
erro de execução linha 8 (função divide): divisão inteira por zero
```
Divisão de `float64` por zero resulta em `±Inf`, como em Go.

---

## 8. Testes

### 8.1 Como rodar

```sh
go test ./...                 # todos os testes
go test ./... -v              # detalhado (111 casos)
nix flake check               # compila e testa num ambiente isolado
```

### 8.2 Programas MicroGo de teste (`testdata/programs/`)

Cada programa tem um arquivo `.esperado` com a saída **gerada pelo compilador Go oficial**
(`go run`). O teste `TestProgramas` compila cada programa, executa as versões sem e com otimização
e exige que as duas saídas sejam iguais ao `.esperado`. Também exige que a otimização não aumente
o código nem o número de instruções executadas.

| Programa | Recursos exercitados |
|---|---|
| `fatorial.go` | recursão, `for` clássico, `*=`, chamadas aninhadas |
| `fibonacci.go` | recursão dupla, `fmt.Print` com vários argumentos |
| `mdc.go` | parâmetros agrupados `(a, b int)`, `for` estilo while, `%` |
| `primos.go` | `for {}` infinito, `break`, `continue`, `bool`, `!` |
| `tipos.go` | `float64`, conversões, `Printf`, concatenação, comparação de strings, `%` negativo |
| `escopo.go` | globais, *shadowing* em 3 níveis, `if` com inicialização |
| `curtocircuito.go` | `&&`/`\|\|` em curto-circuito com chamadas que têm efeito colateral |
| `series.go` | acumulador `float64`, laços aninhados, construção de string |
| `precedencia.go` | precedência e associatividade de todos os operadores |
| `otimizacao.go` | exercita todas as técnicas de otimização |

### 8.3 Testes de erro

| Arquivo | Etapa | Erros esperados |
|---|---|---|
| `lexer/fail/invalidos.go` | léxica | `@` (l.7), string não fechada (l.8), `$` (l.9), escape `\q` (l.10) |
| `parser/fail/brokenfor.go` | sintática | expressão solta como instrução (l.10) |
| `parser/fail/brokenifelse.go` | sintática | `else` na linha seguinte ao `}` (l.21) |
| `parser/fail/brokenvardecl.go` | sintática | declaração malformada (l.11) |
| `parser/fail/missingbraces.go` | sintática | `}` faltando no fim do arquivo |
| `semantic/fail/undeclared.go` | semântica | variável não declarada |
| `semantic/fail/duplicate.go` | semântica | declaração duplicada |
| `semantic/fail/typemismatch.go` | semântica | `int` recebendo `string` |
| `semantic/fail/divzero.go` | semântica | divisão por zero literal |
| `semantic/fail/funcoes.go` | semântica | nº/tipo de argumentos, função inexistente, `void` em expressão, atribuição a função |
| `semantic/fail/retorno.go` | semântica | `return` faltando, tipo de retorno, `return` com valor em `void` |
| `semantic/fail/condicao.go` | semântica | condição não `bool`, operadores inválidos, `int`→`float64`, `break` fora de laço |
| `semantic/fail/deadcode.go`, `unused.go`, `uninitialized.go` | — (avisos) | código inalcançável, variável não usada, uso antes de atribuir |
| `runtime/divzero.go` | execução | divisão por zero com divisor variável (l.8), com e sem otimização |

### 8.4 Testes unitários

- `lexer/scanner_test.go`: tokens, inserção de `;`, palavras reservadas, contagem de linhas em
  comentários, acúmulo de erros.
- `optimizer/optimizer_test.go`: cada técnica isolada sobre TAC montado à mão, incluindo os
  casos em que a otimização **não** pode ser aplicada (divisão por zero, propagação para dentro
  de laço, expressão com operando alterado, globais depois de `call`).
- `compiler/checks_test.go`: uma regra por caso, com trechos curtos — 9 erros sintáticos, 25
  semânticos, 12 programas válidos executados com e sem otimização (constantes não tipadas,
  valores zero, runes, atribuições compostas, `+Inf`, globais alteradas por chamadas…), erros de
  execução (estouro de pilha, `main` ausente, limite de instruções) e validade do JSON gerado.

**Resultado obtido:** os 111 casos passam; cobertura de 80% das instruções do compilador
(`go test ./... -coverpkg=./...`).


---

## 9. Decisões e dificuldades de desenvolvimento

| Problema | Solução adotada |
|---|---|
| Go dispensa `;` no fim das linhas | o lexer implementa a regra de inserção automática da especificação. Efeito colateral documentado: `}` + quebra de linha + `else` é erro, como no Go oficial |
| Um erro sintático gerava cascata de erros, ou o parser travava em laço infinito | modo pânico com sincronização em `;`/`}` e salvaguarda "se não consumiu token, avança" |
| As primeiras versões descartavam tipos declarados, parâmetros e argumentos de chamadas (chamadas viravam `VariableNode("f(...)")`) | AST refeita com `TypeName`, `Params`, `ReturnType` e `CallNode` com argumentos. Sem isso não seria possível gerar código para funções |
| O `for` clássico colocava `init` e `post` dentro do corpo, o que fazia `continue` pular o `post` | `ForNode` ganhou campos `Init`/`Post` separados, e o `post` tem um rótulo próprio, destino do `continue` |
| `&&` e `\|\|` tinham a mesma precedência, e relacionais não associavam (`1 < 2 == true` falhava): diferente do Go | duas camadas `parseOrExpr`/`parseAndExpr` e laço nos relacionais. Os dois bugs foram encontrados ao comparar a saída com a do Go oficial (`precedencia.go`) |
| `var f float64 = 1` é válido em Go, mas o interpretador faria divisão inteira | a análise semântica reconhece constantes não tipadas e reescreve o literal como `float64` na AST |
| *Shadowing* (`x` dentro de `if` escondendo outro `x`) num código intermediário sem blocos | o gerador renomeia (`x.2`, `x.3`) |
| Garantir que o otimizador não muda o resultado | otimizador e interpretador usam a mesma função de avaliação; a propagação é local ao bloco; testes comparam as três saídas (sem otimização, com otimização, Go oficial) |
| Os programas de teste são `.go` e o `go test ./...`/`go vet` tentava compilá-los como pacotes | a pasta `tests/` foi renomeada para `testdata/`, que o toolchain Go ignora |
| Ambiente de desenvolvimento reprodutível (NixOS) | `flake.nix` com *devShell*, *build*, *app* e *check* |

**Decisões técnicas relevantes:**

- **Sem geradores de analisadores (yacc/ANTLR):** lexer e parser escritos à mão, para que cada
  decisão seja explicável e as mensagens de erro fiquem sob nosso controle.
- **Interpretador como back-end:** ver seção 7.1.
- **Avisos × erros:** em Go, variável não usada é erro de compilação. Aqui é aviso, para não
  atrapalhar os exemplos didáticos.

---

## 10. Conclusão

### 10.1 O que foi implementado

Um compilador completo para um subconjunto de Go, cobrindo todas as etapas do enunciado:

- análise léxica com inserção automática de `;` e acúmulo de erros;
- análise sintática descendente recursiva com recuperação de erros;
- análise semântica com tabela de símbolos em escopos, sistema de tipos estático, verificação
  de funções e constantes não tipadas;
- código intermediário de três endereços;
- 12 técnicas de otimização, incluindo análise de vivacidade global;
- interpretador do código de três endereços.

O resultado é validado contra o compilador Go oficial.

### 10.2 Limitações conhecidas

- Tipos suportados: `int`, `float64`/`float32` (`float32` é tratado com precisão de `float64`),
  `string` e `bool`. Não há arrays, slices, maps, structs, ponteiros, `switch`, `const`, closures,
  múltiplos retornos nem declaração múltipla (`a, b := 1, 2`).
- O pacote `fmt` é o único suportado, e só com `Println`, `Print` e `Printf`. Não há entrada de
  dados (`fmt.Scan`).
- `import` é aceito, mas não verificado.
- Variável não usada é aviso (em Go é erro).
- As otimizações de propagação são locais ao bloco básico. A propagação global de constantes
  (*reaching definitions*) e as otimizações de laço não foram implementadas.

### 10.3 Melhorias futuras

- Geração de código para uma arquitetura real (assembly x86-64 ou WebAssembly) a partir do mesmo
  TAC.
- Otimizações globais: propagação de constantes entre blocos, movimentação de código invariante
  para fora de laços e *inlining* de funções pequenas.
- Arrays/slices e `fmt.Scan`, para permitir programas com entrada de dados.
- Forma SSA como segunda representação intermediária.
