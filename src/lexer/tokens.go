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
	STRING = "STRING"
	IF     = "IF"
	ELSE   = "ELSE"
	TRUE   = "TRUE"
	FALSE  = "FALSE"
	RETURN = "RETURN"

	PLUS   = "PLUS"
	HYPHEN = "HYPHEN"
	SLASH  = "SLASH"

	SCOLON = "SCOLON" // Semi colon
	COLON  = "COLON"
	ASSIGN = "ASSIGN"

	// Boolean

	LT   = "LT"
	GT   = "GT"
	LTE  = "LTE"
	GTE  = "GTE"
	EQ   = "EQ"
	NEQ  = "NEQ"
	BANG = "BANG"
	AND  = "AND" // [D]ouble ampersand
	OR   = "OR"  // [D]ouble vertical line

	// Bitwise operation

	BAND = "BAND"
	BOR  = "BOR"
	// TODO: implement the rest of bitwise operators
	BXOR    = "BXOR"
	BLSHIFT = "BLSHIFT"
	BRSHIFT = "BRSHIFT"

	// Binary

	BLEFT    = "BLEFT"   // [B]rackets left
	BRIGHT   = "BRIGHT"  // [B]rackets right
	SBLEFT   = "SBLEFT"  // [S]quare [B]rackets left
	SBRIGHT  = "SBRIGHT" // [S]quare [B]rackets right
	BRLEFT   = "BRLEFT"  // Curly [Br]aces left
	BRRIGHT  = "BRRIGHT" // Curly [Br]aces left
	ASTERISK = "ASTERISK"

	COMA    = "COMA"
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"
	NIL     = "NIL"
)

var keywords = map[string]TokenType{
	"fn":     FUNC,
	"let":    LET,
	"if":     IF,
	"else":   ELSE,
	"true":   TRUE,
	"false":  FALSE,
	"return": RETURN,
	"<<":     BLEFT,
	">>":     BRIGHT,
}

func LookupKeywordOrIdent(ident string) TokenType {
	if token, ok := keywords[ident]; ok {
		return token
	}

	return IDENT
}
