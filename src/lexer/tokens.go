package lexer

type TokenType string

type Token struct {
	Token   TokenType
	Line    int
	Column  int
	Literal string
}

const (
	// Basic syntax

	LET    = "LET"
	FUNC   = "FUNC"
	IDENT  = "INDENT"
	NUMBER = "NUMBER"
	IF     = "IF"
	ELSE   = "ELSE"
	TRUE   = "TRUE"
	FALSE  = "FALSE"
	RETURN = "RETURN"

	PLUS   = "PLUS"
	HYPHEN = "HYPHEN"
	SLASH  = "SLASH"

	// Special token

	DQUOTE  = "DQUOTE"  // Double quote
	SCOLUMN = "SCOLUMN" // Semi column

	ASSIGN = "ASSIGN"

	// Boolean

	LT         = "LT"
	GT         = "GT"
	LTE        = "LTE"
	GTE        = "GTE"
	EQ         = "EQ"
	NEQ        = "NEQ"
	BANG       = "BANG"
	DAMPERSAND = "DAMPERSAND" // [D]ouble ampersand
	DVERTLINE  = "DVERTLINE"  // [D]ouble vertical line

	//TODO add binary

	// Binary

	BLEFT    = "BLEFT"   // [B]rackets left
	BRIGHT   = "BRIGHT"  // [B]rackets right
	SBLEFT   = "SBLEFT"  // [S]quare [B]rackets left
	SBRIGHT  = "SBRIGHT" // [S]quare [B]rackets right
	BRLEFT   = "BRLEFT"  // [Br]aces left
	BRRIGHT  = "BRRIGHT" // [Br]aces left
	ASTERISK = "ASTERISK"

	COMA    = "COMA"
	ILLEGAL = "ILLEGAL"
	NIL     = "NIL"
	EOF     = "EOF"
)

func NewToken(token TokenType, line, column int, literal string) Token {
	return Token{
		Token:   token,
		Line:    line,
		Column:  column,
		Literal: literal,
	}
}

var keywords = map[string]TokenType{
	"fn":     FUNC,
	"let":    LET,
	"if":     IF,
	"else":   ELSE,
	"true":   TRUE,
	"false":  FALSE,
	"return": RETURN,
}

func LookupKeywordOrIdent(ident string) TokenType {
	if token, ok := keywords[ident]; ok {
		return token
	}

	return IDENT
}
