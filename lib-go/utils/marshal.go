package utils

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// This interface stands for serde of derivation-style algebraic
// data types. For example, to serde the Option type defined
// in option.go, we need to implement Kind method for Some
// and None. Then Some{x} will be marshaled by Marshal to
//
//	{"kind": "Some", "val": {"x": x}}.
//
// And None will be marshaled to {"kind": "None"}.
//
// The Unmarshal function will correctly
// unmarshal the json to Some{x} or None.
// Note that the Unmarshal function needs an argument named
// marshalMap which can be created by MakeMarshalMapping.
// Example use:
//
//	utils.MakeMarshalMapping([]Option{&Some{}, &None{}})
type Marshalable interface {
	Kind() string
}

func Clone[T any](t T) T {
	newObj := reflect.New(reflect.TypeOf(t).Elem())
	oldVal := reflect.ValueOf(t).Elem()
	newVal := newObj.Elem()
	for i := 0; i < oldVal.NumField(); i++ {
		newValField := newVal.Field(i)
		if newValField.CanSet() {
			newValField.Set(oldVal.Field(i))
		}
	}

	return newObj.Interface().(T)
}

func Marshal(v Marshalable) ([]byte, error) {
	type Container struct {
		Kind string `json:"kind"`
		Val  any    `json:"val"`
	}
	wrapper := Container{
		Kind: v.Kind(),
		Val:  v,
	}
	res, err := json.Marshal(&wrapper)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func Unmarshal[T Marshalable](input []byte, marshalMap map[string]T) (T, error) {
	wrapper := map[string]json.RawMessage{}
	err := json.Unmarshal(input, &wrapper)
	if err != nil {
		// panic("Unmarshal error: this should never happend. ")
		return marshalMap[""], err
	}
	kind := ""
	err = json.Unmarshal(wrapper["kind"], &kind)
	if err != nil {
		panic("Unmarshal error: arguments has no field kind. ")
	}
	val, ok := marshalMap[kind]
	if !ok {
		panic(fmt.Sprintf("Unmarshal error: marshalMap has not been implemented completely: %v", kind))
	}
	newVal := Clone(val)
	err = json.Unmarshal(wrapper["val"], newVal)
	if err != nil {
		return newVal, err
	}
	return newVal, nil
}

func MakeMarshalMapping[T Marshalable](items []T) map[string]T {
	return MakeMapping(func(k T) string { return k.Kind() }, items)
}
