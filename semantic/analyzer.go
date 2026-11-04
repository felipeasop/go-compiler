package semantic

import (
	"fmt"
	"strconv"
	"strings"

	"go-compiler/ast"
)

// =====================================================================
// ANALISADOR SEMÂNTICO
// =====================================================================
// Percorre a AST produzida pelo Parser (padrão Visitor) e valida o
// programa segundo as regras semânticas da linguagem:
//   - variáveis e funções declaradas antes do uso; declaração duplicada
//   - tipos declarados válidos; compatibilidade de tipos em atribuições,
//     operadores, condições, argumentos e retornos
//   - chamadas com número/tipo de argumentos corretos
//   - função com tipo de retorno que pode terminar sem 'return'
//   - break/continue fora de laço
//   - uso antes da inicialização, variáveis não utilizadas (avisos)
//   - divisão por zero com divisor literal e código morto
//
// Conversão implícita de constantes: em Go, um literal inteiro é uma
// constante "não tipada" e pode ser usado onde se espera float64
// (ex.: var f float64 = 1, f * 2). Quando isso acontece o analisador
// reescreve o literal na AST como float (1 → 1.0), para que as etapas
// seguintes (código intermediário e execução) operem com o tipo correto.
// =====================================================================

// FuncSig é a assinatura de uma função declarada no programa.
type FuncSig struct {
	Name       string
	ParamNames []string
	Params     []SymbolType
	Return     SymbolType // TypeVoid quando não retorna valor
	Line       int
}

func (f *FuncSig) String() string {
	parts := make([]string, len(f.Params))
	for i, p := range f.Params {
		parts[i] = f.ParamNames[i] + " " + p.String()
	}
	s := fmt.Sprintf("func(%s)", strings.Join(parts, ", "))
	if f.Return != TypeVoid {
		s += " " + f.Return.String()
	}
	return s
}

// SemanticAnalyzer percorre a AST e produz um Report com erros e avisos.
type SemanticAnalyzer struct {
	symbols *SymbolTable
	report  *Report
	funcs   map[string]*FuncSig

	currentFunc *FuncSig
	loopDepth   int

	// allDeclared guarda TODOS os símbolos já vistos (mesmo após o escopo
	// deles fechar) para a checagem final de "declarada e nunca utilizada"
	// e para exportação da Tabela de Símbolos completa em JSON.
	allDeclared []*Symbol
}

// NewSemanticAnalyzer cria um analisador pronto para uso.
func NewSemanticAnalyzer() *SemanticAnalyzer {
	return &SemanticAnalyzer{
		symbols: NewSymbolTable(),
		report:  &Report{},
		funcs:   make(map[string]*FuncSig),
	}
}

// Analyze é o ponto de entrada da análise semântica. Retorna o relatório
// completo (erros + avisos) ao final do percurso da AST.
func (s *SemanticAnalyzer) Analyze(program *ast.ProgramNode) *Report {
	s.visitProgram(program)
	s.checkUnusedInScope(0) // variáveis globais não utilizadas
	return s.report
}

// Symbols retorna todos os símbolos vistos durante a análise (de todos
// os escopos, mesmo os já fechados). Usado pelo main para exportar a
// Tabela de Símbolos completa em JSON ao final da execução.
func (s *SemanticAnalyzer) Symbols() []*Symbol {
	return s.allDeclared
}

// ─── visitProgram ──────────────────────────────────────────────────

func (s *SemanticAnalyzer) visitProgram(n *ast.ProgramNode) {
	// 1ª passada: registra as assinaturas de todas as funções, para que
	// uma função possa chamar outra declarada mais abaixo (ou a si mesma).
	for _, f := range n.Functions {
		if fn, ok := f.(*ast.FuncNode); ok {
			s.declareFunc(fn)
		}
	}

	// Declarações globais (var no nível do pacote) entram no escopo 0.
	for _, g := range n.Globals {
		if vd, ok := g.(*ast.VarDeclNode); ok {
			s.visitDeclaration(vd)
		}
	}

	// 2ª passada: analisa o corpo de cada função em seu próprio escopo.
	for _, f := range n.Functions {
		if fn, ok := f.(*ast.FuncNode); ok {
			s.visitFunc(fn)
		}
	}
}

func (s *SemanticAnalyzer) declareFunc(fn *ast.FuncNode) {
	if prev, exists := s.funcs[fn.Name]; exists {
		s.report.AddError(fn.Line, "função '%s' já declarada na linha %d", fn.Name, prev.Line)
		return
	}
	sig := &FuncSig{Name: fn.Name, Return: TypeVoid, Line: fn.Line}
	for _, p := range fn.Params {
		sig.ParamNames = append(sig.ParamNames, p.Name)
		sig.Params = append(sig.Params, s.resolveType(p.TypeName, p.Line))
	}
	if fn.ReturnType != "" {
		sig.Return = s.resolveType(fn.ReturnType, fn.Line)
	}
	if fn.Name == "main" && (len(sig.Params) > 0 || sig.Return != TypeVoid) {
		s.report.AddError(fn.Line, "func main não pode ter parâmetros nem tipo de retorno")
	}
	s.funcs[fn.Name] = sig
	s.allDeclared = append(s.allDeclared, &Symbol{
		Name: fn.Name, Kind: KindFunc, Type: sig.Return, Signature: sig.String(),
		Scope: 0, DeclLine: fn.Line, Initialized: true, Used: true,
	})
}

// resolveType converte o nome de um tipo do código-fonte em SymbolType.
func (s *SemanticAnalyzer) resolveType(name string, line int) SymbolType {
	t, ok := TypeFromName(name)
	if !ok {
		s.report.AddError(line, "tipo desconhecido '%s' (suportados: int, float64, float32, string, bool)", name)
		return TypeUnknown
	}
	return t
}

// visitFunc entra em um novo escopo para o corpo da função.
func (s *SemanticAnalyzer) visitFunc(n *ast.FuncNode) {
	sig := s.funcs[n.Name]
	if sig == nil || sig.Line != n.Line { // função duplicada: assinatura é da primeira
		sig = &FuncSig{Name: n.Name, Return: TypeVoid, Line: n.Line}
	}
	s.currentFunc = sig
	s.loopDepth = 0

	s.symbols.EnterScope()
	level := s.symbols.currentScopeLevel()

	for i, p := range n.Params {
		typ := TypeUnknown
		if i < len(sig.Params) {
			typ = sig.Params[i]
		}
		if sym := s.declare(p.Name, typ, p.Line, true); sym != nil {
			sym.Kind = KindParam
			sym.Used = true // Go não exige o uso de parâmetros
		}
	}

	s.visitBlock(n.Body)

	if sig.Return != TypeVoid && !terminates(n.Body) {
		s.report.AddError(n.Line, "função '%s' deve retornar '%s', mas pode terminar sem 'return'", n.Name, sig.Return)
	}

	s.checkUnusedInScope(level)
	s.symbols.ExitScope()
	s.currentFunc = nil
}

// terminates indica se uma lista de instruções sempre termina com
// 'return' (regra de "terminating statement" simplificada do Go).
func terminates(stmts []ast.ASTNode) bool {
	if len(stmts) == 0 {
		return false
	}
	switch n := stmts[len(stmts)-1].(type) {
	case *ast.ReturnNode:
		return true
	case *ast.IfNode:
		return len(n.ElseBranch) > 0 && terminates(n.ThenBranch) && terminates(n.ElseBranch)
	case *ast.ForNode:
		return n.Condition == nil && !hasBreak(n.Body)
	}
	return false
}

// hasBreak procura um 'break' que pertença ao laço atual (não entra em
// laços aninhados, cujo break não afeta o laço externo).
func hasBreak(stmts []ast.ASTNode) bool {
	for _, st := range stmts {
		switch n := st.(type) {
		case *ast.BreakNode:
			return true
		case *ast.IfNode:
			if hasBreak(n.ThenBranch) || hasBreak(n.ElseBranch) {
				return true
			}
		}
	}
	return false
}

// ─── código morto ─────────────────────────────────────────────────

// visitBlock percorre uma sequência de statements de um mesmo bloco.
// Além de visitar cada um, verifica código morto: qualquer instrução
// que apareça depois de um return/break/continue no mesmo bloco nunca
// será alcançada em tempo de execução.
func (s *SemanticAnalyzer) visitBlock(stmts []ast.ASTNode) {
	terminated := false
	var lastTerminator string

	for _, stmt := range stmts {
		if terminated {
			if line := ast.LineOf(stmt); line > 0 {
				s.report.AddWarning(line, "código inalcançável (instrução após '%s')", lastTerminator)
			}
			// Segue analisando o statement morto mesmo assim: se ele usar
			// uma variável inexistente, por exemplo, isso continua sendo
			// um erro real que vale a pena reportar.
		}

		s.visitStatement(stmt)

		if kind, ok := terminatorKind(stmt); ok {
			terminated = true
			lastTerminator = kind
		}
	}
}

// visitScopedBlock analisa um bloco { ... } em um escopo próprio.
func (s *SemanticAnalyzer) visitScopedBlock(stmts []ast.ASTNode) {
	s.symbols.EnterScope()
	level := s.symbols.currentScopeLevel()
	s.visitBlock(stmts)
	s.checkUnusedInScope(level)
	s.symbols.ExitScope()
}

// terminatorKind reporta se um statement encerra o fluxo do bloco
// (return, break, continue) e qual é a palavra-chave responsável.
func terminatorKind(node ast.ASTNode) (string, bool) {
	switch node.(type) {
	case *ast.ReturnNode:
		return "return", true
	case *ast.BreakNode:
		return "break", true
	case *ast.ContinueNode:
		return "continue", true
	}
	return "", false
}

// ─── dispatch genérico de statement ─────────────────────────────────

func (s *SemanticAnalyzer) visitStatement(node ast.ASTNode) {
	switch n := node.(type) {
	case *ast.VarDeclNode:
		s.visitDeclaration(n)
	case *ast.AssignNode:
		s.visitAssignment(n)
	case *ast.PrintCallNode:
		s.visitPrint(n)
	case *ast.IfNode:
		s.visitIf(n)
	case *ast.ForNode:
		s.visitFor(n)
	case *ast.ReturnNode:
		s.visitReturn(n)
	case *ast.BreakNode:
		if s.loopDepth == 0 {
			s.report.AddError(n.Line, "'break' fora de um laço")
		}
	case *ast.ContinueNode:
		if s.loopDepth == 0 {
			s.report.AddError(n.Line, "'continue' fora de um laço")
		}
	case *ast.CallNode:
		s.visitCall(n, false)
	default:
		if node != nil {
			s.visitExpression(node)
		}
	}
}

// ─── 1. Declaração de variável ──────────────────────────────────────

func (s *SemanticAnalyzer) visitDeclaration(n *ast.VarDeclNode) {
	typ := TypeUnknown
	if n.TypeName != "" {
		typ = s.resolveType(n.TypeName, n.Line)
	}

	if n.Initializer != nil {
		if n.TypeName != "" {
			s.checkAssignable(typ, &n.Initializer, n.Line, fmt.Sprintf("variável '%s'", n.Name))
		} else {
			typ = s.visitExpression(n.Initializer)
			if typ == TypeVoid {
				s.report.AddError(n.Line, "expressão sem valor usada para inicializar '%s'", n.Name)
				typ = TypeUnknown
			}
		}
	}

	s.declare(n.Name, typ, n.Line, n.Initializer != nil)
}

// declare insere uma variável no escopo atual, reportando duplicidade.
func (s *SemanticAnalyzer) declare(name string, typ SymbolType, line int, initialized bool) *Symbol {
	if err := s.symbols.Declare(name, typ, line, initialized); err != nil {
		s.report.AddError(line, "variável '%s' já declarada neste escopo", name)
		return nil
	}
	sym, _ := s.symbols.Lookup(name)
	s.allDeclared = append(s.allDeclared, sym)
	return sym
}

// ─── 2. Atribuição ───────────────────────────────────────────────────

func (s *SemanticAnalyzer) visitAssignment(n *ast.AssignNode) {
	// _ = expr: identificador vazio, descarta o valor (marca variáveis como usadas).
	if n.Name == "_" {
		s.visitExpression(n.Expr)
		return
	}

	sym, found := s.symbols.Lookup(n.Name)
	if !found {
		if _, isFunc := s.funcs[n.Name]; isFunc {
			s.report.AddError(n.Line, "não é possível atribuir valor à função '%s'", n.Name)
		} else {
			s.report.AddError(n.Line, "variável '%s' não declarada", n.Name)
		}
		// Ainda assim analisamos o lado direito para não perder outros erros.
		s.visitExpression(n.Expr)
		return
	}

	s.checkAssignable(sym.Type, &n.Expr, n.Line, fmt.Sprintf("variável '%s'", n.Name))
	sym.Initialized = true
}

// checkAssignable verifica se a expressão em *expr pode ser atribuída a
// um destino do tipo target. Aplica a conversão implícita de constantes
// inteiras para float64 quando permitido.
func (s *SemanticAnalyzer) checkAssignable(target SymbolType, expr *ast.ASTNode, line int, what string) {
	got := s.visitExpression(*expr)
	switch {
	case got == TypeVoid:
		s.report.AddError(line, "expressão sem valor atribuída a %s", what)
	case target == TypeUnknown || got == TypeUnknown || target == got:
		// ok (ou erro já reportado em outro ponto)
	case target == TypeFloat && got == TypeInt && isUntypedConst(*expr):
		coerceToFloat(expr)
	default:
		s.report.AddError(line, "tipos incompatíveis: %s é do tipo '%s', mas recebeu valor do tipo '%s'",
			what, target, got)
	}
}

// ─── 3. print (fmt.Println / Print / Printf) ────────────────────────

func (s *SemanticAnalyzer) visitPrint(n *ast.PrintCallNode) {
	for i, arg := range n.Args {
		t := s.visitExpression(arg)
		if t == TypeVoid {
			s.report.AddError(n.Line, "argumento %d de fmt.%s não tem valor", i+1, n.Func)
		}
		if n.Func == "Printf" && i == 0 && t != TypeString && t != TypeUnknown {
			s.report.AddError(n.Line, "o primeiro argumento de fmt.Printf deve ser string, recebeu '%s'", t)
		}
	}
	if n.Func == "Printf" && len(n.Args) == 0 {
		s.report.AddError(n.Line, "fmt.Printf exige ao menos o argumento de formato")
	}
}

// ─── 4. if / else ────────────────────────────────────────────────────

func (s *SemanticAnalyzer) visitIf(n *ast.IfNode) {
	// Escopo do if: contém a variável do init (visível no then e no else).
	s.symbols.EnterScope()
	level := s.symbols.currentScopeLevel()

	if n.Init != nil {
		s.visitStatement(n.Init)
	}
	s.checkCondition(n.Condition, n.Line, "if")
	s.visitScopedBlock(n.ThenBranch)
	if len(n.ElseBranch) > 0 {
		s.visitScopedBlock(n.ElseBranch)
	}

	s.checkUnusedInScope(level)
	s.symbols.ExitScope()
}

// checkCondition exige que a condição de if/for seja booleana.
func (s *SemanticAnalyzer) checkCondition(cond ast.ASTNode, line int, stmt string) {
	if cond == nil {
		return
	}
	if t := s.visitExpression(cond); t != TypeBool && t != TypeUnknown {
		s.report.AddError(line, "condição do %s deve ser 'bool', mas é '%s'", stmt, t)
	}
}

// ─── 5. for (while / clássico / infinito) ───────────────────────────

func (s *SemanticAnalyzer) visitFor(n *ast.ForNode) {
	s.symbols.EnterScope()
	level := s.symbols.currentScopeLevel()

	if n.Init != nil {
		s.visitStatement(n.Init)
	}
	s.checkCondition(n.Condition, n.Line, "for")
	if n.Post != nil {
		s.visitStatement(n.Post)
	}

	s.loopDepth++
	s.visitScopedBlock(n.Body)
	s.loopDepth--

	s.checkUnusedInScope(level)
	s.symbols.ExitScope()
}

// ─── 6. return ───────────────────────────────────────────────────────

func (s *SemanticAnalyzer) visitReturn(n *ast.ReturnNode) {
	fn := s.currentFunc
	if fn == nil {
		return
	}
	switch {
	case fn.Return == TypeVoid && n.Value != nil:
		s.report.AddError(n.Line, "função '%s' não retorna valor, mas 'return' tem expressão", fn.Name)
		s.visitExpression(n.Value)
	case fn.Return != TypeVoid && n.Value == nil:
		s.report.AddError(n.Line, "função '%s' deve retornar um valor do tipo '%s'", fn.Name, fn.Return)
	case n.Value != nil:
		s.checkAssignable(fn.Return, &n.Value, n.Line, fmt.Sprintf("retorno de '%s'", fn.Name))
	}
}

// ─── 7. Chamadas de função ───────────────────────────────────────────

// conversions são os tipos que podem ser usados como função de conversão.
var conversions = map[string]SymbolType{"int": TypeInt, "float64": TypeFloat, "float32": TypeFloat}

// visitCall valida uma chamada e retorna o tipo do resultado. wantValue
// indica se a chamada aparece dentro de uma expressão (precisa de valor).
func (s *SemanticAnalyzer) visitCall(n *ast.CallNode, wantValue bool) SymbolType {
	if target, ok := conversions[n.Name]; ok {
		if len(n.Args) != 1 {
			s.report.AddError(n.Line, "conversão %s(...) exige exatamente 1 argumento", n.Name)
			return target
		}
		if t := s.visitExpression(n.Args[0]); t != TypeInt && t != TypeFloat && t != TypeUnknown {
			s.report.AddError(n.Line, "não é possível converter '%s' para %s", t, n.Name)
		}
		return target
	}

	sig, ok := s.funcs[n.Name]
	if !ok {
		if _, isVar := s.symbols.Lookup(n.Name); isVar {
			s.report.AddError(n.Line, "'%s' é uma variável, não uma função", n.Name)
		} else {
			s.report.AddError(n.Line, "função '%s' não declarada", n.Name)
		}
		for _, a := range n.Args {
			s.visitExpression(a)
		}
		return TypeUnknown
	}

	if len(n.Args) != len(sig.Params) {
		s.report.AddError(n.Line, "função '%s' espera %d argumento(s), mas recebeu %d",
			n.Name, len(sig.Params), len(n.Args))
	}
	for i := range n.Args {
		if i < len(sig.Params) {
			s.checkAssignable(sig.Params[i], &n.Args[i], n.Line,
				fmt.Sprintf("o parâmetro '%s' de '%s'", sig.ParamNames[i], n.Name))
		} else {
			s.visitExpression(n.Args[i])
		}
	}

	if wantValue && sig.Return == TypeVoid {
		s.report.AddError(n.Line, "função '%s' não retorna valor e não pode ser usada em expressão", n.Name)
		return TypeUnknown
	}
	return sig.Return
}

// ─── 8. Expressões ───────────────────────────────────────────────────
// Percorre a expressão, verifica uso de variáveis (declaradas? já
// inicializadas?) e retorna o tipo resultante para checagem de tipos
// no nó pai (ex.: atribuição, condição de if/for).

func (s *SemanticAnalyzer) visitExpression(node ast.ASTNode) SymbolType {
	switch n := node.(type) {

	case *ast.LiteralNode:
		t := literalType(n.Value)
		if t == TypeUnknown {
			s.report.AddError(n.Line, "literal '%s' não suportado", n.Value)
		}
		return t

	case *ast.VariableNode:
		sym, found := s.symbols.Lookup(n.Name)
		if !found {
			if _, isFunc := s.funcs[n.Name]; isFunc {
				s.report.AddError(n.Line, "função '%s' usada como valor (faltou chamar com '()'?)", n.Name)
			} else {
				s.report.AddError(n.Line, "variável '%s' não declarada", n.Name)
			}
			return TypeUnknown
		}

		if !sym.Initialized {
			s.report.AddWarning(n.Line, "variável '%s' utilizada antes de receber valor (vale o valor zero do tipo)", n.Name)
		}
		sym.Used = true
		return sym.Type

	case *ast.BinaryOpNode:
		return s.visitBinary(n)

	case *ast.UnaryOpNode:
		t := s.visitExpression(n.Operand)
		switch {
		case t == TypeUnknown:
			return t
		case n.OpStr == "-" && (t == TypeInt || t == TypeFloat):
			return t
		case n.OpStr == "!" && t == TypeBool:
			return t
		}
		s.report.AddError(n.Line, "operador unário '%s' inválido para o tipo '%s'", n.OpStr, t)
		return TypeUnknown

	case *ast.CallNode:
		return s.visitCall(n, true)

	default:
		return TypeUnknown
	}
}

func (s *SemanticAnalyzer) visitBinary(n *ast.BinaryOpNode) SymbolType {
	lt := s.visitExpression(n.Left)
	rt := s.visitExpression(n.Right)

	// Constante inteira misturada com float: converte a constante.
	if lt == TypeFloat && rt == TypeInt && isUntypedConst(n.Right) {
		coerceToFloat(&n.Right)
		rt = TypeFloat
	} else if rt == TypeFloat && lt == TypeInt && isUntypedConst(n.Left) {
		coerceToFloat(&n.Left)
		lt = TypeFloat
	}

	// Divisão por zero detectável em tempo de compilação: só o caso em
	// que o divisor é um literal numérico zero. Divisão por variável só
	// é conhecida em tempo de execução (o interpretador reporta).
	if n.OpStr == "/" || n.OpStr == "%" {
		if lit, ok := n.Right.(*ast.LiteralNode); ok && isZeroLiteral(lit.Value) {
			s.report.AddError(n.Line, "divisão por zero")
		}
	}

	isRelational := false
	ok := false
	switch n.OpStr {
	case "+":
		ok = lt == rt && (isNumeric(lt) || lt == TypeString)
	case "-", "*", "/":
		ok = lt == rt && isNumeric(lt)
	case "%":
		ok = lt == rt && lt == TypeInt
	case "==", "!=":
		isRelational, ok = true, lt == rt
	case "<", ">", "<=", ">=":
		isRelational, ok = true, lt == rt && (isNumeric(lt) || lt == TypeString)
	case "&&", "||":
		isRelational, ok = true, lt == TypeBool && rt == TypeBool
	}

	if lt == TypeUnknown || rt == TypeUnknown {
		// Erro já reportado em um operando: evita mensagens em cascata.
		if isRelational {
			return TypeBool
		}
		return TypeUnknown
	}
	if !ok {
		s.report.AddError(n.Line, "operador '%s' inválido entre '%s' e '%s'", n.OpStr, lt, rt)
		if isRelational {
			return TypeBool
		}
		return TypeUnknown
	}
	if isRelational {
		return TypeBool
	}
	return lt
}

// ─── Auxiliares ──────────────────────────────────────────────────────

// checkUnusedInScope varre os símbolos de um escopo específico e emite
// o aviso de "declarada mas nunca utilizada" para os que não foram
// lidos em nenhuma expressão.
func (s *SemanticAnalyzer) checkUnusedInScope(level int) {
	for _, sym := range s.symbols.AllSymbolsInScope(level) {
		if !sym.Used {
			s.report.AddWarning(sym.DeclLine, "variável '%s' declarada mas nunca utilizada", sym.Name)
		}
	}
}

func isNumeric(t SymbolType) bool { return t == TypeInt || t == TypeFloat }

// isUntypedConst reconhece expressões formadas apenas por literais
// inteiros (ex.: 2, -3, 10*2+5), que em Go são constantes não tipadas.
func isUntypedConst(node ast.ASTNode) bool {
	switch n := node.(type) {
	case *ast.LiteralNode:
		return literalType(n.Value) == TypeInt
	case *ast.UnaryOpNode:
		return n.OpStr == "-" && isUntypedConst(n.Operand)
	case *ast.BinaryOpNode:
		switch n.OpStr {
		case "+", "-", "*", "/", "%":
			return isUntypedConst(n.Left) && isUntypedConst(n.Right)
		}
	}
	return false
}

// coerceToFloat converte uma constante inteira para float64. Literais são
// reescritos diretamente (2 → 2.0); expressões constantes são envolvidas
// em float64(...), preservando a semântica do Go (a expressão é avaliada
// como inteira e depois convertida).
func coerceToFloat(node *ast.ASTNode) {
	if lit, ok := (*node).(*ast.LiteralNode); ok {
		if v, ok := intLiteralValue(lit.Value); ok {
			lit.Value = strconv.FormatInt(v, 10) + ".0"
			return
		}
	}
	*node = &ast.CallNode{Name: "float64", Args: []ast.ASTNode{*node}, Line: ast.LineOf(*node)}
}

// literalType infere o SymbolType de um LiteralNode a partir do próprio
// lexema salvo pelo parser.
func literalType(value string) SymbolType {
	if value == "true" || value == "false" {
		return TypeBool
	}
	if len(value) >= 2 && (value[0] == '"' || value[0] == '`') {
		return TypeString
	}
	if len(value) >= 3 && value[0] == '\'' {
		return TypeInt // rune: em Go é um alias de int32
	}
	if strings.Contains(value, ".") {
		if _, err := strconv.ParseFloat(value, 64); err == nil {
			return TypeFloat
		}
	}
	if _, err := strconv.ParseInt(value, 10, 64); err == nil {
		return TypeInt
	}
	return TypeUnknown
}

// intLiteralValue devolve o valor de um literal inteiro ou de caractere.
func intLiteralValue(value string) (int64, bool) {
	if len(value) >= 3 && value[0] == '\'' {
		r, _, _, err := strconv.UnquoteChar(value[1:len(value)-1], '\'')
		return int64(r), err == nil
	}
	v, err := strconv.ParseInt(value, 10, 64)
	return v, err == nil
}

// isZeroLiteral reconhece "0", "0.0" e variações equivalentes escritas
// como literal direto no código-fonte.
func isZeroLiteral(value string) bool {
	if f, err := strconv.ParseFloat(value, 64); err == nil {
		return f == 0
	}
	return false
}
