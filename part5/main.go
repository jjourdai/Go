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
	LEFT = 9
	RIGHT = 10
)

type token struct {
	typ int
	value string
}

func (t *token) String() string {
	return fmt.Sprintf("type %d, value [%c]", t.typ, t.value)
}

type interpreter struct {
	index int
	tok []*token
}

func (i *interpreter) next() *token {
	i.index++
	if i.index < len(i.tok) {
		return i.tok[i.index]
	} else {
		return nil
	}
}

func (i *interpreter) cur() *token {
	if i.index < len(i.tok) {
		return i.tok[i.index]
	} else {
		return nil
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
		"(" : LEFT,
		")" : RIGHT,
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

func isPlus(typ int) bool {
	if typ == MINUS || typ == PLUS {
		return true
	}
	return false
}

func isMult(typ int) bool {
	if typ == MULT || typ == DIV {
		return true
	}
	return false
}

/*
	expres : term ((PLUS | MINUS) term) *
	term : factor ((MUL | DIV) Factor) *
	factor : integer | ( expres )
*/

func (i *interpreter) factor(cur *token) int {
	if cur.typ == INTEGER {
		value, _ := strconv.Atoi(cur.value)
		return value
	} else if cur.typ == LEFT {
		i.next()
		result := i.expre()
		return result
	} else {
		fmt.Fprintf(os.Stderr, "Syntax Error has %d\n", i.cur().typ)
		os.Exit(-1)
	}
	return 0
}

func (i *interpreter) term() int {
	var result int
	result = i.factor(i.cur())
	for token := i.next(); token != nil && isMult(token.typ) == true ; {
		switch token.typ {
		case MULT:
			result = result * i.factor(i.next())
		case DIV:
			result = result / i.factor(i.next())
		}
		token = i.next()
	}
	return result
}

func (i *interpreter) expre() int {
	result := i.term()
	for ; i.cur() != nil && i.cur().typ != EOF; {
		switch i.cur().typ {
		case MINUS:
			i.next()
			result = result - i.term()
		case PLUS:
			i.next()
			result = result + i.term()
		case RIGHT:
			return result
		default:
			fmt.Fprintf(os.Stderr, "Error parsing need + or - has %d\n", i.cur().typ)
			os.Exit(-1)
		}
	}
	return result
}

func (i *interpreter) parse() {
	result := i.expre()
	fmt.Println("Result := ", result)
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
	interpreter := interpreter{0, lexemes}
	interpreter.parse()
}
