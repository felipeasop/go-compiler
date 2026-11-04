package optimizer

import (
	"go-compiler/ir"
)

// =====================================================================
// PASSES DE FLUXO DE CONTROLE
// =====================================================================

// removeUnreachable remove instruções entre um goto/return e o próximo
// rótulo: nenhum caminho de execução chega até elas.
func removeUnreachable(fn *ir.Function, rep *Report) bool {
	changed := false
	out := make([]ir.Instr, 0, len(fn.Code))
	dead := false
	for _, in := range fn.Code {
		if in.Op == ir.OpLabel {
			dead = false
		}
		if dead {
			rep.record(Unreachable, fn.Name, "%s removida", in.String())
			changed = true
			continue
		}
		out = append(out, in)
		if in.Op == ir.OpGoto || in.Op == ir.OpReturn {
			dead = true
		}
	}
	fn.Code = out
	return changed
}

// optimizeJumps faz três simplificações de desvios:
//   - jump threading: um desvio para um rótulo cuja primeira instrução é
//     'goto L2' passa a desviar direto para L2;
//   - desvio para a instrução seguinte (goto L / L:) é removido;
//   - rótulos que nenhum desvio referencia são removidos.
func optimizeJumps(fn *ir.Function, rep *Report) bool {
	changed := false

	// posição de cada rótulo
	labelAt := map[string]int{}
	for i, in := range fn.Code {
		if in.Op == ir.OpLabel {
			labelAt[in.Label] = i
		}
	}
	// finalTarget segue cadeias "L1: goto L2" (com limite contra ciclos).
	finalTarget := func(l string) string {
		for hops := 0; hops < len(fn.Code); hops++ {
			i, ok := labelAt[l]
			if !ok {
				return l
			}
			j := i + 1
			for j < len(fn.Code) && fn.Code[j].Op == ir.OpLabel {
				j++
			}
			if j >= len(fn.Code) || fn.Code[j].Op != ir.OpGoto || fn.Code[j].Label == l {
				return l
			}
			l = fn.Code[j].Label
		}
		return l
	}

	for i, in := range fn.Code {
		switch in.Op {
		case ir.OpGoto, ir.OpIfFalse, ir.OpIfTrue:
			if t := finalTarget(in.Label); t != in.Label {
				rep.record(JumpOpt, fn.Name, "%s → desvio direto para %s", in.String(), t)
				fn.Code[i].Label = t
				changed = true
			}
		}
	}

	// desvio para a instrução imediatamente seguinte
	out := make([]ir.Instr, 0, len(fn.Code))
	for i, in := range fn.Code {
		if in.Op == ir.OpGoto || in.Op == ir.OpIfFalse || in.Op == ir.OpIfTrue {
			if fallsInto(fn.Code, i+1, in.Label) {
				rep.record(Redundant, fn.Name, "%s removido (desvio para a instrução seguinte)", in.String())
				changed = true
				continue
			}
		}
		out = append(out, in)
	}
	fn.Code = out

	// rótulos não utilizados
	used := map[string]bool{}
	for _, in := range fn.Code {
		switch in.Op {
		case ir.OpGoto, ir.OpIfFalse, ir.OpIfTrue:
			used[in.Label] = true
		}
	}
	out = make([]ir.Instr, 0, len(fn.Code))
	for _, in := range fn.Code {
		if in.Op == ir.OpLabel && !used[in.Label] {
			rep.record(UnusedLabels, fn.Name, "%s:", in.Label)
			changed = true
			continue
		}
		out = append(out, in)
	}
	fn.Code = out
	return changed
}

// fallsInto indica se, a partir da posição i, só há rótulos até o rótulo l.
func fallsInto(code []ir.Instr, i int, l string) bool {
	for ; i < len(code) && code[i].Op == ir.OpLabel; i++ {
		if code[i].Label == l {
			return true
		}
	}
	return false
}

// ─── Grafo de fluxo de controle e análise de vivacidade ─────────────

type block struct {
	start, end int // instruções [start, end)
	succ       []int
	use, def   map[string]bool
	in, out    map[string]bool
}

// buildCFG divide a função em blocos básicos. Um bloco começa na
// primeira instrução, em cada rótulo e logo após cada desvio/return.
func buildCFG(code []ir.Instr) []*block {
	var blocks []*block
	start := 0
	for i, in := range code {
		if in.Op == ir.OpLabel && i > start {
			blocks = append(blocks, &block{start: start, end: i})
			start = i
		}
		if in.IsJump() {
			blocks = append(blocks, &block{start: start, end: i + 1})
			start = i + 1
		}
	}
	if start < len(code) {
		blocks = append(blocks, &block{start: start, end: len(code)})
	}

	blockOfLabel := map[string]int{}
	for bi, b := range blocks {
		if code[b.start].Op == ir.OpLabel {
			blockOfLabel[code[b.start].Label] = bi
		}
	}
	for bi, b := range blocks {
		last := code[b.end-1]
		switch last.Op {
		case ir.OpGoto:
			b.succ = []int{blockOfLabel[last.Label]}
		case ir.OpIfFalse, ir.OpIfTrue:
			b.succ = []int{blockOfLabel[last.Label]}
			if bi+1 < len(blocks) {
				b.succ = append(b.succ, bi+1)
			}
		case ir.OpReturn:
		default:
			if bi+1 < len(blocks) {
				b.succ = []int{bi + 1}
			}
		}
	}
	return blocks
}

func isVar(op string) bool { return op != "" && !ir.IsConst(op) }

// liveness calcula, para cada bloco, as variáveis vivas na entrada e na
// saída (análise de fluxo de dados para trás, iterada até estabilizar):
//
//	out[B] = ∪ in[S] para S sucessor de B
//	in[B]  = use[B] ∪ (out[B] − def[B])
func liveness(code []ir.Instr, blocks []*block) {
	for _, b := range blocks {
		b.use, b.def = map[string]bool{}, map[string]bool{}
		b.in, b.out = map[string]bool{}, map[string]bool{}
		for i := b.start; i < b.end; i++ {
			for _, u := range code[i].Uses() {
				if isVar(u) && !b.def[u] {
					b.use[u] = true
				}
			}
			if d := code[i].Def(); d != "" {
				b.def[d] = true
			}
		}
	}
	for changed := true; changed; {
		changed = false
		for i := len(blocks) - 1; i >= 0; i-- {
			b := blocks[i]
			for _, s := range b.succ {
				for v := range blocks[s].in {
					if !b.out[v] {
						b.out[v] = true
						changed = true
					}
				}
			}
			for v := range b.use {
				if !b.in[v] {
					b.in[v] = true
					changed = true
				}
			}
			for v := range b.out {
				if !b.def[v] && !b.in[v] {
					b.in[v] = true
					changed = true
				}
			}
		}
	}
}

// mayTrap indica se a instrução pode gerar erro de execução (e portanto
// não pode ser removida mesmo que seu resultado não seja usado).
func mayTrap(in ir.Instr) bool {
	if in.Op != ir.OpBinary || (in.Operator != "/" && in.Operator != "%") {
		return false
	}
	if ir.IsConst(in.Arg2) {
		v, _ := ir.ParseConst(in.Arg2)
		return v == 0
	}
	return true // divisor desconhecido: pode ser zero
}

// removeDeadStores remove atribuições a variáveis locais/temporárias
// cujo valor não é lido em nenhum caminho posterior. Variáveis globais
// são sempre consideradas vivas (outras funções podem lê-las).
func removeDeadStores(fn *ir.Function, rep *Report) bool {
	if len(fn.Code) == 0 {
		return false
	}
	blocks := buildCFG(fn.Code)
	liveness(fn.Code, blocks)

	remove := make([]bool, len(fn.Code))
	changed := false
	for _, b := range blocks {
		live := map[string]bool{}
		for v := range b.out {
			live[v] = true
		}
		for i := b.end - 1; i >= b.start; i-- {
			in := fn.Code[i]
			d := in.Def()
			if d != "" && !ir.IsGlobal(d) && !live[d] {
				switch {
				case in.Op == ir.OpCall: // a chamada fica (efeitos), só o destino some
					rep.record(DeadStore, fn.Name, "%s → resultado descartado", in.String())
					fn.Code[i].Dst = ""
					changed = true
				case !mayTrap(in):
					rep.record(DeadStore, fn.Name, "%s removida", in.String())
					remove[i] = true
					changed = true
					continue // não conta os usos de uma instrução removida
				}
			}
			if d != "" {
				delete(live, d)
			}
			for _, u := range in.Uses() {
				if isVar(u) {
					live[u] = true
				}
			}
		}
	}

	if changed {
		out := make([]ir.Instr, 0, len(fn.Code))
		for i, in := range fn.Code {
			if !remove[i] {
				out = append(out, in)
			}
		}
		fn.Code = out
	}
	return changed
}

// ─── Coalescência de cópias ─────────────────────────────────────────

// coalesceCopies elimina temporários intermediários:
//
//	t3 = i + 1          →     i = i + 1
//	i = t3
//
// Só se aplica quando o temporário é lido uma única vez em toda a função
// (exatamente pela cópia seguinte), então ele não é necessário depois.
func coalesceCopies(fn *ir.Function, rep *Report) bool {
	reads := map[string]int{}
	for _, in := range fn.Code {
		for _, u := range in.Uses() {
			reads[u]++
		}
	}
	changed := false
	out := make([]ir.Instr, 0, len(fn.Code))
	for i := 0; i < len(fn.Code); i++ {
		in := fn.Code[i]
		if i+1 < len(fn.Code) {
			next := fn.Code[i+1]
			t := in.Def()
			if t != "" && isTemp(t) && in.Op != ir.OpAssign &&
				next.Op == ir.OpAssign && next.Arg1 == t && reads[t] == 1 {
				rep.record(TempElim, fn.Name, "%s; %s → %s = ...", in.String(), next.String(), next.Dst)
				in.Dst = next.Dst
				out = append(out, in)
				i++
				changed = true
				continue
			}
		}
		out = append(out, in)
	}
	fn.Code = out
	return changed
}

// isTemp reconhece os temporários criados pelo gerador (t1, t2, ...).
// Variáveis do usuário com esse formato são renomeadas pelo gerador.
func isTemp(name string) bool {
	if len(name) < 2 || name[0] != 't' {
		return false
	}
	for _, c := range name[1:] {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
