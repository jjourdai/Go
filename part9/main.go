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

    variable: WORD
*/

package main

import (
	"fmt"
	"os"
	"flag"
	"strings"
	"strconv"
	"unicode"
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
	WORD = 14
	EOL = 15
	EOF = 16
)

/*
	Interpreter
*/

type lexemes struct {
	ttype int
	tvalue string
}

func (n *lexemes) String() string {
	return fmt.Sprintf("type [%d] value '%s'", n.ttype, n.tvalue)
}

func init_lex() func(key string) int {
	var lex = map[string]int {
		"+" : PLUS,
		"-" : MINUS,
//		"/" : DIV,
		"%" : MOD,
		"*" : MUL,
		"(" : LPAR,
		")" : RPAR,
		";" : SEMI,
		"." : DOT,
		"\n" : EOF,
	}
	return func(key string) int {
		return lex[key]
	}
}

func reserved_keyword() func(key string) int {
	var keyword = map[string]int {
		"BEGIN" : BEGIN,
		"END" : END,
		"DIV" : DIV,
	}
	return func(key string) int {
		return keyword[key]
	}
}

func store_new_token(tokens *[]lexemes, new_token **lexemes, slice *[]string) {
	keyword := reserved_keyword()
	if *new_token != nil {
		(*new_token).tvalue = strings.ToUpper(strings.Join(*slice, ""))
		if kword := keyword((*new_token).tvalue); kword != 0 {
			(*new_token).ttype = kword
		}
		*tokens = append(*tokens, **new_token)
		*new_token = nil
		*slice = nil
	}
}

func lexer(expr string) []lexemes {
	lex := init_lex()
	var tokens []lexemes
	var new_token *lexemes
	var slice []string
	length := len(expr)
	for index := 0 ; index < length; index++ {
		switch {
		case unicode.IsSpace(rune(expr[index])):
			store_new_token(&tokens, &new_token, &slice)
			continue
		case expr[index] >= '0' && expr[index] <= '9':
			if new_token == nil {
				new_token = &lexemes{INTEGER, ""}
			}
			slice = append(slice, string(expr[index]))
		case expr[index] >= 65 && expr[index] <= 90 || expr[index] >= 97 && expr[index] <= 122 || expr[index] == '_':
			if new_token != nil && new_token.ttype != WORD {
				store_new_token(&tokens, &new_token, &slice)
			}
			if new_token == nil {
				new_token = &lexemes{WORD, ""}
			}
			slice = append(slice, string(expr[index]))
		case expr[index] == ':' && index < length - 1 && expr[index + 1] == '=':
			store_new_token(&tokens, &new_token, &slice)
			tokens = append(tokens, lexemes{ASSIGN, ":="})
			index++
		default:
			store_new_token(&tokens, &new_token, &slice)
			new_val := lex(string(expr[index]))
			tokens = append(tokens, lexemes{new_val, string(expr[index])})
			if new_val == 0 {
				fmt.Fprintf(os.Stderr, "Lexer Error: unexpected character '%c'\n", expr[index])
				os.Exit(-1)
			}
		}
//		fmt.Printf("index [%d] = '%c'\n", index, expr[index])
	}
	store_new_token(&tokens, &new_token, &slice)
	tokens = append(tokens, lexemes{EOF, "EOF"})
	return tokens
}

func prior1(current_token int) bool {
	if current_token == MOD || current_token == DIV || current_token == MUL {
		return true
	}
	return false
}

func prior2(current_token string) bool {
	available_character := "+-"
	return strings.Contains(available_character, current_token)
}

/*
	Interpreter
*/

type interpreter struct {
	index int
	length int
	tokens []lexemes
}

type Element interface {
	Resolve()
}

type Compound struct {
	elem []Element
	scope map[string]int
}

func (c Compound) Resolve() {
	for _, elem := range c.elem {
		if elem != nil {
			switch v := elem.(type) {
			case *Compound:
				v.Resolve()
			case *Var:
				fmt.Println("Type Var")
			case *Assign:
				fmt.Println("Type Assign")
				var_name := v.variable.token.tvalue
				c.scope[var_name] = run(v.expr, c)
				fmt.Println("result :=", c.scope[var_name])
			case *Node:
				fmt.Println("Type Node")
			default:
				fmt.Println("Type unknown")
			}
		}
	}
}

type Elem_list struct {
	elem []Element
}

func (c Elem_list) Resolve() {

}


type Var struct {
	token *lexemes
}

func (c *Var) Resolve() {
}

type Assign struct {
	variable Var
	token *lexemes
	expr *Node
}

func (a *Assign) Resolve() {
	//var_name := a.variable.token.tvalue
	//GLOBAL_SCOPE[var_name] = run(a.expr)
	//fmt.Println("dwa;ljdlawkjdalwjdakw")
}

type Node struct {
	left *Node
	token *lexemes
	right *Node
}

func (n *Node) Resolve() {
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
		os.Exit(-1)
	}
}

func (i *interpreter) factor() *Node {
	var node *Node
	token := i.Cur()
	switch token.ttype {
	case INTEGER:
		i.digest(INTEGER)
		node = &Node{nil, token, nil}
	case WORD:
		i.digest(WORD)
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
	for ; prior1(i.Cur().ttype) == true; {
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

func (i *interpreter) assignment_statement() Element {
	token := i.Cur()
	i.digest(WORD)
	variable := Var{token}
	token = i.Cur()
	i.digest(ASSIGN)
	return &Assign{variable, token, i.expr()}
}

func (i *interpreter) statement() Element {
	ttype := i.Cur().ttype
	var node Element
	if ttype == BEGIN {
		node = i.compound_statement()
	} else if ttype == WORD {
		node = i.assignment_statement()
	} else {
		return nil
	}
	return node
}

func (i *interpreter) statement_list() Elem_list {
	node := i.statement()
	list := Elem_list{}
	list.elem = append(list.elem, node)
	for ; i.Cur().ttype == SEMI ; {
		i.digest(SEMI)
		list.elem = append(list.elem, i.statement())
	}
	return list
}

func (i *interpreter) compound_statement() Element {
	i.digest(BEGIN)
	node := i.statement_list()
	i.digest(END)
	root := Compound{node.elem, make(map[string]int)}
	return &root
}

func (i *interpreter) program() Element {
	node := i.compound_statement()
	i.digest(DOT)
	return node
}

func run(elem Element, c Compound) int {
	var result, left, right int
	var test *Node
	test, ok := elem.(*Node)
	if ok == false {
		fmt.Fprintln(os.Stderr, "need node", ok)
		os.Exit(-1)
	}
	if test.left != nil {
		left = run(test.left, c)
	}
	if test.right != nil {
		right = run(test.right, c)
	}
	switch test.token.ttype {
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
			result, _ = strconv.Atoi(test.token.tvalue)
		case WORD:
			var_name := test.token.tvalue
			value, ok := c.scope[var_name]
			if ok == true {
				result = value
			} else {
				fmt.Fprintf(os.Stderr, "Semantic Error: %s undeclared \n", var_name)
			}
	}
	return result
}

func (i *interpreter) Parse() {
	node := i.program()
	if i.Cur().ttype == EOF {
		fmt.Println("Parsing FINISHED")
		node.Resolve()
	} else {
		fmt.Fprintf(os.Stderr, "Unexpected token %d '%s'\n", i.Cur().ttype, i.Cur().tvalue)
		os.Exit(-1)
	}
}

func reverse_lex() func(key int) string {
	var lex = map[int]string {
		PLUS : "PLUS",
		MINUS : "MINUS",
		MUL : "MUL",
		DIV : "DIV",
		MOD : "MOD",
		LPAR : "LPAR",
		RPAR : "RPAR",
		INTEGER : "INTEGER",
		DOT : "DOT",
		BEGIN : "BEGIN",
		END : "END",
		SEMI : "SEM",
		ASSIGN : "ASSIGN",
		WORD : "WORD",
		EOL : "EOL",
		EOF : "EOF",
	}
	return func(key int) string {
		return lex[key]
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Need 1 parameter")
		os.Exit(-1)
	}
	flag.Parse()
	data, err := ioutil.ReadFile(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, data)
	}
	tokens := lexer(string(data))

	reverse_lex := reverse_lex()
	for i, j := range tokens {
		fmt.Printf("Token[%d] := {%s} '%s'\n", i, reverse_lex(j.ttype), j.tvalue)
	}

	interpreter := interpreter{0, len(tokens), tokens}
	interpreter.Parse()
}
