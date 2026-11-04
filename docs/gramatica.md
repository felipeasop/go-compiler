# Gramática da linguagem (subconjunto de Go)

A linguagem compilada é um subconjunto da linguagem Go (a troca de MicroC por Go foi
aprovada pelo professor). Todo programa aceito pelo nosso compilador também é um programa Go
válido. A recíproca não vale: só o subconjunto descrito aqui é suportado.

Notação EBNF: `{ x }` repete zero ou mais vezes, `[ x ]` é opcional, `|` indica alternativa e
`"x"` é um terminal literal.

## 1. Estrutura léxica

```ebnf
letra        = "a" … "z" | "A" … "Z" | "_" | letra Unicode ;
digito       = "0" … "9" ;

IDENT        = letra { letra | digito } ;
INT          = digito { digito } ;
FLOAT        = digito { digito } "." { digito } ;
STRING       = '"' { caractere | escape } '"'        (* não pode conter quebra de linha *)
             | "`" { qualquer caractere exceto "`" } "`" ;   (* raw string *)
CHAR         = "'" ( caractere | escape ) "'" ;
escape       = "\" ( "n" | "t" | "r" | "\" | '"' | "'" ) ;

comentario   = "//" { qualquer caractere exceto quebra de linha }
             | "/*" { qualquer caractere } "*/" ;
```

**Palavras reservadas:** `break case const continue default defer else for func go if import
interface map package range return select struct switch type var`. As que estão fora do
subconjunto (`switch`, `struct`, `map`…) são reconhecidas pelo lexer e rejeitadas pelo parser
com a mensagem "não suportado nesta versão".

**Identificadores predeclarados** (não são palavras reservadas, assim como em Go): `true`,
`false`, `int`, `float64`, `float32`, `string`, `bool`.

**Operadores e delimitadores:**
`+ - * / %  += -= *= /= %=  ++ --  == != < <= > >=  && || !  = :=  ( ) { } , ; .`

**Inserção automática de `;` (regra do Go):** ao encontrar uma quebra de linha, o lexer insere
`;` se o último token foi um identificador, um literal, `break`, `continue`, `return`, `++`,
`--`, `)`, `]` ou `}`.

## 2. Estrutura sintática

```ebnf
Program      = [ "package" IDENT ";" ] { ImportDecl } { TopLevelDecl } ;
ImportDecl   = "import" ( STRING | "(" { STRING ";" } ")" ) ";" ;
TopLevelDecl = FuncDecl | VarDecl ;

(* ── Funções ─────────────────────────────────────────────── *)
FuncDecl     = "func" IDENT "(" [ ParamList ] ")" [ Type ] Block ";" ;
ParamList    = ParamGroup { "," ParamGroup } ;
ParamGroup   = IDENT { "," IDENT } Type ;            (* (a, b int) ou (a int, b int) *)
Type         = IDENT ;                                (* int | float64 | float32 | string | bool *)
Block        = "{" { Statement } "}" ;

(* ── Instruções ──────────────────────────────────────────── *)
Statement    = VarDecl
             | SimpleStmt ";"
             | PrintStmt
             | IfStmt
             | ForStmt
             | ReturnStmt
             | "break" ";"
             | "continue" ";" ;

VarDecl      = "var" IDENT [ Type ] [ "=" Expr ] ";" ;   (* exige tipo ou valor *)

SimpleStmt   = IDENT ":=" Expr                        (* declaração curta *)
             | IDENT "=" Expr                         (* atribuição; IDENT pode ser "_" *)
             | IDENT AssignOp Expr                    (* x += e  ≡  x = x + e *)
             | IDENT ( "++" | "--" )                  (* x++     ≡  x = x + 1 *)
             | CallExpr ;
AssignOp     = "+=" | "-=" | "*=" | "/=" | "%=" ;

PrintStmt    = "fmt" "." ( "Println" | "Print" | "Printf" ) "(" [ ExprList ] ")" ";" ;

IfStmt       = "if" [ SimpleStmt ";" ] Expr Block
               [ "else" ( IfStmt | Block ) ] ";" ;

ForStmt      = "for" [ ForClause | Expr ] Block ";" ;
ForClause    = [ SimpleStmt ] ";" [ Expr ] ";" [ SimpleStmt ] ;

ReturnStmt   = "return" [ Expr ] ";" ;

(* ── Expressões (precedência crescente) ─────────────────── *)
Expr         = OrExpr ;
OrExpr       = AndExpr { "||" AndExpr } ;
AndExpr      = RelExpr { "&&" RelExpr } ;
RelExpr      = AddExpr { ( "==" | "!=" | "<" | "<=" | ">" | ">=" ) AddExpr } ;
AddExpr      = MulExpr { ( "+" | "-" ) MulExpr } ;
MulExpr      = Unary { ( "*" | "/" | "%" ) Unary } ;
Unary        = ( "!" | "-" ) Unary | Factor ;
Factor       = INT | FLOAT | STRING | CHAR | "true" | "false"
             | IDENT
             | CallExpr
             | "(" Expr ")" ;
CallExpr     = IDENT "(" [ ExprList ] ")" ;           (* também int(x), float64(x) *)
ExprList     = Expr { "," Expr } ;
```

### Observações

- **Precedência:** a gramática tem uma camada por nível, igual ao Go: `||` < `&&` < relacionais <
  `+ -` < `* / %` < unários. Então `2 + 3 * 4` vale `14`, `a || b && c` é `a || (b && c)` e
  `10 - 3 - 2` vale `5` (associatividade à esquerda, dada pelas repetições `{ … }`).
- **Relacionais associam à esquerda:** `1 < 2 == true` é `(1 < 2) == true`, como em Go (já
  `a < b < c` com inteiros é rejeitado na análise semântica: compara `bool` com `int`).
- **`else` na mesma linha do `}`:** por causa da inserção automática de `;`, um `else` na linha
  seguinte gera erro de sintaxe, exatamente como no compilador Go oficial.
- **Ambiguidade de `for`:** o parser procura um `;` antes do `{` para decidir entre
  `for cond {` e `for init; cond; post {`.
