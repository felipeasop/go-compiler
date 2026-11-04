package optimizer

import (
	"go-compiler/ir"
)

// =====================================================================
// PASSE LOCAL (por bloco básico)
// =====================================================================
// Percorre as instruções em ordem, mantendo três tabelas que valem
// apenas até o fim do bloco básico atual:
//
//   consts: variável → constante que ela contém      (x = 5)
//   copies: variável → outra variável que ela copia  (x = y)
//   exprs:  expressão (op, a, b) → variável que já guarda seu resultado
//
// Toda vez que uma variável v é redefinida, as entradas que dependem de
// v são invalidadas (kill). Em rótulos e desvios as tabelas são
// esvaziadas, pois ali o fluxo de controle pode vir de outro caminho.
// =====================================================================

type exprKey struct {
	op, a, b string
}

type localState struct {
	consts map[string]string
	copies map[string]string
	exprs  map[exprKey]string
}

func newLocalState() *localState {
	s := &localState{}
	s.reset()
	return s
}

func (s *localState) reset() {
	s.consts = map[string]string{}
	s.copies = map[string]string{}
	s.exprs = map[exprKey]string{}
}

// kill invalida tudo o que depende do valor atual de v.
func (s *localState) kill(v string) {
	delete(s.consts, v)
	delete(s.copies, v)
	for k, src := range s.copies {
		if src == v {
			delete(s.copies, k)
		}
	}
	for k, holder := range s.exprs {
		if holder == v || k.a == v || k.b == v {
			delete(s.exprs, k)
		}
	}
}

// killGlobals esquece tudo o que envolve variáveis globais (após 'call').
func (s *localState) killGlobals() {
	for k := range s.consts {
		if ir.IsGlobal(k) {
			delete(s.consts, k)
		}
	}
	for k, src := range s.copies {
		if ir.IsGlobal(k) || ir.IsGlobal(src) {
			delete(s.copies, k)
		}
	}
	for k, holder := range s.exprs {
		if ir.IsGlobal(holder) || ir.IsGlobal(k.a) || ir.IsGlobal(k.b) {
			delete(s.exprs, k)
		}
	}
}

// subst substitui um operando pela constante ou variável original que ele contém.
func (s *localState) subst(op string, fn string, rep *Report) string {
	if c, ok := s.consts[op]; ok {
		rep.record(ConstProp, fn, "%s → %s", op, c)
		return c
	}
	if src, ok := s.copies[op]; ok {
		rep.record(CopyProp, fn, "%s → %s", op, src)
		return src
	}
	return op
}

// remember registra o efeito de "dst = src" nas tabelas.
func (s *localState) remember(dst, src string) {
	if ir.IsConst(src) {
		s.consts[dst] = src
	} else if src != dst {
		s.copies[dst] = src
	}
}

func localPass(fn *ir.Function, rep *Report) bool {
	s := newLocalState()
	changed := false
	out := make([]ir.Instr, 0, len(fn.Code))

	for _, in := range fn.Code {
		orig := in.String()

		switch in.Op {
		case ir.OpLabel:
			s.reset()

		case ir.OpAssign:
			in.Arg1 = s.subst(in.Arg1, fn.Name, rep)
			if in.Arg1 == in.Dst {
				rep.record(Redundant, fn.Name, "%s removida", orig)
				changed = true
				continue
			}
			s.kill(in.Dst)
			s.remember(in.Dst, in.Arg1)

		case ir.OpBinary:
			in.Arg1 = s.subst(in.Arg1, fn.Name, rep)
			in.Arg2 = s.subst(in.Arg2, fn.Name, rep)
			in = simplifyBinary(in, fn.Name, rep)
			if in.Op == ir.OpBinary {
				key := exprKey{in.Operator, in.Arg1, in.Arg2}
				if holder, ok := s.exprs[key]; ok {
					rep.record(CSE, fn.Name, "%s → %s = %s", orig, in.Dst, holder)
					in = ir.Instr{Op: ir.OpAssign, Dst: in.Dst, Arg1: holder, Line: in.Line}
				}
			}
			s.kill(in.Dst)
			if in.Op == ir.OpAssign {
				s.remember(in.Dst, in.Arg1)
			} else if in.Dst != in.Arg1 && in.Dst != in.Arg2 {
				s.exprs[exprKey{in.Operator, in.Arg1, in.Arg2}] = in.Dst
			}

		case ir.OpUnary:
			in.Arg1 = s.subst(in.Arg1, fn.Name, rep)
			if ir.IsConst(in.Arg1) {
				if folded, ok := foldUnary(in); ok {
					rep.record(ConstFolding, fn.Name, "%s → %s", orig, folded)
					in = folded
				}
			}
			s.kill(in.Dst)
			if in.Op == ir.OpAssign {
				s.remember(in.Dst, in.Arg1)
			}

		case ir.OpIfFalse, ir.OpIfTrue:
			in.Arg1 = s.subst(in.Arg1, fn.Name, rep)
			if ir.IsConst(in.Arg1) {
				cond, _ := ir.ParseConst(in.Arg1)
				jumps := cond == (in.Op == ir.OpIfTrue)
				changed = true
				if jumps {
					rep.record(ConstBranch, fn.Name, "%s → goto %s", orig, in.Label)
					in = ir.Instr{Op: ir.OpGoto, Label: in.Label, Line: in.Line}
				} else {
					rep.record(ConstBranch, fn.Name, "%s removido (nunca desvia)", orig)
					continue
				}
			}
			s.reset()

		case ir.OpGoto:
			s.reset()

		case ir.OpReturn:
			if in.Arg1 != "" {
				in.Arg1 = s.subst(in.Arg1, fn.Name, rep)
			}
			s.reset()

		case ir.OpParam:
			in.Arg1 = s.subst(in.Arg1, fn.Name, rep)

		case ir.OpCall:
			s.killGlobals()
			if in.Dst != "" {
				s.kill(in.Dst)
			}
		}

		if in.String() != orig {
			changed = true
		}
		out = append(out, in)
	}

	fn.Code = out
	return changed
}

// simplifyBinary aplica dobramento de constantes, simplificação
// algébrica e redução de força a uma instrução binária.
func simplifyBinary(in ir.Instr, fn string, rep *Report) ir.Instr {
	orig := in.String()
	asAssign := func(src string) ir.Instr {
		return ir.Instr{Op: ir.OpAssign, Dst: in.Dst, Arg1: src, Line: in.Line}
	}

	// Dobramento: os dois operandos são constantes.
	if ir.IsConst(in.Arg1) && ir.IsConst(in.Arg2) {
		a, _ := ir.ParseConst(in.Arg1)
		b, _ := ir.ParseConst(in.Arg2)
		if v, err := ir.EvalBinary(in.Operator, a, b); err == nil { // div. por zero: não dobra
			if c, ok := ir.FormatConst(v); ok {
				rep.record(ConstFolding, fn, "%s → %s = %s", orig, in.Dst, c)
				return asAssign(c)
			}
		}
		return in
	}

	a, b := in.Arg1, in.Arg2
	switch in.Operator {
	case "+":
		if b == "0" || b == `""` {
			rep.record(Algebraic, fn, "%s → %s = %s", orig, in.Dst, a)
			return asAssign(a)
		}
		if a == "0" || a == `""` {
			rep.record(Algebraic, fn, "%s → %s = %s", orig, in.Dst, b)
			return asAssign(b)
		}
	case "-":
		if b == "0" {
			rep.record(Algebraic, fn, "%s → %s = %s", orig, in.Dst, a)
			return asAssign(a)
		}
	case "*":
		switch {
		case b == "1" || b == "1.0":
			rep.record(Algebraic, fn, "%s → %s = %s", orig, in.Dst, a)
			return asAssign(a)
		case a == "1" || a == "1.0":
			rep.record(Algebraic, fn, "%s → %s = %s", orig, in.Dst, b)
			return asAssign(b)
		case a == "0" || b == "0": // apenas inteiros: para float, 0 * Inf = NaN
			rep.record(Algebraic, fn, "%s → %s = 0", orig, in.Dst)
			return asAssign("0")
		case b == "2":
			rep.record(Strength, fn, "%s → %s = %s + %s", orig, in.Dst, a, a)
			in.Operator, in.Arg2 = "+", a
			return in
		case a == "2":
			rep.record(Strength, fn, "%s → %s = %s + %s", orig, in.Dst, b, b)
			in.Operator, in.Arg1 = "+", b
			return in
		}
	case "/":
		if b == "1" || b == "1.0" {
			rep.record(Algebraic, fn, "%s → %s = %s", orig, in.Dst, a)
			return asAssign(a)
		}
	case "%":
		if b == "1" {
			rep.record(Algebraic, fn, "%s → %s = 0", orig, in.Dst)
			return asAssign("0")
		}
	}
	return in
}

func foldUnary(in ir.Instr) (ir.Instr, bool) {
	a, err := ir.ParseConst(in.Arg1)
	if err != nil {
		return in, false
	}
	v, err := ir.EvalUnary(in.Operator, a)
	if err != nil {
		return in, false
	}
	c, ok := ir.FormatConst(v)
	if !ok {
		return in, false
	}
	return ir.Instr{Op: ir.OpAssign, Dst: in.Dst, Arg1: c, Line: in.Line}, true
}
