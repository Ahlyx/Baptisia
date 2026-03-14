package lexer

type Lexer struct {
	input string
	pos   int
	line  int
}

func New(input string) *Lexer {
	return &Lexer{input: input, line: 1}
}

func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	if l.pos >= len(l.input) {
		return Token{Type: TOKEN_EOF, Literal: "", Line: l.line}
	}

	ch := l.input[l.pos]

	switch ch {
	case '{':
		l.pos++
		return Token{Type: TOKEN_LBRACE, Literal: "{", Line: l.line}
	case '}':
		l.pos++
		return Token{Type: TOKEN_RBRACE, Literal: "}", Line: l.line}
	case '(':
		l.pos++
		return Token{Type: TOKEN_LPAREN, Literal: "(", Line: l.line}
	case ')':
		l.pos++
		return Token{Type: TOKEN_RPAREN, Literal: ")", Line: l.line}
	case ',':
		l.pos++
		return Token{Type: TOKEN_COMMA, Literal: ",", Line: l.line}
	case ':':
		l.pos++
		return Token{Type: TOKEN_COLON, Literal: ":", Line: l.line}
	case '=':
		l.pos++
		return Token{Type: TOKEN_ASSIGN, Literal: "=", Line: l.line}
	case '<':
		l.pos++
		return Token{Type: TOKEN_LT, Literal: "<", Line: l.line}
	case '>':
		if l.peek() == '=' {
			l.pos += 2
			return Token{Type: TOKEN_GTE, Literal: ">=", Line: l.line}
		}
		l.pos++
		return Token{Type: TOKEN_ILLEGAL, Literal: ">", Line: l.line}
	case '-':
		if l.isDigit(l.peek()) {
			return l.readNumber()
		}
		l.pos++
		return Token{Type: TOKEN_ILLEGAL, Literal: "-", Line: l.line}
	case '/':
		if l.peek() == '/' {
			l.skipComment()
			return l.NextToken()
		}
		l.pos++
		return Token{Type: TOKEN_ILLEGAL, Literal: "/", Line: l.line}
	}

	if l.isLetter(ch) {
		return l.readWord()
	}

	if l.isDigit(ch) {
		return l.readNumber()
	}

	l.pos++
	return Token{Type: TOKEN_ILLEGAL, Literal: string(ch), Line: l.line}
}

func (l *Lexer) skipComment() {
	for l.pos < len(l.input) && l.input[l.pos] != '\n' {
		l.pos++
	}
}

func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.input) {
		switch l.input[l.pos] {
		case ' ', '\t', '\r':
			l.pos++
		case '\n':
			l.line++
			l.pos++
		default:
			return
		}
	}
}

func (l *Lexer) peek() byte {
	if l.pos+1 >= len(l.input) {
		return 0
	}
	return l.input[l.pos+1]
}

func (l *Lexer) readWord() Token {
	start := l.pos
	for l.pos < len(l.input) && (l.isLetter(l.input[l.pos]) || l.isDigit(l.input[l.pos]) || l.input[l.pos] == '_') {
		l.pos++
	}
	word := l.input[start:l.pos]
	tokenType := l.lookupKeyword(word)
	return Token{Type: tokenType, Literal: word, Line: l.line}
}

func (l *Lexer) readNumber() Token {
	start := l.pos

	if l.pos < len(l.input) && l.input[l.pos] == '-' {
		l.pos++
	}

	for l.pos < len(l.input) && l.isDigit(l.input[l.pos]) {
		l.pos++
	}

	if l.pos < len(l.input) && l.input[l.pos] == '.' {
		l.pos++
		for l.pos < len(l.input) && l.isDigit(l.input[l.pos]) {
			l.pos++
		}
		return Token{Type: TOKEN_FLOAT_LIT, Literal: l.input[start:l.pos], Line: l.line}
	}

	if l.pos+1 < len(l.input) && l.input[l.pos] == 'm' && l.input[l.pos+1] == 's' {
		l.pos += 2
		return Token{Type: TOKEN_TIME_LIT, Literal: l.input[start:l.pos], Line: l.line}
	}

	return Token{Type: TOKEN_INT_LIT, Literal: l.input[start:l.pos], Line: l.line}
}

func (l *Lexer) lookupKeyword(word string) TokenType {
	if t, ok := keywords[word]; ok {
		return t
	}
	return TOKEN_IDENT
}

var keywords = map[string]TokenType{
	"device":   TOKEN_DEVICE,
	"vars":     TOKEN_VARS,
	"boot":     TOKEN_BOOT,
	"inputs":   TOKEN_INPUTS,
	"outputs":  TOKEN_OUTPUTS,
	"safety":   TOKEN_SAFETY,
	"failsafe": TOKEN_FAILSAFE,
	"control":  TOKEN_CONTROL,
	"vol":      TOKEN_VOL,
	"watchdog": TOKEN_WATCHDOG,
	"cycle":    TOKEN_CYCLE,
	"b":        TOKEN_BOOL,
	"i32":      TOKEN_I32,
	"f32":      TOKEN_F32,
	"if":       TOKEN_IF,
	"else":     TOKEN_ELSE,
	"AND":      TOKEN_AND,
	"OR":       TOKEN_OR,
	"sensor":   TOKEN_SENSOR,
	"actuator": TOKEN_ACTUATOR,
	"val":      TOKEN_VAL,
	"output":   TOKEN_OUTPUT,
	"true":     TOKEN_TRUE,
	"false":    TOKEN_FALSE,
	"off":      TOKEN_OFF,
	"const":    TOKEN_CONST,
	"states":   TOKEN_STATES,
}

func (l *Lexer) isLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

func (l *Lexer) isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}
