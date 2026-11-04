package compiler

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go-compiler/ir"
)

// =====================================================================
// TESTES DE INTEGRAÇÃO (fonte → execução)
// =====================================================================
// Os programas em testdata/programs são Go válido. A saída esperada
// (.esperado) foi gerada com o compilador oficial (go run), então estes
// testes comparam o nosso compilador com o Go de verdade.
// =====================================================================

const stepLimit = 50_000_000

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// TestProgramas compila e executa cada programa com e sem otimização e
// compara as duas saídas com a saída do compilador Go oficial.
func TestProgramas(t *testing.T) {
	files, _ := filepath.Glob("../testdata/programs/*.go")
	if len(files) == 0 {
		t.Fatal("nenhum programa de teste encontrado")
	}
	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			expected := readFile(t, strings.TrimSuffix(file, ".go")+".esperado")
			res := Compile(readFile(t, file))
			if !res.OK() {
				t.Fatalf("falhou na etapa %s: %v %v %v", res.FailedAt, res.LexErrors, res.ParseErrors, res.Semantic)
			}

			plain, stepsPlain, err := RunToString(res.IR, stepLimit)
			if err != nil {
				t.Fatalf("erro sem otimização: %v", err)
			}
			opt, stepsOpt, err := RunToString(res.Optimized, stepLimit)
			if err != nil {
				t.Fatalf("erro com otimização: %v", err)
			}

			if plain != expected {
				t.Errorf("saída SEM otimização difere do Go oficial:\n--- obtido\n%s--- esperado\n%s", plain, expected)
			}
			if opt != expected {
				t.Errorf("saída COM otimização difere do Go oficial:\n--- obtido\n%s--- esperado\n%s", opt, expected)
			}
			if res.Optimized.InstrCount() > res.IR.InstrCount() {
				t.Errorf("otimização aumentou o código: %d → %d", res.IR.InstrCount(), res.Optimized.InstrCount())
			}
			if stepsOpt > stepsPlain {
				t.Errorf("otimização aumentou as instruções executadas: %d → %d", stepsPlain, stepsOpt)
			}
		})
	}
}

// TestProgramasValidos garante que os exemplos "pass" passam por todas
// as etapas de análise sem erros.
func TestProgramasValidos(t *testing.T) {
	var files []string
	for _, pattern := range []string{"../testdata/parser/pass/*.go", "../testdata/semantic/pass/*.go", "../testdata/lexer/*.go"} {
		m, _ := filepath.Glob(pattern)
		files = append(files, m...)
	}
	for _, file := range files {
		t.Run(file, func(t *testing.T) {
			if res := Compile(readFile(t, file)); !res.OK() {
				t.Fatalf("falhou na etapa %s: %v %v %v", res.FailedAt, res.LexErrors, res.ParseErrors, res.Semantic)
			}
		})
	}
}

// TestErros verifica que cada programa inválido é rejeitado na etapa
// correta, com as mensagens (e linhas) esperadas.
func TestErros(t *testing.T) {
	cases := []struct {
		file  string
		stage Stage
		want  []string // trechos que devem aparecer nas mensagens
	}{
		{"lexer/fail/invalidos.go", StageLexer, []string{
			"linha 7: caractere invalido '@'", "linha 8: string nao fechada",
			"linha 9: caractere invalido '$'", "linha 10: sequencia de escape invalida"}},
		{"parser/fail/brokenfor.go", StageParser, []string{"linha 10"}},
		{"parser/fail/brokenifelse.go", StageParser, []string{"linha 21", "else"}},
		{"parser/fail/brokenvardecl.go", StageParser, []string{"linha 11"}},
		{"parser/fail/missingbraces.go", StageParser, []string{"esperava }"}},
		{"semantic/fail/divzero.go", StageSemantic, []string{"Linha 8", "divisão por zero"}},
		{"semantic/fail/duplicate.go", StageSemantic, []string{"Linha 5", "'x' já declarada"}},
		{"semantic/fail/typemismatch.go", StageSemantic, []string{"Linha 5", "tipos incompatíveis"}},
		{"semantic/fail/undeclared.go", StageSemantic, []string{"Linha 5", "'y' não declarada"}},
		{"semantic/fail/funcoes.go", StageSemantic, []string{
			"espera 2 argumento(s), mas recebeu 1", "parâmetro 'b' de 'soma'",
			"função 'subtrai' não declarada", "'avisa' não retorna valor", "atribuir valor à função 'soma'"}},
		{"semantic/fail/retorno.go", StageSemantic, []string{
			"pode terminar sem 'return'", "retorno de 'tipoErrado'", "'vazia' não retorna valor"}},
		{"semantic/fail/condicao.go", StageSemantic, []string{
			"condição do if deve ser 'bool'", "operador '*' inválido entre 'string' e 'int'",
			"operador '&&' inválido", "'f' é do tipo 'float64'", "'break' fora de um laço"}},
	}

	for _, c := range cases {
		t.Run(c.file, func(t *testing.T) {
			res := Compile(readFile(t, "../testdata/"+c.file))
			if res.FailedAt != c.stage {
				t.Fatalf("esperava falha na etapa %q, obteve %q", c.stage, res.FailedAt)
			}
			var msgs []string
			for _, e := range append(res.LexErrors, res.ParseErrors...) {
				msgs = append(msgs, e.Error())
			}
			if res.Semantic != nil {
				msgs = append(msgs, res.Semantic.String())
			}
			all := strings.Join(msgs, "\n")
			for _, w := range c.want {
				if !strings.Contains(all, w) {
					t.Errorf("mensagem esperada não encontrada: %q\nmensagens:\n%s", w, all)
				}
			}
		})
	}
}

// TestAvisos: programas válidos, mas com avisos semânticos.
func TestAvisos(t *testing.T) {
	cases := map[string]string{
		"semantic/fail/deadcode.go":      "código inalcançável",
		"semantic/fail/uninitialized.go": "utilizada antes de receber valor",
		"semantic/fail/unused.go":        "declarada mas nunca utilizada",
	}
	for file, want := range cases {
		t.Run(file, func(t *testing.T) {
			res := Compile(readFile(t, "../testdata/"+file))
			if !res.OK() {
				t.Fatalf("não deveria falhar: %v", res.Semantic)
			}
			if !strings.Contains(res.Semantic.String(), want) {
				t.Errorf("aviso %q não encontrado em:\n%s", want, res.Semantic)
			}
		})
	}
}

// TestErroDeExecucao: divisão por zero com divisor variável só é
// detectada pelo interpretador — com e sem otimização.
func TestErroDeExecucao(t *testing.T) {
	res := Compile(readFile(t, "../testdata/runtime/divzero.go"))
	if !res.OK() {
		t.Fatal("deveria compilar")
	}
	for name, prog := range map[string]*ir.Program{"sem otimização": res.IR, "com otimização": res.Optimized} {
		out, _, err := RunToString(prog, stepLimit)
		if !errors.Is(err, ir.ErrDivZero) {
			t.Errorf("%s: esperava divisão por zero, obteve %v", name, err)
		}
		if !strings.Contains(err.Error(), "linha 8") {
			t.Errorf("%s: erro sem a linha correta: %v", name, err)
		}
		if want := "12 / 3 = 4\n12 / 2 = 6\n12 / 1 = 12\n"; out != want {
			t.Errorf("%s: saída antes do erro = %q, esperava %q", name, out, want)
		}
	}
}

// TestExemploEnunciado: o exemplo do enunciado (x = 10 * 2 + 5) deve ser
// reduzido a uma única constante.
func TestExemploEnunciado(t *testing.T) {
	res := Compile("package main\nimport \"fmt\"\nfunc main() {\n\tx := 10*2 + 5\n\tfmt.Println(x)\n}\n")
	before := res.IR.Func("main").String()
	after := res.Optimized.Func("main").String()
	for _, want := range []string{"t1 = 10 * 2", "t2 = t1 + 5", "x = t2"} {
		if !strings.Contains(before, want) {
			t.Errorf("código intermediário deveria conter %q:\n%s", want, before)
		}
	}
	if !strings.Contains(after, "param 25") || strings.Contains(after, "*") {
		t.Errorf("código otimizado deveria usar a constante 25:\n%s", after)
	}
}
