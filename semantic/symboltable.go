package semantic

import (
	"fmt"
	"sort"
)

// =====================================================================
// TABELA DE SÍMBOLOS
// =====================================================================
// Implementada como uma pilha de escopos (scope stack). Cada escopo é um
// mapa nome → *Symbol. Ao entrar em um bloco (func, if, for) empilhamos
// um novo escopo; ao sair, desempilhamos. A busca de símbolo percorre a
// pilha de dentro para fora, permitindo shadowing (variável interna
// "esconder" uma externa de mesmo nome), como o Go permite.
//
// As funções ficam em uma tabela à parte (SemanticAnalyzer.funcs), pois
// são todas registradas antes da análise dos corpos.
// =====================================================================

// SymbolType representa o tipo inferido/declarado de um símbolo.
type SymbolType int

const (
	TypeUnknown SymbolType = iota
	TypeInt
	TypeFloat
	TypeString
	TypeBool
	TypeVoid // "tipo" de retorno de funções sem valor
)

func (t SymbolType) String() string {
	switch t {
	case TypeInt:
		return "int"
	case TypeFloat:
		return "float64"
	case TypeString:
		return "string"
	case TypeBool:
		return "bool"
	case TypeVoid:
		return "void"
	default:
		return "desconhecido"
	}
}

// TypeFromName converte o nome de um tipo do código-fonte em SymbolType.
func TypeFromName(name string) (SymbolType, bool) {
	switch name {
	case "int":
		return TypeInt, true
	case "float64", "float32":
		return TypeFloat, true
	case "string":
		return TypeString, true
	case "bool":
		return TypeBool, true
	}
	return TypeUnknown, false
}

// SymbolKind classifica o que um símbolo representa.
type SymbolKind string

const (
	KindVar   SymbolKind = "variável"
	KindParam SymbolKind = "parâmetro"
	KindFunc  SymbolKind = "função"
)

// Symbol representa uma entrada na Tabela de Símbolos.
type Symbol struct {
	Name        string     // nome do identificador
	Kind        SymbolKind // variável, parâmetro ou função
	Type        SymbolType // tipo da variável (ou de retorno, para funções)
	Signature   string     // assinatura completa (apenas funções)
	Scope       int        // nível do escopo em que foi declarada (0 = global)
	DeclLine    int        // linha de declaração
	Initialized bool       // se já recebeu algum valor
	Used        bool       // se foi lida em alguma expressão após ser declarada
}

// scope é um único nível de escopo: nome → símbolo.
type scope map[string]*Symbol

// SymbolTable é a pilha de escopos usada durante a análise semântica.
type SymbolTable struct {
	scopes []scope
}

// NewSymbolTable cria a tabela já com o escopo global (nível 0) aberto.
func NewSymbolTable() *SymbolTable {
	return &SymbolTable{scopes: []scope{make(scope)}}
}

// EnterScope empilha um novo escopo (entrada em bloco: func, if, for...).
func (st *SymbolTable) EnterScope() {
	st.scopes = append(st.scopes, make(scope))
}

// ExitScope desempilha o escopo atual (saída de bloco).
func (st *SymbolTable) ExitScope() {
	if len(st.scopes) > 1 {
		st.scopes = st.scopes[:len(st.scopes)-1]
	}
}

// currentScopeLevel retorna o índice do escopo mais interno.
func (st *SymbolTable) currentScopeLevel() int {
	return len(st.scopes) - 1
}

// Declare adiciona um símbolo ao escopo atual (mais interno).
// Retorna erro se já existir uma declaração do mesmo nome NESTE escopo
// (declaração duplicada no mesmo nível — shadowing entre escopos
// diferentes é permitido).
func (st *SymbolTable) Declare(name string, typ SymbolType, line int, initialized bool) error {
	current := st.scopes[st.currentScopeLevel()]
	if _, exists := current[name]; exists {
		return fmt.Errorf("variavel '%s' ja declarada", name)
	}
	current[name] = &Symbol{
		Name:        name,
		Kind:        KindVar,
		Type:        typ,
		Scope:       st.currentScopeLevel(),
		DeclLine:    line,
		Initialized: initialized,
		Used:        false,
	}
	return nil
}

// Lookup procura um símbolo a partir do escopo mais interno até o global.
func (st *SymbolTable) Lookup(name string) (*Symbol, bool) {
	for i := st.currentScopeLevel(); i >= 0; i-- {
		if sym, ok := st.scopes[i][name]; ok {
			return sym, true
		}
	}
	return nil, false
}

// AllSymbolsInScope retorna todos os símbolos de um nível de escopo
// específico, ordenados pela linha de declaração — usado ao fechar o
// escopo para checar variáveis não usadas.
func (st *SymbolTable) AllSymbolsInScope(level int) []*Symbol {
	if level < 0 || level >= len(st.scopes) {
		return nil
	}
	var out []*Symbol
	for _, sym := range st.scopes[level] {
		out = append(out, sym)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].DeclLine < out[j].DeclLine })
	return out
}
