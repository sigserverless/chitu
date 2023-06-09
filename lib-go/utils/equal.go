package utils

import "reflect"

func DeepEqual(a, b interface{}) bool {
	valA := reflect.ValueOf(a)
	valB := reflect.ValueOf(b)

	if valA.Kind() != valB.Kind() {
		return false
	}

	switch valA.Kind() {
	case reflect.Map:
		if valA.Len() != valB.Len() {
			return false
		}

		for _, key := range valA.MapKeys() {
			valAElement := valA.MapIndex(key)
			valBElement := valB.MapIndex(key)

			if !DeepEqual(valAElement.Interface(), valBElement.Interface()) {
				return false
			}
		}
		return true

	case reflect.Slice:
		if valA.Len() != valB.Len() {
			return false
		}

		for i := 0; i < valA.Len(); i++ {
			if !DeepEqual(valA.Index(i).Interface(), valB.Index(i).Interface()) {
				return false
			}
		}
		return true

	default:
		return reflect.DeepEqual(a, b)
	}
}
