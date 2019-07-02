package main

import (
	"fmt"
	"os"
	"flag"
	"strings"
	"strconv"
)

const (
	PLUS = 1
	MINUS = 2
	MUL = 3
	DIV = 4
	MOD = 5
	LPAR = 6
	RPAR = 7
	INTEGER = 8
	EOF = 10
)

type lexemes struct {
	ttype int
	tvalue string
}

func init_lex() func(key string) int {
	var lex = map[string]int {
		"+" : PLUS,
		"-" : MINUS,
		"/" : DIV,
		"%" : MOD,
		"*" : MUL,
		"(" : LPAR,
		")" : RPAR,
	}
	return func(key string) int {
		return lex[key]
	}
}

func lexer(expr string) []lexemes {
	lex := init_lex()
	var tokens []lexemes
	start := -1
	for index, value := range expr {
		switch {
		case value == ' ':
			if start != -1 {
				tokens = append(tokens, lexemes{INTEGER, expr[start:index]})
				start = -1
			}
			continue
		case value >= '0' && value <= '9':
			if start == -1 {
				start = index
			}
		default:
			if start != -1 {
				tokens = append(tokens, lexemes{INTEGER, expr[start:index]})
				start = -1
			}
			new_val := lex(string(value))
			tokens = append(tokens, lexemes{new_val, string(value)})
			if new_val == 0 {
				fmt.Fprintf(os.Stderr, "Lexer Error: unexpected character '%c'\n", value)
				os.Exit(-1)
			}
		}
//		fmt.Printf("index [%d] = '%c'\n", index, value)
	}
	if start != -1 {
		tokens = append(tokens, lexemes{INTEGER, expr[start:]})
	}
	tokens = append(tokens, lexemes{EOF, "EOF"})
	return tokens
}

/*
	expr := term (( '+' | '-') term) *
	term := factor (( '+' | '-' | '%') factor) *
	factor := INTEGER | '(' expres ')'
*/

func prior1(current_token string) bool {
	available_character := "*/%"
	return strings.Contains(available_character, current_token)
}

func prior2(current_token string) bool {
	available_character := "+-"
	return strings.Contains(available_character, current_token)
}

type interpreter struct {
	index int
	length int
	tokens []lexemes
}

func (i *interpreter) Cur() *lexemes {
	if i.index == i.length {
		return nil
	} else {
		return &i.tokens[i.index]
	}
}

func (i *interpreter) Next() *lexemes {
	i.index++
	if i.index == i.length {
		return nil
	} else {
		return &i.tokens[i.index]
	}
}

func (i *interpreter) digest(needed int) {
	if needed == i.Cur().ttype {
		fmt.Printf("Digest := [%d] '%s'\n", i.Cur().ttype, i.Cur().tvalue)
		i.Next()
	} else {
		fmt.Fprintf(os.Stderr, "Syntax Error need %d has %d for %s\n",
			needed, i.Cur().ttype, i.Cur().tvalue)
	}
}

func (i *interpreter) factor() int {
	var result int
	if i.Cur().ttype == INTEGER {
		value, _ := strconv.Atoi(i.Cur().tvalue)
		i.digest(INTEGER)
		return value
	} else if (i.Cur().ttype == LPAR) {
		i.digest(LPAR)
		result = i.expr()
		i.digest(RPAR)
	}
	return result
}

func (i *interpreter) term() int {
	var result int
	result = i.factor()
	for ; prior1(i.Cur().tvalue) == true; {
		switch i.Cur().ttype {
		case MUL:
			i.digest(MUL)
			result = result * i.factor()
		case DIV:
			i.digest(DIV)
			result = result / i.factor()
		case MOD:
			i.digest(MOD)
			result = result % i.factor()
		}
	}
	return result
}

func (i *interpreter) expr() int {
	var result int
	result = i.term()
	for ; prior2(i.Cur().tvalue) == true; {
		switch i.Cur().ttype {
		case MINUS:
			i.digest(MINUS)
			result = result - i.term()
		case PLUS:
			i.digest(PLUS)
			result = result + i.term()
		}
	}
	return result
}

func (i *interpreter) Parse() {
	result := i.expr()
	if i.Cur().ttype == EOF {
		fmt.Println("Result :=", result)
	} else {
		fmt.Fprintf(os.Stderr, "Unexpected token %d '%s'\n", i.Cur().ttype, i.Cur().tvalue)
		os.Exit(-1)
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Need 2 parameters")
		os.Exit(-1)
	}
	flag.Parse()
	tokens := lexer(flag.Arg(0))
/*
	for i, j := range tokens {
		fmt.Println("Token :=", i, j)
	}
*/
	interpreter := interpreter{0, len(tokens), tokens}
	interpreter.Parse()
}
