package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go-compiler/compiler"
	"go-compiler/lexer"
	"go-compiler/semantic"
)

var (
	showTokens  = flag.Bool("tokens", false, "mostra os tokens (análise léxica)")
	showAST     = flag.Bool("ast", false, "mostra a árvore sintática abstrata")
	showSem     = flag.Bool("sem", false, "mostra a tabela de símbolos e o relatório semântico")
	showIR      = flag.Bool("ir", false, "mostra o código intermediário (três endereços)")
	showOpt     = flag.Bool("opt", false, "mostra o código otimizado e o resumo das otimizações")
	doRun       = flag.Bool("run", false, "executa o programa")
	noOptimize  = flag.Bool("O0", false, "executa o código intermediário SEM otimização")
	compare     = flag.Bool("compare", false, "executa as versões sem e com otimização e compara as saídas")
	explain     = flag.Bool("explain", false, "lista cada transformação feita pelo otimizador")
	outDir      = flag.String("out", "outputs", "diretório onde os artefatos de cada etapa são gravados")
	maxSteps    = flag.Int("max-steps", 0, "limite de instruções executadas (0 = sem limite)")
	exitCompile = 1
	exitRuntime = 2
)

func usage() {
	fmt.Fprintf(os.Stderr, "Uso: go run . [opções] <arquivo.go>\n\n")
	fmt.Fprintf(os.Stderr, "Sem opções de etapa, mostra todas as etapas e executa o programa.\n\nOpções:\n")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	// Permite as opções antes ou depois do arquivo: go run . prog.go -ir
	var flags, files []string
	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		a := args[i]
		if strings.HasPrefix(a, "-") {
			flags = append(flags, a)
			name := strings.TrimLeft(a, "-")
			if (name == "out" || name == "max-steps") && i+1 < len(args) {
				i++
				flags = append(flags, args[i])
			}
		} else {
			files = append(files, a)
		}
	}
	flag.CommandLine.Parse(flags)
	if len(files) != 1 {
		usage()
		os.Exit(exitCompile)
	}
	if !*showTokens && !*showAST && !*showSem && !*showIR && !*showOpt && !*doRun && !*compare {
		*showTokens, *showAST, *showSem, *showIR, *showOpt, *doRun = true, true, true, true, true, true
	}

	filePath := files[0]
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao ler o arquivo %s: %v\n", filePath, err)
		os.Exit(exitCompile)
	}
	os.MkdirAll(*outDir, 0755)

	res := compiler.Compile(string(data))
	fmt.Printf("Compilando %s\n", filePath)

	// ── 1. Análise léxica ──────────────────────────────────────────
	if *showTokens {
		header("ANÁLISE LÉXICA")
		printTokens(res.Tokens)
	}
	tokJSON, _ := json.MarshalIndent(res.Tokens, "", "  ")
	write("tokens.json", string(tokJSON))
	if res.FailedAt == compiler.StageLexer {
		fail("léxica", res.LexErrors)
	}

	// ── 2. Análise sintática ───────────────────────────────────────
	if res.FailedAt == compiler.StageParser {
		header("ANÁLISE SINTÁTICA")
		fail("sintática", res.ParseErrors)
	}
	write("ast.json", res.AST.ToJSON(0))
	if *showAST {
		header("ANÁLISE SINTÁTICA — AST")
		res.AST.Print(0)
	}

	// ── 3. Análise semântica ───────────────────────────────────────
	write("relatorio_semantico.txt", res.Semantic.String())
	write("tabela_simbolos.json", semantic.SymbolsToJSON(res.Symbols, 0))
	if *showSem || res.FailedAt == compiler.StageSemantic {
		header("ANÁLISE SEMÂNTICA")
		fmt.Println("Tabela de símbolos:")
		fmt.Println(semantic.SymbolsTable(res.Symbols))
		fmt.Print(res.Semantic.String())
	}
	if res.FailedAt == compiler.StageSemantic {
		fmt.Fprintln(os.Stderr, "\nCompilação interrompida na etapa semântica.")
		os.Exit(exitCompile)
	}

	// ── 4. Código intermediário ────────────────────────────────────
	write("codigo_intermediario.txt", res.IR.String())
	if *showIR {
		header(fmt.Sprintf("CÓDIGO INTERMEDIÁRIO (três endereços, %d instruções)", res.IR.InstrCount()))
		fmt.Print(res.IR.String())
	}

	// ── 5. Otimização ──────────────────────────────────────────────
	optText := "Resumo:\n" + res.OptReport.Summary() + "\nTransformações:\n" + strings.Join(res.OptReport.Log, "\n") + "\n"
	write("codigo_otimizado.txt", res.Optimized.String())
	write("otimizacao.txt", optText)
	if *showOpt {
		header(fmt.Sprintf("CÓDIGO OTIMIZADO (%d instruções)", res.Optimized.InstrCount()))
		fmt.Print(res.Optimized.String())
		fmt.Println("\nOtimizações aplicadas:")
		fmt.Print(res.OptReport.Summary())
	}
	if *explain {
		header("TRANSFORMAÇÕES DO OTIMIZADOR")
		for _, l := range res.OptReport.Log {
			fmt.Println("  " + l)
		}
	}

	// ── 6. Execução ────────────────────────────────────────────────
	if !*doRun && !*compare {
		return
	}
	if !res.HasMain() {
		header("EXECUÇÃO")
		fmt.Println("O programa não possui func main(): nada a executar.")
		return
	}
	if *compare {
		runCompare(res)
		return
	}

	prog, label := res.Optimized, "código otimizado"
	if *noOptimize {
		prog, label = res.IR, "código NÃO otimizado"
	}
	header("EXECUÇÃO (" + label + ")")
	steps, err := compiler.Run(prog, os.Stdout, *maxSteps)
	fmt.Printf("\n[%d instruções executadas]\n", steps)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(exitRuntime)
	}
}

// runCompare executa o programa com e sem otimização e compara as saídas,
// demonstrando que a otimização preserva o comportamento observável.
func runCompare(res *compiler.Result) {
	header("COMPARAÇÃO: SEM × COM OTIMIZAÇÃO")
	outA, stepsA, errA := compiler.RunToString(res.IR, *maxSteps)
	outB, stepsB, errB := compiler.RunToString(res.Optimized, *maxSteps)

	fmt.Println("Saída sem otimização:")
	fmt.Print(indent(outA))
	fmt.Println("Saída com otimização:")
	fmt.Print(indent(outB))

	fmt.Printf("\nInstruções no código:      %5d → %5d\n", res.IR.InstrCount(), res.Optimized.InstrCount())
	fmt.Printf("Instruções executadas:     %5d → %5d\n", stepsA, stepsB)
	same := outA == outB && errString(errA) == errString(errB)
	if same {
		fmt.Println("Resultado: saídas IDÊNTICAS — a otimização preservou o comportamento.")
	} else {
		fmt.Println("Resultado: saídas DIFERENTES!")
		fmt.Printf("  erro sem otimização: %v\n  erro com otimização: %v\n", errA, errB)
		os.Exit(exitRuntime)
	}
	if errA != nil {
		fmt.Fprintln(os.Stderr, errA)
		os.Exit(exitRuntime)
	}
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	var target interface{ Unwrap() error }
	if errors.As(err, &target) {
		return target.Unwrap().Error()
	}
	return err.Error()
}

func indent(s string) string {
	if s == "" {
		return "    (nenhuma saída)\n"
	}
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	return "    " + strings.Join(lines, "\n    ") + "\n"
}

func header(title string) {
	fmt.Printf("\n=== %s ===\n", title)
}

func printTokens(tokens []lexer.TokenData) {
	fmt.Printf("%-20s %-20s %s\n", "TIPO DE TOKEN", "LEXEMA", "LINHA")
	fmt.Println(strings.Repeat("-", 50))
	for _, tok := range tokens {
		if tok.Type != lexer.EOF {
			fmt.Printf("%-20s %-20s %d\n", tok.Type.String(), tok.Lexeme, tok.Line)
		}
	}
}

func fail(stage string, errs []error) {
	fmt.Printf("\n%d erro(s) na análise %s:\n", len(errs), stage)
	for _, e := range errs {
		fmt.Printf("  - %v\n", e)
	}
	fmt.Fprintf(os.Stderr, "\nCompilação interrompida na etapa %s.\n", stage)
	os.Exit(exitCompile)
}

func write(name, content string) {
	os.WriteFile(filepath.Join(*outDir, name), []byte(content), 0644)
}
