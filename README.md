# Scanner Léxico de Go

Analisador léxico escrito em C++ para uma linguagem inspirada em Go. Lê um código-fonte e transforma o texto em uma sequência de tokens identificados.

## Como compilar e rodar

```bash
g++ -o main main.cpp
./main
```

## Tokens reconhecidos

| Categoria | Tokens |
|---|---|
| Estrutura | `package`, `import`, `func`, `main` |
| Tipos | `var`, `int`, `float`, `bool`, `string` |
| Controle | `if`, `else`, `for` |
| Operadores | `+` `-` `*` `/` `=` `:=` `==` `<` `>` |
| Delimitadores | `(` `)` `{` `}` `;` `:` |
| Literais | números inteiros, floats, strings com `"` |
| Comentários | `//` linha e `/* bloco */` |

## Trocando o teste

No final do `main()`, descomente o teste que quiser rodar:

```cpp
string code = code_teste1;   // tipos, aritmética, if/else
// string code = code_teste2; // strings, for, comentários
// string code = code_teste3; // underscores, aninhamento, todos os operadores
```

## Saída esperada

```
T_PACKAGE -> package (linha 2)
T_MAIN -> main (linha 2)
T_FUNC -> func (linha 6)
T_VAR -> var (linha 7)
T_ID -> x (linha 7)
T_INT -> int (linha 7)
...
Fim da analise lexica.
```

## Erros léxicos tratados

- Caractere inválido → `Erro Lexico: caractere invalido 'X' na linha N`
- String não fechada → `Erro lexico: string nao fechada na linha N`
- Float inválido (ex: `1.2.3`) → `Erro lexico: float invalido`
- Comentário de bloco não fechado → `Erro lexico: comentario de bloco nao fechado na linha N`
