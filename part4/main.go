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
		return &token{EOF, ""}, nil
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

/*

*/

func isRight(typ int) bool {
	if typ == MINUS || typ == PLUS || typ == MULT || typ == DIV {
		return true
	}
	return false
}

/*
	expres : factor ((MUL | DIV | PLUS | MINUS) Factor) *
*/

func integer(cur *token, needed int) int {
	if cur.typ != needed {
		fmt.Printf("Error parsing need %d has %d\n", cur.typ, needed)
		os.Exit(-1)
	}
	value, _ := strconv.Atoi(cur.value)
	return value
}

func factor(cur *token) int {
	return integer(cur, INTEGER)
}

func expre(lexems []*token) {
	var result, index int
	result = factor(lexems[0])
	for {
		index++
		if index == len(lexems) {
			fmt.Println(len(lexems))
		}
		if isRight(lexems[index].typ) == true {
			switch lexems[index].typ {
			case MULT:
				index++
				result = result * factor(lexems[index])
			case DIV:
				index++
				result = result / factor(lexems[index])
			case MINUS:
				index++
				result = result - factor(lexems[index])
			case PLUS:
				index++
				result = result + factor(lexems[index])
			}
		} else if lexems[index].typ == EOF {
				break
		} else {
			fmt.Println("Error parsing need Operator has", lexems[index].typ)
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
	/*
	left, _ := strconv.Atoi(expre[0].value)
	right, _ := strconv.Atoi(expre[2].value)
	fmt.Println(left, right)
	switch expre[1].typ {
	case MINUS:
		fmt.Println(left - right)
	case PLUS:
		fmt.Println(left + right)
	case MULT:
		fmt.Println(left * right)
	case DIV:
		fmt.Println(left / right)
	}
	*/
}
