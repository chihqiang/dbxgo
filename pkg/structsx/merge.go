package structsx

import (
	"fmt"
	"reflect"
)

// MergeWithDefaults 先生成默认值结构体，再用用户传入的值覆盖非零字段。
// 返回合并后的结构体。
func MergeWithDefaults[T any](v T) (T, error) {
	var def T
	// 生成默认值结构体
	if err := EnvDefaultsSet(&def); err != nil {
		return def, err
	}
	// 将用户传入的值覆盖默认值
	return MergeStructs[T](def, v)
}

// MergeStructs 合并多个结构体，后面的非零字段覆盖前面。
// 支持值类型和指针类型 struct。
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
	// 合并后续元素
	for _, item := range v[1:] {
		val := reflect.ValueOf(item)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}
		if val.Kind() != reflect.Struct {
			continue
		}
		// 遍历字段并合并非零值
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
