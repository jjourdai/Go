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
	INTEGER_DIV = 4
	MOD = 5
	LPAR = 6
	RPAR = 7
	DOT = 9
	BEGIN = 10
	END = 11
	SEMI = 12
	ASSIGN = 13
	ID = 14
	COLON = 15
	COMMA = 16
	REAL_CONST = 17
	INTEGER_CONST = 18
	EOL = 19
	EOF = 20
	VAR = 21
	PROGRAM = 22
	OCOMMENT = 23
	CCOMMENT = 24
	FLOAT_DIV = 25
)

/* STATIC VALUE */

var reverse_lex = map[int]string {
		PLUS : "PLUS",
		MINUS : "MINUS",
		MUL : "MUL",
		INTEGER_DIV : "INTEGER_DIV",
		FLOAT_DIV : "FLOAT_DIV",
		MOD : "MOD",
		LPAR : "LPAR",
		RPAR : "RPAR",
		DOT : "DOT",
		BEGIN : "BEGIN",
		END : "END",
		SEMI : "SEMI",
		ASSIGN : "ASSIGN",
		ID : "ID",
		EOL : "EOL",
		EOF : "EOF",
		VAR : "VAR",
		REAL_CONST : "REAL_CONST",
		INTEGER_CONST : "INTEGER_CONST",
		COLON : "COLON",
		COMMA : "COMMA",
		PROGRAM : "PROGRAM",
}

var lex = map[string]int {
		"+" : PLUS,
		"-" : MINUS,
		"%" : MOD,
		"*" : MUL,
		"(" : LPAR,
		")" : RPAR,
		"/" : FLOAT_DIV,
		";" : SEMI,
		"." : DOT,
		"," : COMMA,
		":" : COLON,
		"\n" : EOF,
		"{" : OCOMMENT,
		"}" : CCOMMENT,
}

var keyword = map[string]int {
		"BEGIN" : BEGIN,
		"END" : END,
		"DIV" : INTEGER_DIV,
		"VAR" : VAR,
		"REAL" : REAL_CONST,
		"INTEGER" : INTEGER_CONST,
		"PROGRAM" : PROGRAM,
}

/* STRUCT */

type lexemes struct {
	ttype int
	tstring string
}

type lexer struct {
	index int
	length int
	tokens []lexemes
}

type rules struct {
	lexer lexer
}

type Compound struct {
	elem []interface{}
}

type Elem_list struct {
	elem []interface{}
}

type Spec int

var global_varlist *VarList = nil
type VarList map[string]*Var

type Var struct {
	token *lexemes
	spec *Spec
	value float64
}

type Block struct {
	varlist *VarList
	compound interface{}
}

type Assign struct {
	variable *Var
	token *lexemes
	expr *Node
}

type Node struct {
	left *Node
	token *lexemes
	right *Node
}

/*
	LEXER
*/

func (n *lexemes) String() string {
	return fmt.Sprintf("type [%d] value '%s'", n.ttype, n.tstring)
}

func store_new_token(tokens *[]lexemes, new_token **lexemes) {
	if *new_token != nil {
		(*new_token).tstring = strings.ToUpper((*new_token).tstring)
		if kword := keyword[(*new_token).tstring]; kword != 0 {
			(*new_token).ttype = kword
		}
		*tokens = append(*tokens, **new_token)
		*new_token = nil
	}
}

func tokenize(expr string) []lexemes {
	var tokens []lexemes
	var new_token *lexemes
	length := len(expr)
	comment_mode := false
	for index := 0 ; index < length; index++ {
		switch {
		case expr[index] == '}':
			comment_mode = false
		case expr[index] == '{' || comment_mode == true:
			store_new_token(&tokens, &new_token)
			comment_mode = true
			continue
		case unicode.IsSpace(rune(expr[index])):
			store_new_token(&tokens, &new_token)
			continue
		case expr[index] >= '0' && expr[index] <= '9':
			if new_token == nil {
				new_token = &lexemes{INTEGER_CONST, ""}
			}
			new_token.tstring += string(expr[index])
		case expr[index] >= 65 && expr[index] <= 90 || expr[index] >= 97 && expr[index] <= 122 || expr[index] == '_':
			if new_token != nil && new_token.ttype != ID {
				store_new_token(&tokens, &new_token)
			}
			if new_token == nil {
				new_token = &lexemes{ID, ""}
			}
			new_token.tstring += string(expr[index])
		case expr[index] == ':' && index < length - 1 && expr[index + 1] == '=':
			store_new_token(&tokens, &new_token)
			tokens = append(tokens, lexemes{ASSIGN, ":="})
			index++
		case expr[index] == '.' && new_token.ttype == INTEGER_CONST:
			new_token.tstring += string(expr[index])
			new_token.ttype = REAL_CONST
		default:
			store_new_token(&tokens, &new_token)
			new_val := lex[string(expr[index])]
			tokens = append(tokens, lexemes{new_val, string(expr[index])})
			if new_val == 0 {
				fmt.Fprintf(os.Stderr, "Lexer Error: unexpected character '%c'\n", expr[index])
				os.Exit(-1)
			}
		}
	}
	store_new_token(&tokens, &new_token)
	tokens = append(tokens, lexemes{EOF, "EOF"})
	return tokens
}

/*
	Parser
*/

func prior1(current_token int) bool {
	if current_token == MOD ||
		current_token == INTEGER_DIV ||
		current_token == MUL ||
		current_token == FLOAT_DIV {
		return true
	}
	return false
}

func prior2(current_token int) bool {
	if current_token == PLUS || current_token == MINUS {
		return true
	}
	return false
}

func (v *Var) String() string {
	if *v.spec == INTEGER_CONST {
		return fmt.Sprintf("%v type := %s := %d", v.token, reverse_lex[int(*v.spec)], int64(v.value))
	} else {
		return fmt.Sprintf("%v type := %s := %f", v.token, reverse_lex[int(*v.spec)], v.value)
	}
}

func (l *lexer) Cur() *lexemes {
	if l.index == l.length {
		return nil
	} else {
		return &l.tokens[l.index]
	}
}

func (l *lexer) Next() *lexemes {
	l.index++
	if l.index == l.length {
		return nil
	} else {
		return &l.tokens[l.index]
	}
}

func (r *rules) digest(needed int) {
	if needed == r.lexer.Cur().ttype {
		fmt.Printf("Digest := [%d] '%s'\n", r.lexer.Cur().ttype, r.lexer.Cur().tstring)
		r.lexer.Next()
	} else {
		fmt.Fprintf(os.Stderr, "Syntax Error waiting for [%s] has [%s] for %s\n",
			reverse_lex[needed], reverse_lex[r.lexer.Cur().ttype], r.lexer.Cur().tstring)
		os.Exit(-1)
	}
}

func (r *rules) factor() *Node {
	var node *Node
	token := r.lexer.Cur()
	switch token.ttype {
	case INTEGER_CONST:
		r.digest(INTEGER_CONST)
		node = &Node{nil, token, nil}
	case REAL_CONST:
		r.digest(REAL_CONST)
		node = &Node{nil, token, nil}
	case ID:
		r.digest(ID)
		node = &Node{nil, token, nil}
	case LPAR:
		r.digest(LPAR)
		node = r.expr()
		r.digest(RPAR)
	case PLUS:
		r.digest(PLUS)
		node = &Node{nil, token, r.factor()}
	case MINUS:
		r.digest(MINUS)
		node = &Node{nil, token, r.factor()}
	default:
		fmt.Fprintf(os.Stderr, "Syntax Error\n")
		os.Exit(-1)
	}
	return node
}

func (r *rules) term() *Node {
	node := r.factor()
	for ; prior1(r.lexer.Cur().ttype) == true; {
		token := r.lexer.Cur()
		switch token.ttype {
		case MUL:
			r.digest(MUL)
		case INTEGER_DIV:
			r.digest(INTEGER_DIV)
		case FLOAT_DIV:
			r.digest(FLOAT_DIV)
//		case MOD:
//			r.digest(MOD)
		}
		node = &Node{node, token,  r.factor()}
	}
	return node
}

func (r *rules) expr() *Node {
	node := r.term()
	for ; prior2(r.lexer.Cur().ttype) == true; {
		token := r.lexer.Cur()
		switch token.ttype {
		case MINUS:
			r.digest(MINUS)
		case PLUS:
			r.digest(PLUS)
		}
		node = &Node{node, token,  r.term()}
	}
	return node
}

func (r *rules) variable() *Var {
	token := r.lexer.Cur()
	r.digest(ID)
	return &Var{token, nil, 0}
}

func (r *rules) assignment_statement() interface{} {
	variable := r.variable()
	token := r.lexer.Cur()
	r.digest(ASSIGN)
	return &Assign{variable, token, r.expr()}
}

func (r *rules) statement() interface{} {
	ttype := r.lexer.Cur().ttype
	var node interface{}
	if ttype == BEGIN {
		node = r.compound_statement()
	} else if ttype == ID {
		node = r.assignment_statement()
	} else {
		return nil
	}
	return node
}

func (r *rules) statement_list() Elem_list {
	node := r.statement()
	list := Elem_list{}
	list.elem = append(list.elem, node)
	for ; r.lexer.Cur().ttype == SEMI ; {
		r.digest(SEMI)
		list.elem = append(list.elem, r.statement())
	}
	return list
}

func (r *rules) compound_statement() interface{} {
	r.digest(BEGIN)
	node := r.statement_list()
	r.digest(END)
	root := Compound{node.elem}
	return &root
}

func (r *rules) type_spec() interface{} {
	token := r.lexer.Cur()
	switch token.ttype {
	case INTEGER_CONST:
		r.digest(INTEGER_CONST)
		return Spec(INTEGER_CONST)
	case REAL_CONST:
		r.digest(REAL_CONST)
		return Spec(REAL_CONST)
	default:
		fmt.Fprintf(os.Stderr, "Semantic Error: %s unknown type\n", token.tstring)
		os.Exit(-1)
	}
	return nil
}

func (r *rules) variable_declaration() Elem_list {
	variable := r.variable()
	list := Elem_list{}
	list.elem = append(list.elem, variable)
	for ; r.lexer.Cur().ttype == COMMA ; {
		r.digest(COMMA)
		variable = r.variable()
		list.elem = append(list.elem, variable)
	}
	r.digest(COLON)
	type_spec := r.type_spec()
	list.elem = append(list.elem, type_spec)
	return list
}

func (r *rules) declaration() *VarList {
	r.digest(VAR)
	varlist := make(VarList)
	for ; r.lexer.Cur().ttype == ID; {
		list := r.variable_declaration()
		length := len(list.elem)
		index := 0
		type_spec, _ := list.elem[length - 1].(Spec)
		for ; index < length - 1; index++ {
			variable, _ := list.elem[index].(*Var)
			varlist[variable.token.tstring] = variable
			variable.spec = &type_spec
		}
		r.digest(SEMI)
	}
	return &varlist
}

func (r *rules) block() *Block {
	return &Block{r.declaration(), r.compound_statement()}
}

func (r *rules) program() *Block {
	r.digest(PROGRAM)
	r.variable()
	r.digest(SEMI)
	node := r.block()
	global_varlist = node.varlist
/*
for index, value := range *node.varlist {
	fmt.Println(index, value)
}
*/
	r.digest(DOT)
	return node
}

/*
	Interpreter
*/

func (c Compound) Resolve() {
	for _, elem := range c.elem {
		if elem != nil {
			switch v := elem.(type) {
			case *Block:
				fmt.Println("Type Block")
			case *Compound:
				v.Resolve()
			case *Var:
				fmt.Println("Type Var")
			case *Assign:
				fmt.Println("Type Assign")
				var_name := v.variable.token.tstring
				_, ok := (*global_varlist)[var_name]
				if ok == true {
					(*global_varlist)[var_name].value = run(v.expr)
				} else {
					fmt.Fprintf(os.Stderr, "Semantic Error: %s undeclared \n", var_name)
					os.Exit(-1)
				}
				run(v.expr)
			case *Node:
				fmt.Println("Type Node")
			default:
				fmt.Println("Type unknown")
			}
		}
	}
}

func (b *Block) Resolve() {
	switch l := b.compound.(type) {
		case *Compound:
			l.Resolve()
		default:
			fmt.Println("Error")
	}
}

func run(elem interface{}) float64 {
	var result, left, right float64
	var test *Node
	test, ok := elem.(*Node)
	if ok == false {
		fmt.Fprintln(os.Stderr, "need node", ok)
		os.Exit(-1)
	}
	if test.left != nil {
		left = run(test.left)
	}
	if test.right != nil {
		right = run(test.right)
	}
	switch test.token.ttype {
		case MINUS:
			result = left - right
		case PLUS:
			result = left + right
		case INTEGER_DIV:
			result = left / right
		case FLOAT_DIV:
			result = left / right
		case MUL:
			result = left * right
//		case MOD:
//			result = left % right
		case INTEGER_CONST:
			tmp, _ := strconv.ParseInt(test.token.tstring, 10, 64)
			result = float64(tmp)
		case REAL_CONST:
			result, _ = strconv.ParseFloat(test.token.tstring, 64)
		case ID:
			var_name := test.token.tstring
			_, ok := (*global_varlist)[var_name]
	//		fmt.Println(var_name)
			if ok == true {
				result = (*global_varlist)[var_name].value
			} else {
				fmt.Fprintf(os.Stderr, "Semantic Error: %s undeclared \n", var_name)
				os.Exit(-1)
			}
	}
	return result
}

func (r *rules) Parse() {
	node := r.program()
	if r.lexer.Cur().ttype == EOF {
		fmt.Println("Parsing FINISHED")
		node.Resolve()
		fmt.Println("\\\\\\\\\\\\\\\\ Result /////////")
		for index, value := range *global_varlist {
			fmt.Println(index, value)
		}
	} else {
		fmt.Fprintf(os.Stderr, "Unexpected token %d '%s'\n", r.lexer.Cur().ttype, r.lexer.Cur().tstring)
		os.Exit(-1)
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
	tokens := tokenize(string(data))

	for i, j := range tokens {
		fmt.Printf("Token[%d] := {%s} '%s'\n", i, reverse_lex[j.ttype], j.tstring)
	}
	rules := rules{lexer{0, len(tokens), tokens}}
	rules.Parse()
}
