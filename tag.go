package main

import (
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
	"strings"
	"text/scanner"
)

var ErrNotAPointer = errors.New("Please insert a pointer")
var ErrMalformedSyntax = errors.New("Malformed tag syntax")

const tagname = "rjson"
const divider = "."
const array_open = "["
const array_close = "]"

type SymbolType = int

const (
	SymbolTypeName SymbolType = iota
	SymbolTypeArrayAccess
	SymbolTypeArray
)

type Symbol struct {
	Type    SymbolType
	Content any
}

func valueFinder(input []byte, tag string) (object json.RawMessage, err error) {
	var s scanner.Scanner
	s.Init(strings.NewReader(tag))

	var symbols []Symbol
	var array_started bool
	var array_has_number bool
	for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
		if s.TokenText() == array_open {
			if array_started {
				err = ErrMalformedSyntax
				return
			}

			array_started = true
		} else if s.TokenText() == array_close {
			if !array_started {
				err = ErrMalformedSyntax
				return
			}

			if !array_has_number {
				symbols = append(symbols, Symbol{Type: SymbolTypeArray})
			}

			array_started = false
			array_has_number = false
		} else if s.TokenText() == divider {
			continue
		} else {
			if array_started {
				var i int
				i, err = strconv.Atoi(s.TokenText())
				if err != nil {
					return
				}

				symbols = append(symbols, Symbol{Type: SymbolTypeArrayAccess, Content: i})
				array_has_number = true
			} else {
				symbols = append(symbols, Symbol{Type: SymbolTypeName, Content: s.TokenText()})
			}
		}
	}

	if err = json.Unmarshal(input, &object); err != nil {
		return
	}

	var last_symbol_was_array bool
	for _, symbol := range symbols {
		switch symbol.Type {
		case SymbolTypeName:
			if last_symbol_was_array {
				var input []json.RawMessage
				if err = json.Unmarshal(object, &input); err != nil {
					return
				}

				var result []json.RawMessage
				for _, row := range input {
					var obj map[string]json.RawMessage
					if err = json.Unmarshal(row, &obj); err != nil {
						return
					}

					var v json.RawMessage
					if err = json.Unmarshal(obj[symbol.Content.(string)], &v); err != nil {
						return
					}

					result = append(result, v)
				}

				var bs []byte
				bs, err = json.Marshal(result)
				if err != nil {
					return
				}

				if err = json.Unmarshal(bs, &object); err != nil {
					return
				}
			} else {
				var obj map[string]json.RawMessage

				if err = json.Unmarshal(object, &obj); err != nil {
					return
				}

				if err = json.Unmarshal(obj[symbol.Content.(string)], &object); err != nil {
					return
				}
			}
		case SymbolTypeArrayAccess:
			var obj []json.RawMessage
			if err = json.Unmarshal(object, &obj); err != nil {
				return
			}

			object = obj[symbol.Content.(int)]
		case SymbolTypeArray:
			last_symbol_was_array = true
		}

		if err != nil {
			return
		}
	}

	return
}

// Unmarshal parses the JSON-encoded data and stores the result in the value pointed to by v. If v is nil or not a pointer, Unmarshal returns an ErrNotAPointer.
func Unmarshal(data []byte, v any) (err error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return ErrNotAPointer
	}

	t := rv.Elem().Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get(tagname)

		if tag != "" && field.IsExported() {
			valuefield := rv.Elem().Field(i)
			test := reflect.New(valuefield.Type())
			inter := test.Interface()

			var res json.RawMessage
			res, err = valueFinder(data, tag)
			if err != nil {
				return
			}

			if err = json.Unmarshal(res, &inter); err != nil {
				return
			}

			if valuefield.CanSet() {
				valuefield.Set(reflect.Indirect(test))
			}
		}
	}

	return
}
