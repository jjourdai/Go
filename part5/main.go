package main

import (
	"fmt"
	"flag"
	"os"
	"unicode"
	"strconv"
)

const (
	ERROR = 0
	INTEGER = 1
	PLUS = 2
	MINUS = 3
	MULT = 4
	DIV = 5
	EOF = 6
	NONE = 7
	OPERATOR = 8
)

type token struct {
	typ int
	value string
}

func (t *token) String() string {
	return fmt.Sprintf("type %d, value [%c]", t.typ, t.value)
}

func next() func(t []*token) *token {
	index := -1
	return func(t []*token) *token {
		index++
		if len(t) <= index {
			return nil
		} else {
			return t[index]
		}
	}
}

type lexer struct {
	pos int
	text string
	current_token *token
	count int
	tokenizer func(string) int
}

type lexer_error string
func (e lexer_error) Error() string {
	return fmt.Sprintf("Error lexer input %s", string(e))
}

func tokenizer() func(string) int {
	tokens := map[string]int {
		"*" : MULT,
		"/" : DIV,
		"+" : PLUS,
		"-" : MINUS,
	}
	return func(key string) int {
		return tokens[key]
	}
}

func (i *lexer) get_next_token() (*token, error) {
	text := i.text
	if i.pos > len(text) - 1 {
		i.count++
		return &token{EOF, "EOF"}, nil
	}
	current_char := text[i.pos]
	for ; i.pos < len(text) ; i.pos++ {
		current_char = text[i.pos]
		if unicode.IsSpace(rune(current_char)) == false {
			break
		}
	}
	if i.pos > len(text) - 1 {
		i.count++
		return &token{EOF, "EOF"}, nil
	}
	var str string
	for ;i.pos < len(text); {
		current_char = text[i.pos]
		if unicode.IsDigit(rune(current_char)) == false {
			break
		}
		str += string(text[i.pos])
		i.pos++
	}
	if len(str) > 0 {
		i.count++
		return &token{INTEGER, str}, nil
	}
	cur_token := i.tokenizer(string(current_char))
	if cur_token != ERROR {
		i.pos++
		i.count++
		return &token{cur_token, string(current_char)}, nil
	} else {
		return nil, lexer_error(text)
	}
}

func isRight(typ int) bool {
	if typ == MINUS || typ == PLUS || typ == MULT || typ == DIV {
		return true
	}
	return false
}

/*
	expres : term ((PLUS | MINUS) term) *
	term : factor ((MUL | DIV) Factor) *
	factor : integer
*/

func integer(cur *token, needed int) int {
	if cur.typ != needed {
		fmt.Printf("Error parsing need %d has %d\n", needed, cur.typ)
		os.Exit(-1)
	}
	value, _ := strconv.Atoi(cur.value)
	return value
}

func factor(cur *token) int {
	return integer(cur, INTEGER)
}

func expre(lexems []*token) {
	var result int
	test := next()
	result = factor(test(lexems))
	for token := test(lexems); token != nil ; token = test(lexems){
		if isRight(token.typ) == true {
			switch token.typ {
			case MULT:
				result = result * factor(test(lexems))
			case DIV:
				result = result / factor(test(lexems))
			case MINUS:
				result = result - factor(test(lexems))
			case PLUS:
				result = result + factor(test(lexems))
			}
		} else if token.typ == EOF {
				break
		} else {
			fmt.Println("Error parsing need Operator has", token.typ)
			os.Exit(-1)
		}
	}
	fmt.Println("Result := ", result)
}

func parse(lexems []*token) {
	expre(lexems)
}

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "Error: need just 1 params")
		os.Exit(-1)
	}
	inter := lexer{0, args[0], nil, 0, tokenizer()}
	lexemes := make([]*token, 0)
	for {
		token, err := inter.get_next_token()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(-1)
		}
		lexemes = append(lexemes, token)
		if token != nil && token.typ == EOF {
			break
		}
	}
	parse(lexemes)
}
