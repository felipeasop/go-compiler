package ir

import (
	"fmt"
	"strings"
)

// =====================================================================
// CÓDIGO INTERMEDIÁRIO — CÓDIGO DE TRÊS ENDEREÇOS (TAC)
// =====================================================================
// Cada instrução tem no máximo um operador e três "endereços"
// (destino + dois operandos), representada como uma quádrupla
// (Op, Dst, Arg1, Arg2). Os operandos são strings com três formas:
//
//   - constante:      25   2.5   "texto"   true   false
//   - variável local: x    soma   x.2  (renomeada quando há shadowing)
//   - variável global: @contador
//   - temporário:     t1   t2 ...  (criados pelo gerador)
//
// Fluxo de controle é feito com rótulos (L1:) e desvios (goto,
// ifFalse, if). Chamadas usam a convenção clássica de 'param' seguido
// de 'call f, n'.
// =====================================================================

// Op é o código de operação de uma instrução.
type Op int

const (
	OpAssign  Op = iota // Dst = Arg1
	OpBinary            // Dst = Arg1 Operator Arg2
	OpUnary             // Dst = Operator Arg1    (-, !, int, float64)
	OpLabel             // Label:
	OpGoto              // goto Label
	OpIfFalse           // ifFalse Arg1 goto Label
	OpIfTrue            // if Arg1 goto Label
	OpParam             // param Arg1
	OpCall              // [Dst =] call Label, N
	OpReturn            // return [Arg1]
	OpPrint             // print Label, N   (Label = Println | Print | Printf)
)

// Instr é uma instrução de três endereços.
type Instr struct {
	Op       Op
	Dst      string
	Arg1     string
	Arg2     string
	Operator string // operador de OpBinary / OpUnary
	Label    string // rótulo (Label/Goto/If*), nome da função (Call) ou tipo de print
	N        int    // número de parâmetros de Call/Print
	Line     int    // linha do código-fonte (para erros de execução)
}

func (i Instr) String() string {
	switch i.Op {
	case OpAssign:
		return fmt.Sprintf("%s = %s", i.Dst, i.Arg1)
	case OpBinary:
		return fmt.Sprintf("%s = %s %s %s", i.Dst, i.Arg1, i.Operator, i.Arg2)
	case OpUnary:
		if i.Operator == "-" || i.Operator == "!" {
			return fmt.Sprintf("%s = %s%s", i.Dst, i.Operator, i.Arg1)
		}
		return fmt.Sprintf("%s = %s(%s)", i.Dst, i.Operator, i.Arg1)
	case OpLabel:
		return i.Label + ":"
	case OpGoto:
		return "goto " + i.Label
	case OpIfFalse:
		return fmt.Sprintf("ifFalse %s goto %s", i.Arg1, i.Label)
	case OpIfTrue:
		return fmt.Sprintf("if %s goto %s", i.Arg1, i.Label)
	case OpParam:
		return "param " + i.Arg1
	case OpCall:
		if i.Dst != "" {
			return fmt.Sprintf("%s = call %s, %d", i.Dst, i.Label, i.N)
		}
		return fmt.Sprintf("call %s, %d", i.Label, i.N)
	case OpReturn:
		if i.Arg1 != "" {
			return "return " + i.Arg1
		}
		return "return"
	case OpPrint:
		return fmt.Sprintf("print %s, %d", i.Label, i.N)
	}
	return "?"
}

// Uses retorna os operandos lidos pela instrução (variáveis e constantes).
func (i Instr) Uses() []string {
	switch i.Op {
	case OpAssign, OpUnary, OpIfFalse, OpIfTrue, OpParam:
		return []string{i.Arg1}
	case OpBinary:
		return []string{i.Arg1, i.Arg2}
	case OpReturn:
		if i.Arg1 != "" {
			return []string{i.Arg1}
		}
	}
	return nil
}

// Def retorna a variável escrita pela instrução ("" se nenhuma).
func (i Instr) Def() string {
	switch i.Op {
	case OpAssign, OpBinary, OpUnary, OpCall:
		return i.Dst
	}
	return ""
}

// IsJump indica se a instrução encerra um bloco básico.
func (i Instr) IsJump() bool {
	switch i.Op {
	case OpGoto, OpIfFalse, OpIfTrue, OpReturn:
		return true
	}
	return false
}

// Function é o código de três endereços de uma função.
type Function struct {
	Name   string
	Params []string
	Code   []Instr
}

// Program é o programa completo em código intermediário. Init contém a
// inicialização das variáveis globais, executada antes de main.
type Program struct {
	Globals   []string
	Init      *Function
	Functions []*Function
}

// InitName é o nome da pseudo-função que inicializa as variáveis globais.
const InitName = "init.globais"

// Func procura uma função pelo nome.
func (p *Program) Func(name string) *Function {
	if p.Init != nil && name == p.Init.Name {
		return p.Init
	}
	for _, f := range p.Functions {
		if f.Name == name {
			return f
		}
	}
	return nil
}

// AllFunctions retorna Init (se houver) seguida das funções do programa.
func (p *Program) AllFunctions() []*Function {
	var out []*Function
	if p.Init != nil && len(p.Init.Code) > 0 {
		out = append(out, p.Init)
	}
	return append(out, p.Functions...)
}

// InstrCount é o total de instruções (sem contar rótulos).
func (p *Program) InstrCount() int {
	n := 0
	for _, f := range p.AllFunctions() {
		n += f.InstrCount()
	}
	return n
}

// InstrCount conta as instruções da função, sem os rótulos.
func (f *Function) InstrCount() int {
	n := 0
	for _, in := range f.Code {
		if in.Op != OpLabel {
			n++
		}
	}
	return n
}

func (f *Function) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("func %s(%s):\n", f.Name, strings.Join(f.Params, ", ")))
	for _, in := range f.Code {
		if in.Op == OpLabel {
			sb.WriteString(in.String() + "\n")
		} else {
			sb.WriteString("    " + in.String() + "\n")
		}
	}
	return sb.String()
}

func (p *Program) String() string {
	var sb strings.Builder
	if len(p.Globals) > 0 {
		sb.WriteString("globais: " + strings.Join(p.Globals, ", ") + "\n\n")
	}
	for i, f := range p.AllFunctions() {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(f.String())
	}
	return sb.String()
}

// Clone faz uma cópia profunda do programa (o otimizador trabalha sobre
// uma cópia para que o código original continue disponível para comparação).
func (p *Program) Clone() *Program {
	cp := &Program{Globals: append([]string(nil), p.Globals...)}
	cloneFn := func(f *Function) *Function {
		if f == nil {
			return nil
		}
		return &Function{
			Name:   f.Name,
			Params: append([]string(nil), f.Params...),
			Code:   append([]Instr(nil), f.Code...),
		}
	}
	cp.Init = cloneFn(p.Init)
	for _, f := range p.Functions {
		cp.Functions = append(cp.Functions, cloneFn(f))
	}
	return cp
}
