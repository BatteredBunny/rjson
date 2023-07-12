package main

import (
	"encoding/json"
	"errors"
	"log"
	"reflect"
	"strings"
)

const tagname = "rjson"
const divider = "."

func valueFinder(input []byte, tag string) (result json.RawMessage) {
	fields := strings.Split(strings.TrimSuffix(tag, divider), divider)

	var object map[string]json.RawMessage
	if err := json.Unmarshal(input, &object); err != nil {
		log.Fatal(err)
	}

	for i, field := range fields {
		log.Println(string(object[field]), i == len(fields)-1)

		var err error
		if i == len(fields)-1 {
			err = json.Unmarshal(object[field], &result)
		} else {
			err = json.Unmarshal(object[field], &object)
		}

		if err != nil {
			log.Fatal(err)
		}
	}

	return
}

var ErrNotAPointer = errors.New("Please insert a pointer")

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

			if err = json.Unmarshal(valueFinder(data, tag), &inter); err != nil {
				return
			}

			if valuefield.CanSet() {
				valuefield.Set(reflect.Indirect(test))
			}
		}
	}

	return
}
