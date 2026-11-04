package ir

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// =====================================================================
// VALORES E AVALIAÇÃO DE OPERAÇÕES
// =====================================================================
// Um valor em tempo de execução é um valor nativo do Go: int, float64,
// string ou bool. EvalBinary/EvalUnary definem a semântica de cada
// operador e são usadas TANTO pelo otimizador (para dobrar constantes)
// QUANTO pelo interpretador (para executar). Assim uma constante
// calculada em tempo de compilação tem sempre o mesmo valor que teria
// em tempo de execução.
// =====================================================================

// Value é um valor de tempo de execução.
type Value = interface{}

// ErrDivZero é retornado em divisão (ou resto) inteira por zero.
var ErrDivZero = errors.New("divisão inteira por zero")

// IsConst indica se o operando é uma constante (e não um nome).
func IsConst(op string) bool {
	if op == "" {
		return false
	}
	if op == "true" || op == "false" {
		return true
	}
	c := op[0]
	return c == '"' || c == '-' || (c >= '0' && c <= '9')
}

// IsGlobal indica se o operando é uma variável global.
func IsGlobal(op string) bool { return strings.HasPrefix(op, "@") }

// ParseConst converte um operando constante em valor.
func ParseConst(op string) (Value, error) {
	switch {
	case op == "true":
		return true, nil
	case op == "false":
		return false, nil
	case op[0] == '"':
		return strconv.Unquote(op)
	case strings.ContainsAny(op, ".eE"):
		return strconv.ParseFloat(op, 64)
	default:
		v, err := strconv.ParseInt(op, 10, 64)
		return int(v), err
	}
}

// FormatConst converte um valor em operando constante. Floats sempre
// levam '.' ou expoente para não serem confundidos com inteiros.
// Retorna false para valores sem representação literal (Inf, NaN).
func FormatConst(v Value) (string, bool) {
	switch x := v.(type) {
	case int:
		return strconv.Itoa(x), true
	case float64:
		if math.IsInf(x, 0) || math.IsNaN(x) {
			return "", false
		}
		s := strconv.FormatFloat(x, 'g', -1, 64)
		if !strings.ContainsAny(s, ".eE") {
			s += ".0"
		}
		return s, true
	case string:
		return strconv.Quote(x), true
	case bool:
		return strconv.FormatBool(x), true
	}
	return "", false
}

// EvalBinary aplica um operador binário a dois valores.
func EvalBinary(op string, a, b Value) (Value, error) {
	// Promoção defensiva int → float64 (a análise semântica já garante
	// operandos do mesmo tipo; isso só cobre conversões de constantes).
	if ai, ok := a.(int); ok {
		if _, ok := b.(float64); ok {
			a = float64(ai)
		}
	}
	if bi, ok := b.(int); ok {
		if _, ok := a.(float64); ok {
			b = float64(bi)
		}
	}

	switch x := a.(type) {
	case int:
		y, ok := b.(int)
		if !ok {
			break
		}
		switch op {
		case "+":
			return x + y, nil
		case "-":
			return x - y, nil
		case "*":
			return x * y, nil
		case "/":
			if y == 0 {
				return nil, ErrDivZero
			}
			return x / y, nil
		case "%":
			if y == 0 {
				return nil, ErrDivZero
			}
			return x % y, nil
		case "==":
			return x == y, nil
		case "!=":
			return x != y, nil
		case "<":
			return x < y, nil
		case "<=":
			return x <= y, nil
		case ">":
			return x > y, nil
		case ">=":
			return x >= y, nil
		}
	case float64:
		y, ok := b.(float64)
		if !ok {
			break
		}
		switch op {
		case "+":
			return x + y, nil
		case "-":
			return x - y, nil
		case "*":
			return x * y, nil
		case "/":
			return x / y, nil // em Go, float / 0 resulta em ±Inf (sem pânico)
		case "==":
			return x == y, nil
		case "!=":
			return x != y, nil
		case "<":
			return x < y, nil
		case "<=":
			return x <= y, nil
		case ">":
			return x > y, nil
		case ">=":
			return x >= y, nil
		}
	case string:
		y, ok := b.(string)
		if !ok {
			break
		}
		switch op {
		case "+":
			return x + y, nil
		case "==":
			return x == y, nil
		case "!=":
			return x != y, nil
		case "<":
			return x < y, nil
		case "<=":
			return x <= y, nil
		case ">":
			return x > y, nil
		case ">=":
			return x >= y, nil
		}
	case bool:
		y, ok := b.(bool)
		if !ok {
			break
		}
		switch op {
		case "==":
			return x == y, nil
		case "!=":
			return x != y, nil
		}
	}
	return nil, fmt.Errorf("operação inválida: %v %s %v", a, op, b)
}

// EvalUnary aplica um operador unário (ou conversão de tipo) a um valor.
func EvalUnary(op string, a Value) (Value, error) {
	switch op {
	case "-":
		switch x := a.(type) {
		case int:
			return -x, nil
		case float64:
			return -x, nil
		}
	case "!":
		if x, ok := a.(bool); ok {
			return !x, nil
		}
	case "int":
		switch x := a.(type) {
		case int:
			return x, nil
		case float64:
			return int(x), nil // trunca em direção a zero, como no Go
		}
	case "float64", "float32":
		switch x := a.(type) {
		case int:
			return float64(x), nil
		case float64:
			return x, nil
		}
	}
	return nil, fmt.Errorf("operação inválida: %s(%v)", op, a)
}

// IsTruthy converte o valor de uma condição para bool.
func IsTruthy(v Value) (bool, error) {
	b, ok := v.(bool)
	if !ok {
		return false, fmt.Errorf("condição não booleana: %v", v)
	}
	return b, nil
}
