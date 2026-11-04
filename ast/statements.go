package ast

import (
	"fmt"
	"strings"
)

// ─── FuncNode ──────────────────────────────────────────────────────
// func Name(p1 T1, p2 T2) ReturnType { Body }
type FuncNode struct {
	Name       string
	Params     []Param
	ReturnType string // "" quando a função não retorna valor
	Body       []ASTNode
	Line       int
}

// Param é um parâmetro formal de função.
type Param struct {
	Name     string
	TypeName string
	Line     int
}

func (n *FuncNode) signature() string {
	params := make([]string, len(n.Params))
	for i, p := range n.Params {
		params[i] = p.Name + " " + p.TypeName
	}
	sig := fmt.Sprintf("func %s(%s)", n.Name, strings.Join(params, ", "))
	if n.ReturnType != "" {
		sig += " " + n.ReturnType
	}
	return sig
}

func (n *FuncNode) Print(indent int) {
	fmt.Printf("%sFuncNode (%s)\n", pad(indent), n.signature())
	for _, s := range n.Body {
		s.Print(indent + 2)
	}
}

func (n *FuncNode) ToJSON(indent int) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s{\n%s\"tipo\": \"FuncNode\",\n%s\"nome\": %q,\n", jpd(indent), jpd(indent+1), jpd(indent+1), n.Name))
	sb.WriteString(fmt.Sprintf("%s\"parametros\": [", jpd(indent+1)))
	for i, p := range n.Params {
		sb.WriteString(fmt.Sprintf("{\"nome\": %q, \"tipo\": %q}", p.Name, p.TypeName))
		if i < len(n.Params)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteString("],\n")
	sb.WriteString(fmt.Sprintf("%s\"retorno\": %q,\n%s\"corpo\": [\n", jpd(indent+1), n.ReturnType, jpd(indent+1)))
	writeJSONList(&sb, n.Body, indent+2)
	sb.WriteString(fmt.Sprintf("%s]\n%s}", jpd(indent+1), jpd(indent)))
	return sb.String()
}

// ─── VarDeclNode ───────────────────────────────────────────────────
// var x T = expr  |  var x T  |  var x = expr  |  x := expr
type VarDeclNode struct {
	Name        string
	TypeName    string // "" quando o tipo é inferido do inicializador
	Initializer ASTNode
	Line        int // usado pelo Semantic Analyzer no relatório de erros
}

func (n *VarDeclNode) Print(indent int) {
	if n.TypeName != "" {
		fmt.Printf("%sVarDeclNode (%s %s)\n", pad(indent), n.Name, n.TypeName)
	} else {
		fmt.Printf("%sVarDeclNode (%s)\n", pad(indent), n.Name)
	}
	if n.Initializer != nil {
		n.Initializer.Print(indent + 4)
	}
}

func (n *VarDeclNode) ToJSON(indent int) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s{\n%s\"tipo\": \"VarDeclNode\",\n%s\"nome\": %q,\n%s\"tipoDeclarado\": %q",
		jpd(indent), jpd(indent+1), jpd(indent+1), n.Name, jpd(indent+1), n.TypeName))
	if n.Initializer != nil {
		sb.WriteString(fmt.Sprintf(",\n%s\"inicializador\":\n%s", jpd(indent+1), n.Initializer.ToJSON(indent+2)))
	}
	sb.WriteString(fmt.Sprintf("\n%s}", jpd(indent)))
	return sb.String()
}

// ─── AssignNode ────────────────────────────────────────────────────
// x = expr. Atribuições compostas (x += e) e x++/x-- são convertidas
// pelo parser em x = x op e.
type AssignNode struct {
	Name string
	Expr ASTNode
	Line int // usado pelo Semantic Analyzer no relatório de erros
}

func (n *AssignNode) Print(indent int) {
	fmt.Printf("%sAssignNode (%s)\n", pad(indent), n.Name)
	n.Expr.Print(indent + 4)
}

func (n *AssignNode) ToJSON(indent int) string {
	return fmt.Sprintf("%s{\n%s\"tipo\": \"AssignNode\",\n%s\"nome\": %q,\n%s\"expr\":\n%s\n%s}",
		jpd(indent), jpd(indent+1), jpd(indent+1), n.Name, jpd(indent+1), n.Expr.ToJSON(indent+2), jpd(indent))
}

// ─── PrintCallNode ─────────────────────────────────────────────────
// fmt.Println(args...) | fmt.Print(args...) | fmt.Printf(formato, args...)
type PrintCallNode struct {
	Func string // "Println", "Print" ou "Printf"
	Args []ASTNode
	Line int
}

func (n *PrintCallNode) Print(indent int) {
	fmt.Printf("%sPrintCallNode (fmt.%s)\n", pad(indent), n.Func)
	for _, a := range n.Args {
		a.Print(indent + 4)
	}
}

func (n *PrintCallNode) ToJSON(indent int) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s{\n%s\"tipo\": \"PrintCallNode\",\n%s\"funcao\": %q,\n%s\"args\": [\n",
		jpd(indent), jpd(indent+1), jpd(indent+1), "fmt."+n.Func, jpd(indent+1)))
	writeJSONList(&sb, n.Args, indent+2)
	sb.WriteString(fmt.Sprintf("%s]\n%s}", jpd(indent+1), jpd(indent)))
	return sb.String()
}

// ─── IfNode ────────────────────────────────────────────────────────
// else if é representado colocando um IfNode dentro de ElseBranch.
type IfNode struct {
	Init       ASTNode
	Condition  ASTNode
	ThenBranch []ASTNode
	ElseBranch []ASTNode
	Line       int
}

func (n *IfNode) Print(indent int) {
	fmt.Printf("%sIfNode\n", pad(indent))
	if n.Init != nil {
		fmt.Printf("%sInit:\n", pad(indent+2))
		n.Init.Print(indent + 4)
	}
	fmt.Printf("%sCondicao:\n", pad(indent+2))
	n.Condition.Print(indent + 4)
	fmt.Printf("%sThen:\n", pad(indent+2))
	for _, s := range n.ThenBranch {
		s.Print(indent + 4)
	}
	if len(n.ElseBranch) > 0 {
		fmt.Printf("%sElse:\n", pad(indent+2))
		for _, s := range n.ElseBranch {
			s.Print(indent + 4)
		}
	}
}

func (n *IfNode) ToJSON(indent int) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s{\n%s\"tipo\": \"IfNode\",\n", jpd(indent), jpd(indent+1)))
	if n.Init != nil {
		sb.WriteString(fmt.Sprintf("%s\"init\":\n%s,\n", jpd(indent+1), n.Init.ToJSON(indent+2)))
	}
	sb.WriteString(fmt.Sprintf("%s\"condicao\":\n%s,\n%s\"then\": [\n",
		jpd(indent+1), n.Condition.ToJSON(indent+2), jpd(indent+1)))
	writeJSONList(&sb, n.ThenBranch, indent+2)
	sb.WriteString(fmt.Sprintf("%s]", jpd(indent+1)))
	if len(n.ElseBranch) > 0 {
		sb.WriteString(fmt.Sprintf(",\n%s\"else\": [\n", jpd(indent+1)))
		writeJSONList(&sb, n.ElseBranch, indent+2)
		sb.WriteString(fmt.Sprintf("%s]", jpd(indent+1)))
	}
	sb.WriteString(fmt.Sprintf("\n%s}", jpd(indent)))
	return sb.String()
}

// ─── ForNode ───────────────────────────────────────────────────────
// for { }  |  for cond { }  |  for init; cond; post { }
// Init e Post ficam separados do corpo para que 'continue' execute o
// Post antes de reavaliar a condição, como na especificação do Go.
type ForNode struct {
	Init      ASTNode // nil se ausente
	Condition ASTNode // nil = laço infinito
	Post      ASTNode // nil se ausente
	Body      []ASTNode
	Line      int
}

func (n *ForNode) Print(indent int) {
	fmt.Printf("%sForNode\n", pad(indent))
	if n.Init != nil {
		fmt.Printf("%sInit:\n", pad(indent+2))
		n.Init.Print(indent + 4)
	}
	fmt.Printf("%sCondicao:\n", pad(indent+2))
	if n.Condition != nil {
		n.Condition.Print(indent + 4)
	} else {
		fmt.Printf("%s(loop infinito)\n", pad(indent+4))
	}
	if n.Post != nil {
		fmt.Printf("%sPost:\n", pad(indent+2))
		n.Post.Print(indent + 4)
	}
	fmt.Printf("%sCorpo:\n", pad(indent+2))
	for _, s := range n.Body {
		s.Print(indent + 4)
	}
}

func (n *ForNode) ToJSON(indent int) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s{\n%s\"tipo\": \"ForNode\",\n", jpd(indent), jpd(indent+1)))
	optional := func(key string, node ASTNode) {
		if node != nil {
			sb.WriteString(fmt.Sprintf("%s%q:\n%s,\n", jpd(indent+1), key, node.ToJSON(indent+2)))
		} else {
			sb.WriteString(fmt.Sprintf("%s%q: null,\n", jpd(indent+1), key))
		}
	}
	optional("init", n.Init)
	optional("condicao", n.Condition)
	optional("post", n.Post)
	sb.WriteString(fmt.Sprintf("%s\"corpo\": [\n", jpd(indent+1)))
	writeJSONList(&sb, n.Body, indent+2)
	sb.WriteString(fmt.Sprintf("%s]\n%s}", jpd(indent+1), jpd(indent)))
	return sb.String()
}

// ─── ReturnNode ────────────────────────────────────────────────────
// return [expr]
type ReturnNode struct {
	Value ASTNode // nil para return sem valor
	Line  int
}

func (n *ReturnNode) Print(indent int) {
	fmt.Printf("%sReturnNode\n", pad(indent))
	if n.Value != nil {
		n.Value.Print(indent + 4)
	}
}

func (n *ReturnNode) ToJSON(indent int) string {
	if n.Value == nil {
		return fmt.Sprintf("%s{\"tipo\": \"ReturnNode\", \"valor\": null}", jpd(indent))
	}
	return fmt.Sprintf("%s{\n%s\"tipo\": \"ReturnNode\",\n%s\"valor\":\n%s\n%s}",
		jpd(indent), jpd(indent+1), jpd(indent+1), n.Value.ToJSON(indent+2), jpd(indent))
}

// ─── BreakNode ─────────────────────────────────────────────────────
type BreakNode struct {
	Line int
}

func (n *BreakNode) Print(indent int) {
	fmt.Printf("%sBreakNode\n", pad(indent))
}
func (n *BreakNode) ToJSON(indent int) string {
	return fmt.Sprintf("%s{\"tipo\": \"BreakNode\"}", jpd(indent))
}

// ─── ContinueNode ──────────────────────────────────────────────────
type ContinueNode struct {
	Line int
}

func (n *ContinueNode) Print(indent int) {
	fmt.Printf("%sContinueNode\n", pad(indent))
}
func (n *ContinueNode) ToJSON(indent int) string {
	return fmt.Sprintf("%s{\"tipo\": \"ContinueNode\"}", jpd(indent))
}
