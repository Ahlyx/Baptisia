package lexer

type TokenType string

const (
	TOKEN_DEVICE    TokenType = "DEVICE"
	TOKEN_VARS      TokenType = "VARS"
	TOKEN_BOOT      TokenType = "BOOT"
	TOKEN_INPUTS    TokenType = "INPUTS"
	TOKEN_OUTPUTS   TokenType = "OUTPUTS"
	TOKEN_SAFETY    TokenType = "SAFETY"
	TOKEN_FAILSAFE  TokenType = "FAILSAFE"
	TOKEN_CONTROL   TokenType = "CONTROL"
	TOKEN_CONST     TokenType = "CONST"
	TOKEN_STATES    TokenType = "STATES"
	TOKEN_VOL       TokenType = "VOL"
	TOKEN_WATCHDOG  TokenType = "WATCHDOG"
	TOKEN_CYCLE     TokenType = "CYCLE"
	TOKEN_BOOL      TokenType = "BOOL"
	TOKEN_I32       TokenType = "I32"
	TOKEN_F32       TokenType = "F32"
	TOKEN_IF        TokenType = "IF"
	TOKEN_ELSE      TokenType = "ELSE"
	TOKEN_AND       TokenType = "AND"
	TOKEN_OR        TokenType = "OR"
	TOKEN_SENSOR    TokenType = "SENSOR"
	TOKEN_ACTUATOR  TokenType = "ACTUATOR"
	TOKEN_VAL       TokenType = "VAL"
	TOKEN_OUTPUT    TokenType = "OUTPUT"
	TOKEN_TRUE      TokenType = "TRUE"
	TOKEN_FALSE     TokenType = "FALSE"
	TOKEN_OFF       TokenType = "OFF"
	TOKEN_INT_LIT   TokenType = "INT_LIT"
	TOKEN_FLOAT_LIT TokenType = "FLOAT_LIT"
	TOKEN_TIME_LIT  TokenType = "TIME_LIT"
	TOKEN_IDENT     TokenType = "IDENT"
	TOKEN_ASSIGN    TokenType = "ASSIGN"
	TOKEN_GTE       TokenType = "GTE"
	TOKEN_GT        TokenType = "GT"
	TOKEN_LTE       TokenType = "LTE"
	TOKEN_NEQ       TokenType = "NEQ"
	TOKEN_EQ        TokenType = "EQ"
	TOKEN_LT        TokenType = "LT"
	TOKEN_COLON     TokenType = "COLON"
	TOKEN_LBRACE    TokenType = "LBRACE"
	TOKEN_RBRACE    TokenType = "RBRACE"
	TOKEN_LPAREN    TokenType = "LPAREN"
	TOKEN_RPAREN    TokenType = "RPAREN"
	TOKEN_COMMA     TokenType = "COMMA"
	TOKEN_EOF       TokenType = "EOF"
	TOKEN_ILLEGAL   TokenType = "ILLEGAL"
)

type Token struct {
	Type    TokenType
	Literal string
	Line    int
}
