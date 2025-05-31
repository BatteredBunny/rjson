package rjson

import (
	"errors"
	"strconv"
	"strings"
	"text/scanner"
	"slices"
)

const Divider = "."
const ArrayOpen = "["
const ArrayClose = "]"
const ArrayLast = "-"

type tokenType = uint

const (
	literalToken tokenType = iota
	arrayIndexToken
	arrayLastToken
	arrayIteratorToken
)

type token struct {
	Type    tokenType
	Content any
}

var ErrFailedToParseTag = errors.New("failed to parse tag")

type unparsedTokens struct {
	Tokens []string
	pos    int
}

func (u *unparsedTokens) Current() (token string) {
	return u.Tokens[u.pos]
}

func (u *unparsedTokens) Next() (token string) {
	u.pos++
	token = u.Tokens[u.pos]

	return
}

func (u *unparsedTokens) Match(match string) bool {
	return slices.Contains(u.Tokens[u.pos:], match)
}

func scanTokens(tag string) (tokens []token, err error) {
	var s scanner.Scanner
	s.Init(strings.NewReader(tag))
	s.Mode ^= scanner.ScanChars | scanner.ScanFloats
	var ut unparsedTokens

	for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
		text := s.TokenText()
		ut.Tokens = append(ut.Tokens, text)
	}

	for ; ut.pos <= len(ut.Tokens)-1; ut.pos++ {
		tok := ut.Current()
		switch tok {
		case ArrayOpen:
			if !ut.Match(ArrayClose) {
				err = ErrFailedToParseTag
				return
			}

			v := ut.Next()
			switch v {
			case ArrayClose:
				tokens = append(tokens, token{Type: arrayIteratorToken})
			case ArrayLast:
				tokens = append(tokens, token{Type: arrayLastToken})
				ut.Next()
			default:
				var i int
				i, err = strconv.Atoi(v)
				if err != nil {
					return
				}

				tokens = append(tokens, token{Type: arrayIndexToken, Content: i})
				ut.Next()
			}
		case Divider:
		case ArrayLast:
			err = ErrFailedToParseTag
		case ArrayClose:
			err = ErrFailedToParseTag
		default:
			tokens = append(tokens, token{Type: literalToken, Content: tok})
		}
	}

	return
}
