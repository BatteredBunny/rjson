package rjson

import (
	"fmt"
	"strconv"
	"unicode"
	"unicode/utf8"
)

type lexer struct {
	input string
	pos   int
	start int
	width int
}

const Divider = '.' // Path divider
const ArrayOpen = '['
const ArrayClose = ']'
const ArrayLast = '-' // Takes last element in array

func newLexer(input string) *lexer {
	return &lexer{
		input: input,
	}
}

func (l *lexer) next() rune {
	if l.pos >= len(l.input) {
		l.width = 0
		return 0 // EOF
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = w
	l.pos += l.width
	return r
}

func (l *lexer) backup() {
	l.pos -= l.width
}

func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

func (l *lexer) ignore() {
	l.start = l.pos
}

func (l *lexer) emit(t int, lval *yySymType) int {
	switch t {
	case IDENTIFIER:
		lval.str = l.input[l.start:l.pos]
	case NUMBER:
		val, _ := strconv.Atoi(l.input[l.start:l.pos])
		lval.num = val
	}
	l.start = l.pos
	return t
}

func (l *lexer) Lex(lval *yySymType) int {
	for {
		r := l.next()
		if r == 0 {
			return 0 // EOF
		}

		switch r {
		case ' ', '\t', '\n', '\r':
			l.ignore()
		case Divider:
			l.ignore()
			return DOT
		case ArrayOpen:
			l.ignore()
			return LBRACKET
		case ArrayClose:
			l.ignore()
			return RBRACKET
		case ArrayLast:
			l.ignore()
			return MINUS
		default:
			if unicode.IsLetter(r) {
				return l.lexIdentifier(lval)
			} else if unicode.IsDigit(r) {
				return l.lexNumber(lval)
			}
			return int(r) // Return the rune as is for unknown characters
		}
	}
}

func (l *lexer) lexIdentifier(lval *yySymType) int {
	for {
		r := l.peek()
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			break
		}
		l.next()
	}
	return l.emit(IDENTIFIER, lval)
}

func (l *lexer) lexNumber(lval *yySymType) int {
	for {
		r := l.peek()
		if !unicode.IsDigit(r) {
			break
		}
		l.next()
	}
	return l.emit(NUMBER, lval)
}

func (l *lexer) Error(s string) {
	if Debug {
		fmt.Printf("Parse error: %s\n", s)
	}
}
