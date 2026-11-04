package optimizer

import (
	"strings"
	"testing"

	"go-compiler/ir"
)

// Testes unitários de cada técnica, sobre código de três endereços
// montado à mão.

func bin(dst, a, op, b string) ir.Instr {
	return ir.Instr{Op: ir.OpBinary, Dst: dst, Arg1: a, Operator: op, Arg2: b}
}
func assign(dst, src string) ir.Instr { return ir.Instr{Op: ir.OpAssign, Dst: dst, Arg1: src} }
func ret(v string) ir.Instr           { return ir.Instr{Op: ir.OpReturn, Arg1: v} }
func label(l string) ir.Instr         { return ir.Instr{Op: ir.OpLabel, Label: l} }
func gotoL(l string) ir.Instr         { return ir.Instr{Op: ir.OpGoto, Label: l} }
func ifFalse(c, l string) ir.Instr    { return ir.Instr{Op: ir.OpIfFalse, Arg1: c, Label: l} }

func optimize(t *testing.T, params []string, code ...ir.Instr) (string, *Report) {
	t.Helper()
	prog := &ir.Program{Functions: []*ir.Function{{Name: "f", Params: params, Code: code}}}
	out, rep := Optimize(prog)
	return out.Functions[0].String(), rep
}

func expectCode(t *testing.T, got string, want ...string) {
	t.Helper()
	body := strings.TrimSpace(strings.SplitN(got, "\n", 2)[1])
	var lines []string
	for _, l := range strings.Split(body, "\n") {
		lines = append(lines, strings.TrimSpace(l))
	}
	if strings.Join(lines, "\n") != strings.Join(want, "\n") {
		t.Errorf("código otimizado:\n%s\nesperado:\n%s", strings.Join(lines, "\n"), strings.Join(want, "\n"))
	}
}

// Exemplo do enunciado: t1 = 10*2; t2 = t1+5; t3 = t2; x = t3 → x = 25
func TestEnunciado(t *testing.T) {
	got, rep := optimize(t, nil,
		bin("t1", "10", "*", "2"),
		bin("t2", "t1", "+", "5"),
		assign("t3", "t2"),
		assign("x", "t3"),
		ret("x"),
	)
	expectCode(t, got, "return 25")
	if rep.Counts[ConstFolding] != 2 {
		t.Errorf("esperava 2 dobramentos, obteve %d", rep.Counts[ConstFolding])
	}
}

func TestSubexpressaoComum(t *testing.T) {
	got, rep := optimize(t, []string{"a", "b"},
		bin("t1", "a", "+", "b"),
		bin("t2", "a", "+", "b"),
		bin("t3", "t1", "*", "t2"),
		ret("t3"),
	)
	expectCode(t, got, "t1 = a + b", "t3 = t1 * t1", "return t3")
	if rep.Counts[CSE] != 1 {
		t.Errorf("esperava 1 CSE, obteve %d", rep.Counts[CSE])
	}
}

// Se um operando muda entre as duas ocorrências, não é a mesma expressão.
func TestSubexpressaoInvalidada(t *testing.T) {
	got, _ := optimize(t, []string{"a", "b"},
		bin("t1", "a", "+", "b"),
		bin("a", "a", "+", "t1"),
		bin("t2", "a", "+", "b"),
		bin("t3", "t1", "*", "t2"),
		ret("t3"),
	)
	expectCode(t, got, "t1 = a + b", "a = a + t1", "t2 = a + b", "t3 = t1 * t2", "return t3")
}

func TestSimplificacaoAlgebrica(t *testing.T) {
	got, _ := optimize(t, []string{"x"},
		bin("t1", "x", "*", "1"),
		bin("t2", "t1", "+", "0"),
		bin("t3", "t2", "*", "2"),
		ret("t3"),
	)
	expectCode(t, got, "t3 = x + x", "return t3")
}

// Divisão por zero nunca é dobrada nem removida: o erro tem de acontecer
// em tempo de execução, como no programa original.
func TestDivisaoPorZeroPreservada(t *testing.T) {
	got, _ := optimize(t, []string{"x"},
		bin("t1", "x", "/", "0"),
		bin("t2", "10", "/", "0"),
		ret("x"),
	)
	expectCode(t, got, "t1 = x / 0", "t2 = 10 / 0", "return x")
}

func TestDesvioConstanteECodigoInalcancavel(t *testing.T) {
	got, _ := optimize(t, nil,
		bin("t1", "4", "<", "5"),
		ifFalse("t1", "L1"),
		assign("x", "1"),
		gotoL("L2"),
		label("L1"),
		assign("x", "2"),
		label("L2"),
		ret("x"),
	)
	expectCode(t, got, "return 1")
}

// Propagação só dentro do bloco básico: num laço, o valor de i muda a
// cada volta, então "i = 0" não pode ser propagado para dentro do laço.
func TestPropagacaoNaoAtravessaLaco(t *testing.T) {
	got, _ := optimize(t, nil,
		assign("i", "0"),
		label("L1"),
		bin("t1", "i", "<", "10"),
		ifFalse("t1", "L2"),
		bin("i", "i", "+", "1"),
		gotoL("L1"),
		label("L2"),
		ret("i"),
	)
	expectCode(t, got,
		"i = 0", "L1:", "t1 = i < 10", "ifFalse t1 goto L2", "i = i + 1", "goto L1", "L2:", "return i")
}

func TestAtribuicaoMortaComLiveness(t *testing.T) {
	got, rep := optimize(t, []string{"a"},
		bin("x", "a", "*", "a"), // morta: x é sobrescrito antes de ser lido
		bin("x", "a", "+", "a"),
		ret("x"),
	)
	expectCode(t, got, "x = a + a", "return x")
	if rep.Counts[DeadStore] == 0 {
		t.Error("esperava eliminação de atribuição morta")
	}
}

// O resultado de uma chamada pode ser descartado, mas a chamada fica
// (ela pode imprimir ou alterar variáveis globais).
func TestChamadaNaoEhRemovida(t *testing.T) {
	got, _ := optimize(t, nil,
		ir.Instr{Op: ir.OpCall, Dst: "t1", Label: "g", N: 0},
		ret(""),
	)
	expectCode(t, got, "call g, 0", "return")
}

// Valores de variáveis globais são esquecidos após uma chamada.
func TestGlobaisInvalidadasPorChamada(t *testing.T) {
	got, _ := optimize(t, nil,
		assign("@g", "1"),
		ir.Instr{Op: ir.OpCall, Label: "muda", N: 0},
		bin("t1", "@g", "+", "1"),
		ret("t1"),
	)
	expectCode(t, got, "@g = 1", "call muda, 0", "t1 = @g + 1", "return t1")
}

func TestCoalescenciaDeCopias(t *testing.T) {
	got, rep := optimize(t, []string{"n"},
		assign("i", "0"),
		label("L1"),
		bin("t1", "i", "<", "n"),
		ifFalse("t1", "L2"),
		bin("t3", "i", "+", "1"),
		assign("i", "t3"),
		gotoL("L1"),
		label("L2"),
		ret("i"),
	)
	expectCode(t, got, "i = 0", "L1:", "t1 = i < n", "ifFalse t1 goto L2", "i = i + 1", "goto L1", "L2:", "return i")
	if rep.Counts[TempElim] != 1 {
		t.Errorf("esperava 1 coalescência, obteve %d", rep.Counts[TempElim])
	}
}
