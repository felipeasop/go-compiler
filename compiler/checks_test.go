package compiler

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"go-compiler/ir"
	"go-compiler/semantic"
	"go-compiler/vm"
)

// =====================================================================
// TESTES POR REGRA (trechos curtos de código)
// =====================================================================

// program monta um programa completo a partir do corpo de main e de
// declarações extras.
func program(decls, body string) string {
	return "package main\nimport \"fmt\"\n" + decls + "\nfunc main() {\n" + body + "\n}\n"
}

func messages(r *Result) string {
	var msgs []string
	for _, e := range append(r.LexErrors, r.ParseErrors...) {
		msgs = append(msgs, e.Error())
	}
	if r.Semantic != nil {
		msgs = append(msgs, r.Semantic.String())
	}
	return strings.Join(msgs, "\n")
}

// TestRegrasSintaticas: cada trecho deve falhar na análise sintática.
func TestRegrasSintaticas(t *testing.T) {
	cases := []struct{ name, decls, body, want string }{
		{"parametro sem tipo", "func f(a) {}", "", "parametro 'a' sem tipo"},
		{"var sem tipo nem valor", "", "var x", "precisa de tipo ou valor inicial"},
		{"var com '=' sem valor", "", "var x int =", "expressao esperada"},
		{"fmt nao suportado", "", "fmt.Scan(x)", "fmt.Scan nao suportado"},
		{"switch nao suportado", "", "switch {}", "nao suportado"},
		{"atribuicao como condicao", "", "x := 1\nif x = 2 {\n}", "atribuicao usada como condicao"},
		{"expressao solta", "", "x := 1\nx + 1", "expressao avaliada mas nao utilizada"},
		{"struct no pacote", "type P struct {}", "", "nao suportado"},
		{"falta parentese", "", "fmt.Println((1 + 2)", "esperava )"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := Compile(program(c.decls, c.body))
			if r.FailedAt != StageParser {
				t.Fatalf("esperava erro sintático, etapa = %q\n%s", r.FailedAt, messages(r))
			}
			if !strings.Contains(messages(r), c.want) {
				t.Errorf("esperava %q em:\n%s", c.want, messages(r))
			}
		})
	}
}

// TestRegrasSemanticas: cada trecho deve falhar na análise semântica
// com a mensagem indicada.
func TestRegrasSemanticas(t *testing.T) {
	cases := []struct{ name, decls, body, want string }{
		{"tipo desconhecido", "", "var x inteiro = 1\n_ = x", "tipo desconhecido 'inteiro'"},
		{"tipo de parametro desconhecido", "func f(a texto) {}", "", "tipo desconhecido 'texto'"},
		{"funcao duplicada", "func f() {}\nfunc f() {}", "", "função 'f' já declarada"},
		{"return sem valor", "func f() int {\n\treturn\n}", "", "deve retornar um valor do tipo 'int'"},
		{"if sem else nao termina", "func f(n int) int {\n\tif n > 0 {\n\t\treturn 1\n\t}\n}", "", "pode terminar sem 'return'"},
		{"for com break nao termina", "func f() int {\n\tfor {\n\t\tbreak\n\t}\n}", "", "pode terminar sem 'return'"},
		{"menos em string", "", `s := -"a"` + "\n_ = s", "operador unário '-' inválido para o tipo 'string'"},
		{"not em int", "", "b := !1\n_ = b", "operador unário '!' inválido para o tipo 'int'"},
		{"resto com float", "", "f := 2.5 % 2.0\n_ = f", "operador '%' inválido"},
		{"comparar tipos diferentes", "", `b := 1 == "1"` + "\n_ = b", "operador '==' inválido"},
		{"variavel chamada como funcao", "", "x := 1\nx()", "'x' é uma variável, não uma função"},
		{"funcao usada como valor", "func f() int { return 1 }", "x := f\n_ = x", "função 'f' usada como valor"},
		{"conversao de string", "", `x := int("10")` + "\n_ = x", "não é possível converter 'string' para int"},
		{"conversao sem argumento", "", "x := float64()\n_ = x", "exige exatamente 1 argumento"},
		{"printf sem formato", "", "fmt.Printf()", "exige ao menos o argumento de formato"},
		{"printf formato nao string", "", "fmt.Printf(1)", "deve ser string"},
		{"void no println", "func f() {}", "fmt.Println(f())", "não retorna valor"},
		{"void atribuido a _", "func f() {}", "_ = f()", "não retorna valor"},
		{"void inicializa var", "func f() {}", "x := f()\n_ = x", "não retorna valor"},
		{"continue fora de laco", "", "continue", "'continue' fora de um laço"},
		{"float literal em int", "", "var x int = 2.5\n_ = x", "'x' é do tipo 'int'"},
		{"variavel int em float", "", "n := 1\nvar f float64 = n\n_ = f", "'f' é do tipo 'float64'"},
		{"divisao por zero resto", "", "x := 10 % 0\n_ = x", "divisão por zero"},
		{"atribuir a nao declarada", "", "z = 1", "variável 'z' não declarada"},
		{"main com parametro", "MAIN", "", "func main não pode ter parâmetros"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			src := program(c.decls, c.body)
			if c.decls == "MAIN" {
				src = "package main\nfunc main(x int) {}\n"
			}
			r := Compile(src)
			if r.FailedAt != StageSemantic {
				t.Fatalf("esperava erro semântico, etapa = %q\n%s", r.FailedAt, messages(r))
			}
			if !strings.Contains(messages(r), c.want) {
				t.Errorf("esperava %q em:\n%s", c.want, messages(r))
			}
		})
	}
}

// TestProgramasAceitos: construções válidas e o resultado da execução.
func TestProgramasAceitos(t *testing.T) {
	cases := []struct{ name, decls, body, want string }{
		{"if/else com return termina", "func sinal(n int) int {\n\tif n < 0 {\n\t\treturn -1\n\t} else {\n\t\treturn 1\n\t}\n}",
			"fmt.Println(sinal(-3), sinal(3))", "-1 1\n"},
		{"for infinito termina", "func f() int {\n\tfor {\n\t\treturn 7\n\t}\n}", "fmt.Println(f())", "7\n"},
		{"constante inteira vira float", "", "var f float64 = 10 / 4\ng := f * 2\nfmt.Println(f, g, 1 + 2.5)", "2 4 3.5\n"},
		{"zero values", "", "var a int\nvar b float64\nvar c string\nvar d bool\nfmt.Println(a, b, c == \"\", d)", "0 0 true false\n"},
		{"rune", "", "c := 'A' + 1\nfmt.Println(c)", "66\n"},
		{"raw string", "", "fmt.Println(`a\\nb`)", "a\\nb\n"},
		{"atribuicoes compostas", "", "x := 10\nx -= 3\nx *= 2\nx /= 4\nx %= 2\nx--\nfmt.Println(x)", "0\n"},
		{"divisao float por zero", "", "z := 0.0\nfmt.Println(1.0 / z)", "+Inf\n"},
		{"Print sem espaco entre strings", "", `fmt.Print("a", "b", 1, 2, "\n")`, "ab1 2\n"},
		{"if com init", "func f() int { return 5 }", "if v := f(); v > 3 {\n\tfmt.Println(\"maior\", v)\n}", "maior 5\n"},
		{"variavel chamada t1", "", "t1 := 3\nt2 := t1 * 2\nfmt.Println(t1, t2)", "3 6\n"},
		{"global modificada por funcao", "var g int = 1\nfunc dobra() {\n\tg = g * 2\n}",
			"x := g\ndobra()\ny := g\nfmt.Println(x, y)", "1 2\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := Compile(program(c.decls, c.body))
			if !r.OK() {
				t.Fatalf("deveria compilar:\n%s", messages(r))
			}
			for name, prog := range map[string]*ir.Program{"sem otimização": r.IR, "com otimização": r.Optimized} {
				out, _, err := RunToString(prog, stepLimit)
				if err != nil {
					t.Fatalf("%s: %v", name, err)
				}
				if out != c.want {
					t.Errorf("%s: saída %q, esperava %q", name, out, c.want)
				}
			}
		})
	}
}

// TestErrosDeExecucaoDiversos: estouro de pilha, programa sem main e
// limite de instruções.
func TestErrosDeExecucaoDiversos(t *testing.T) {
	r := Compile(program("func infinita(n int) int {\n\treturn infinita(n + 1)\n}", "fmt.Println(infinita(0))"))
	if _, _, err := RunToString(r.IR, 0); err == nil || !strings.Contains(err.Error(), "estouro de pilha") {
		t.Errorf("esperava estouro de pilha, obteve %v", err)
	}

	r = Compile("package main\nfunc f() {}\n")
	if r.HasMain() {
		t.Error("programa sem main não deveria ter main")
	}
	if _, _, err := RunToString(r.IR, 0); err == nil || !strings.Contains(err.Error(), "não possui func main") {
		t.Errorf("esperava erro de main ausente, obteve %v", err)
	}

	r = Compile(program("", "for {\n}"))
	if _, _, err := RunToString(r.IR, 1000); !errors.Is(err, vm.ErrStepLimit) {
		t.Errorf("esperava limite de instruções, obteve %v", err)
	}
}

// TestArtefatosJSON: os arquivos JSON gravados em outputs/ são válidos.
func TestArtefatosJSON(t *testing.T) {
	r := Compile(readFile(t, "../testdata/programs/escopo.go"))
	tokens, _ := json.Marshal(r.Tokens)
	artefatos := map[string]string{
		"ast.json":             r.AST.ToJSON(0),
		"tabela_simbolos.json": semantic.SymbolsToJSON(r.Symbols, 0),
		"tokens.json":          string(tokens),
	}
	for name, content := range artefatos {
		if !json.Valid([]byte(content)) {
			t.Errorf("%s não é JSON válido:\n%s", name, content)
		}
	}

	var syms []map[string]interface{}
	json.Unmarshal([]byte(artefatos["tabela_simbolos.json"]), &syms)
	if len(syms) != 10 || syms[0]["categoria"] != "função" {
		t.Errorf("tabela de símbolos inesperada: %v", syms)
	}
	if table := semantic.SymbolsTable(r.Symbols); !strings.Contains(table, "func()") {
		t.Errorf("tabela de texto sem assinatura de função:\n%s", table)
	}
}

// TestASTDeTodosOsProgramas: todo programa de teste gera AST em JSON válido
// (cobre ToJSON de todos os tipos de nó).
func TestASTDeTodosOsProgramas(t *testing.T) {
	for _, file := range []string{"fatorial", "tipos", "escopo", "primos", "curtocircuito", "series"} {
		r := Compile(readFile(t, "../testdata/programs/"+file+".go"))
		if js := r.AST.ToJSON(0); !json.Valid([]byte(js)) {
			t.Errorf("%s: ast.json inválido", file)
		}
	}
}
