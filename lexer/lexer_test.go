package lexer

import (
	"reflect"
	"testing"
)

func TestLexerNextToken(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []Token
	}{
		{
			name:  "simple keyword sequence",
			input: "device motor : PLC",
			want: []Token{
				{Type: TOKEN_DEVICE, Literal: "device", Line: 1},
				{Type: TOKEN_IDENT, Literal: "motor", Line: 1},
				{Type: TOKEN_COLON, Literal: ":", Line: 1},
				{Type: TOKEN_IDENT, Literal: "PLC", Line: 1},
				{Type: TOKEN_EOF, Literal: "", Line: 1},
			},
		},
		{
			name:  "variable declaration",
			input: "vol i32 speed = 0",
			want: []Token{
				{Type: TOKEN_VOL, Literal: "vol", Line: 1},
				{Type: TOKEN_I32, Literal: "i32", Line: 1},
				{Type: TOKEN_IDENT, Literal: "speed", Line: 1},
				{Type: TOKEN_ASSIGN, Literal: "=", Line: 1},
				{Type: TOKEN_INT_LIT, Literal: "0", Line: 1},
				{Type: TOKEN_EOF, Literal: "", Line: 1},
			},
		},
		{
			name:  "time literal",
			input: "watchdog: 500ms",
			want: []Token{
				{Type: TOKEN_WATCHDOG, Literal: "watchdog", Line: 1},
				{Type: TOKEN_COLON, Literal: ":", Line: 1},
				{Type: TOKEN_TIME_LIT, Literal: "500ms", Line: 1},
				{Type: TOKEN_EOF, Literal: "", Line: 1},
			},
		},
		{
			name:  "boolean literals",
			input: "true false",
			want: []Token{
				{Type: TOKEN_TRUE, Literal: "true", Line: 1},
				{Type: TOKEN_FALSE, Literal: "false", Line: 1},
				{Type: TOKEN_EOF, Literal: "", Line: 1},
			},
		},
		{
			name:  "comment followed by token",
			input: "// comment\ndevice",
			want: []Token{
				{Type: TOKEN_DEVICE, Literal: "device", Line: 2},
				{Type: TOKEN_EOF, Literal: "", Line: 2},
			},
		},
		{
			name:  "or condition",
			input: "if psi >= max_psi OR flow >= max_flow",
			want: []Token{
				{Type: TOKEN_IF, Literal: "if", Line: 1},
				{Type: TOKEN_IDENT, Literal: "psi", Line: 1},
				{Type: TOKEN_GTE, Literal: ">=", Line: 1},
				{Type: TOKEN_IDENT, Literal: "max_psi", Line: 1},
				{Type: TOKEN_OR, Literal: "OR", Line: 1},
				{Type: TOKEN_IDENT, Literal: "flow", Line: 1},
				{Type: TOKEN_GTE, Literal: ">=", Line: 1},
				{Type: TOKEN_IDENT, Literal: "max_flow", Line: 1},
				{Type: TOKEN_EOF, Literal: "", Line: 1},
			},
		},
		{
			name:  "line tracking second line token",
			input: "\nboot",
			want: []Token{
				{Type: TOKEN_BOOT, Literal: "boot", Line: 2},
				{Type: TOKEN_EOF, Literal: "", Line: 2},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			l := New(tc.input)
			var got []Token
			for {
				tok := l.NextToken()
				got = append(got, tok)
				if tok.Type == TOKEN_EOF {
					break
				}
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("tokens mismatch\n got: %#v\nwant: %#v", got, tc.want)
			}
		})
	}
}
