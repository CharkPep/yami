package lexer

import (
	"bufio"
	"fmt"
	"io"
)

type TokenReader interface {
	Read(token *Token) error
}

type Lexer struct {
	r      *bufio.Reader
	column int
	line   int
}

func New(r io.Reader) *Lexer {
	l := &Lexer{
		r: bufio.NewReader(r),
	}

	return l
}

func (l *Lexer) Read(t *Token) error {
	l.skipWhitespace()
	cur, err := l.r.ReadByte()
	if err == io.EOF {
		l.assignToken(t, EOF, l.line, l.column, "")
		return nil
	}

	if err != nil {
		return err
	}

	l.column += 1
	switch cur {
	case '*':
		l.assignToken(t, ASTERISK, l.line, l.column, "*")
		return nil
	case '=':
		ok := l.peekAndAssert(byte('='))
		if !ok {
			l.assignToken(t, ASSIGN, l.line, l.column, "=")
			return nil
		}

		l.r.ReadByte()
		l.column++
		l.assignToken(t, EQ, l.line, l.column, "==")
		return nil
	case ';':
		l.assignToken(t, SCOLON, l.line, l.column, ";")
		return nil
	case '+':
		l.assignToken(t, PLUS, l.line, l.column, "+")
		return nil
	case '>':
		switch {
		case l.peekAndAssert(byte('=')):
			l.r.ReadByte()
			l.column++
			l.assignToken(t, GTE, l.line, l.column, ">=")
		case l.peekAndAssert(byte('>')):
			l.r.ReadByte()
			l.column++
			l.assignToken(t, BRSHIFT, l.line, l.column, ">>")
		default:
			l.assignToken(t, GT, l.line, l.column, ">")
		}

		return nil
	case '<':
		switch {
		case l.peekAndAssert(byte('=')):
			l.r.ReadByte()
			l.column++
			l.assignToken(t, LTE, l.line, l.column, "<=")
		case l.peekAndAssert(byte('<')):
			l.r.ReadByte()
			l.column++
			l.assignToken(t, BLSHIFT, l.line, l.column, "<<")
		default:
			l.assignToken(t, LT, l.line, l.column, "<")
		}

		return nil
	case '-':
		l.assignToken(t, HYPHEN, l.line, l.column, "-")
		return nil
	case '!':
		ok := l.peekAndAssert(byte('='))
		if !ok {
			l.assignToken(t, BANG, l.line, l.column, "!")
			return nil
		}

		l.r.ReadByte()
		l.column++
		l.assignToken(t, NEQ, l.line, l.column, "!=")
		return nil
	case '&':
		ok := l.peekAndAssert(byte('&'))
		if !ok {
			l.assignToken(t, BAND, l.line, l.column, "&")
			return nil
		}

		l.r.ReadByte()
		l.column++
		l.assignToken(t, AND, l.line, l.column, "&&")
		return nil
	case '|':
		ok := l.peekAndAssert(byte('|'))
		if !ok {
			l.assignToken(t, BOR, l.line, l.column, "|")
			return nil
		}
		l.r.ReadByte()
		l.column++
		l.assignToken(t, OR, l.line, l.column, "||")
		return nil
	case '/':
		ok := l.peekAndAssert(byte('/'))
		if !ok {
			l.assignToken(t, SLASH, l.line, l.column, "/")
			return nil
		}

		if _, _, err = l.r.ReadLine(); err != nil && err != io.EOF {
			return err
		}

		l.line++
		l.column = 0
		return l.Read(t)
	case '"':
		str, err := l.r.ReadSlice('"')
		if err == io.EOF {
			return fmt.Errorf("given string is invalid")
		}

		if err != nil {
			return err
		}

		l.column += len(str)
		l.assignToken(t, STRING, l.line, l.column, string(str[:len(str)-1]))
		return nil
	case '{':
		l.assignToken(t, BRLEFT, l.line, l.column, `{`)
		return nil
	case '}':
		l.assignToken(t, BRRIGHT, l.line, l.column, `}`)
		return nil
	case '(':
		l.assignToken(t, BLEFT, l.line, l.column, `(`)
		return nil
	case ')':
		l.assignToken(t, BRIGHT, l.line, l.column, `)`)
		return nil
	case '[':
		l.assignToken(t, SBLEFT, l.line, l.column, "[")
		return nil
	case ']':
		l.assignToken(t, SBRIGHT, l.line, l.column, "]")
		return nil
	case ',':
		l.assignToken(t, COMA, l.line, l.column, ",")
		return nil
	case ':':
		l.assignToken(t, COLON, l.line, l.column, ":")
		return nil
	default:
		var literal []byte
		if IsDigit(cur) {
			literal = l.readNumber()
			l.column += len(literal)
			l.assignToken(t, NUMBER, l.line, l.column, string(cur)+string(literal))
			return nil
		}

		literal = l.readIndent()
		l.column += len(literal)
		tokenLiteral := string(cur) + string(literal)
		tokenType := LookupKeywordOrIdent(tokenLiteral)
		l.assignToken(t, tokenType, l.line, l.column, tokenLiteral)
		return nil
	}

}
func (l *Lexer) assignToken(t *Token, tType TokenType, line, column int, literal string) {
	t.Token = tType
	t.Line = line
	t.Column = column
	t.Literal = literal
}

func (l *Lexer) peekAndAssert(assert byte) bool {
	peek, err := l.r.Peek(1)
	if err != nil {
		return false
	}

	if peek[0] == assert {
		return true
	}

	return false
}

func (l *Lexer) readIndent() []byte {
	var buf []byte
	for peek, err := l.r.Peek(1); err == nil && IsIdentCh(peek[0]); peek, err = l.r.Peek(1) {
		buf = append(buf, peek[0])
		l.r.ReadByte()
	}

	return buf
}

func (l *Lexer) readNumber() []byte {
	var buf []byte
	for digit, err := l.r.Peek(1); err == nil && IsDigit(digit[0]); digit, err = l.r.Peek(1) {
		buf = append(buf, digit[0])
		l.r.ReadByte()
	}

	return buf
}

func (l *Lexer) skipWhitespace() {
	for peek, err := l.r.Peek(1); err == nil && IsWhitespace(peek[0]); peek, err = l.r.Peek(1) {
		l.column++
		if ch, _ := l.r.ReadByte(); ch == '\n' {
			l.line++
			l.column = 0
		}
	}

	return
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

func IsIdentCh(ch byte) bool {
	if (ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z') || ch == '_' {
		return true
	}

	return false
}
