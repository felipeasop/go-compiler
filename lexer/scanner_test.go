package lexer

import (
	"strings"
	"testing"
)

func lexemes(t *testing.T, src string) []string {
	t.Helper()
	tokens, errs := NewScanner(src).Tokenize()
	if len(errs) > 0 {
		t.Fatalf("erros inesperados: %v", errs)
	}
	var out []string
	for _, tok := range tokens {
		out = append(out, tok.Lexeme)
	}
	return out
}

func TestTokensBasicos(t *testing.T) {
	got := strings.Join(lexemes(t, `x := 10*2 + 3.5 // comentário
y += "a\tb"`), " ")
	want := `x := 10 * 2 + 3.5 ; y += "a\tb" ; EOF`
	if got != want {
		t.Errorf("obtido %q\nesperado %q", got, want)
	}
}

// Inserção automática de ';': depois de identificador/literal/)/} no fim
// da linha, mas não depois de um operador.
func TestInsercaoDePontoEVirgula(t *testing.T) {
	got := strings.Join(lexemes(t, "a := x +\n  y\nreturn\n}"), " ")
	want := "a := x + y ; return ; } ; EOF"
	if got != want {
		t.Errorf("obtido %q\nesperado %q", got, want)
	}
}

func TestPalavrasReservadas(t *testing.T) {
	tokens, _ := NewScanner("func for if else return var break continue").Tokenize()
	want := []Token{FUNC, FOR, IF, ELSE, RETURN, VAR, BREAK, CONTINUE}
	for i, w := range want {
		if tokens[i].Type != w {
			t.Errorf("token %d: obtido %s, esperado %s", i, tokens[i].Type, w)
		}
	}
}

func TestComentarioDeBlocoContaLinhas(t *testing.T) {
	tokens, _ := NewScanner("/* a\nb\nc */ x").Tokenize()
	if tokens[0].Lexeme != "x" || tokens[0].Line != 3 {
		t.Errorf("esperava x na linha 3, obteve %q na linha %d", tokens[0].Lexeme, tokens[0].Line)
	}
}

// O lexer não para no primeiro erro: reporta todos.
func TestVariosErros(t *testing.T) {
	_, errs := NewScanner("a @ b\nc $ d\n\"aberta\n/* sem fim").Tokenize()
	if len(errs) != 4 {
		t.Fatalf("esperava 4 erros, obteve %d: %v", len(errs), errs)
	}
	for i, want := range []string{"linha 1", "linha 2", "linha 3", "linha 4"} {
		if !strings.Contains(errs[i].Error(), want) {
			t.Errorf("erro %d = %q, esperava %q", i, errs[i], want)
		}
	}
}
