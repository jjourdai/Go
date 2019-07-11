package main

import (
	"fmt"
	"os"
	"flag"
	"strings"
	"strconv"
	"unicode"
	"bufio"
	"log"
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

type Symbol interface {
	getName() string
}

type BuiltinSymbol struct {
	name string
}
func (b *BuiltinSymbol) getName() string {
	return b.name
}
func (b *BuiltinSymbol) String() string {
	return fmt.Sprintf("%s", b.name)
}

type ProcedureSymbol struct {
	name string
	params []*VarSymbol
}

func (p *ProcedureSymbol) getName() string {
	return p.name
}

func (p *ProcedureSymbol) String() string {
	expr := fmt.Sprintf("%s: <", p.name)
	for _, elem := range p.params {
		expr += fmt.Sprintf("%v, ", elem)
	}
	expr += fmt.Sprintf(">")
	return expr
}

type VarSymbol struct {
	name string
	stype *BuiltinSymbol
}

func (v *VarSymbol) getName() string {
	return v.name
}

func (v *VarSymbol) String() string {
	return fmt.Sprintf("%s: <%s>", v.name, v.stype)
}

type SemanticsAnalyser struct {
	scope *ScopedSymbolTable
}

type ScopedSymbolTable struct {
	symbols map[string]Symbol
	scope_name string
	scope_level int
	enclosing_scope *ScopedSymbolTable
}

func (s ScopedSymbolTable) String() string {
	repr := fmt.Sprintf("SymbolTable := %s at scope %d\n", s.scope_name, s.scope_level)
	for _, value := range s.symbols {
			repr += fmt.Sprintf("	%v\n", value)
	}
	return repr
}

func (s ScopedSymbolTable) insert(new Symbol) {
	(s.symbols)[new.getName()] = new
}

func (s ScopedSymbolTable) lookup(name string, current_scope_only bool) (Symbol, bool) {
	symbol, ok := (s.symbols)[name]
	if ok == true {
		return symbol, ok
	}
	if current_scope_only == false && s.enclosing_scope != nil {
		symbol, ok = s.enclosing_scope.lookup(name, false)
	}
	return symbol, ok
}

type lexemes struct {
	ttype int
	tstring string
	line int
	column int
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

type Param struct {
	var_name *Var
	var_type *Spec
}

type ProcedureDecl struct {
	proc_name string
	params []Param
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

func tokenize(file *os.File) []lexemes {
	scanner := bufio.NewScanner(file)
	var tokens []lexemes
	var new_token *lexemes
	line := 1
	for scanner.Scan() {
		expr := scanner.Text()
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
					new_token = &lexemes{INTEGER_CONST, "", line, index}
				}
				new_token.tstring += string(expr[index])
			case expr[index] >= 65 && expr[index] <= 90 || expr[index] >= 97 && expr[index] <= 122 || expr[index] == '_':
				if new_token != nil && new_token.ttype != ID {
					store_new_token(&tokens, &new_token)
				}
				if new_token == nil {
					new_token = &lexemes{ID, "", line, index}
				}
				new_token.tstring += string(expr[index])
			case expr[index] == ':' && index < length - 1 && expr[index + 1] == '=':
				store_new_token(&tokens, &new_token)
				tokens = append(tokens, lexemes{ASSIGN, ":=", line, index})
				index++
			case expr[index] == '.' && new_token.ttype == INTEGER_CONST:
				new_token.tstring += string(expr[index])
				new_token.ttype = REAL_CONST
			default:
				store_new_token(&tokens, &new_token)
				new_val := lex[string(expr[index])]
				tokens = append(tokens, lexemes{new_val, string(expr[index]), line, index})
				if new_val == 0 {
					fmt.Fprintf(os.Stderr, "Lexer Error: unexpected character '%c' line:[%d:%d]\n", expr[index], line, index)
					os.Exit(-1)
				}
			}
		}
		line++
	}
	store_new_token(&tokens, &new_token)
	tokens = append(tokens, lexemes{EOF, "EOF", line, 0})
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
		token := r.lexer.Cur()
		fmt.Fprintf(os.Stderr, "Unexpected token %s '%s' line [%d:%d] wait for '%s'\n", reverse_lex[token.ttype], token.tstring, token.line, token.column, reverse_lex[needed])
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

func (r *rules) type_spec() *Spec {
	token := r.lexer.Cur()
	switch token.ttype {
	case INTEGER_CONST:
		r.digest(INTEGER_CONST)
		return &Spec{INTEGER_CONST, "INTEGER_CONST"}
	case REAL_CONST:
		r.digest(REAL_CONST)
		return &Spec{REAL_CONST, "REAL_CONST"}
	default:
		fmt.Fprintf(os.Stderr, "Semantic Error: %s unknown type\n", token.tstring)
		os.Exit(-1)
		return nil
	}
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

func (r *rules) formal_parameters() []Param {
	token := r.lexer.Cur()
	r.digest(ID)
	new_var := Var{token, 0}
	list := []Var{}
	list = append(list, new_var)
	for token = r.lexer.Cur(); token.ttype == COMMA; token = r.lexer.Cur() {
		r.digest(COMMA)
		token = r.lexer.Cur()
		new_var = Var{token, 0}
		r.digest(ID)
		list = append(list, new_var)
	}
	r.digest(COLON)
	type_spec := r.type_spec()
	param_list := []Param{}
	for _, val := range list {
		param_list = append(param_list, Param{&val, type_spec})
	}
	return param_list
}

func (r *rules) formal_parameters_list() []Param {
	if r.lexer.Cur().ttype != ID {
		return []Param{}
	}
	param_list := r.formal_parameters()
	for token := r.lexer.Cur(); token.ttype == SEMI; token = r.lexer.Cur() {
		r.digest(SEMI)
		param_list = append(param_list, r.formal_parameters()...)
		fmt.Println(param_list)
	}
	return param_list
}

func (r *rules) procedure_declaration() *ProcedureDecl {
	r.digest(PROCEDURE)
	proc_name := r.lexer.Cur().tstring
	r.digest(ID)
	token := r.lexer.Cur()
	var params []Param
	if token.ttype == LPAR {
		r.digest(LPAR)
		params = r.formal_parameters_list()
		r.digest(RPAR)
	}
	r.digest(SEMI)
	block := r.block()
	procedure := &ProcedureDecl{proc_name, params, block}
	r.digest(SEMI)
	return procedure
}

func (r *rules) declaration() Elem_list {
	token := r.lexer.Cur()
	declare_list := Elem_list{}
	token = r.lexer.Cur()
	for {
		token = r.lexer.Cur()
		if token.ttype == VAR {
			r.digest(VAR)
			for ; r.lexer.Cur().ttype == ID; {
				list := r.variable_declaration()
				length := len(list.elem)
				index := 0
				type_spec, _ := list.elem[length - 1].(*Spec)
				for ; index < length - 1; index++ {
					variable, _ := list.elem[index].(*VarDeclaration)
					variable.spec = type_spec
					declare_list.elem = append(declare_list.elem, variable)
				}
				r.digest(SEMI)
			}
		} else if token.ttype == PROCEDURE {
			declare_list.elem = append(declare_list.elem, r.procedure_declaration())
		} else {
			break
		}
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

func (s SemanticsAnalyser) check(i interface{}) {
	switch v := i.(type) {
	case *ProcedureDecl:
		fmt.Print(s.scope)
		fmt.Println("Type ProcedurDecl")
		fmt.Printf("ENTER scope: %s\n", v.proc_name)
		proc_symbol := ProcedureSymbol{v.proc_name, []*VarSymbol{}}
		s.scope.insert(&proc_symbol)
		new_scope := ScopedSymbolTable{make(map[string]Symbol), v.proc_name, 1, s.scope}
		s.scope = &new_scope
		for _, param := range v.params {
			type_symbol, _ := s.scope.lookup(param.var_type.sstring, false)
			builtin_symbol := (type_symbol).(*BuiltinSymbol)
			var_symbol := VarSymbol{param.var_name.token.tstring, builtin_symbol}
			s.scope.insert(&var_symbol)
			proc_symbol.params = append(proc_symbol.params, &var_symbol)
		}
		fmt.Print(new_scope)
		s.check(v.block)
		s.scope = s.scope.enclosing_scope
		fmt.Printf("LEAVE scope: %s\n", v.proc_name)
	case *Block:
		fmt.Println("Type Block")
		list := v.declaration_list.elem
		for _, variable := range list {
			s.check(variable)
		}
		s.check(v.compound)
	case *Compound:
		fmt.Println("Type Compound")
		for _, elem := range v.elem {
			s.check(elem)
		}
	case *VarDeclaration:
		fmt.Println("Type VarDeclaration")
		type_name := v.spec.sstring
		type_symbol, _ := s.scope.lookup(type_name, false)
		builtin_symbol := (type_symbol).(*BuiltinSymbol)
		var_name := v.token.tstring
		if _, found := s.scope.lookup(var_name, true); found == true {
			fmt.Fprintf(os.Stderr, "Semantic Error: %s already declaed \n", var_name, v.token.line, v.token.column)
			os.Exit(-1)
		}
		new_var_symbol := &VarSymbol{var_name, builtin_symbol}
		s.scope.insert(new_var_symbol)
	case *Var:
		fmt.Println("Type Var")
		var_name := v.token.tstring
		_, ok := s.scope.lookup(var_name, false)
		if ok == false {
			fmt.Fprintf(os.Stderr, "Semantic Error: %s undeclared line [%d:%d]\n", var_name, v.token.line, v.token.column);
			os.Exit(-1)
		}
	case *Assign:
		fmt.Println("Type Assign")
		s.check(v.variable)
		s.check(v.expr)
	case *Node:
		fmt.Println("Type Node")
		s.check(v.token)
		if v.left != nil {
			s.check(v.left)
		}
		if v.right != nil {
			s.check(v.right)
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
		symbol_table := ScopedSymbolTable{make(map[string]Symbol), "Global", 0, nil}
		symbol_table.insert(&BuiltinSymbol{"INTEGER_CONST"})
		symbol_table.insert(&BuiltinSymbol{"REAL_CONST"})
		semantics_analyser := SemanticsAnalyser{&symbol_table}
		semantics_analyser.check(tree)
		fmt.Println("=========================================")
		interpret_tree(tree)
		for index, value := range *global_scope {
			fmt.Println(index, value)
		}
		fmt.Println(">>>>>>>>>>>>>SYMBOL_TABLE<<<<<<<<<<<<<<<<<<<")
		fmt.Print(symbol_table)
		fmt.Println("=========================================")
	} else {
		token := r.lexer.Cur()
		fmt.Fprintf(os.Stderr, "Unexpected token %d '%s' line [%d:%d]\n", token.ttype, token.tstring, token.line, token.column)
		os.Exit(-1)
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Need 1 parameter")
		os.Exit(-1)
	}
	flag.Parse()
	file, err := os.Open(flag.Arg(0))
	if err != nil {
		    log.Fatal(err)
	}
	defer file.Close()
	tokens := tokenize(file)
	for i, j := range tokens {
		fmt.Printf("Token[%d] := {%s} '%s'\n", i, reverse_lex[j.ttype], j.tstring)
	}
	rules := rules{lexer{0, len(tokens), tokens}}
	rules.Parse()
}
