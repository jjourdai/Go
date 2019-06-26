package main

import (
	"fmt"
	"flag"
	"os"
	"unicode"
	"strconv"
)

const (
	NONE = 0
	INTEGER = 1
	PLUS = 2
	MINUS = 3
	EOF = 4
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
}

type syntax_error string
func (e syntax_error) Error() string {
	return fmt.Sprintf("Error Parsing input %s", string(e))
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
	if current_char == '+' {
		i.pos++
		i.count++
		return &token{PLUS, string(current_char)}, nil
	} else if current_char == '-' {
		i.pos++
		i.count++
		return &token{MINUS, string(current_char)}, nil
	}
	return nil, syntax_error(text)
}

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "Error: need just 1 params")
		os.Exit(-1)
	}
	inter := interpreter{0, args[0], nil, 0}
	expre := make([]*token, 3)
	for {
		token, err := inter.get_next_token()
		if token != nil && token.typ == EOF {
			break
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(-1)
		}
		expre[inter.count - 1] = token
	}
	left, _ := strconv.Atoi(expre[0].value)
	right, _ := strconv.Atoi(expre[2].value)
	fmt.Println(left, right)
	switch expre[1].typ {
	case MINUS:
		fmt.Println(left - right)
	case PLUS:
		fmt.Println(left + right)
	}
}
