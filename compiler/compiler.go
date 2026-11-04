package compiler

import (
	"bytes"
	"io"

	"go-compiler/ast"
	"go-compiler/ir"
	"go-compiler/lexer"
	"go-compiler/optimizer"
	"go-compiler/parser"
	"go-compiler/semantic"
	"go-compiler/vm"
)

// =====================================================================
// PIPELINE DO COMPILADOR
// =====================================================================
//   código-fonte → [léxico] → tokens → [sintático] → AST
//     → [semântico] → AST validada + tabela de símbolos
//     → [gerador] → código de três endereços
//     → [otimizador] → código de três endereços otimizado
//     → [interpretador] → execução
//
// Cada etapa só roda se a anterior terminou sem erros.
// =====================================================================

// Stage identifica a etapa em que a compilação parou.
type Stage string

const (
	StageLexer    Stage = "léxica"
	StageParser   Stage = "sintática"
	StageSemantic Stage = "semântica"
	StageDone     Stage = ""
)

// Result guarda a saída de cada etapa da compilação.
type Result struct {
	Tokens      []lexer.TokenData
	LexErrors   []error
	AST         *ast.ProgramNode
	ParseErrors []error
	Semantic    *semantic.Report
	Symbols     []*semantic.Symbol
	IR          *ir.Program
	Optimized   *ir.Program
	OptReport   *optimizer.Report
	FailedAt    Stage
}

// OK indica se o programa passou por todas as etapas de análise.
func (r *Result) OK() bool { return r.FailedAt == StageDone }

// Compile executa todas as etapas de análise, geração de código
// intermediário e otimização.
func Compile(source string) *Result {
	r := &Result{}

	r.Tokens, r.LexErrors = lexer.NewScanner(source).Tokenize()
	if len(r.LexErrors) > 0 {
		r.FailedAt = StageLexer
		return r
	}

	p := parser.NewParser(r.Tokens)
	r.AST = p.ParseProgram()
	r.ParseErrors = p.Errors()
	if len(r.ParseErrors) > 0 {
		r.FailedAt = StageParser
		return r
	}

	analyzer := semantic.NewSemanticAnalyzer()
	r.Semantic = analyzer.Analyze(r.AST)
	r.Symbols = analyzer.Symbols()
	if r.Semantic.HasErrors() {
		r.FailedAt = StageSemantic
		return r
	}

	r.IR = ir.Generate(r.AST)
	r.Optimized, r.OptReport = optimizer.Optimize(r.IR)
	return r
}

// HasMain indica se o programa possui uma função main para executar.
func (r *Result) HasMain() bool {
	return r.IR != nil && r.IR.Func("main") != nil
}

// Run executa um programa em código intermediário, escrevendo a saída
// em out. Retorna o número de instruções executadas.
func Run(prog *ir.Program, out io.Writer, maxSteps int) (int, error) {
	machine := vm.New(prog, out)
	machine.MaxSteps = maxSteps
	err := machine.Run()
	return machine.Steps, err
}

// RunToString executa o programa e devolve a saída como string.
func RunToString(prog *ir.Program, maxSteps int) (string, int, error) {
	var buf bytes.Buffer
	steps, err := Run(prog, &buf, maxSteps)
	return buf.String(), steps, err
}
