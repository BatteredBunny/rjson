package rjson

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"text/scanner"
)

var ErrNotAPointer = errors.New("Please insert a pointer")
var ErrMalformedSyntax = errors.New("Malformed tag syntax")

const TagName = "rjson"
const Divider = "."
const ArrayOpen = "["
const ArrayClose = "]"

type symbolType = int

const (
	symbolTypeName symbolType = iota
	symbolTypeArrayAccess
	symbolTypeArray
)

type symbol struct {
	Type    symbolType
	Content any
}

func valueFinder(input []byte, tag string) (object json.RawMessage, err error) {
	var s scanner.Scanner
	s.Init(strings.NewReader(tag))

	var symbols []symbol
	var array_started bool
	var array_has_number bool
	for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
		if s.TokenText() == ArrayOpen {
			if array_started {
				err = ErrMalformedSyntax
				return
			}

			array_started = true
		} else if s.TokenText() == ArrayClose {
			if !array_started {
				err = ErrMalformedSyntax
				return
			}

			if !array_has_number {
				symbols = append(symbols, symbol{Type: symbolTypeArray})
			}

			array_started = false
			array_has_number = false
		} else if s.TokenText() == Divider {
			continue
		} else {
			if array_started {
				var i int
				i, err = strconv.Atoi(s.TokenText())
				if err != nil {
					return
				}

				symbols = append(symbols, symbol{Type: symbolTypeArrayAccess, Content: i})
				array_has_number = true
			} else {
				symbols = append(symbols, symbol{Type: symbolTypeName, Content: s.TokenText()})
			}
		}
	}

	if err = json.Unmarshal(input, &object); err != nil {
		return
	}

	var last_symbol_was_array bool
	for _, s := range symbols {
		switch s.Type {
		case symbolTypeName:
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
					if err = json.Unmarshal(obj[s.Content.(string)], &v); err != nil {
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

				if err = json.Unmarshal(obj[s.Content.(string)], &object); err != nil {
					return
				}
			}
		case symbolTypeArrayAccess:
			var obj []json.RawMessage
			if err = json.Unmarshal(object, &obj); err != nil {
				return
			}

			object = obj[s.Content.(int)]
		case symbolTypeArray:
			last_symbol_was_array = true
		}

		if err != nil {
			return
		}
	}

	return
}

func handleStructFields(data []byte, tag string, rv reflect.Value, not_nested bool) (err error) {
	t := reflect.Indirect(rv).Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		currenttag := field.Tag.Get(TagName)
		valuefield := reflect.Indirect(rv).Field(i)

		var ct string
		if currenttag != "" && t.Field(i).IsExported() {
			if field.Type.Kind() == reflect.Struct {
				if not_nested {
					ct = fmt.Sprintf("%s.%s", tag, currenttag)
				} else {
					ct = fmt.Sprintf("%s[%d]%s", tag, i, currenttag)
				}

				if err = handleStructFields(data, ct, valuefield, false); err != nil {
					return
				}
			} else if field.Type.Kind() == reflect.Slice && field.Type.Elem().Kind() == reflect.Struct {
				if not_nested {
					ct = fmt.Sprintf("%s.%s", tag, currenttag)
				} else {
					ct = fmt.Sprintf("%s[%d]%s", tag, i, currenttag)
				}

				var res json.RawMessage
				res, err = valueFinder(data, ct)
				if err != nil {
					return
				}

				// var bs []byte
				// bs, err = res.MarshalJSON()
				// if err != nil {
				// 	return
				// }

				var arr []json.RawMessage
				if err = json.Unmarshal(res, &arr); err != nil {
					return
				}

				for j := 0; j < len(arr); j++ {
					var bs []byte
					bs, err = arr[j].MarshalJSON()
					if err != nil {
						return
					}

					valuefield.Set(reflect.Append(valuefield, reflect.Indirect(reflect.New(valuefield.Type().Elem()))))
					sv := valuefield.Index(j)

					if err = handleStructFields(bs, "", sv, true); err != nil {
						return
					}
				}
			} else {
				if not_nested {
					ct = currenttag
				} else {
					ct = tag + "." + currenttag
				}

				if err = handleFields(data, ct, valuefield); err != nil {
					return
				}
			}
		}
	}

	return
}

func handleFields(data []byte, tag string, rv reflect.Value) (err error) {
	empty_value := reflect.New(rv.Type())
	inter := empty_value.Interface()

	var res json.RawMessage
	res, err = valueFinder(data, tag)
	if err != nil {
		return
	}

	if err = json.Unmarshal(res, &inter); err != nil {
		return
	}

	if rv.CanSet() {
		rv.Set(reflect.Indirect(empty_value))
	}

	return
}

// Unmarshal parses the JSON-encoded data and stores the result in the value pointed to by v. If v is nil or not a pointer, Unmarshal returns an ErrNotAPointer.
func Unmarshal(data []byte, v any) (err error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return ErrNotAPointer
	}

	handleStructFields(data, "", rv, true)

	return
}
