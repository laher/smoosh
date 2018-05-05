/*
Package lexer handles the interpreter's first phase of operation. Lexer takes some input and interpret it into 'tokens'.

Smoosh (and Monkey) builds up its input into an ordered list of tokens without attmpting to make sense of the structure and relationship.
These tokens are operators, keywords, identifiers or values (strings, numbers, etc).
*/
package lexer

import (
	"unicode/utf8"

	"github.com/laher/smoosh/token"
)

// Lexer tokenises input
type Lexer struct {
	input        string
	position     int  // current position in input (points to current rune)
	readPosition int  // current reading position in input (after current rune)
	ru           rune // current rune under examination
	line         int  //current line number
}

// New creates and initialises a new lexer
func New(input string) *Lexer {
	l := &Lexer{input: input, line: 1}
	l.readRune()
	return l
}

// NextToken attempts to find the next token in the program's input
func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespace()

	switch l.ru {
	case '=':
		if l.peekChar() == '=' {
			ch := l.ru
			l.readRune()
			literal := string(ch) + string(l.ru)
			tok = token.Token{Type: token.EQ, Literal: literal, Line: l.line}
		} else {
			tok = newToken(token.ASSIGN, l.ru, l.line)
		}
	case '+':
		tok = newToken(token.PLUS, l.ru, l.line)
	case '-':
		tok = newToken(token.MINUS, l.ru, l.line)
	case '!':
		if l.peekChar() == '=' {
			ch := l.ru
			l.readRune()
			literal := string(ch) + string(l.ru)
			tok = token.Token{Type: token.NOT_EQ, Literal: literal, Line: l.line}
		} else {
			tok = newToken(token.BANG, l.ru, l.line)
		}
	case '/':
		tok = newToken(token.SLASH, l.ru, l.line)
	case '*':
		tok = newToken(token.ASTERISK, l.ru, l.line)
	case '<':
		tok = newToken(token.LT, l.ru, l.line)
	case '>':
		tok = newToken(token.GT, l.ru, l.line)
	case ';':
		tok = newToken(token.SEMICOLON, l.ru, l.line)
	case ':':
		tok = newToken(token.COLON, l.ru, l.line)
	case ',':
		tok = newToken(token.COMMA, l.ru, l.line)
	case '{':
		tok = newToken(token.LBRACE, l.ru, l.line)
	case '}':
		tok = newToken(token.RBRACE, l.ru, l.line)
	case '(':
		tok = newToken(token.LPAREN, l.ru, l.line)
	case ')':
		tok = newToken(token.RPAREN, l.ru, l.line)
	case '"':
		tok.Type = token.STRING
		tok.Literal = l.readString()
	case '[':
		tok = newToken(token.LBRACKET, l.ru, l.line)
	case ']':
		tok = newToken(token.RBRACKET, l.ru, l.line)
	case '|':
		tok = newToken(token.PIPE, l.ru, l.line)
	case '#':
		tok = newToken(token.HASH, l.ru, l.line)
		tok.Literal = l.readLine()
	case 0, utf8.RuneError:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if isLetter(l.ru) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			return tok
		}
		if isDigit(l.ru) {
			tok.Type = token.INT
			tok.Literal = l.readNumber()
			return tok
		}
		tok = newToken(token.ILLEGAL, l.ru, l.line)
	}

	l.readRune()
	return tok
}

func (l *Lexer) skipWhitespace() {
	for l.ru == ' ' || l.ru == '\t' || l.ru == '\n' || l.ru == '\r' {
		if l.ru == '\n' {
			l.line++
		}
		l.readRune()
	}
}

func (l *Lexer) readRune() {
	var ln int
	l.ru, ln = utf8.DecodeRuneInString(l.input[l.readPosition:])
	l.position = l.readPosition
	l.readPosition += ln
}

func (l *Lexer) peekChar() rune {
	r, _ := utf8.DecodeRuneInString(l.input[l.readPosition:])
	return r
}

func (l *Lexer) readIdentifier() string {
	return l.read(isLetter)
}

func (l *Lexer) readNumber() string {
	return l.read(isDigit)
}

func (l *Lexer) readString() string {
	position := l.position + 1
	for {
		l.readRune()
		if l.ru == '"' || l.ru == utf8.RuneError || l.ru == 0 {
			break
		}
	}
	return l.input[position:l.position]
}

func (l *Lexer) readLine() string {
	position := l.position + 1
	for {
		l.readRune()
		if l.ru == '\n' || l.ru == utf8.RuneError || l.ru == 0 { //TODO is this cross-platform?
			break
		}
	}
	return l.input[position:l.position]
}

func (l *Lexer) read(checkFn func(rune) bool) string {
	position := l.position
	for checkFn(l.ru) {
		l.readRune()
		if l.ru == utf8.RuneError || l.ru == 0 {
			break
		}
	}
	return l.input[position:l.position]
}

func isLetter(ch rune) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_' || ch == '$' || ch == '.'
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9'
}

func newToken(tokenType token.TokenType, ch rune, line int) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch), Line: line}
}
