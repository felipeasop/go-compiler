package ir

import (
	"fmt"
	"regexp"
	"strconv"

	"go-compiler/ast"
)

// =====================================================================
// GERADOR DE CÓDIGO INTERMEDIÁRIO
// =====================================================================
// Percorre a AST (já validada pela análise semântica) e emite código
// de três endereços. Esquemas de tradução:
//
//   x := a + b * c        t1 = b * c
//                         t2 = a + t1
//                         x = t2
//
//   if c { A } else { B }     t = c
//                             ifFalse t goto L1
//                             A
//                             goto L2
//                         L1: B
//                         L2:
//
//   for i; c; p { A }         i
//                         L1: t = c
//                             ifFalse t goto L3
//                             A
//                         L2: p              ← destino do 'continue'
//                             goto L1
//                         L3:                ← destino do 'break'
//
//   a && b  (curto-circuito)  t = a
//                             ifFalse t goto L1
//                             t = b
//                         L1:
//
// Nomes: variáveis globais recebem prefixo '@'. Uma variável local que
// "esconde" (shadowing) outra de mesmo nome na mesma função é renomeada
// (x, x.2, x.3...), pois o código intermediário não tem blocos aninhados.
// =====================================================================

type loopLabels struct {
	brk, cont string
}

// Generator traduz a AST em um Program de código de três endereços.
type Generator struct {
	prog *Program
	fn   *Function

	temps, labels int
	scopes        []map[string]string // nome no fonte → nome no código intermediário
	used          map[string]int      // quantas vezes cada nome-base já foi declarado na função
	globals       map[string]bool
	loops         []loopLabels
}

// Generate gera o código intermediário de um programa.
func Generate(program *ast.ProgramNode) *Program {
	g := &Generator{prog: &Program{}, globals: map[string]bool{}}

	// Variáveis globais: inicializadas na pseudo-função init.globais.
	g.beginFunction(InitName)
	for _, gl := range program.Globals {
		if vd, ok := gl.(*ast.VarDeclNode); ok {
			value := g.declValue(vd)
			name := "@" + vd.Name
			g.globals[vd.Name] = true
			g.prog.Globals = append(g.prog.Globals, name)
			g.emit(Instr{Op: OpAssign, Dst: name, Arg1: value, Line: vd.Line})
		}
	}
	if len(g.fn.Code) > 0 {
		g.emit(Instr{Op: OpReturn})
	}
	g.prog.Init = g.fn

	for _, f := range program.Functions {
		if fn, ok := f.(*ast.FuncNode); ok {
			g.genFunc(fn)
		}
	}
	return g.prog
}

// ─── infraestrutura ────────────────────────────────────────────────

func (g *Generator) beginFunction(name string) {
	g.fn = &Function{Name: name}
	g.temps, g.labels = 0, 0
	g.scopes = []map[string]string{{}}
	g.used = map[string]int{}
	g.loops = nil
}

func (g *Generator) emit(in Instr) { g.fn.Code = append(g.fn.Code, in) }

func (g *Generator) newTemp() string {
	g.temps++
	return "t" + strconv.Itoa(g.temps)
}

func (g *Generator) newLabel() string {
	g.labels++
	return "L" + strconv.Itoa(g.labels)
}

func (g *Generator) label(l string) { g.emit(Instr{Op: OpLabel, Label: l}) }

func (g *Generator) pushScope() { g.scopes = append(g.scopes, map[string]string{}) }
func (g *Generator) popScope()  { g.scopes = g.scopes[:len(g.scopes)-1] }

var tempLike = regexp.MustCompile(`^t\d+$`)

// declareLocal cria o nome único de uma variável local no escopo atual.
func (g *Generator) declareLocal(name string) string {
	g.used[name]++
	irName := name
	if tempLike.MatchString(name) { // evita colisão com temporários
		g.used[name]++
	}
	if n := g.used[name]; n > 1 {
		irName = name + "." + strconv.Itoa(n)
	}
	g.scopes[len(g.scopes)-1][name] = irName
	return irName
}

// resolve encontra o nome no código intermediário de uma variável.
func (g *Generator) resolve(name string) string {
	for i := len(g.scopes) - 1; i >= 0; i-- {
		if irName, ok := g.scopes[i][name]; ok {
			return irName
		}
	}
	return "@" + name // análise semântica garante que é uma global
}

// ─── funções ───────────────────────────────────────────────────────

func (g *Generator) genFunc(n *ast.FuncNode) {
	g.beginFunction(n.Name)
	for _, p := range n.Params {
		g.fn.Params = append(g.fn.Params, g.declareLocal(p.Name))
	}
	g.genBlock(n.Body)

	// Garante que toda função termina com return (funções void podem
	// simplesmente "cair" no fim do corpo).
	if len(g.fn.Code) == 0 || g.fn.Code[len(g.fn.Code)-1].Op != OpReturn {
		g.emit(Instr{Op: OpReturn})
	}
	g.prog.Functions = append(g.prog.Functions, g.fn)
}

func (g *Generator) genBlock(stmts []ast.ASTNode) {
	for _, s := range stmts {
		g.genStmt(s)
	}
}

func (g *Generator) genScopedBlock(stmts []ast.ASTNode) {
	g.pushScope()
	g.genBlock(stmts)
	g.popScope()
}

// ─── instruções ────────────────────────────────────────────────────

func (g *Generator) genStmt(node ast.ASTNode) {
	switch n := node.(type) {

	case *ast.VarDeclNode:
		value := g.declValue(n) // avaliado ANTES de declarar: x := x + 1
		g.emit(Instr{Op: OpAssign, Dst: g.declareLocal(n.Name), Arg1: value, Line: n.Line})

	case *ast.AssignNode:
		value := g.genExpr(n.Expr)
		if n.Name == "_" { // valor descartado: só os efeitos da expressão importam
			return
		}
		g.emit(Instr{Op: OpAssign, Dst: g.resolve(n.Name), Arg1: value, Line: n.Line})

	case *ast.PrintCallNode:
		args := make([]string, len(n.Args))
		for i, a := range n.Args {
			args[i] = g.genExpr(a)
		}
		for _, a := range args {
			g.emit(Instr{Op: OpParam, Arg1: a, Line: n.Line})
		}
		g.emit(Instr{Op: OpPrint, Label: n.Func, N: len(args), Line: n.Line})

	case *ast.CallNode:
		g.genCall(n, false)

	case *ast.IfNode:
		g.genIf(n)

	case *ast.ForNode:
		g.genFor(n)

	case *ast.ReturnNode:
		in := Instr{Op: OpReturn, Line: n.Line}
		if n.Value != nil {
			in.Arg1 = g.genExpr(n.Value)
		}
		g.emit(in)

	case *ast.BreakNode:
		g.emit(Instr{Op: OpGoto, Label: g.loops[len(g.loops)-1].brk, Line: n.Line})

	case *ast.ContinueNode:
		g.emit(Instr{Op: OpGoto, Label: g.loops[len(g.loops)-1].cont, Line: n.Line})

	default:
		g.genExpr(node)
	}
}

// declValue gera o valor inicial de uma declaração: o inicializador ou o
// valor zero do tipo declarado.
func (g *Generator) declValue(n *ast.VarDeclNode) string {
	if n.Initializer != nil {
		return g.genExpr(n.Initializer)
	}
	switch n.TypeName {
	case "float64", "float32":
		return "0.0"
	case "string":
		return `""`
	case "bool":
		return "false"
	}
	return "0"
}

func (g *Generator) genIf(n *ast.IfNode) {
	g.pushScope()
	if n.Init != nil {
		g.genStmt(n.Init)
	}
	cond := g.genExpr(n.Condition)
	elseLabel := g.newLabel()
	g.emit(Instr{Op: OpIfFalse, Arg1: cond, Label: elseLabel, Line: n.Line})
	g.genScopedBlock(n.ThenBranch)

	if len(n.ElseBranch) > 0 {
		endLabel := g.newLabel()
		g.emit(Instr{Op: OpGoto, Label: endLabel})
		g.label(elseLabel)
		g.genScopedBlock(n.ElseBranch)
		g.label(endLabel)
	} else {
		g.label(elseLabel)
	}
	g.popScope()
}

func (g *Generator) genFor(n *ast.ForNode) {
	g.pushScope()
	if n.Init != nil {
		g.genStmt(n.Init)
	}
	start, cont, end := g.newLabel(), g.newLabel(), g.newLabel()

	g.label(start)
	if n.Condition != nil {
		cond := g.genExpr(n.Condition)
		g.emit(Instr{Op: OpIfFalse, Arg1: cond, Label: end, Line: n.Line})
	}

	g.loops = append(g.loops, loopLabels{brk: end, cont: cont})
	g.genScopedBlock(n.Body)
	g.loops = g.loops[:len(g.loops)-1]

	g.label(cont)
	if n.Post != nil {
		g.genStmt(n.Post)
	}
	g.emit(Instr{Op: OpGoto, Label: start})
	g.label(end)
	g.popScope()
}

// ─── expressões ────────────────────────────────────────────────────

// genExpr gera o código de uma expressão e retorna o operando que
// contém seu valor (constante, variável ou temporário).
func (g *Generator) genExpr(node ast.ASTNode) string {
	switch n := node.(type) {

	case *ast.LiteralNode:
		return literalOperand(n.Value)

	case *ast.VariableNode:
		return g.resolve(n.Name)

	case *ast.BinaryOpNode:
		if n.OpStr == "&&" || n.OpStr == "||" {
			return g.genShortCircuit(n)
		}
		left := g.genExpr(n.Left)
		right := g.genExpr(n.Right)
		t := g.newTemp()
		g.emit(Instr{Op: OpBinary, Dst: t, Arg1: left, Operator: n.OpStr, Arg2: right, Line: n.Line})
		return t

	case *ast.UnaryOpNode:
		operand := g.genExpr(n.Operand)
		t := g.newTemp()
		g.emit(Instr{Op: OpUnary, Dst: t, Operator: n.OpStr, Arg1: operand, Line: n.Line})
		return t

	case *ast.CallNode:
		return g.genCall(n, true)
	}
	panic(fmt.Sprintf("gerador: expressão não suportada %T", node))
}

// genShortCircuit traduz && e || com avaliação em curto-circuito: o
// operando da direita só é avaliado quando necessário.
func (g *Generator) genShortCircuit(n *ast.BinaryOpNode) string {
	t := g.newTemp()
	end := g.newLabel()
	left := g.genExpr(n.Left)
	g.emit(Instr{Op: OpAssign, Dst: t, Arg1: left, Line: n.Line})
	op := OpIfFalse // a && b: se a é falso, o resultado já é falso
	if n.OpStr == "||" {
		op = OpIfTrue // a || b: se a é verdadeiro, o resultado já é verdadeiro
	}
	g.emit(Instr{Op: op, Arg1: t, Label: end, Line: n.Line})
	right := g.genExpr(n.Right)
	g.emit(Instr{Op: OpAssign, Dst: t, Arg1: right, Line: n.Line})
	g.label(end)
	return t
}

func (g *Generator) genCall(n *ast.CallNode, wantValue bool) string {
	switch n.Name {
	case "int", "float64", "float32": // conversão de tipo
		operand := g.genExpr(n.Args[0])
		t := g.newTemp()
		g.emit(Instr{Op: OpUnary, Dst: t, Operator: n.Name, Arg1: operand, Line: n.Line})
		return t
	}

	args := make([]string, len(n.Args))
	for i, a := range n.Args {
		args[i] = g.genExpr(a)
	}
	for _, a := range args {
		g.emit(Instr{Op: OpParam, Arg1: a, Line: n.Line})
	}
	in := Instr{Op: OpCall, Label: n.Name, N: len(args), Line: n.Line}
	if wantValue {
		in.Dst = g.newTemp()
	}
	g.emit(in)
	return in.Dst
}

// literalOperand normaliza um literal do código-fonte em operando
// constante: strings em formato Go com aspas duplas, runes como inteiros.
func literalOperand(lex string) string {
	switch {
	case lex[0] == '"' || lex[0] == '`':
		if s, err := strconv.Unquote(lex); err == nil {
			return strconv.Quote(s)
		}
	case lex[0] == '\'':
		if r, _, _, err := strconv.UnquoteChar(lex[1:len(lex)-1], '\''); err == nil {
			return strconv.Itoa(int(r))
		}
	case lex == "true" || lex == "false":
		return lex
	default:
		if v, err := ParseConst(lex); err == nil {
			if s, ok := FormatConst(v); ok {
				return s
			}
		}
	}
	return lex
}
