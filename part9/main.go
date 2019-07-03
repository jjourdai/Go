/*
	GRAMMAR 

    program : compound_statement DOT

    compound_statement : BEGIN statement_list END

    statement_list : statement
                   | statement SEMI statement_list

    statement : compound_statement
              | assignment_statement
              | empty

    assignment_statement : variable ASSIGN expr

    empty :

    expr: term ((PLUS | MINUS) term)*

    term: factor ((MUL | DIV) factor)*

    factor : PLUS factor
           | MINUS factor
           | INTEGER
           | LPAREN expr RPAREN
           | variable

    variable: ID
*/

package main

import (
	"fmt"
	"os"
	"flag"
	"strings"
	"strconv"
	"io/ioutil"
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
	DOT = 9
	BEGIN = 10
	END = 11
	SEMI = 12
	ASSIGN = 13
	EOL = 14
	EOF = 15
)

type lexemes struct {
	ttype int
	tvalue []string
}


func (n *lexemes) String() string {
	return fmt.Sprintf("type [%d] value '%s'", n.ttype, n.tvalue)
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
		";" : RPAR,
		"\n" : EOF,
	}
	return func(key string) int {
		return lex[key]
	}
}

func lexer(expr string) []lexemes {
	lex := init_lex()
	var tokens []lexemes
	var new_token *lexemes
	var slice []string
	for _, value := range expr {
		switch {
		case value == ' ':
			if new_token != nil {
				new_token.tvalue = slice
				tokens = append(tokens, *new_token)
				new_token = nil
				slice = nil
			}
			continue
		case value >= '0' && value <= '9':
			if new_token == nil {
				new_token = &lexemes{INTEGER, make([]string, 10)}
			}
			slice = append(slice, string(value))
		default:
			if new_token != nil {
				new_token.tvalue = slice
				tokens = append(tokens, *new_token)
				new_token = nil
				slice = nil
			}
			new_val := lex(string(value))
			tokens = append(tokens, lexemes{new_val, []string{string(value)}})
			if new_val == 0 {
				fmt.Fprintf(os.Stderr, "Lexer Error: unexpected character '%c'\n", value)
				os.Exit(-1)
			}
		}
//		fmt.Printf("index [%d] = '%c'\n", index, value)
	}
	if new_token != nil {
		new_token.tvalue = slice
		tokens = append(tokens, *new_token)
		new_token = nil
		slice = nil
	}

	tokens = append(tokens, lexemes{EOF, []string{"EOF"}})
	return tokens
}

func prior1(current_token []string) bool {
	available_character := "*/%"
	for _, value := range current_token {
		if strings.Contains(available_character, value) {
			return true
		}
	}
	return false
}

func prior2(current_token []string) bool {
	available_character := "+-"
	for _, value := range current_token {
		if strings.Contains(available_character, value) {
			return true
		}
	}
	return false
}

type interpreter struct {
	index int
	length int
	tokens []lexemes
}

type Node struct {
	left *Node
	token *lexemes
	right *Node
}

type Ast struct {
	root *Node
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

func (i *interpreter) factor() *Node {
	var node *Node
	token := i.Cur()
	switch token.ttype {
	case INTEGER:
		i.digest(INTEGER)
		node = &Node{nil, token, nil}
	case LPAR:
		i.digest(LPAR)
		node = i.expr()
		i.digest(RPAR)
	case PLUS:
		i.digest(PLUS)
		node = &Node{nil, token, i.factor()}
	case MINUS:
		i.digest(MINUS)
		node = &Node{nil, token, i.factor()}
	default:
		fmt.Fprintf(os.Stderr, "Syntax Error\n")
		os.Exit(-1)
	}
	return node
}

func (i *interpreter) term() *Node {
	node := i.factor()
	for ; prior1(i.Cur().tvalue) == true; {
		token := i.Cur()
		switch token.ttype {
		case MUL:
			i.digest(MUL)
		case DIV:
			i.digest(DIV)
		case MOD:
			i.digest(MOD)
		}
		node = &Node{node, token,  i.factor()}
	}
	return node
}

func (i *interpreter) expr() *Node {
	node := i.term()
	for ; prior2(i.Cur().tvalue) == true; {
		token := i.Cur()
		switch token.ttype {
		case MINUS:
			i.digest(MINUS)
		case PLUS:
			i.digest(PLUS)
		}
		node = &Node{node, token,  i.term()}
	}
	return node
}

func rpn_notation(node *Node) {
	if node.left != nil {
		rpn_notation(node.left)
	}
	if node.right != nil {
		rpn_notation(node.right)
	}
	fmt.Print(node.token.tvalue, " ")
}

func lisp_notation(node *Node) {
	fmt.Print(node.token.tvalue, " ")
	if node.left != nil {
		lisp_notation(node.left)
	}
	if node.right != nil {
		lisp_notation(node.right)
	}
}

func run(node *Node) int {
	var result, left, right int
	if node.left != nil {
		left = run(node.left)
	}
	if node.right != nil {
		right = run(node.right)
	}
	switch node.token.ttype {
		case MINUS:
			result = left - right
		case PLUS:
			result = left + right
		case DIV:
			result = left / right
		case MUL:
			result = left * right
		case MOD:
			result = left % right
		case INTEGER:
			result, _ = strconv.Atoi(strings.Join(node.token.tvalue," "))
	}
	return result
}

func (i *interpreter) Parse() {
	node := i.expr()
	if i.Cur().ttype == EOF {
		fmt.Println("Result :=", run(node))
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
	data, err := ioutil.ReadFile(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, data)
	}
	tokens := lexer(string(data))
/*
	for i, j := range tokens {
		fmt.Println("Token :=", i, j)
	}
*/
	interpreter := interpreter{0, len(tokens), tokens}
	interpreter.Parse()
}
