package main

import (
	"aspire/are/io"
)

type Lexer struct {
	source string
	path   string
	// index of lexer in source (must be <len)
	pos uint
	len uint

	line uint
	col  uint
}

type Token struct {
	loc    Loc
	_type_ TokenType
}

type TokenType int

const (
	start TokenType = iota

	double_quote_string
	single_quote_string

	// literals
	number_lit
	string_lit
	ident_lit
	hex
	binary
	octal

	// keywords
	decl_keywords

	var_decl_keyword

	var_keyword
	spawn_keyword
	immortal_keyword
	static_keyword

	var_decl_keywords_end

	fn_keyword
	async_keyword

	decl_keywords_end

	class_keyword

	if_keyword
	else_keyword
	while_keyword
	break_keyword
	continue_keyword
	do_keyword
	for_keyword
	await_keyword
	typeof_keyword
	from_keyword
	import_keyword
	export_keyword
	return_keyword
	throw_keyword
	of_keyword
	in_keyword
	default_keyword
	switch_keyword
	case_keyword
	private_keyword
	public_keyword
	extends_keyword
	super_keyword
	new_keyword
	match_keyword
	ctor_keyword
	get_keyword
	set_keyword
	fallthrough_keyword
	void_keyword
	instanceof_keyword
	gt_keyword

	// symbols
	at
	colon
	semicolon
	comma
	open_brace
	close_brace
	open_bracket
	close_bracket
	open_paren
	close_paren
	forward_slash
	number_sep
	float
	dot
	rest_spread
	question
	nullish_coalesce
	arrow

	logical_ops

	not_op
	or_op
	and_op

	logical_ops_end

	or2_op

	assignment_ops

	equals
	// plus_equals
	// minus_equals

	assignment_ops_end

	plus_plus
	minus_minus

	binary_ops

	plus_op
	minus_op
	divide_op
	multiply_op
	modulo_op
	exponent_op

	binary_ops_end

	// groupings
	binary_op
	comparison_op
	assignment_op
	logical_op

	quote
	doc_comment
	line_comment
	newline
	EOF
	unknown
)

func newLexer(source string, path string) *Lexer {
	lexer := &Lexer{
		source: source,
		path:   path,
		pos:    0,
		len:    uint(len(source)),
		line:   1,
		col:    1,
	}
	return lexer
}

var TokenTypeNames map[TokenType]string = nil

func initTkNames() {
	if TokenTypeNames != nil {
		return
	}
	TokenTypeNames = map[TokenType]string{
		start:               "start",
		number_lit:          "number",
		string_lit:          "string",
		ident_lit:           "identifier",
		hex:                 "hex literal",
		binary:              "binary literal",
		octal:               "octal literal",
		var_keyword:         "var",
		spawn_keyword:       "spawn",
		immortal_keyword:    "immortal",
		static_keyword:      "static",
		fn_keyword:          "function",
		async_keyword:       "async",
		if_keyword:          "if",
		else_keyword:        "else",
		while_keyword:       "while",
		break_keyword:       "break",
		continue_keyword:    "continue",
		do_keyword:          "do",
		for_keyword:         "for",
		await_keyword:       "await",
		typeof_keyword:      "typeof",
		from_keyword:        "from",
		import_keyword:      "import",
		export_keyword:      "export",
		return_keyword:      "return",
		throw_keyword:       "throw",
		of_keyword:          "of",
		in_keyword:          "in",
		class_keyword:       "class",
		default_keyword:     "default",
		switch_keyword:      "switch",
		case_keyword:        "case",
		private_keyword:     "private",
		public_keyword:      "public",
		extends_keyword:     "extends",
		super_keyword:       "super",
		new_keyword:         "new",
		match_keyword:       "match",
		ctor_keyword:        "constructor",
		get_keyword:         "get",
		set_keyword:         "set",
		fallthrough_keyword: "fallthrough",
		void_keyword:        "void",
		instanceof_keyword:  "instanceof",
		gt_keyword:          "globalThis",
		at:                  "@",
		colon:               ":",
		semicolon:           ";",
		comma:               ",",
		open_brace:          "{",
		close_brace:         "}",
		open_bracket:        "[",
		close_bracket:       "]",
		open_paren:          "(",
		close_paren:         ")",
		forward_slash:       "/",
		number_sep:          "_",
		dot:                 ".",
		rest_spread:         "...",
		question:            "?",
		nullish_coalesce:    "??",
		not_op:              "!",
		or_op:               "|",
		and_op:              "&",
		or2_op:              "||",
		equals:              "=",
		plus_plus:           "++",
		minus_minus:         "--",
		plus_op:             "+",
		minus_op:            "-",
		divide_op:           "/",
		multiply_op:         "*",
		modulo_op:           "%",
		exponent_op:         "**",
		quote:               "quote",
		doc_comment:         "doc comment",
		line_comment:        "line comment",
		newline:             "newline",
		EOF:                 "EOF",
		unknown:             "unknown token",
	}
}

// TokenLexeme returns a human-readable name for a token type
func TokenLexeme(tt TokenType) string {
	initTkNames()
	if name, ok := TokenTypeNames[tt]; ok {
		return name
	}
	return io.Sprintf("token(%d)", tt)
}

func (l *Lexer) src(tk Token) string {
	return l.source[tk.loc.start:tk.loc.end]
}

func (l *Lexer) print(t Token) {
	src := ""
	if t._type_ == EOF {
		src = "<eof>"
	} else {
		src = l.src(t)
	}
	io.Printf("\x1b[32mToken\x1b[0m {\r\n - \x1b[36m%#v\x1b[0m\r\n - line: %d\r\n - col: %d\r\n}\r\n", src, t.loc.line, t.loc.col)
}

func (l *Lexer) tokenize() []Token {
	tokens := []Token{}
	for l.pos < l.len {
		tokens = append(tokens, l.next())
	}
	eof := l.EOFToken()
	tokens = append(tokens, eof)
	return tokens
}

func (l *Lexer) EOFToken() Token {
	eof := Token{
		loc: Loc{
			start: l.pos,
			end:   l.pos + 1,
			col:   l.col,
			line:  l.line,
		},
		_type_: EOF,
	}
	return eof
}

func (l *Lexer) advance() {
	l.pos++
	l.col++
}

func (l *Lexer) throw(msg string) {
	print("\x1b[31mParse Error\x1b[0m: ")
	loc := Loc{
		start: l.pos,
		end:   l.pos + 1,
		col:   l.col,
		line:  l.line,
	}
	io.Print(io.Sprintf("%s%s%s", msg, dbg.SourceWithinRange(l.path, loc), dbg.SourceAtPosition(l.path, loc)))
	exit_with_error()
}

func (l *Lexer) isAlpha(c byte) bool {
	if (c >= 65 && c <= 90) ||
		(c >= 97 && c <= 122) ||
		c == '_' || c == '#' {
		return true
	} else {
		return false
	}
}

func (l *Lexer) token() Token {
	return Token{
		loc: Loc{
			start: l.pos,
			end:   l.pos + 1,
			col:   l.col,
			line:  l.line,
		},
		_type_: unknown,
	}
}

func (l *Lexer) char() byte {
	if l.pos < l.len {
		return l.source[l.pos]
	}
	return 0
}

var keywords map[string]TokenType = nil

func initKeywords() {
	if keywords != nil {
		return
	}
	keywords = map[string]TokenType{
		"var":         var_keyword,
		"spawn":       spawn_keyword,
		"immortal":    immortal_keyword,
		"static":      static_keyword,
		"if":          if_keyword,
		"else":        else_keyword,
		"while":       while_keyword,
		"continue":    continue_keyword,
		"break":       break_keyword,
		"do":          do_keyword,
		"for":         for_keyword,
		"function":    fn_keyword,
		"async":       async_keyword,
		"await":       await_keyword,
		"typeof":      typeof_keyword,
		"from":        from_keyword,
		"import":      import_keyword,
		"export":      export_keyword,
		"return":      return_keyword,
		"throw":       throw_keyword,
		"of":          of_keyword,
		"in":          in_keyword,
		"class":       class_keyword,
		"super":       super_keyword,
		"new":         new_keyword,
		"extends":     extends_keyword,
		"default":     default_keyword,
		"case":        case_keyword,
		"switch":      switch_keyword,
		"match":       match_keyword,
		"constructor": ctor_keyword,
		"private":     private_keyword,
		"public":      public_keyword,
		"get":         get_keyword,
		"set":         set_keyword,
		"fallthrough": fallthrough_keyword,
		"void":        void_keyword,
		"instanceof":  instanceof_keyword,
		"globalThis":  gt_keyword,
	}
}

func (l *Lexer) next() Token {
	state := start
	var result Token = l.token()
state:
	switch state {
	case start:
		switch l.char() {
		case '$':
			{
				state = doc_comment
				goto state
			}
		case '\r', ' ', '\t':
			{
				l.advance()
				return l.next()
			}
		case '*', '-', '+', '%':
			{
				state = binary_op
				goto state
			}
		case '<', '>':
			{
				state = comparison_op
				goto state
			}
		case '&', '|':
			{
				state = logical_op
				goto state
			}
		case ':':
			{
				result._type_ = colon
				state = colon
				goto state
			}
		case ';':
			{
				result._type_ = semicolon
				state = semicolon
				goto state
			}
		case ',':
			{
				result._type_ = comma
				state = comma
				goto state
			}
		case '@':
			{
				result._type_ = at
				state = at
				goto state
			}
		case '{':
			{
				result._type_ = open_brace
				state = open_brace
				goto state
			}
		case '}':
			{
				result._type_ = close_brace
				state = close_brace
				goto state
			}
		case '(':
			{
				result._type_ = open_paren
				state = open_paren
				goto state
			}
		case ')':
			{
				result._type_ = close_paren
				state = close_paren
				goto state
			}
		case '[':
			{
				result._type_ = open_bracket
				state = open_bracket
				goto state
			}
		case ']':
			{
				result._type_ = close_bracket
				state = close_bracket
				goto state
			}
		case '\n':
			{
				state = newline
				goto state
			}
		case '/':
			{
				state = forward_slash
				goto state
			}
		case '"':
			{
				state = quote
				goto state
			}
		case '\'':
			{
				state = quote
				goto state
			}
		case '.':
			{
				state = dot
				goto state
			}
		case '=':
			{
				state = equals
				goto state
			}
		case '?':
			{
				state = question
				goto state
			}
		case '!':
			{
				state = not_op
				goto state
			}
		default:
			{
				ch := l.char()
				if l.isAlpha(ch) {
					state = ident_lit
					goto state
				} else if isInt(ch) {
					state = number_lit
					goto state
				}
				l.throw(io.Sprintf("unrecognised character: \x1b[33m%c\x1b[0m code: %v", l.char(), l.char()))
			}
		}
	case ident_lit:
		{
			l.advance()
			ch := l.char()
			if l.isAlpha(ch) || isInt(ch) || ch == '_' {
				state = ident_lit
				goto state
			} else {
				ident := l.source[result.loc.start:l.pos]
				result._type_ = ident_lit
				initKeywords()
				if t, ok := keywords[ident]; ok {
					result._type_ = t
				}
			}
		}
	case number_lit:
		{
			result._type_ = number_lit
			ch := l.char()
			l.advance()
			if isInt(l.char()) {
				state = number_lit
				goto state
			} else if l.char() == '_' {
				state = number_sep
				goto state
			} else if l.char() == '.' {
				l.advance()
				if isInt(l.char()) {
					state = float
					goto state
				} else {
					l.pos--
					l.col--
				}
			} else {
				// TODO: review
				if result.loc.start == l.pos-1 && l.source[l.pos-1] == '0' {
					switch ch {
					case 'x', 'X':
						state = hex
					case 'b', 'B':
						state = binary
					case 'o', 'O':
						state = octal
					default:
						goto end
					}
					goto state
				}
			}
		end:
		}
	case number_sep:
		{
			l.advance()
			if isInt(l.char()) {
				state = number_lit
				goto state
			} else {
				l.throw("'_' must separate successive digits")
			}
		}
	case float:
		{
		start:
			l.advance()
			if isInt(l.char()) {
				goto start
			}
		}
	case binary:
		{
			l.advance()
			ch := l.char()
			if isBinary(ch) {
				state = binary
				goto state
			}
			state = number_lit
		}
	case octal:
		{
			l.advance()
			ch := l.char()
			if isOctal(ch) {
				state = octal
				goto state
			}
			state = number_lit
		}
	case hex:
		{
			l.advance()
			ch := l.char()
			if isHex(ch) {
				state = hex
				goto state
			}
			state = number_lit
		}
	case line_comment:
		{
			l.advance()
			switch l.char() {
			case '\n':
				{
					result.loc.start = l.pos
					state = newline
					goto state
				}
			default:
				{
					if l.pos < l.len {
						state = line_comment
						goto state
					} else {
						return l.EOFToken()
					}
				}
			}
		}
	case doc_comment:
		{
			l.advance()
			switch l.char() {
			case '\n':
				{
					result.loc.start = l.pos
					state = newline
					goto state
				}
			case '$':
				{
					l.advance()
				}
			default:
				{
					if l.pos < l.len {
						state = doc_comment
						goto state
					}
				}
			}
		}
	case newline:
		{
			l.advance()
			l.col = 1
			l.line++
			result._type_ = newline
		}
	case colon:
		{
			l.advance()
		}
	case semicolon:
		{
			l.advance()
		}
	case comma:
		{
			l.advance()
		}
	case at:
		{
			l.advance()
		}
	case open_brace:
		{
			l.advance()
		}
	case close_brace:
		{
			l.advance()
		}
	case open_bracket:
		{
			l.advance()
		}
	case close_bracket:
		{
			l.advance()
		}
	case open_paren:
		{
			l.advance()
		}
	case close_paren:
		{
			l.advance()
		}
	case question:
		{
			l.advance()
			result._type_ = question
			if l.char() == '?' {
				l.advance()
				state = nullish_coalesce
				goto state
			}
		}
	case nullish_coalesce:
		{
			result._type_ = nullish_coalesce
			if l.char() == '=' {
				l.advance()
				result._type_ = assignment_op
			}
		}
	case forward_slash:
		{
			l.advance()
			switch l.char() {
			case '/':
				{
					state = line_comment
					goto state
				}
			default:
				{
					result._type_ = divide_op
				}
			}
		}
	case binary_op:
		{
			c := l.char()
			l.advance()
			switch c {
			case '*':
				result._type_ = multiply_op
				if l.char() == '*' {
					l.advance()
					result._type_ = exponent_op
				}
			case '-':
				result._type_ = minus_op
				ch := l.char()
				if isInt(ch) {
					state = number_lit
					goto state
				} else if ch == '-' {
					result._type_ = minus_minus
				}
			case '+':
				result._type_ = plus_op
				if l.char() == '+' {
					l.advance()
					result._type_ = plus_plus
				}
			case '/':
				result._type_ = divide_op
			case '%':
				result._type_ = modulo_op
			}
			if l.char() == '=' {
				l.advance()
				result._type_ = assignment_op
			}
		}
	case dot:
		{
			l.advance()
			result._type_ = dot
			if l.char() == '.' && l.pos+1 < l.len && l.source[l.pos+1] == '.' {
				l.advance()
				l.advance()
				result._type_ = rest_spread
			}
		}
	case equals:
		{
			l.advance()
			switch l.char() {
			case '=':
				{
					state = comparison_op
					goto state
				}
			case '>':
				{
					l.advance()
					result._type_ = arrow
				}
			default:
				{
					result._type_ = equals
				}
			}
		}
	case logical_op:
		{
			result._type_ = unknown
			del := l.char()
			l.advance()
			if l.char() == del {
				l.advance()
				if del == '&' {
					result._type_ = and_op
				} else {
					result._type_ = or_op
				}
			} else if del == '|' {
				result._type_ = or2_op
			}
		}
	case not_op:
		{
			l.advance()
			result._type_ = not_op
			switch l.char() {
			case '=':
				{
					state = comparison_op
					goto state
				}
			}
		}
	case comparison_op:
		{
			l.advance()
			result._type_ = comparison_op
			switch l.char() {
			case '=':
				{
					l.advance()
				}
			}
		}
	case quote:
		{
			q := l.char()
			l.advance()
			result._type_ = string_lit
			if q == '"' {
				state = double_quote_string
			} else {
				state = single_quote_string
			}
			goto state
		}

	case double_quote_string:
		{
			switch l.char() {
			case '"':
				{
					l.advance()
					break
				}
			case '\\':
				{
					l.advance() // \
					if l.pos+1 >= l.len {
						l.throw("invalid escape in string literal")
					}
					l.advance() // next character
					state = double_quote_string
					goto state
				}
			case '\n':
				{
					l.pos = result.loc.start
					l.throw("unclosed string literal")
					break
				}
			default:
				{
					if l.pos >= l.len {
						l.pos = result.loc.start
						l.throw("unclosed string literal")
					}
					l.advance()
					state = double_quote_string
					goto state
				}
			}
		}
	case single_quote_string:
		{
			switch l.char() {
			case '\'':
				{
					l.advance()
					break
				}
			case '\\':
				{
					l.advance() // \
					if l.pos+1 >= l.len {
						l.throw("invalid escape in string literal")
					}
					l.advance() // next character
					state = single_quote_string
					goto state
				}
			case '\n':
				{
					l.pos = result.loc.start
					l.throw("unclosed string literal")
					break
				}
			default:
				{
					if l.pos >= l.len {
						l.pos = result.loc.start
						l.throw("unclosed string literal")
					}
					l.advance()
					state = single_quote_string
					goto state
				}
			}
		}

	default:
		l.throw("unhandled token state in lexer")
	}
	result.loc.end = l.pos
	return result
}
