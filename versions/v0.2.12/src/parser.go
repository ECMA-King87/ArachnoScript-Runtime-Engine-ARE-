package main

import (
	"aspire/are/io"
	"slices"
)

var nodes *Array[Node] = newArray[Node]()
var DumbyNode = Node{tag: "Node", children: nil, data: nil, loc: DumbyLoc}
var DumbyLoc Loc = Loc{
	start: 0,
	end:   0,
	col:   1,
	line:  1,
}

type Parser struct {
	path          string
	token_index   int
	tokens        *Array[Token]
	lexer         *Lexer
	precedenceFns map[int]func(int) NodeIndex
}

func newParser(path string) *Parser {
	// env_vars.debug = true
	buffer := fs.readTextFile(path)
	return newParserInstance(buffer, path)
}

func newParserInstance(buffer string, path string) *Parser {
	path = fs.Abs(path)
	lexer := newLexer(buffer, path)
	tokens := lexer.tokenize()
	p := &Parser{
		path:          path,
		token_index:   0,
		tokens:        newArray(tokens...),
		lexer:         lexer,
		precedenceFns: map[int]func(int) NodeIndex{},
	}
	p.registerPrecedence(0, p.parse_expr)
	p.registerPrecedence(1, p.parse_ternary_expr)
	p.registerPrecedence(2, p.parse_globalThis_expr)
	p.registerPrecedence(3, p.parse_assignment_expr)
	p.registerPrecedence(4, p.parse_instanceof_expr)
	p.registerPrecedence(5, p.parse_void_expr)
	p.registerPrecedence(6, p.parse_await_expr)
	p.registerPrecedence(7, p.parse_logical_expr)
	p.registerPrecedence(8, p.parse_comparison_expr)
	p.registerPrecedence(9, p.parse_typeof_expr)
	p.registerPrecedence(10, p.parse_module_expr)
	p.registerPrecedence(11, p.parse_from_expr)
	p.registerPrecedence(12, p.parse_binary_expr)
	p.registerPrecedence(13, p.parse_new_expr)
	p.registerPrecedence(14, p.parse_fn_expr)
	p.registerPrecedence(15, p.parse_class_expr)
	p.registerPrecedence(16, p.parse_member_expr)
	p.registerPrecedence(17, p.parse_call_expr)
	p.registerPrecedence(18, p.parse_object)
	p.registerPrecedence(19, p.parse_array)
	return p
}

func (p *Parser) registerPrecedence(pre int, fn func(int) NodeIndex) {
	p.precedenceFns[pre] = fn
}

func (p *Parser) at(index int) Token {
	index += p.token_index
	return p.tokens.at(index)
}

func (p *Parser) atType(index int) TokenType {
	return p.at(index)._type_
}

func (p *Parser) isAt(t TokenType) bool {
	curr_t := p.atType(0)
	switch t {
	case binary_op:
		return curr_t > binary_ops && curr_t < binary_ops_end
	case logical_op:
		return curr_t > logical_ops && curr_t < logical_ops_end
	case var_decl_keyword:
		return curr_t > var_decl_keyword && curr_t < var_decl_keywords_end
	case decl_keywords:
		return curr_t > decl_keywords && curr_t < decl_keywords_end
	case assignment_op:
		return (curr_t > assignment_ops && curr_t < assignment_ops_end) || curr_t == assignment_op
	}
	return curr_t == t
}

func (p *Parser) notAt(t TokenType) bool {
	return !p.isAt(t)
}

func (p *Parser) not_eof() bool {
	return p.notAt(EOF)
}

func (p *Parser) expect(t TokenType) Token {
	if p.notAt(t) {
		tk := p.at(0)
		expectedName := TokenLexeme(t)
		gotName := TokenLexeme(tk._type_)
		p.throwParseError(io.Sprintf("expected a token of type \x1b[32m%s\x1b[0m, but got %s%s", expectedName, gotName, SourceLog(p.path, tk.loc)))
	}
	return p.eat()
}

func (p *Parser) eat() Token {
	tk := p.at(0)
	p.next()
	return tk
}

func (p *Parser) next() {
	p.token_index++
}

func (p *Parser) eatSemiColon() bool {
	var _bool_ = false
	for p.isAt(semicolon) {
		p.next()
		_bool_ = true
	}
	return _bool_
}

func (p *Parser) eatNewLine() bool {
	var _bool_ = false
	for p.isAt(newline) {
		p.next()
		_bool_ = true
	}
	return _bool_
}

// func (p *Parser) eatComment() bool {
// 	var _bool_ = false
// 	for p.isAt(doc_comment) || p.isAt(line_comment) {
// 		p.next()
// 		_bool_ = true
// 	}
// 	return _bool_
// }

func (p *Parser) getLoc(index NodeIndex) Loc {
	return p.getNode(index).loc
}

func (p *Parser) getNode(index NodeIndex) Node {
	return nodes.at(int(index))
}

func (p *Parser) Parse(main bool) *Node {
	program := &Node{
		tag:      "Program",
		children: []NodeIndex{},
		data: Program{
			path: p.path,
			main: main,
		},
		loc: Loc{
			start: 0,
			end:   0,
			col:   1,
			line:  1,
		},
	}
	for p.not_eof() {
		if p.eatNewLine() {
			// || p.eatComment() {
			continue
		}
		program.children = append(program.children, p.parse_stmt())
	}
	return program
}

const (
	member_expr_tag = "Member Expr"
	block_stmt_tag  = "Block"
	call_expr_tag   = "Call Expr"
)

func (p *Parser) parse_stmt() NodeIndex {
	defer p.eatSemiColon()
	switch p.atType(0) {
	case var_keyword, immortal_keyword, spawn_keyword, static_keyword:
		return p.parse_var_decl()
	case open_brace:
		return p.parse_block_stmt()
	case if_keyword:
		return p.parse_if_stmt()
	case async_keyword, fn_keyword:
		return p.parse_fn_decl(false)
	case import_keyword:
		return p.parse_import_stmt()
	case while_keyword:
		return p.parse_while_stmt()
	case do_keyword:
		return p.parse_do_while_stmt()
	case return_keyword:
		return p.parse_return_stmt()
	case throw_keyword:
		return p.parse_throw_stmt()
	case for_keyword:
		return p.parse_for_stmt()
	case class_keyword:
		return p.parse_class_decl(false)
	case at:
		return p.parse_label()
	case switch_keyword:
		return p.parse_switch_stmt()
	case fallthrough_keyword:
		return p.parse_fallthrough_stmt()
	case break_keyword:
		return p.parse_break_stmt()
	case continue_keyword:
		return p.parse_continue_stmt()
	default:
		return p.parse_expr(1)
	}
}

func (p *Parser) parse_continue_stmt() NodeIndex {
	loc := p.expect(continue_keyword).loc
	return nodes.push(Node{
		tag:      "Continue",
		children: nil,
		data:     nil,
		loc:      loc,
	})
}

func (p *Parser) parse_break_stmt() NodeIndex {
	loc := p.expect(break_keyword).loc
	return nodes.push(Node{
		tag:      "Break",
		children: nil,
		data:     nil,
		loc:      loc,
	})
}

func (p *Parser) parse_fallthrough_stmt() NodeIndex {
	loc := p.expect(fallthrough_keyword).loc
	return nodes.push(Node{
		tag:      "Fallthrough",
		children: nil,
		data:     nil,
		loc:      loc,
	})
}

func (p *Parser) parse_switch_stmt() NodeIndex {
	loc := p.expect(switch_keyword).loc
	defaultCase := []NodeIndex{}
	cases := []Case{}
	operand := p.parse_condition()
	p.expect(open_brace)
	for p.not_eof() && p.notAt(close_brace) {
		// p.eatComment()
		p.eatNewLine()
		if p.isAt(case_keyword) {
			p.next()
			cond := []NodeIndex{p.parse_expr(1)}
			for p.isAt(comma) {
				p.next()
				cond = append(cond, p.parse_expr(1))
			}
			p.expect(colon)
			body := []NodeIndex{}
			if p.notAt(open_brace) {
				body = append(body, p.parse_stmt())
			} else {
				body = p.parse_block()
			}
			cases = append(cases, Case{
				conditions: cond,
				body:       body,
			})
		} else if p.isAt(close_brace) {
		} else {
			p.expect(default_keyword)
			p.expect(colon)
			if p.notAt(open_brace) {
				defaultCase = append(defaultCase, p.parse_stmt())
			} else {
				defaultCase = p.parse_block()
			}
		}
	}
	p.expect(close_brace)
	return nodes.push(Node{
		tag:      "Switch Stmt",
		children: nil,
		data: SwitchStmt{
			cases:       cases,
			defaultCase: defaultCase,
			operand:     operand,
		},
		loc: loc,
	})
}

func (p *Parser) parse_label() NodeIndex {
	loc := p.expect(at).loc
	return nodes.push(Node{
		tag:      "Label",
		children: nil,
		data:     Label{p.getNode(p.parse_identifier()).data.(Identifier).symbol},
		loc:      loc,
	})
}

func (p *Parser) parse_class_decl(expr bool) NodeIndex {
	loc := p.expect(class_keyword).loc

	name := ""
	anonymous, has_ctor := false, false
	if expr {
		anonymous = true
	} else {
		name = p.lexer.src(p.expect(ident_lit))
	}
	props := []ClassProp{}
	extends := NodeIndex(-1)
	if p.isAt(extends_keyword) {
		p.next()
		extends = p.parse_by_precedence(parse_member_pre)
	}
	ctor := NodeIndex(-1)
	p.expect(open_brace)
	for p.not_eof() {
		p.eatNewLine()
		// p.eatComment()
		if p.isAt(ctor_keyword) {
			if has_ctor {
				p.throwParseError("multiple constructor implementations are not allowed." + SourceLog(p.path, p.at(0).loc))
			}
			ctor = p.parse_ctor()
		} else if p.isAt(close_brace) {
			break
		} else {
			props = append(props, p.parse_class_member())
		}
	}
	p.expect(close_brace)
	return nodes.push(Node{
		tag:      "Class Decl",
		children: nil,
		data: ClassDecl{
			name:      name,
			anonymous: anonymous,
			extends:   extends,
			ctor:      ctor,
			body:      props,
		},
		loc: loc,
	})
}

func (p *Parser) parse_class_member() ClassProp {
	modifiers := p.parse_modifiers()
	var async, dynamic_key = false, false
	if p.isAt(async_keyword) {
		p.next()
		async = true
	}
	var _type_ PropType = DataProp
	if slices.Contains(modifiers, default_keyword) {
		_type_ = DefaultProp
	}
	accessor := Accessor{
		get: false,
		set: false,
	}
	if p.isAt(get_keyword) || p.isAt(set_keyword) {
		if async {
			p.throwParseError("async modifier cannot be used here." + SourceLog(p.path, p.at(0).loc))
		}
		if _type_._default {
			p.throwParseError("default modifier cannot be used here." + SourceLog(p.path, p.at(0).loc))
		}
		src := p.lexer.src(p.eat())
		if src == "get" {
			accessor.get = true
		} else {
			accessor.set = true
		}
		_type_ = AccessorProp
	}
	var name_exp NodeIndex = -1
	var value NodeIndex = -1
	if p.isAt(open_bracket) {
		// dynamic key
		dynamic_key = true
		p.next()
		name_exp = p.parse_expr(1)
		p.expect(close_bracket)
	} else {
		name_exp = p.parse_primary_expr()
		switch p.getNode(name_exp).tag {
		case "Identifier", "String", "Number":
		default:
			p.throwParseError("invalid property key in declaration" + SourceLog(p.path, p.getLoc(name_exp)))
		}
	}
	if p.isAt(open_paren) {
		params := p.parse_params()
		body := p.parse_block()
		value = nodes.push(Node{
			tag:      "Function Decl",
			children: nil,
			data: FnDecl{
				name:      "",
				body:      body,
				params:    params,
				async:     async,
				anonymous: true,
				arrow:     true,
			},
			loc: p.getLoc(name_exp),
		})
	} else {
		if accessor.get || accessor.set {
			p.throwParseError("(get / set) accessors must be declared as functions." + SourceLog(p.path, p.at(0).loc))
		}
		if async {
			p.UnexpectedTokenError(p.at(0))
		}
		if p.isAt(equals) {
			p.next()
			value = p.parse_expr(1)
		}
		p.eatSemiColon()
	}
	return ClassProp{
		modifiers: modifiers,
		accessor:  accessor,
		key: struct {
			node    NodeIndex
			dynamic bool
		}{name_exp, dynamic_key},
		value:  value,
		_type_: _type_,
	}
}

func (p *Parser) parse_ctor() NodeIndex {
	loc := p.expect(ctor_keyword).loc
	params := []NodeIndex{}
	p.expect(open_paren)
	for p.not_eof() && p.notAt(close_paren) {
		modifiers := p.parse_modifiers()
		var param NodeIndex
		if p.isAt(rest_spread) {
			if len(modifiers) == 0 {
				param = p.parse_rest_spread(true)
			} else {
				p.throwParseError("modifiers cannot be used with rest parameters." + SourceLog(p.path, p.at(0).loc))
			}
		} else {
			param = p.parse_param()
		}
		params = append(params, nodes.push(Node{
			tag:      "Ctor Param",
			children: nil,
			data: CtorParam{
				modifiers: modifiers,
				param:     param,
			},
			loc: loc,
		}))
		if p.notAt(close_paren) {
			p.expect(comma)
		}
	}
	p.expect(close_paren)
	body := p.parse_block()
	return nodes.push(Node{
		tag:      "Function Decl",
		children: nil,
		data: FnDecl{
			name:   "constructor",
			body:   body,
			params: params,
			async:  false,
			// prevent being declared in scope
			anonymous: true,
			arrow:     false,
		},
		loc: loc,
	})
}

func (p *Parser) parse_modifiers() []TokenType {
	modifiers := []TokenType{}
	if p.isAt(public_keyword) {
		modifiers = append(modifiers, p.eat()._type_)
	}
	if p.isAt(private_keyword) {
		if len(modifiers) == 1 {
			p.throwParseError("private and public modifiers cannot be used together." + SourceLog(p.path, p.at(0).loc))
		}
		modifiers = append(modifiers, p.eat()._type_)
	}
	if p.isAt(static_keyword) {
		modifiers = append(modifiers, p.eat()._type_)
	}
	if p.isAt(default_keyword) {
		modifiers = append(modifiers, p.eat()._type_)
	}
	return modifiers
}

func (p *Parser) parse_for_stmt() NodeIndex {
	loc := p.expect(for_keyword).loc
	var trad_for_loop any
	var for_loop any
	p.expect(open_paren)
	if p.isAt(var_decl_keyword) {
		_, _type_ := p.parse_var_decl_keywords()
		lhs := p.parse_var_decl_lhs()
		op := int8(0)
		if p.isAt(of_keyword) || p.isAt(in_keyword) {
			t := p.eat()._type_
			if t == in_keyword {
				op = 1
			}
		} else {
			if p.notAt(equals) {
				p.UnexpectedTokenError(p.at(0))
			}
			p.next()
			rhs := p.parse_expr(1)
			// Construct a Var Declaration node in the same shape as parse_var_decl()
			inner := nodes.push(Node{
				tag:      "Var Declaration",
				children: nil,
				data: VarDecl{
					left:   lhs,
					right:  rhs,
					_type_: _type_,
				},
				loc: loc,
			})
			before := nodes.push(Node{
				tag:      "Var Declaration",
				children: []NodeIndex{inner},
				data:     nil,
				loc:      loc,
			})
			p.expect(semicolon)
			condition := p.parse_expr(1)
			p.expect(semicolon)
			after := p.parse_expr(1)
			trad_for_loop = TradForLoop{
				before:    before,
				condition: condition,
				after:     after,
			}
			goto end
		}
		rhs := p.parse_by_precedence(2)
		for_loop = ForLoop{
			lhs:    lhs,
			rhs:    rhs,
			op:     op,
			_type_: _type_,
		}
	} else {
		before := p.parse_expr(1)
		p.expect(semicolon)
		condition := p.parse_expr(1)
		p.expect(semicolon)
		after := p.parse_expr(1)
		trad_for_loop = TradForLoop{
			before:    before,
			condition: condition,
			after:     after,
		}
	}
end:
	p.expect(close_paren)
	var children []NodeIndex
	if !p.eatSemiColon() {
		if p.isAt(open_brace) {
			children = p.parse_block()
		} else if !p.eatNewLine() {
			children = []NodeIndex{p.parse_stmt()}
		} else {
			p.expect(open_brace)
		}
	}
	return nodes.push(Node{
		tag:      "For Loop",
		children: children,
		data: func() any {
			if trad_for_loop != nil {
				return trad_for_loop
			}
			return for_loop
		}(),
		loc: loc,
	})
}

func (p *Parser) parse_throw_stmt() NodeIndex {
	loc := p.expect(throw_keyword).loc
	value := NodeIndex(-1)
	value = p.parse_expr(1)
	return nodes.push(Node{
		tag:      "Throw Stmt",
		children: nil,
		data: ThrowStmt{
			value: value,
		},
		loc: loc,
	})
}

func (p *Parser) parse_return_stmt() NodeIndex {
	loc := p.expect(return_keyword).loc
	value := NodeIndex(-1)
	var has_value = false
	if p.notAt(semicolon) && p.notAt(close_brace) {
		value = p.parse_expr(1)
		has_value = true
	}
	return nodes.push(Node{
		tag:      "Return Stmt",
		children: nil,
		data: ReturnStmt{
			value:     value,
			has_value: has_value,
		},
		loc: loc,
	})
}

func (p *Parser) parse_do_while_stmt() NodeIndex {
	loc := p.expect(do_keyword).loc
	children := p.parse_block()
	p.expect(while_keyword)
	return nodes.push(Node{
		tag:      "While Stmt",
		children: children,
		data: WhileStmt{
			condition: p.parse_condition(),
			do:        true,
		},
		loc: loc,
	})
}

func (p *Parser) parse_while_stmt() NodeIndex {
	loc := p.expect(while_keyword).loc
	condition := p.parse_condition()
	var body = []NodeIndex{}
	if !p.eatSemiColon() {
		body = p.parse_block()
	}
	return nodes.push(Node{
		tag:      "While Stmt",
		children: body,
		data: WhileStmt{
			condition: condition,
			do:        false,
		},
		loc: loc,
	})
}

func (p *Parser) parse_import_stmt() NodeIndex {
	loc := p.expect(import_keyword).loc
	path := ""
	namespace := ""
	named := ObjectLiteral{
		props: NewMap[ObjectLitKey, NodeIndex](),
	}
	has_namespace := false
	script := false
	if p.isAt(string_lit) {
		path = p.getNode(p.parse_string()).data.(StringLiteral).value
		script = true
	} else {
		if p.isAt(ident_lit) {
			namespace = p.lexer.src(p.eat())
			has_namespace = true
			if p.notAt(from_keyword) {
				p.expect(comma)
			}
		}
		if p.notAt(from_keyword) || !has_namespace {
			named = p.getNode(p.parse_object_destructuring()).data.(ObjectLiteral)
		}
		p.expect(from_keyword)
		path = p.getNode(p.parse_string()).data.(StringLiteral).value
	}
	return nodes.push(Node{
		tag:      "Import Stmt",
		children: nil,
		data: ImportStmt{
			namespace: namespace,
			named:     named,
			path:      path,
			script:    script,
		},
		loc: loc,
	})
}

func (p *Parser) parse_fn_decl(expr bool) NodeIndex {
	async := false
	if p.isAt(async_keyword) {
		p.next()
		async = true
	}
	if p.isAt(open_paren) {
		// anonymous arrow function
		expr := p.parse_grouping_expr()
		if p.getNode(expr).tag != "Function Decl" {
			p.throwParseError("invalid arrow function syntax" + SourceLog(p.path, p.getLoc(expr)))
		}
	}
	tk := p.expect(fn_keyword)
	name := ""
	if !expr {
		name = p.lexer.src(p.expect(ident_lit))
	}
	params := p.parse_params()
	body := p.parse_block()
	return nodes.push(Node{
		tag:      "Function Decl",
		children: nil,
		data: FnDecl{
			name:      name,
			body:      body,
			params:    params,
			async:     async,
			anonymous: expr,
			arrow:     false,
		},
		loc: tk.loc,
	})
}

func (p *Parser) parse_params() []NodeIndex {
	p.expect(open_paren)
	params := []NodeIndex{}
	for p.not_eof() && p.notAt(close_paren) {
		param := NodeIndex(-1)
		if p.isAt(rest_spread) {
			param = p.parse_rest_spread(true)
		} else {
			param = p.parse_param()
		}
		params = append(params, param)
		if p.notAt(close_paren) {
			p.expect(comma)
		}
	}
	p.expect(close_paren)
	return params
}

func (p *Parser) parse_param() NodeIndex {
	param := p.parse_identifier()
	if p.isAt(equals) {
		loc := p.getLoc(param)
		right := p.parse_by_precedence(2)
		param = nodes.push(Node{
			tag:      "Assignment Expr",
			children: nil,
			data: AssignmentExpr{
				left:  param,
				right: right,
				op:    "=",
			},
			loc: Loc{
				start: loc.start,
				end:   p.getLoc(right).end,
				col:   loc.col,
				line:  loc.line,
			},
		})
	}
	return param
}

func (p *Parser) parse_if_stmt() NodeIndex {
	loc := p.expect(if_keyword).loc
	cond := p.parse_condition()
	if_block := []NodeIndex{}
	if p.isAt(open_brace) {
		if_block = p.parse_block()
	} else {
		if_block = append(if_block, p.parse_stmt())
	}
	else_block := []NodeIndex{}
	p.eatNewLine()
	if p.isAt(else_keyword) {
		p.next()
		p.eatNewLine()
		if p.isAt(open_brace) {
			else_block = p.parse_block()
		} else {
			else_block = append(else_block, p.parse_stmt())
		}
	}
	return nodes.push(Node{
		tag:      "If Stmt",
		children: if_block,
		data: IfStmt{
			cond:   cond,
			_else_: else_block,
		},
		loc: loc,
	})
}

func (p *Parser) parse_condition() NodeIndex {
	p.expect(open_paren)
	p.eatNewLine()
	cond := NodeIndex(-1)
	if p.isAt(var_decl_keyword) {
		cond = p.parse_var_decl()
	} else {
		cond = p.parse_expr(1)
	}
	p.eatNewLine()
	p.expect(close_paren)
	return cond
}

func (p *Parser) parse_block_stmt() NodeIndex {
	loc := p.at(0).loc
	return nodes.push(Node{
		tag:      block_stmt_tag,
		children: p.parse_block(),
		data:     nil,
		loc:      loc,
	})
}

func (p *Parser) parse_block() []NodeIndex {
	p.expect(open_brace)
	block := []NodeIndex{}
	for p.notAt(close_brace) && p.not_eof() {
		if p.eatNewLine() {
			//  || p.eatComment() {
			continue
		}
		block = append(block, p.parse_stmt())
	}
	p.expect(close_brace)
	return block
}

func (p *Parser) parse_var_decl() NodeIndex {
	tk, _type_ := p.parse_var_decl_keywords()
	loc := tk.loc
	decls := []NodeIndex{}
start:
	left, right, has_value := p.var_decl()
	if _type_ != "mutable" && _type_ != "var" && !has_value {
		p.throwParseError(io.Sprintf("%s declarations must be initialized.%s",
			_type_, SourceLog(p.path, p.tokens.at(p.token_index-1).loc)))
	}
	decls = append(decls, nodes.push(Node{
		tag:      "Var Declaration",
		children: nil,
		data: VarDecl{
			left:   left,
			right:  right,
			_type_: _type_,
		},
		loc: loc,
	}))
	if !p.eatSemiColon() {
		p.expect(comma)
		goto start
	}

	if len(decls) == 0 {
		p.UnexpectedTokenError(p.tokens.at(p.token_index - 1))
	}
	return nodes.push(Node{
		tag:      "Var Declaration",
		children: decls,
		data:     nil,
		loc:      loc,
	})
}

func (p *Parser) parse_var_decl_keywords() (Token, string) {
	tk := p.expect(var_decl_keyword)
	if tk._type_ != spawn_keyword && tk._type_ != var_keyword {
		p.expect(spawn_keyword)
	}
	_type_ := "mutable"
	switch tk._type_ {
	case var_keyword:
		_type_ = "var"
	case immortal_keyword:
		_type_ = "constant"
	case static_keyword:
		_type_ = "static"
	}
	return tk, _type_
}

// NodeIndex, NodeIndex, has_value
func (p *Parser) var_decl() (NodeIndex, NodeIndex, bool) {
	left := p.parse_var_decl_lhs()
	l_node := p.getNode(left)
	switch l_node.tag {
	case "Identifier", "Object", "Array":
	default:
		p.throwParseError(io.Sprintf("expected identifier after declaration keyword. but got %s%s", l_node.tag, SourceLog(p.path, l_node.loc)))
	}
	if p.notAt(equals) {
		return left, 0, false
	}
	p.next()
	right := p.parse_by_precedence(2)
	return left, right, true
}

func (p *Parser) parse_var_decl_lhs() NodeIndex {
	var left NodeIndex
	if p.isAt(open_brace) {
		left = p.parse_object_destructuring()
	} else if p.isAt(open_bracket) {
		left = p.parse_array(0)
	} else {
		left = p.parse_primary_expr()
	}
	return left
}

func (p *Parser) parse_object_destructuring() NodeIndex {
	loc := p.expect(open_brace).loc
	props := NewMap[ObjectLitKey, NodeIndex]()
	for p.notAt(close_brace) && p.not_eof() {
		key_node := NodeIndex(-1)
		value_node := NodeIndex(-1)
		require_value := false
		if p.isAt(ident_lit) {
			key_node = p.parse_identifier()
		} else if p.isAt(string_lit) {
			require_value = true
			key_node = p.parse_string()
		} else if p.isAt(number_lit) {
			require_value = true
			key_node = p.parse_number()
		}
		if require_value {
			p.expect(colon)
			value_node = p.parse_identifier()
		}
		props.set(ObjectLitKey{
			node:    key_node,
			dynamic: false,
			useKey:  !require_value,
		}, value_node)
		if p.notAt(close_brace) {
			p.expect(comma)
		}
	}
	p.expect(close_brace)
	return nodes.push(Node{
		tag:      "Object",
		children: nil,
		data: ObjectLiteral{
			props: props,
		},
		loc: loc,
	})
}

func (p *Parser) parse_expr(pre int) NodeIndex {
	if pre == 0 {
		pre++
	}
	switch p.atType(0) {
	case newline:
		p.next()
		return p.parse_expr(pre)
	case doc_comment, line_comment:
		p.next()
		return p.parse_expr(pre)
	default:
		return p.parse_by_precedence(pre)
	}
}

func (p *Parser) parse_instanceof_expr(pre int) NodeIndex {
	left := p.parse_by_precedence(pre + 1)
	if p.notAt(instanceof_keyword) {
		return left
	}
	p.next()
	right := p.parse_by_precedence(pre + 1)
	loc := Loc{
		start: p.getLoc(left).start,
		end:   p.getLoc(right).end,
		col:   p.getLoc(left).col,
		line:  p.getLoc(left).line,
	}
	return nodes.push(Node{
		tag:      "Instanceof Expr",
		children: nil,
		data: InstanceofExpr{
			left:  left,
			right: right,
		},
		loc: loc,
	})
}

func (p *Parser) parse_void_expr(pre int) NodeIndex {
	if p.notAt(void_keyword) {
		return p.parse_by_precedence(pre + 1)
	}
	loc := p.eat().loc
	return nodes.push(Node{
		tag:      "Void Expr",
		children: []NodeIndex{p.parse_by_precedence(pre + 1)},
		data:     nil,
		loc:      loc,
	})
}

func (p *Parser) parse_assignment_expr(pre int) NodeIndex {
	left := p.parse_by_precedence(pre + 1)
	if p.isAt(assignment_op) {
		op := p.lexer.src(p.eat())
		right := p.parse_expr(1)
		loc := p.getLoc(left)
		return nodes.push(Node{
			tag:      "Assignment Expr",
			children: nil,
			data: AssignmentExpr{
				left:  left,
				right: right,
				op:    op,
			},
			loc: Loc{
				start: loc.start,
				end:   p.getLoc(right).end,
				col:   loc.col,
				line:  loc.line,
			},
		})
	} else {
		return left
	}
}

func (p *Parser) parse_ternary_expr(pre int) NodeIndex {
	cond := p.parse_by_precedence(pre + 1)
	if p.notAt(question) {
		return cond
	}
	p.next()
	l_loc := p.getLoc(cond)
	then := p.parse_expr(1)
	p.expect(colon)
	_else := p.parse_expr(1)
	r_loc := p.getLoc(_else)
	return nodes.push(Node{
		tag:      "Ternary Expr",
		children: nil,
		loc: Loc{
			start: l_loc.start,
			end:   r_loc.end,
			col:   l_loc.col,
			line:  l_loc.line,
		},
		data: TernaryExpr{
			condition: cond,
			then:      then,
			_else:     _else,
		},
	})
}

func (p *Parser) parse_new_expr(pre int) NodeIndex {
	if p.notAt(new_keyword) {
		return p.parse_by_precedence(pre + 1)
	}
	loc := p.eat().loc

	return nodes.push(Node{
		tag:      "New Expr",
		children: []NodeIndex{p.parse_by_precedence(pre + 1)},
		loc:      loc,
		// data:     NewExpr{},
	})
}

func (p *Parser) parse_by_precedence(pre int) NodeIndex {
	if fn, ok := p.precedenceFns[pre]; ok {
		return fn(pre)
	}
	return p.parse_primary_expr()
}

func (p *Parser) parse_await_expr(pre int) NodeIndex {
	if p.isAt(await_keyword) {
		loc := p.eat().loc
		operand := p.parse_by_precedence(pre + 1)
		op_node := p.getNode(operand)
		if op_node.tag != "Call Expr" {
			p.throwParseError(io.Sprintf("await expects a function call as its operand.%s", SourceLog(p.path, op_node.loc)))
		}
		return nodes.push(Node{
			tag:      "Await Expr",
			children: []NodeIndex{operand},
			data:     nil,
			loc:      loc,
		})
	}
	return p.parse_by_precedence(pre + 1)
}

const parse_member_pre = 17

func (p *Parser) parse_rest_spread(param bool) NodeIndex {
	loc := p.expect(rest_spread).loc
	var operand NodeIndex
	if param {
		operand = p.parse_identifier()
	} else {
		operand = p.parse_call_expr(0)
	}
	return nodes.push(Node{
		tag:      "Rest or Spread Expr",
		children: []NodeIndex{operand},
		data:     nil,
		loc:      loc,
	})
}

func (p *Parser) parse_from_expr(pre int) NodeIndex {
	if p.notAt(from_keyword) {
		return p.parse_by_precedence(pre + 1)
	}
	loc := p.eat().loc
	return nodes.push(Node{
		tag:      "From Expr",
		children: nil,
		data:     FromExpr{p.getNode(p.parse_string()).data.(StringLiteral).value},
		loc:      loc,
	})
}

func (p *Parser) parse_typeof_expr(pre int) NodeIndex {
	if p.notAt(typeof_keyword) {
		return p.parse_by_precedence(pre + 1)
	}
	loc := p.eat().loc
	return nodes.push(Node{
		tag:      "Typeof Expr",
		children: []NodeIndex{p.parse_by_precedence(pre + 1)},
		data:     nil,
		loc:      loc,
	})
}

func (p *Parser) parse_module_expr(pre int) NodeIndex {
	curr_t := p.at(0)
	if curr_t._type_ == ident_lit && p.lexer.src(curr_t) == "module" {
		return p.parse_member_expr(parse_member_pre)
	}
	return p.parse_by_precedence(pre + 1)
}

func (p *Parser) parse_globalThis_expr(pre int) NodeIndex {
	curr_t := p.at(0)
	if curr_t._type_ == gt_keyword {
		object := nodes.push(Node{
			tag:      "Identifier",
			children: nil,
			data: Identifier{
				symbol: "globalThis",
			},
			loc: p.eat().loc,
		})
		if p.atType(0) == dot {
			p.eat() // dot
			member := p.parse_identifier()
			return nodes.push(Node{
				tag:      member_expr_tag,
				children: nil,
				data: MemberExpr{
					object:   object,
					member:   member,
					computed: false,
				},
				loc: curr_t.loc,
			})
		}
		return object
	}
	return p.parse_by_precedence(pre + 1)
}

func (p *Parser) parse_fn_expr(pre int) NodeIndex {
	if p.notAt(fn_keyword) {
		return p.parse_by_precedence(pre + 1)
	}
	return p.parse_fn_decl(true)
}

func (p *Parser) parse_class_expr(pre int) NodeIndex {
	if p.notAt(class_keyword) {
		return p.parse_by_precedence(pre + 1)
	}
	return p.parse_class_decl(true)
}

func (p *Parser) parse_member_expr(pre int) NodeIndex {
	object := p.parse_by_precedence(pre + 1)
	if p.notAt(dot) && p.notAt(open_bracket) {
		return object
	}
	for (p.isAt(dot) || p.isAt(open_bracket)) && p.not_eof() {
		computed := false
		member := NodeIndex(-1)
		if p.isAt(open_bracket) {
			computed = true
			p.next()
			member = p.parse_expr(1)
			p.expect(close_bracket)
		} else {
			// dot
			p.next()
			member = p.parse_identifier()
		}
		loc := p.getLoc(object)
		object = nodes.push(Node{
			tag:      member_expr_tag,
			children: []NodeIndex{},
			data: MemberExpr{
				object:   object,
				member:   member,
				computed: computed,
			},
			loc: Loc{
				start: loc.start,
				end:   p.getLoc(member).end,
				col:   loc.col,
				line:  loc.line,
			},
		})
		if p.isAt(open_paren) {
			object = p.parse_call(object)
		}
	}
	if p.isAt(plus_plus) || p.isAt(minus_minus) {
		tk := p.eat()
		o_loc := p.getLoc(object)
		loc := Loc{
			start: o_loc.start,
			end:   tk.loc.end,
			col:   o_loc.col,
			line:  o_loc.line,
		}
		t := tk._type_
		return nodes.push(Node{
			tag:      "Incre Expr",
			children: nil,
			data: IncreExpr{
				operand: object,
				op:      t,
				pre:     false,
			},
			loc: loc,
		})
	}
	return object
}

func (p *Parser) parse_array(pre int) NodeIndex {
	if p.notAt(open_bracket) {
		return p.parse_by_precedence(pre + 1)
	}
	loc := p.eat().loc
	elements := []NodeIndex{}
	for p.notAt(close_bracket) && p.not_eof() {
		elements = append(elements, p.parse_expr(1))
		if p.notAt(close_bracket) {
			p.expect(comma)
		}
	}
	p.expect(close_bracket)
	return nodes.push(Node{
		tag:      "Array",
		children: elements,
		data:     nil,
		loc:      loc,
	})
}

func (p *Parser) parse_object(pre int) NodeIndex {
	if p.notAt(open_brace) {
		return p.parse_by_precedence(pre + 1)
	}
	loc := p.eat().loc
	props := NewMap[ObjectLitKey, NodeIndex]()
	for p.notAt(close_brace) && p.not_eof() {
		p.eatNewLine()
		// p.eatComment()
		// p.eatNewLine()
		var async = false
		if p.isAt(async_keyword) {
			p.next()
			async = true
		}
		var key_exp NodeIndex = 0
		if p.isAt(open_bracket) {
			// dynamic key
			p.next()
			key_exp = p.parse_expr(1)
			p.expect(close_bracket)
		} else {
			key_exp = p.parse_primary_expr()
			switch p.getNode(key_exp).tag {
			case "Identifier", "String", "Number":
			default:
				p.throwParseError("invalid property key in object literal" + SourceLog(p.path, p.getLoc(key_exp)))
			}
			if (p.isAt(comma) || p.isAt(close_brace)) && p.getNode(key_exp).tag == "Identifier" {
				if p.isAt(comma) {
					p.eat()
				}
				props.set(ObjectLitKey{
					node:    key_exp,
					dynamic: false,
					useKey:  true,
				}, 0)
				continue
			}
		}
		if p.isAt(open_paren) {
			params := p.parse_params()
			body := p.parse_block()
			props.set(ObjectLitKey{
				node:    key_exp,
				dynamic: false,
				useKey:  false,
			}, nodes.push(Node{
				tag:      "Function Decl",
				children: []NodeIndex{},
				data: FnDecl{
					name:      "",
					body:      body,
					params:    params,
					async:     async,
					anonymous: true,
					arrow:     false,
				},
				loc: loc,
			}))
		} else {
			if async {
				p.UnexpectedTokenError(p.at(0))
			}
			p.expect(colon)
			p.eatNewLine()
			value := p.parse_expr(1)
			p.eatNewLine()
			props.set(ObjectLitKey{
				node:    key_exp,
				dynamic: false,
				useKey:  false,
			}, value)
			// if p.notAt(close_brace) {
			// 	p.expect(comma)
			// }
			if p.isAt(comma) {
				p.eat()
			}
		}
		p.eatNewLine()
	}
	p.expect(close_brace)
	return nodes.push(Node{
		tag:      "Object",
		children: nil,
		data: ObjectLiteral{
			props: props,
		},
		loc: loc,
	})
}

func (p *Parser) parse_call_expr(_ int) NodeIndex {
	caller := p.parse_member_expr(parse_member_pre)
	return p.parse_call(caller)
}

func (p *Parser) parse_call(caller NodeIndex) NodeIndex {
	if p.notAt(open_paren) {
		return caller
	}
	args := p.parse_expr_list()
	loc := p.getLoc(caller)
	loc.end = p.getLoc(args).end
	return nodes.push(Node{
		tag:      call_expr_tag,
		children: nil,
		data: CallExpr{
			caller: caller,
			args:   args,
		},
		loc: Loc{
			start: loc.start,
			end:   loc.end,
			col:   loc.col,
			line:  loc.line,
		},
	})
}

func (p *Parser) parse_comparison_expr(pre int) NodeIndex {
	left := p.parse_by_precedence(pre + 1)
	p.eatNewLine()
	if p.notAt(comparison_op) {
		return left
	}
	loc := p.getLoc(left)
	op := p.eat()
	right := NodeIndex(0)
	if p.isAt(open_paren) {
		right = p.parse_grouping_or2()
	} else {
		right = p.parse_by_precedence(pre)
	}
	return nodes.push(Node{
		tag:      "Comparison Expr",
		children: nil,
		data: ComparisonExpr{
			left:   left,
			right:  right,
			op:     op,
			op_src: p.lexer.src(op),
		},
		loc: Loc{
			start: loc.start,
			end:   p.getLoc(right).end,
			col:   loc.col,
			line:  loc.line,
		},
	})
}

func (p *Parser) parse_logical_expr(pre int) NodeIndex {
	var left NodeIndex
	var right NodeIndex
	var loc Loc
	var op TokenType
	if p.isAt(not_op) {
		loc = p.at(0).loc
		op = p.eat()._type_
		left = p.parse_by_precedence(pre + 1)
		right = -1
	} else {
		left = p.parse_by_precedence(pre + 1)
		p.eatNewLine()
		if p.notAt(logical_op) {
			return left
		}
		loc = p.getLoc(left)
		op = p.eat()._type_
		right = p.parse_expr(1)
	}
	return nodes.push(Node{
		tag:      "Logical Expr",
		children: nil,
		data: LogicalExpr{
			left:  left,
			right: right,
			op:    op,
		},
		loc: loc,
	})
}

func (p *Parser) parse_binary_expr(pre int) NodeIndex {
	return p.parse_additive(pre + 1)
}

func (p *Parser) parse_additive(pre int) NodeIndex {
	left := p.parse_multiplicative(pre)
	p.eatNewLine()
	if p.notAt(plus_op) && p.notAt(minus_op) {
		return left
	}
	op := p.eat()._type_
	right := p.parse_multiplicative(pre)
	l_loc := p.getLoc(left)
	r_loc := p.getLoc(right)
	start := l_loc.start
	end := r_loc.end
	left = nodes.push(Node{
		tag:      "Binary Expr",
		children: nil,
		data: BinaryExpr{
			left:  left,
			right: right,
			op:    op,
		},
		loc: Loc{
			start: start,
			end:   end,
			col:   l_loc.col,
			line:  l_loc.line,
		},
	})
	for p.eatNewLine(); p.isAt(plus_op) || p.isAt(minus_op); p.eatNewLine() {
		op := p.eat()._type_
		right := p.parse_multiplicative(pre)
		l_loc := p.getLoc(left)
		r_loc := p.getLoc(right)
		start := l_loc.start
		end := r_loc.end
		left = nodes.push(Node{
			tag:      "Binary Expr",
			children: nil,
			data: BinaryExpr{
				left:  left,
				right: right,
				op:    op,
			},
			loc: Loc{
				start: start,
				end:   end,
				col:   l_loc.col,
				line:  l_loc.line,
			},
		})
	}
	return left
}

func (p *Parser) parse_multiplicative(pre int) NodeIndex {
	left := p.parse_by_precedence(pre)
	p.eatNewLine()
	if p.notAt(multiply_op) && p.notAt(divide_op) && p.notAt(modulo_op) && p.notAt(exponent_op) {
		return left
	}
	op := p.eat()._type_
	right := p.parse_by_precedence(pre)
	l_loc := p.getLoc(left)
	r_loc := p.getLoc(right)
	start := l_loc.start
	end := r_loc.end
	left = nodes.push(Node{
		tag:      "Binary Expr",
		children: nil,
		data: BinaryExpr{
			left:  left,
			right: right,
			op:    op,
		},
		loc: Loc{
			start: start,
			end:   end,
			col:   l_loc.col,
			line:  l_loc.line,
		},
	})
	for p.eatNewLine(); p.isAt(multiply_op) || p.isAt(divide_op) || p.isAt(modulo_op) || p.isAt(exponent_op); p.eatNewLine() {
		op := p.eat()._type_
		right := p.parse_by_precedence(pre)
		l_loc := p.getLoc(left)
		r_loc := p.getLoc(right)
		start := l_loc.start
		end := r_loc.end
		left = nodes.push(Node{
			tag:      "Binary Expr",
			children: nil,
			data: BinaryExpr{
				left:  left,
				right: right,
				op:    op,
			},
			loc: Loc{
				start: start,
				end:   end,
				col:   l_loc.col,
				line:  l_loc.line,
			},
		})
	}
	return left
}

func (p *Parser) parse_identifier() NodeIndex {
	n := p.expect(ident_lit)
	return nodes.push(Node{
		tag:      "Identifier",
		children: nil,
		data:     Identifier{p.lexer.src(n)},
		loc:      n.loc,
	})
}

func (p *Parser) parse_number() NodeIndex {
	n := p.expect(number_lit)
	return nodes.push(Node{
		tag:      "Number",
		children: nil,
		data:     NumericLiteral{parseFloat(p.lexer.src(n))},
		loc:      n.loc,
	})
}

func (p *Parser) parse_string() NodeIndex {
	str := p.expect(string_lit)
	start := str.loc.start
	end := str.loc.end
	if start < end-1 {
		// omit opening quote
		start++
		// omit closing quote
		end--
	}
	tk := Token{
		loc: Loc{
			start: start,
			end:   end,
			col:   str.loc.col,
			line:  str.loc.line,
		},
		_type_: string_lit,
	}
	return nodes.push(Node{
		tag:      "String",
		children: nil,
		data:     StringLiteral{p.lexer.src(tk)},
		loc:      str.loc,
	})
}

func (p *Parser) parse_primary_expr() NodeIndex {
	switch p.atType(0) {
	case plus_plus, minus_minus:
		tk := p.eat()
		loc := tk.loc
		t := tk._type_
		operand := p.parse_member_expr(-2) // -2 to make sure that we parse either member or primary
		return nodes.push(Node{
			tag:      "Incre Expr",
			children: nil,
			data: IncreExpr{
				operand: operand,
				op:      t,
				pre:     true,
			},
			loc: loc,
		})
	case number_lit:
		return p.parse_number()
	case string_lit:
		return p.parse_string()
	case ident_lit:
		ident := p.parse_identifier()
		if p.isAt(plus_plus) || p.isAt(minus_minus) {
			tk := p.eat()
			i_loc := p.getLoc(ident)
			loc := Loc{
				start: i_loc.start,
				end:   tk.loc.end,
				col:   i_loc.col,
				line:  i_loc.line,
			}
			t := tk._type_
			return nodes.push(Node{
				tag:      "Incre Expr",
				children: nil,
				data: IncreExpr{
					operand: ident,
					op:      t,
					pre:     false,
				},
				loc: loc,
			})
		}
		return ident
	case open_paren:
		return p.parse_grouping_expr()
	default:
		p.UnexpectedTokenError(p.at(0))
	}
	return 0
}

func (p *Parser) parse_grouping_expr() NodeIndex {
	open_p := p.expect(open_paren)
	can_parse_arrow := p.isAt(close_paren)
	expressions := make([]NodeIndex, 0)
	valid_params := []string{"Identifier", "Assignment Expr", "Object", "Rest or Spread Expr"}
	for p.notAt(close_paren) && p.not_eof() {
		if len(expressions) != 0 {
			p.expect(comma)
		}
		idx := p.parse_expr(1)
		exp_node := p.getNode(idx)
		can_parse_arrow = slices.Contains(valid_params, exp_node.tag)
		expressions = append(expressions, idx)
	}
	close_p := p.expect(close_paren)
	if can_parse_arrow && p.isAt(arrow) {
		p.next()
		var body []NodeIndex
		if p.isAt(open_brace) {
			body = p.parse_block()
		} else {
			loc := p.at(0).loc
			body = []NodeIndex{nodes.push(Node{
				tag:      "Return Stmt",
				children: nil,
				data: ReturnStmt{
					value:     p.parse_expr(1),
					has_value: true,
				},
				loc: loc,
			})}
		}
		return nodes.push(Node{
			tag:      "Function Decl",
			children: nil,
			data: FnDecl{
				name:      "",
				body:      body,
				params:    expressions,
				async:     false,
				anonymous: true,
				arrow:     true,
			},
			loc: open_p.loc,
		})
	}
	return nodes.push(Node{
		tag:      "Grouping Expr",
		children: expressions,
		data:     nil,
		loc: Loc{
			start: open_p.loc.start,
			end:   close_p.loc.end,
			col:   open_p.loc.col,
			line:  open_p.loc.line,
		},
	})
}

func (p *Parser) parse_grouping_or2() NodeIndex {
	open_p := p.expect(open_paren)
	expressions := []NodeIndex{p.parse_expr(1)}
	if p.isAt(or2_op) {
		return p.parse_or2(expressions, open_p.loc)
	}
	for p.notAt(close_paren) && p.not_eof() {
		p.expect(comma)
		expressions = append(expressions, p.parse_expr(1))
	}
	close_p := p.expect(close_paren)
	return nodes.push(Node{
		tag:      "Grouping Expr",
		children: expressions,
		data:     nil,
		loc: Loc{
			start: open_p.loc.start,
			end:   close_p.loc.end,
			col:   open_p.loc.col,
			line:  open_p.loc.line,
		},
	})
}

func (p *Parser) parse_or2(exprs []NodeIndex, loc Loc) NodeIndex {
	for p.not_eof() && p.isAt(or2_op) {
		p.next()
		exprs = append(exprs, p.parse_expr(1))
	}
	p.expect(close_paren)
	return nodes.push(Node{
		tag:      "Or2 Expr",
		children: exprs,
		data:     nil,
		loc:      loc,
	})
}

func (p *Parser) parse_expr_list() NodeIndex {
	open_p := p.expect(open_paren)
	expressions := []NodeIndex{}
	for p.notAt(close_paren) && p.not_eof() {
		expr := NodeIndex(-1)
		if p.isAt(rest_spread) {
			expr = p.parse_rest_spread(false)
		} else {
			expr = p.parse_expr(1)
		}
		expressions = append(expressions, expr)
		p.eatNewLine()
		if p.notAt(close_paren) {
			p.expect(comma)
		}
	}
	close_p := p.expect(close_paren)
	return nodes.push(Node{
		tag:      "Grouping Expr",
		children: expressions,
		data:     nil,
		loc: Loc{
			start: open_p.loc.start,
			end:   close_p.loc.end,
			col:   open_p.loc.col,
			line:  open_p.loc.line,
		},
	})
}

func (p *Parser) UnexpectedTokenError(tk Token) {
	tokenName := TokenLexeme(tk._type_)
	p.throwParseError(io.Sprintf("unexpected token reached: %s%s", tokenName, SourceLog(p.path, tk.loc)))
}

func (p *Parser) throwParseError(msg string) {
	err_msg := io.Sprintf("\x1b[31mSyntaxError\x1b[0m: %s", msg)
	if env_vars.debug {
		panic(err_msg)
	}
	io.Println(err_msg)
	exit_with_error()
}

// ---------------------------------------------------------------
// ---------------------------- NODES ----------------------------

type (
	NodeType  = string
	NodeIndex = int
)

type Node struct {
	tag      NodeType
	children []NodeIndex
	data     any
	loc      Loc
}

// Node Data

type Program struct {
	path string
	main bool
}

type FnDecl struct {
	// expr
	name      string
	body      []NodeIndex
	params    []NodeIndex
	async     bool
	anonymous bool
	arrow     bool
}

type IfStmt struct {
	cond   NodeIndex
	_else_ []NodeIndex
}

type VarDecl struct {
	left   NodeIndex
	right  NodeIndex
	_type_ string
}

type ImportStmt struct {
	namespace string
	named     ObjectLiteral
	path      string
	script    bool
}

type WhileStmt struct {
	condition NodeIndex
	do        bool
}

type ReturnStmt struct {
	value     NodeIndex
	has_value bool
}

type ThrowStmt struct {
	value NodeIndex
}

type TradForLoop struct {
	before    NodeIndex
	condition NodeIndex
	after     NodeIndex
}

type ForLoop struct {
	lhs    NodeIndex
	rhs    NodeIndex
	op     int8 // (0 = of | 1 = in)
	_type_ string
}

type ClassDecl struct {
	name      string
	anonymous bool
	extends   NodeIndex
	ctor      NodeIndex
	body      []ClassProp
}

type Accessor struct {
	get bool
	set bool
}

type PropType struct {
	data     bool
	accessor bool
	_default bool
}

var DataProp = PropType{
	data:     true,
	accessor: false,
	_default: false,
}

var AccessorProp = PropType{
	data:     false,
	accessor: true,
	_default: false,
}

var DefaultProp = PropType{
	data:     false,
	accessor: false,
	_default: true,
}

type ClassProp struct {
	modifiers []TokenType
	accessor  Accessor
	_type_    PropType
	key       struct {
		node    NodeIndex
		dynamic bool
	}
	value NodeIndex
}

type CtorParam struct {
	modifiers []TokenType
	param     NodeIndex
}

type Label struct {
	ident string
}

type SwitchStmt struct {
	operand     NodeIndex
	cases       []Case
	defaultCase []NodeIndex
}

type Case struct {
	conditions []NodeIndex
	body       []NodeIndex
}

// ------------------------------------------

type Module = MemberExpr

type GlobalThis = Module

type InstanceofExpr struct {
	left, right NodeIndex
}

type TernaryExpr struct {
	condition, then, _else NodeIndex
}

// type NewExpr struct {
// 	operand NodeIndex
// }

type IncreExpr struct {
	operand NodeIndex
	op      TokenType
	pre     bool
}

type FromExpr struct {
	path string
}

type MemberExpr struct {
	object   NodeIndex
	member   NodeIndex
	computed bool
}

// type ArrayLiteral struct {
// 	elements []Node
// }

type ObjectLitKey struct {
	node    NodeIndex
	dynamic bool
	useKey  bool
}

type ObjectLiteral struct {
	props *Map[ObjectLitKey, NodeIndex]
}

type CallExpr struct {
	caller NodeIndex
	// grouping expression
	args NodeIndex
}

type ComparisonExpr struct {
	left   NodeIndex
	right  NodeIndex
	op     Token
	op_src string
}

type LogicalExpr struct {
	left  NodeIndex
	right NodeIndex
	op    TokenType
}

type AssignmentExpr struct {
	left  NodeIndex
	right NodeIndex
	op    string
}

type BinaryExpr struct {
	left  NodeIndex
	right NodeIndex
	op    TokenType
}

type Identifier struct {
	symbol string
}

type StringLiteral struct {
	value string
}

type NumericLiteral struct {
	value float64
}
