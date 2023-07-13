package rjson

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"text/scanner"

	"github.com/goccy/go-json"
)

var ErrNotAPointer = errors.New("please insert a pointer")
var ErrMalformedSyntax = errors.New("malformed tag syntax")
var ErrCantFindField = errors.New("cant find field")

const TagName = "rjson"
const Divider = "."
const ArrayOpen = "["
const ArrayClose = "]"
const ArrayLast = "-"

type symbolType = int

const (
	symbolTypeName symbolType = iota
	symbolTypeArrayAccess
	symbolTypeArrayLast
	symbolTypeArray
)

type symbol struct {
	Type    symbolType
	Content any
}

// The underlying function powering the tag, accepts json as bytes
func QueryJson(data []byte, tag string) (object json.RawMessage, err error) {
	var s scanner.Scanner
	s.Init(strings.NewReader(tag))

	var symbols []symbol
	var arrayStarted bool
	var arrayFilled bool
	for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
		token := s.TokenText()
		switch token {
		case ArrayOpen:
			if arrayStarted {
				err = ErrMalformedSyntax
				return
			}

			arrayStarted = true
		case ArrayClose:
			if !arrayStarted {
				err = ErrMalformedSyntax
				return
			} else if !arrayFilled {
				symbols = append(symbols, symbol{Type: symbolTypeArray})
			}

			arrayStarted = false
			arrayFilled = false
		case Divider:
			continue
		case ArrayLast:
			if arrayStarted && !arrayFilled {
				symbols = append(symbols, symbol{Type: symbolTypeArrayLast})
				arrayFilled = true
			}
		default:
			if arrayStarted && !arrayFilled {
				var i int
				i, err = strconv.Atoi(token)
				if err != nil {
					return
				}

				symbols = append(symbols, symbol{Type: symbolTypeArrayAccess, Content: i})
				arrayFilled = true
			} else {
				symbols = append(symbols, symbol{Type: symbolTypeName, Content: token})
			}
		}

	}

	if err = json.Unmarshal(data, &object); err != nil {
		return
	}

	var lastSymbolWasArray bool
	for _, sym := range symbols {
		switch sym.Type {
		case symbolTypeName:
			if lastSymbolWasArray {
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

					if err = json.Unmarshal(obj[sym.Content.(string)], &v); err != nil {
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

				if err = json.Unmarshal(obj[sym.Content.(string)], &object); err != nil {
					err = fmt.Errorf("%w %s: %s", ErrCantFindField, sym.Content, err)
					return
				}
			}
		case symbolTypeArrayAccess:
			var obj []json.RawMessage
			if err = json.Unmarshal(object, &obj); err != nil {
				return
			}

			object = obj[sym.Content.(int)]
		case symbolTypeArrayLast:
			var obj []json.RawMessage
			if err = json.Unmarshal(object, &obj); err != nil {
				return
			}

			object = obj[len(obj)-1]
		case symbolTypeArray:
			lastSymbolWasArray = true
		}
	}

	return
}

func handleStructFields(data []byte, tag string, rv reflect.Value, notNested bool) (err error) {
	t := reflect.Indirect(rv).Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		currentTag := field.Tag.Get(TagName)
		valueField := reflect.Indirect(rv).Field(i)

		if Debug {
			fmt.Printf("Handling field %s with %s tag name\n", field.Name, currentTag)
		}

		var ct string
		if currentTag != "" && t.Field(i).IsExported() {
			if field.Type.Kind() == reflect.Struct {
				if notNested {
					ct = fmt.Sprintf("%s.%s", tag, currentTag)
				} else {
					ct = fmt.Sprintf("%s[%d]%s", tag, i, currentTag)
				}

				if err = handleStructFields(data, ct, valueField, false); errors.Unwrap(err) == ErrCantFindField {
					continue
				} else if err != nil {
					return
				}
			} else if field.Type.Kind() == reflect.Slice && field.Type.Elem().Kind() == reflect.Struct {
				if notNested {
					ct = fmt.Sprintf("%s.%s", tag, currentTag)
				} else {
					ct = fmt.Sprintf("%s[%d]%s", tag, i, currentTag)
				}

				if err = handleStructSlices(data, ct, valueField); errors.Unwrap(err) == ErrCantFindField {
					continue
				} else if err != nil {
					return
				}
			} else {
				if notNested {
					ct = currentTag
				} else {
					ct = tag + "." + currentTag
				}

				if err = handleFields(data, ct, valueField); errors.Unwrap(err) == ErrCantFindField {
					continue
				} else if err != nil {
					return
				}
			}
		}
	}

	return
}

func handleStructSlices(data []byte, tag string, rv reflect.Value) (err error) {
	var res json.RawMessage
	res, err = QueryJson(data, tag)
	if err != nil {
		return
	}

	var arr []json.RawMessage
	if err = json.Unmarshal(res, &arr); err != nil {
		return
	}

	for j := 0; j < len(arr); j++ {
		rv.Set(reflect.Append(rv, reflect.Indirect(reflect.New(rv.Type().Elem()))))
		sv := rv.Index(j)

		var bs []byte
		bs, err = arr[j].MarshalJSON()
		if err != nil {
			return
		}

		if err = handleStructFields(bs, "", sv, true); errors.Unwrap(err) == ErrCantFindField {
			continue
		} else if err != nil {
			return
		}
	}

	return
}

func handleFields(data []byte, tag string, rv reflect.Value) (err error) {
	emptyValue := reflect.New(rv.Type())
	inter := emptyValue.Interface()

	var res json.RawMessage
	res, err = QueryJson(data, tag)
	if err != nil {
		return
	}

	if err = json.Unmarshal(res, &inter); err != nil {
		return
	}

	if rv.CanSet() {
		rv.Set(reflect.Indirect(emptyValue))
	}

	return
}

// Unmarshal parses the JSON-encoded data and stores the result in the value pointed to by v. If v is nil or not a pointer, Unmarshal returns an ErrNotAPointer.
func Unmarshal(data []byte, v any) (err error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return ErrNotAPointer
	}

	if err = handleStructFields(data, "", rv, true); err != nil {
		return
	}

	return
}
