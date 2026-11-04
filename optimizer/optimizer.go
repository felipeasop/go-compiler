package optimizer

import (
	"fmt"
	"strings"

	"go-compiler/ir"
)

// =====================================================================
// OTIMIZADOR DE CÓDIGO INTERMEDIÁRIO
// =====================================================================
// Aplica, sobre o código de três endereços, um conjunto de passes que
// são repetidos até que nenhum deles altere mais o código (ponto fixo):
//
//   1. Passe local (dentro de cada bloco básico, em uma única varredura):
//        - propagação de constantes     t1 = 20; t2 = t1 + 5 → t2 = 20 + 5
//        - propagação de cópias         t3 = t2; x = t3      → x = t2
//        - dobramento de constantes     t2 = 20 + 5          → t2 = 25
//        - simplificação algébrica      x + 0, x * 1, x / 1  → x
//        - redução de força             x * 2                → x + x
//        - subexpressões comuns         t5 = a * b (repetida) → t5 = t3
//        - desvios com condição constante  ifFalse true goto L → (removido)
//        - instruções redundantes       x = x                → (removida)
//   2. Eliminação de código inalcançável (após goto/return).
//   3. Otimização de desvios: goto para a instrução seguinte, desvio
//      para outro desvio (jump threading) e rótulos não utilizados.
//   4. Eliminação de atribuições mortas com análise de vivacidade
//      (liveness) global sobre o grafo de fluxo de controle.
//   5. Coalescência de cópias: t3 = i + 1; i = t3  →  i = i + 1.
//
// Por que a semântica é preservada:
//   - O dobramento usa ir.EvalBinary/EvalUnary — as MESMAS funções que o
//     interpretador usa —, então o valor calculado é idêntico ao que
//     seria calculado em tempo de execução.
//   - Operações que podem falhar em execução (divisão inteira por zero)
//     nunca são dobradas nem removidas.
//   - As propagações são locais ao bloco básico: o conhecimento é
//     descartado em todo rótulo (ponto de junção de fluxos) e desvio, e
//     variáveis globais são esquecidas a cada 'call' (a função chamada
//     pode alterá-las).
//   - Só se remove uma atribuição cujo valor nunca é lido em nenhum
//     caminho de execução, e que não tem efeitos colaterais.
// =====================================================================

// Technique identifica uma técnica de otimização (para o relatório).
type Technique string

const (
	ConstFolding  Technique = "Dobramento de constantes (constant folding)"
	ConstProp     Technique = "Propagação de constantes"
	CopyProp      Technique = "Propagação de cópias"
	Algebraic     Technique = "Simplificação algébrica"
	Strength      Technique = "Redução de força"
	CSE           Technique = "Eliminação de subexpressões comuns"
	ConstBranch   Technique = "Desvios com condição constante"
	Redundant     Technique = "Eliminação de instruções redundantes"
	Unreachable   Technique = "Eliminação de código inalcançável"
	JumpOpt       Technique = "Otimização de desvios"
	UnusedLabels  Technique = "Remoção de rótulos não utilizados"
	DeadStore     Technique = "Eliminação de código morto (atribuições mortas)"
	TempElim      Technique = "Eliminação de temporários (coalescência de cópias)"
	maxIterations           = 50
)

var techniqueOrder = []Technique{ConstFolding, ConstProp, CopyProp, Algebraic, Strength, CSE,
	ConstBranch, Redundant, Unreachable, JumpOpt, UnusedLabels, DeadStore, TempElim}

// Report registra o que o otimizador fez.
type Report struct {
	Counts map[Technique]int
	Log    []string // cada transformação aplicada, em ordem
	Before int      // nº de instruções antes
	After  int      // nº de instruções depois
	Passes int      // iterações até o ponto fixo
}

func (r *Report) record(t Technique, fn string, format string, args ...interface{}) {
	r.Counts[t]++
	r.Log = append(r.Log, fmt.Sprintf("[%s] %s: %s", fn, t, fmt.Sprintf(format, args...)))
}

// Summary devolve um resumo das técnicas aplicadas.
func (r *Report) Summary() string {
	var sb strings.Builder
	for _, t := range techniqueOrder {
		if n := r.Counts[t]; n > 0 {
			sb.WriteString(fmt.Sprintf("  %-50s %3d\n", t, n))
		}
	}
	if sb.Len() == 0 {
		sb.WriteString("  (nenhuma otimização aplicável)\n")
	}
	reduction := 0.0
	if r.Before > 0 {
		reduction = 100 * float64(r.Before-r.After) / float64(r.Before)
	}
	sb.WriteString(fmt.Sprintf("  Instruções: %d → %d (%.1f%% a menos), %d iteração(ões) até o ponto fixo\n",
		r.Before, r.After, reduction, r.Passes))
	return sb.String()
}

// Optimize devolve uma cópia otimizada do programa e o relatório.
// O programa original não é modificado.
func Optimize(prog *ir.Program) (*ir.Program, *Report) {
	out := prog.Clone()
	rep := &Report{Counts: map[Technique]int{}, Before: prog.InstrCount()}

	for _, fn := range out.AllFunctions() {
		for i := 0; i < maxIterations; i++ {
			changed := false
			for _, pass := range []func(*ir.Function, *Report) bool{
				localPass, removeUnreachable, optimizeJumps, removeDeadStores, coalesceCopies,
			} {
				if pass(fn, rep) {
					changed = true
				}
			}
			if i+1 > rep.Passes {
				rep.Passes = i + 1
			}
			if !changed {
				break
			}
		}
	}
	rep.After = out.InstrCount()
	return out, rep
}
