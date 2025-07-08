package token

import "unicode"

type Token uint64

const (
	UNEXPECTED Token = iota
	EOF
	COMMENT

	literal_init
	IDENTIFIER
	INTEGER
	FLOAT
	STRING
	literal_end

	operator_init
	ASSIGN
	QUOTE
	COMMA
	METHOD_CALL
	DOT
	LPAREN
	RPAREN
	LCURLY
	RCURLY
	LBRACKET
	RBRACKET
	ARROW
	DOUBLE_DOT
	ELLIPSIS
	TILDE
	operator_end

	binary_op_init
	ADD
	SUB
	MUL
	DIV
	REM
	LT
	LE
	GT
	GE
	EQ
	NEQ
	BOR
	BAND
	BXOR
	BSHL
	BSHR
	binary_op_end

	keyword_init
	TRUE
	FALSE
	NOT
	NIL
	LET
	VAR
	AND
	OR
	FOR
	IF
	ELSE
	WHILE
	BREAK
	CONTINUE
	IN
	FUN
	RET
	IMPORT
	EXPORT
	ENUM
	REC
	keyword_end
)

var Tokens = [...]string{
	UNEXPECTED:  "Unrecognizable character",
	EOF:         "EOF",
	COMMENT:     "Comment",
	IDENTIFIER:  "Identifier",
	INTEGER:     "Integer",
	FLOAT:       "Float",
	STRING:      "String",
	ASSIGN:      "=",
	QUOTE:       "\"",
	COMMA:       ",",
	METHOD_CALL: ":",
	DOT:         ".",
	LPAREN:      "(",
	RPAREN:      ")",
	LCURLY:      "{",
	RCURLY:      "}",
	LBRACKET:    "[",
	RBRACKET:    "]",
	ARROW:       "=>",
	DOUBLE_DOT:  "..",
	ELLIPSIS:    "...",
	TILDE:       "~",
	ADD:         "+",
	SUB:         "-",
	MUL:         "*",
	DIV:         "/",
	REM:         "%",
	LT:          "<",
	LE:          "<=",
	GT:          ">",
	GE:          ">=",
	EQ:          "==",
	NEQ:         "!=",
	BOR:         "|",
	BAND:        "&",
	BXOR:        "^",
	BSHL:        "<<",
	BSHR:        ">>",
	AND:         "and",
	OR:          "or",
	FOR:         "for",
	NOT:         "not",
	TRUE:        "true",
	FALSE:       "false",
	NIL:         "nil",
	LET:         "let",
	VAR:         "var",
	IF:          "if",
	ELSE:        "else",
	WHILE:       "while",
	BREAK:       "break",
	CONTINUE:    "continue",
	IN:          "in",
	FUN:         "fun",
	RET:         "ret",
	IMPORT:      "import",
	EXPORT:      "export",
	ENUM:        "enum",
	REC:         "rec",
}

type TokenInfo struct {
	Lit   string
	Line  uint
	Token Token
}

var keywords map[string]Token

func init() {
	keywords = make(map[string]Token, keyword_end-(keyword_init+1))
	for i := keyword_init + 1; i < keyword_end; i++ {
		keywords[Tokens[i]] = i
	}
}

func (token Token) String() string {
	return Tokens[token]
}

func (token Token) IsLiteral() bool {
	return literal_init < token && token < literal_end
}

func (token Token) IsKeyword() bool {
	return keyword_init < token && token < keyword_end
}

func (token Token) IsBinaryOperator() bool {
	return (binary_op_init < token && token < binary_op_end) || token == AND || token == OR || token == IN
}

func IsKeyword(name string) bool {
	_, ok := keywords[name]
	return ok
}

func LookUp(name string) Token {
	if token, is_keyword := keywords[name]; is_keyword {
		return token
	}
	return IDENTIFIER
}

func IsIdentifier(name string) bool {
	if name == "" || IsKeyword(name) {
		return false
	}
	for i, c := range name {
		if !unicode.IsLetter(c) && c != '_' && (i == 0 || !unicode.IsDigit(c)) {
			return false
		}
	}
	return true
}

const (
	LowestPrec  = 0
	PrefixPrec  = 6
	HighestPrec = 7
)

func (op Token) Precedence() int {
	switch op {
	case OR:
		return 1
	case AND:
		return 2
	case EQ, NEQ, LT, LE, GT, GE, IN:
		return 3
	case ADD, SUB, BOR, BXOR:
		return 4
	case MUL, DIV, REM, BSHL, BSHR, BAND:
		return 5
	}
	return LowestPrec
}
