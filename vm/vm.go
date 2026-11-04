package vm

import (
	"errors"
	"fmt"
	"io"

	"go-compiler/ir"
)

// =====================================================================
// BACK-END: INTERPRETADOR DE CÓDIGO DE TRÊS ENDEREÇOS
// =====================================================================
// Executa diretamente o código intermediário (otimizado ou não).
//
//   - Cada função é pré-processada uma vez: os rótulos viram índices
//     (label → posição da instrução) e as constantes são convertidas.
//   - Cada chamada cria um registro de ativação (frame) com as variáveis
//     locais e temporários; a pilha de chamadas do próprio Go guarda os
//     frames, o que permite recursão.
//   - Variáveis globais (@x) ficam em um mapa compartilhado.
//   - 'param' acumula argumentos; 'call' e 'print' os consomem.
//   - Erros de execução (divisão inteira por zero, estouro de pilha)
//     interrompem o programa com a linha do código-fonte.
// =====================================================================

// MaxDepth limita a profundidade de recursão (estouro de pilha).
const MaxDepth = 10000

// RuntimeError é um erro ocorrido durante a execução do programa.
type RuntimeError struct {
	Func string
	Line int
	Err  error
}

func (e *RuntimeError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("erro de execução linha %d (função %s): %v", e.Line, e.Func, e.Err)
	}
	return fmt.Sprintf("erro de execução (função %s): %v", e.Func, e.Err)
}

func (e *RuntimeError) Unwrap() error { return e.Err }

// ErrStepLimit indica que o limite de instruções executadas foi atingido.
var ErrStepLimit = errors.New("limite de instruções executadas atingido")

type compiled struct {
	fn     *ir.Function
	labels map[string]int
}

// VM é o interpretador.
type VM struct {
	funcs   map[string]*compiled
	globals map[string]ir.Value
	consts  map[string]ir.Value
	out     io.Writer
	depth   int

	// MaxSteps limita o total de instruções executadas (0 = sem limite).
	MaxSteps int
	// Steps conta as instruções executadas (útil para comparar o
	// desempenho do código otimizado com o não otimizado).
	Steps int
}

// New prepara um programa para execução, escrevendo a saída em out.
func New(prog *ir.Program, out io.Writer) *VM {
	m := &VM{
		funcs:   map[string]*compiled{},
		globals: map[string]ir.Value{},
		consts:  map[string]ir.Value{},
		out:     out,
	}
	for _, f := range prog.AllFunctions() {
		c := &compiled{fn: f, labels: map[string]int{}}
		for i, in := range f.Code {
			if in.Op == ir.OpLabel {
				c.labels[in.Label] = i
			}
		}
		m.funcs[f.Name] = c
	}
	return m
}

// Run inicializa as variáveis globais e executa main.
func (m *VM) Run() error {
	if _, ok := m.funcs[ir.InitName]; ok {
		if _, err := m.call(ir.InitName, nil); err != nil {
			return err
		}
	}
	if _, ok := m.funcs["main"]; !ok {
		return errors.New("o programa não possui func main()")
	}
	_, err := m.call("main", nil)
	return err
}

// call executa uma função com os argumentos dados e devolve o retorno.
func (m *VM) call(name string, args []ir.Value) (ir.Value, error) {
	c, ok := m.funcs[name]
	if !ok {
		return nil, &RuntimeError{Func: name, Err: fmt.Errorf("função '%s' inexistente", name)}
	}
	m.depth++
	defer func() { m.depth-- }()
	if m.depth > MaxDepth {
		return nil, &RuntimeError{Func: name, Err: fmt.Errorf("estouro de pilha (recursão com mais de %d níveis)", MaxDepth)}
	}

	locals := make(map[string]ir.Value, len(c.fn.Params)+8)
	for i, p := range c.fn.Params {
		locals[p] = args[i]
	}
	var params []ir.Value

	fail := func(in ir.Instr, err error) (ir.Value, error) {
		var re *RuntimeError
		if errors.As(err, &re) {
			return nil, err // erro já contextualizado em uma chamada interna
		}
		return nil, &RuntimeError{Func: name, Line: in.Line, Err: err}
	}

	code := c.fn.Code
	for pc := 0; pc < len(code); pc++ {
		in := code[pc]
		m.Steps++
		if m.MaxSteps > 0 && m.Steps > m.MaxSteps {
			return fail(in, ErrStepLimit)
		}

		switch in.Op {
		case ir.OpLabel:
			m.Steps-- // rótulos não são instruções executáveis

		case ir.OpAssign:
			v, err := m.read(locals, in.Arg1)
			if err != nil {
				return fail(in, err)
			}
			m.write(locals, in.Dst, v)

		case ir.OpBinary:
			a, err := m.read(locals, in.Arg1)
			if err != nil {
				return fail(in, err)
			}
			b, err := m.read(locals, in.Arg2)
			if err != nil {
				return fail(in, err)
			}
			v, err := ir.EvalBinary(in.Operator, a, b)
			if err != nil {
				return fail(in, err)
			}
			m.write(locals, in.Dst, v)

		case ir.OpUnary:
			a, err := m.read(locals, in.Arg1)
			if err != nil {
				return fail(in, err)
			}
			v, err := ir.EvalUnary(in.Operator, a)
			if err != nil {
				return fail(in, err)
			}
			m.write(locals, in.Dst, v)

		case ir.OpGoto:
			pc = c.labels[in.Label]

		case ir.OpIfFalse, ir.OpIfTrue:
			v, err := m.read(locals, in.Arg1)
			if err != nil {
				return fail(in, err)
			}
			b, err := ir.IsTruthy(v)
			if err != nil {
				return fail(in, err)
			}
			if b == (in.Op == ir.OpIfTrue) {
				pc = c.labels[in.Label]
			}

		case ir.OpParam:
			v, err := m.read(locals, in.Arg1)
			if err != nil {
				return fail(in, err)
			}
			params = append(params, v)

		case ir.OpCall:
			args := params[len(params)-in.N:]
			params = params[:len(params)-in.N]
			v, err := m.call(in.Label, append([]ir.Value(nil), args...))
			if err != nil {
				return fail(in, err)
			}
			if in.Dst != "" {
				m.write(locals, in.Dst, v)
			}

		case ir.OpPrint:
			args := params[len(params)-in.N:]
			params = params[:len(params)-in.N]
			if err := m.print(in.Label, args); err != nil {
				return fail(in, err)
			}

		case ir.OpReturn:
			if in.Arg1 == "" {
				return nil, nil
			}
			v, err := m.read(locals, in.Arg1)
			if err != nil {
				return fail(in, err)
			}
			return v, nil
		}
	}
	return nil, nil
}

// print implementa fmt.Println/Print/Printf usando o próprio pacote fmt
// do Go — a saída é idêntica à de um programa Go compilado.
func (m *VM) print(kind string, args []ir.Value) error {
	var err error
	switch kind {
	case "Println":
		_, err = fmt.Fprintln(m.out, args...)
	case "Print":
		_, err = fmt.Fprint(m.out, args...)
	case "Printf":
		format, _ := args[0].(string)
		_, err = fmt.Fprintf(m.out, format, args[1:]...)
	default:
		err = fmt.Errorf("print desconhecido: %s", kind)
	}
	return err
}

func (m *VM) read(locals map[string]ir.Value, op string) (ir.Value, error) {
	if ir.IsConst(op) {
		if v, ok := m.consts[op]; ok {
			return v, nil
		}
		v, err := ir.ParseConst(op)
		if err != nil {
			return nil, fmt.Errorf("constante inválida %s", op)
		}
		m.consts[op] = v
		return v, nil
	}
	if ir.IsGlobal(op) {
		if v, ok := m.globals[op]; ok {
			return v, nil
		}
		return nil, fmt.Errorf("variável global %s sem valor", op)
	}
	if v, ok := locals[op]; ok {
		return v, nil
	}
	return nil, fmt.Errorf("variável %s sem valor", op)
}

func (m *VM) write(locals map[string]ir.Value, name string, v ir.Value) {
	if ir.IsGlobal(name) {
		m.globals[name] = v
		return
	}
	locals[name] = v
}
