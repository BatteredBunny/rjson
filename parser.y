%{
//go:generate goyacc -o parser.go parser.y
package rjson

import (
    "fmt"
)

type query struct {
    Tokens []token
}

type token struct {
    Type    tokenType
    Content interface{}
}

type tokenType int

const (
    literalToken tokenType = iota
    arrayIndexToken
    arrayLastToken
    arrayIteratorToken
)

var parseResult query
%}

%union {
    str string
    num int
    query query
    token token
    tokens []token
}

%token <str> IDENTIFIER
%token <num> NUMBER
%token DOT LBRACKET RBRACKET MINUS

%type <query> query
%type <tokens> path_elements
%type <token> path_element array_access

%start query

%%

query:
    path_elements {
        parseResult = query{Tokens: $1}
        $$ = parseResult
    }

path_elements:
    path_element {
        $$ = []token{$1}
    }
|   path_elements DOT path_element {
        $$ = append($1, $3)
    }
|   path_elements array_access {
        $$ = append($1, $2)
    }

path_element:
    IDENTIFIER {
        $$ = token{Type: literalToken, Content: $1}
    }

array_access:
    LBRACKET RBRACKET {
        $$ = token{Type: arrayIteratorToken}
    }
|   LBRACKET NUMBER RBRACKET {
        $$ = token{Type: arrayIndexToken, Content: $2}
    }
|   LBRACKET MINUS RBRACKET {
        $$ = token{Type: arrayLastToken}
    }

%%

func parse(input string) (query, error) {
    parseResult = query{}
    lexer := newLexer(input)
    result := yyParse(lexer)
    if result != 0 {
        return query{}, fmt.Errorf("parse error")
    }
    return parseResult, nil
}
