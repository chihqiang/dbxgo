package structx

import (
	"fmt"
	"reflect"
)

// MergeWithDefaults First, generate a struct with default values, then overwrite non-zero fields with values from the user.
// Returns the merged struct.
func MergeWithDefaults[T any](v T) (T, error) {
	var def T
	// Generate a struct with default values
	if err := SetEnvDefault(&def); err != nil {
		return def, err
	}
	// Overwrite the default values with the user's values
	return MergeStructs[T](def, v)
}

// MergeStructs Merges multiple structs, where later non-zero fields overwrite earlier ones.
// Supports both value types and pointer types of structs.
func MergeStructs[T any](v ...T) (T, error) {
	var zero T
	if len(v) == 0 {
		return zero, fmt.Errorf("no values provided")
	}
	firstVal := reflect.ValueOf(v[0])
	var result reflect.Value
	switch firstVal.Kind() {
	case reflect.Ptr:
		if firstVal.Elem().Kind() != reflect.Struct {
			return zero, fmt.Errorf("pointer must point to struct")
		}
		result = reflect.New(firstVal.Elem().Type())
		result.Elem().Set(firstVal.Elem())
	case reflect.Struct:
		result = reflect.New(firstVal.Type()).Elem()
		result.Set(firstVal)
	default:
		return zero, fmt.Errorf("unsupported type: %s", firstVal.Kind())
	}
	// Merge subsequent elements
	for _, item := range v[1:] {
		val := reflect.ValueOf(item)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}
		if val.Kind() != reflect.Struct {
			continue
		}
		// Iterate through fields and merge non-zero values
		dst := result
		if dst.Kind() == reflect.Ptr {
			dst = dst.Elem()
		}
		for i := 0; i < val.NumField(); i++ {
			field := val.Field(i)
			zeroField := reflect.Zero(field.Type())
			if !reflect.DeepEqual(field.Interface(), zeroField.Interface()) {
				dst.Field(i).Set(field)
			}
		}
	}
	return result.Interface().(T), nil
}
