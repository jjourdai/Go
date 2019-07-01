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
type interpreter struct {
	pos int
	text string
	current_token *token
	count int
	tokenizer func(string) int
}

type syntax_error string
func (e syntax_error) Error() string {
	return fmt.Sprintf("Error Parsing input %s", string(e))
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

func (i *interpreter) get_next_token() (*token, error) {
	text := i.text
	if i.pos > len(text) - 1 {
		i.count++
		return &token{EOF, ""}, nil
	}
	current_char := text[i.pos]
	for ; i.pos < len(text) ; i.pos++ {
		current_char = text[i.pos]
		if unicode.IsSpace(rune(current_char)) == false {
			break
		}
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
		return nil, syntax_error(text)
	}
}

/*

*/

func isRight(typ, index int) bool {
	grammar := [3]int {
		INTEGER,
		OPERATOR,
		INTEGER,
	}
	if grammar[index] == OPERATOR {
		if typ == MINUS || typ == PLUS || typ == MULT || typ == DIV {
			return true
		}
	} else if grammar[index] == typ {
		return true
	}
	return false
}

func parse(lexems []*token) {
	var left, right, operator, index int
	for _, j := range lexems {
		if isRight(j.typ, index) == true {
			switch j.typ {
			case INTEGER:
				if index == 0 {
					left, _ = strconv.Atoi(j.value)
				} else if index == 2 {
					right, _ = strconv.Atoi(j.value)
				}
			default:
				operator = j.typ
			}
			if index == 2 {
				switch operator {
				case MINUS:
					left = left - right
				case PLUS:
					left = left + right
				case MULT:
					left = left * right
				case DIV:
					left = left / right
				}
			}
			index++;
			if index == 3 {
				index = 1
			}
		} else {
			fmt.Fprintf(os.Stderr, "Error: type == %d\n", j.typ);
			os.Exit(-1)
		}
	}
	if (index != 1) {
		fmt.Fprintln(os.Stderr, "Syntax Error");
		os.Exit(-1)
	}
	fmt.Println(left)
}

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "Error: need just 1 params")
		os.Exit(-1)
	}
	inter := interpreter{0, args[0], nil, 0, tokenizer()}
	lexemes := make([]*token, 0)
	for {
		token, err := inter.get_next_token()
		if token != nil && token.typ == EOF {
			break
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(-1)
		}
		lexemes = append(lexemes, token)
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
