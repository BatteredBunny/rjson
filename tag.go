package rjson

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"reflect"

	"github.com/goccy/go-json"
)

var ErrNotAPointer = errors.New("please insert a pointer")
var ErrMalformedSyntax = errors.New("malformed tag syntax")
var ErrCantFindField = errors.New("cant find field")
var ErrInvalidIndex = errors.New("invalid slice index")
var ErrNotAnObject = errors.New("failed to parse as json object:")

const TagName = "rjson"

func iteratorExecutor(input []json.RawMessage, iteratorLiterals []string, iteratorLevel int, totalIteratorLevel int) (result []json.RawMessage) {
	for _, row := range input {
		if iteratorLevel > 1 {
			var obj []json.RawMessage
			if err := json.Unmarshal(row, &obj); err != nil {
				log.Fatal(err)
			}

			bs, err := json.Marshal(iteratorExecutor(obj, iteratorLiterals[1:], iteratorLevel-1, totalIteratorLevel))
			if err != nil {
				log.Fatal(err)
			}

			var res json.RawMessage
			if err := json.Unmarshal(bs, &res); err != nil {
				log.Fatal(err)
			}

			result = append(result, res)
		} else {
			var obj map[string]json.RawMessage
			if err := json.Unmarshal(row, &obj); err != nil {
				continue
			}

			var v json.RawMessage
			if err := json.Unmarshal(obj[iteratorLiterals[len(iteratorLiterals)-1]], &v); err != nil {
				continue
			}

			result = append(result, v)
		}
	}

	return
}

// QueryJson is the underlying function powering the tag, accepts json as bytes
func QueryJson(data []byte, tag string) (object json.RawMessage, err error) {
	tokens, err := scanTokens(tag)
	if err != nil {
		return
	}

	if err = json.Unmarshal(data, &object); err != nil {
		return
	}

	var totalIteratorLevel int
	for _, tok := range tokens {
		switch tok.Type {
		case arrayIteratorToken:
			totalIteratorLevel++
		}
	}

	var iteratorLevel int
	var iteratorLiterals []string
	for _, tok := range tokens {
		switch tok.Type {
		case literalToken:
			if iteratorLevel > 0 {
				iteratorLiterals = append(iteratorLiterals, tok.Content.(string))

				var input []json.RawMessage
				if err = json.Unmarshal(object, &input); err != nil {
					return
				}

				var bs []byte
				bs, err = json.Marshal(iteratorExecutor(input, iteratorLiterals, iteratorLevel, totalIteratorLevel))
				if err != nil {
					return
				}

				if err = json.Unmarshal(bs, &object); err != nil {
					return
				}
			} else {
				var obj map[string]json.RawMessage
				if err = json.Unmarshal(object, &obj); err != nil {
					err = fmt.Errorf("%w %s: %s", ErrNotAnObject, tok.Content, err)
					return
				}

				if err = json.Unmarshal(obj[tok.Content.(string)], &object); err != nil {
					err = fmt.Errorf("%w %s: %s", ErrCantFindField, tok.Content, err)
					return
				}
			}
		case arrayIndexToken:
			var obj []json.RawMessage
			if err = json.Unmarshal(object, &obj); err != nil {
				return
			}

			i := tok.Content.(int)
			if i >= len(obj) {
				err = fmt.Errorf("%w %d", ErrInvalidIndex, i)
				return
			} else {
				object = obj[i]
			}
		case arrayLastToken:
			var obj []json.RawMessage
			if err = json.Unmarshal(object, &obj); err != nil {
				return
			}

			i := len(obj) - 1
			if i >= len(obj) {
				err = fmt.Errorf("%w %d", ErrInvalidIndex, i)
				return
			} else {
				object = obj[i]
			}
		case arrayIteratorToken:
			iteratorLevel++
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
			fmt.Printf("Handling field %s with tag name: %s, type: %v\n", field.Name, currentTag, field.Type.Kind())
		}

		var ct string
		if currentTag != "" && field.IsExported() {
			if field.Type.Kind() == reflect.Struct {
				if notNested {
					ct = fmt.Sprintf("%s.%s", tag, currentTag)
				} else {
					ct = fmt.Sprintf("%s[%d]%s", tag, i, currentTag)
				}

				if err = handleStructFields(data, ct, valueField, false); errors.Is(err, ErrCantFindField) || errors.Is(err, ErrInvalidIndex) {
					if Debug {
						fmt.Println("WARNING:", err)
					}
					err = nil
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

				if err = handleStructSlices(data, ct, valueField); errors.Is(err, ErrCantFindField) || errors.Is(err, ErrInvalidIndex) {
					if Debug {
						fmt.Println("WARNING:", err)
					}
					err = nil
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

				if err = handleFields(data, ct, valueField); errors.Is(err, ErrCantFindField) || errors.Is(err, ErrInvalidIndex) {
					if Debug {
						fmt.Println("WARNING:", err)
					}
					err = nil
					continue
				} else if err != nil {
					return
				}
			}
		} else if Debug && len(currentTag) > 0 && !field.IsExported() {
			fmt.Println("WARNING: rjson tag on an unexported field")
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

		if err = handleStructFields(bs, "", sv, true); errors.Is(err, ErrCantFindField) || errors.Is(err, ErrInvalidIndex) {
			if Debug {
				fmt.Println("WARNING:", err)
			}
			continue
		} else if err != nil {
			return
		}
	}

	return
}

func recursiveNumCheck(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Slice, reflect.Array:
		return recursiveNumCheck(t.Elem())
	case reflect.Float64, reflect.Float32, reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int8, reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint8, reflect.Uintptr:
		return true
	default:
		return false
	}
}

func handleFields(data []byte, tag string, rv reflect.Value) (err error) {
	emptyValue := reflect.New(rv.Type())
	inter := emptyValue.Interface()

	var res json.RawMessage
	res, err = QueryJson(data, tag)
	if err != nil {
		return
	}

	// TODO: make this work recursively for slices and such
	if recursiveNumCheck(rv.Type()) {
		res = bytes.TrimSuffix(bytes.TrimPrefix(res, []byte("\"")), []byte("\""))
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
