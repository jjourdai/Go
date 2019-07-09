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
	PROCEDURE = 26
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
		PROCEDURE : "PROCEDURE",
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
		"PROCEDURE" : PROCEDURE,
}

/* STRUCT */

type Symbol struct {
	name string
	stype interface{}
}

type SymbolTable map[string]Symbol
func (s SymbolTable) define(new Symbol) {
	s[new.name] = new
}

func (s SymbolTable) lookup(name string) (Symbol, bool) {
	symbol, ok := s[name]
	return symbol, ok
}

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

type Spec struct {
	val int
	sstring string
}

type Number struct {
	token *lexemes
}

type Op struct {
	token *lexemes
}

var global_scope *VarInit = nil
type VarInit map[string]*Var

type VarDeclaration struct {
	token *lexemes
	spec *Spec
}

type ProcedureDecl struct {
	proc_name string
	block *Block
}

type Var struct {
	token *lexemes
	value float64
}

type Block struct {
	declaration_list Elem_list
	compound interface{}
}

type Assign struct {
	variable *Var
	token *lexemes
	expr *Node
}

type Node struct {
	left *Node
	token interface{}
	right *Node
}

/*
	LEXER
*/

func (n *lexemes) String() string {
	return fmt.Sprintf("type [%d] '%s'", n.ttype, n.tstring)
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
	return fmt.Sprintf("%v := %d", v.token, int64(v.value))
}
func (v *VarDeclaration) String() string {
	return fmt.Sprintf("%v type := %s", v.token, (*v.spec).sstring)
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
		node = &Node{nil, &Number{token}, nil}
	case REAL_CONST:
		r.digest(REAL_CONST)
		node = &Node{nil, &Number{token}, nil}
	case ID:
		r.digest(ID)
		node = &Node{nil, &Var{token, 0}, nil}
	case LPAR:
		r.digest(LPAR)
		node = r.expr()
		r.digest(RPAR)
	case PLUS:
		r.digest(PLUS)
		node = &Node{nil, &Op{token}, r.factor()}
	case MINUS:
		r.digest(MINUS)
		node = &Node{nil, &Op{token}, r.factor()}
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
		node = &Node{node, &Op{token},  r.factor()}
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
		node = &Node{node, &Op{token},  r.term()}
	}
	return node
}

func (r *rules) variable() *Var {
	token := r.lexer.Cur()
	r.digest(ID)
	return &Var{token, 0}
}

func (r *rules) declare_variable() *VarDeclaration {
	token := r.lexer.Cur()
	r.digest(ID)
	return &VarDeclaration{token, nil}
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
		return Spec{INTEGER_CONST, "INTEGER_CONST"}
	case REAL_CONST:
		r.digest(REAL_CONST)
		return Spec{REAL_CONST, "REAL_CONST"}
	default:
		fmt.Fprintf(os.Stderr, "Semantic Error: %s unknown type\n", token.tstring)
		os.Exit(-1)
	}
	return nil
}

func (r *rules) variable_declaration() Elem_list {
	variable := r.declare_variable()
	list := Elem_list{}
	list.elem = append(list.elem, variable)
	for ; r.lexer.Cur().ttype == COMMA ; {
		r.digest(COMMA)
		variable = r.declare_variable()
		list.elem = append(list.elem, variable)
	}
	r.digest(COLON)
	type_spec := r.type_spec()
	list.elem = append(list.elem, type_spec)
	return list
}

func (r *rules) declaration() Elem_list {
	token := r.lexer.Cur()
	declare_list := Elem_list{}
	if token.ttype == VAR {
		r.digest(VAR)
		for ; r.lexer.Cur().ttype == ID; {
			list := r.variable_declaration()
			length := len(list.elem)
			index := 0
			type_spec, _ := list.elem[length - 1].(Spec)
			for ; index < length - 1; index++ {
				variable, _ := list.elem[index].(*VarDeclaration)
				variable.spec = &type_spec
				declare_list.elem = append(declare_list.elem, variable)
			}
			r.digest(SEMI)
		}
	}
	for token = r.lexer.Cur(); token.ttype == PROCEDURE ; token = r.lexer.Cur() {
		r.digest(PROCEDURE)
		proc_name := r.lexer.Cur().tstring
		r.digest(ID)
		r.digest(SEMI)
		block := r.block()
		procedure := &ProcedureDecl{proc_name, block}
		declare_list.elem = append(declare_list.elem, procedure)
		r.digest(SEMI)
	}
	return declare_list
}

func (r *rules) block() *Block {
	test := r.declaration()
	return &Block{test, r.compound_statement()}
}

func (r *rules) program() *Block {
	r.digest(PROGRAM)
	r.variable()
	r.digest(SEMI)
	declarations := r.block()
	r.digest(DOT)
	return declarations
}

/*
	Interpreter
*/

func interpret_tree(node interface{}) float64 {
	var result float64
	switch v := node.(type) {
	case *ProcedureDecl:
		fmt.Println("Type ProcedurDecl")
	case *Block:
		fmt.Println("Type Block")
		interpret_tree(v.compound)
	case *Compound:
		fmt.Println("Type Compound")
		for _, elem := range v.elem {
			interpret_tree(elem)
		}
	case *VarDeclaration:
		fmt.Println("Type VarDeclaration")
	case *Var:
		fmt.Println("Type Var")
		result = (*global_scope)[v.token.tstring].value
		return result
	case *Assign:
		fmt.Println("Type Assign")
		var_name := v.variable.token.tstring
		(*global_scope)[var_name] = v.variable
		(*global_scope)[var_name].value = interpret_tree(v.expr)
	case *Node:
		fmt.Println("Type Node")
		var result, left, right float64
		var test *Node
		test, ok := node.(*Node)
		if ok == false {
			fmt.Fprintln(os.Stderr, "need node", ok)
			os.Exit(-1)
		}
		if test.left != nil {
			left = interpret_tree(test.left)
		}
		if test.right != nil {
			right = interpret_tree(test.right)
		}
		switch cur := test.token.(type) {
		case *Op:
			switch cur.token.ttype {
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
			case ID:
				var_name := cur.token.tstring
				result = (*global_scope)[var_name].value
			}
		default:
			result = interpret_tree(cur)
		}
		return result
	case *Op:
		fmt.Println("Type Op")
	case *Number:
		fmt.Println("Type Number")
		switch v.token.ttype {
		case INTEGER_CONST:
			tmp, _ := strconv.ParseInt(v.token.tstring, 10, 64)
			result = float64(tmp)
		case REAL_CONST:
			result, _ = strconv.ParseFloat(v.token.tstring, 64)
		}
		return result
	default:
		fmt.Printf("Type unknown %T\n", v)
		val, ok := node.(*lexemes)
		fmt.Println(val, ok)
	}
	return 0
}

func (s SymbolTable) fill_symbol_table(i interface{}) {
	switch v := i.(type) {
	case *ProcedureDecl:
		fmt.Println("Type ProcedurDecl")
	case *Block:
		fmt.Println("Type Block")
		list := v.declaration_list.elem
		for _, variable := range list {
			s.fill_symbol_table(variable)
		}
		s.fill_symbol_table(v.compound)
	case *Compound:
		fmt.Println("Type Compound")
		for _, elem := range v.elem {
			s.fill_symbol_table(elem)
		}
	case *VarDeclaration:
		fmt.Println("Type VarDeclaration")
		type_name := v.spec.sstring
		type_symbol, _ := s.lookup(type_name)
		var_name := v.token.tstring
		s.define(Symbol{var_name, type_symbol})
	case *Var:
		fmt.Println("Type Var")
		var_name := v.token.tstring
		_, ok := s.lookup(var_name)
		if ok == false {
			fmt.Fprintf(os.Stderr, "Semantic Error: %s undeclared \n", var_name)
			os.Exit(-1)
		}
	case *Assign:
		fmt.Println("Type Assign")
		var_name := v.variable.token.tstring
		_, ok := s.lookup(var_name)
		if ok == false {
			fmt.Fprintf(os.Stderr, "Semantic Error: %s undeclared \n", var_name)
			os.Exit(-1)
		}
		s.fill_symbol_table(v.expr)
	case *Node:
		fmt.Println("Type Node")
		s.fill_symbol_table(v.token)
		if v.left != nil {
			s.fill_symbol_table(v.left)
		}
		if v.right != nil {
			s.fill_symbol_table(v.right)
		}
	case *Op:
		fmt.Println("Type Op")
	case *Number:
		fmt.Println("Type Number")
	default:
		fmt.Printf("Type unknown %T\n", v)
		val, ok := i.(*lexemes)
		fmt.Println(val, ok)
	}
}

func (r *rules) Parse() {
	tree := r.program()
	if r.lexer.Cur().ttype == EOF {
		fmt.Println("Parsing FINISHED")
		init := make(VarInit)
		global_scope = &init
		fmt.Println("=========================================")
		symbol_table := SymbolTable{}
		symbol_table.define(Symbol{"INTEGER_CONST", INTEGER_CONST})
		symbol_table.define(Symbol{"REAL_CONST", REAL_CONST})
		symbol_table.fill_symbol_table(tree)
		fmt.Println("=========================================")
		interpret_tree(tree)
		for index, value := range *global_scope {
			fmt.Println(index, value)
		}
		fmt.Println(">>>>>>>>>>>>>SYMBOL_TABLE<<<<<<<<<<<<<<<<<<<")
		for _, value := range symbol_table {
			fmt.Printf("%s -->>> %s\n", value.name, value.stype)
		}
		fmt.Println("=========================================")
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
