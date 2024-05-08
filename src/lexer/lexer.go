package lexer

import (
	"bufio"
	"io"
)

// TODO: rewrite lexer to be more like bufio.Reader with Peek and Read (Advance) methods only

type ILexer interface {
	Advance() (Token, error)
	CurToken() Token
	PeekToken() Token
}

type BufferedLexer struct {
	r      *bufio.Reader
	peek   *Token
	cur    *Token
	ch     byte
	column int
	line   int
}

func New(r io.Reader) *BufferedLexer {
	l := &BufferedLexer{
		r: bufio.NewReader(r),
	}

	return l
}

func (l *BufferedLexer) advanceToken() error {
	l.cur = l.peek
	token, err := l.nextToken()
	l.peek = &token
	return err
}

func (l *BufferedLexer) CurToken() Token {
	if l.cur == nil {
		return NewToken(NIL, l.line, l.column, "")
	}

	return *l.cur
}

func (l *BufferedLexer) PeekToken() Token {
	if l.peek == nil {
		return NewToken(EOF, l.line, l.column, "")
	}

	return *l.peek
}

func (l *BufferedLexer) Advance() (Token, error) {
	if err := l.advanceToken(); err != nil {
		return NewToken(ILLEGAL, l.line, l.column, ""), err
	}

	for l.cur == nil {
		if err := l.advanceToken(); err != nil {
			return NewToken(ILLEGAL, l.line, l.column, ""), err
		}
	}

	return *l.cur, nil
}

func (l *BufferedLexer) nextToken() (Token, error) {
	l.skipWhitespace()
	peek, err := l.r.ReadByte()
	if err == io.EOF {
		return NewToken(EOF, l.line, l.column, "EOF"), nil
	}

	l.ch = peek
	l.column += 1
	if err != nil {
		return NewToken(ILLEGAL, l.line, l.column, ""), err
	}

	switch peek {
	case '*':
		return NewToken(ASTERISK, l.line, l.column, "*"), nil
	case '=':
		ok, err := l.peekAndAssert(byte('='))
		if err != nil || !ok {
			return NewToken(ASSIGN, l.line, l.column, "="), nil
		}

		return NewToken(EQ, l.line, l.column, "=="), nil
	case ';':
		return NewToken(SCOLUMN, l.line, l.column, ";"), nil
	case '+':
		return NewToken(PLUS, l.line, l.column, "+"), nil
	case '>':
		ok, err := l.peekAndAssert(byte('='))
		if err != nil || !ok {
			return NewToken(GT, l.line, l.column, ">"), nil
		}

		return NewToken(GTE, l.line, l.column, ">="), nil
	case '<':
		ok, err := l.peekAndAssert(byte('='))
		if err != nil || !ok {
			return NewToken(LT, l.line, l.column, "<"), nil
		}

		return NewToken(LTE, l.line, l.column, "<="), nil
	case '-':
		return NewToken(HYPHEN, l.line, l.column, "-"), nil
	case '!':
		ok, err := l.peekAndAssert(byte('='))
		if err != nil || !ok {
			return NewToken(BANG, l.line, l.column, "!"), nil
		}

		return NewToken(NEQ, l.line, l.column, "!="), nil
	case '&':
		ok, err := l.peekAndAssert(byte('&'))
		if err != nil || !ok {
			return NewToken(ILLEGAL, l.line, l.column, "&"), nil
		}

		return NewToken(DAMPERSAND, l.line, l.column, "&&"), nil
	case '|':
		ok, err := l.peekAndAssert(byte('|'))
		if err != nil || !ok {
			return NewToken(ILLEGAL, l.line, l.column, "|"), nil
		}

		return NewToken(DVERTLINE, l.line, l.column, "||"), nil
	case '/':
		ok, err := l.peekAndAssert(byte('/'))
		if err != nil || !ok {
			return NewToken(SLASH, l.line, l.column, "/"), nil
		}

		if _, _, err := l.r.ReadLine(); err != nil {
			return NewToken(ILLEGAL, l.line, l.column, ""), err
		}

		l.line++
		l.column = 0

		return l.nextToken()
	case '"':
		return NewToken(DQUOTE, l.line, l.column, `"`), nil
	case '{':
		return NewToken(BRLEFT, l.line, l.column, `{`), nil
	case '}':
		return NewToken(BRRIGHT, l.line, l.column, `}`), nil
	case '(':
		return NewToken(BLEFT, l.line, l.column, `(`), nil
	case ')':
		return NewToken(BRIGHT, l.line, l.column, `)`), nil
	case '[':
		return NewToken(SBLEFT, l.line, l.column, "["), nil
	case ']':
		return NewToken(SBRIGHT, l.line, l.column, "]"), nil
	case ',':
		return NewToken(COMA, l.line, l.column, ","), nil
	default:
		var literal []byte
		if IsDigit(peek) {
			literal, err = l.readNumber()
			l.column += len(literal)
			return NewToken(NUMBER, l.line, l.column, string(peek)+string(literal)), nil
		}

		literal, err = l.readIndent()
		if err != nil && err != io.EOF {
			return NewToken(ILLEGAL, l.line, l.column, ""), err
		}

		l.column += len(literal)
		tokenLiteral := string(peek) + string(literal)
		tokenType := LookupKeywordOrIdent(tokenLiteral)
		return NewToken(tokenType, l.line, l.column, tokenLiteral), nil
	}

}

func (l *BufferedLexer) peekAndAssert(assert byte) (bool, error) {
	peek, err := l.r.Peek(1)
	if err != nil {
		return false, err
	}

	if peek[0] == assert {
		l.r.ReadByte()
		l.column++
		return true, err
	}

	return false, err
}

func (l *BufferedLexer) readIndent() (buf []byte, err error) {
	var (
		peek []byte
	)

	for peek, err = l.r.Peek(1); err == nil && IsAllowedInIdent(peek[0]); peek, err = l.r.Peek(1) {
		buf = append(buf, peek[0])
		l.r.ReadByte()
	}

	return
}

func (l *BufferedLexer) readNumber() (buf []byte, err error) {
	for digit, err := l.r.Peek(1); err == nil && IsDigit(digit[0]); digit, err = l.r.Peek(1) {
		buf = append(buf, digit[0])
		l.r.ReadByte()
	}

	return
}

func (l *BufferedLexer) skipWhitespace() error {
	var (
		peek []byte
		err  error
	)

	for peek, err = l.r.Peek(1); err == nil && IsWhitespace(peek[0]); peek, err = l.r.Peek(1) {
		l.column++
		if ch, _ := l.r.ReadByte(); ch == '\n' {
			l.line++
			l.column = 0
		}
	}

	return err
}

func IsLetter(ch byte) bool {
	return ch >= 'A' && ch <= 'Z' || ch >= 'a' && ch <= 'z'
}

func IsDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func IsWhitespace(ch byte) bool {
	if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
		return true
	}

	return false
}

func IsAllowedInIdent(ch byte) bool {
	if (ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z') || ch == '_' {
		return true
	}

	return false
}
