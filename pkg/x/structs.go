package x

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
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

// EnvDefaultsSet 给结构体应用默认值，默认值来源于 tag。
func EnvDefaultsSet(s any) error {
	return (&DefaultSetter{TagName: "envDefault", Separator: ","}).Set(s)
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

// DefaultSetter 用于统一管理默认值应用
type DefaultSetter struct {
	TagName   string // 用哪个 tag 名称
	Separator string // slice 分隔符
}

// Set 给结构体应用默认值，s 必须是 struct 指针
func (sd *DefaultSetter) Set(s interface{}) error {
	v := reflect.ValueOf(s)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return fmt.Errorf("expected pointer to struct, got %T", s)
	}
	v = v.Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		structField := t.Field(i)

		defaultTag := structField.Tag.Get(sd.TagName)
		if !field.CanSet() || defaultTag == "" {
			continue
		}

		switch field.Kind() {
		case reflect.String:
			if field.String() == "" {
				field.SetString(defaultTag)
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if field.Int() == 0 {
				val, err := strconv.ParseInt(defaultTag, 10, 64)
				if err != nil {
					return fmt.Errorf("field %s: %w", structField.Name, err)
				}
				field.SetInt(val)
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if field.Uint() == 0 {
				val, err := strconv.ParseUint(defaultTag, 10, 64)
				if err != nil {
					return fmt.Errorf("field %s: %w", structField.Name, err)
				}
				field.SetUint(val)
			}
		case reflect.Float32, reflect.Float64:
			if field.Float() == 0 {
				val, err := strconv.ParseFloat(defaultTag, 64)
				if err != nil {
					return fmt.Errorf("field %s: %w", structField.Name, err)
				}
				field.SetFloat(val)
			}
		case reflect.Bool:
			if !field.Bool() {
				val, err := strconv.ParseBool(defaultTag)
				if err != nil {
					return fmt.Errorf("field %s: %w", structField.Name, err)
				}
				field.SetBool(val)
			}
		case reflect.Struct:
			if err := sd.Set(field.Addr().Interface()); err != nil {
				return err
			}
		case reflect.Ptr:
			if field.IsNil() {
				newField := reflect.New(field.Type().Elem())
				field.Set(newField)
			}
			if err := sd.Set(field.Interface()); err != nil {
				return err
			}
		case reflect.Slice:
			if field.Len() == 0 {
				parts := strings.Split(defaultTag, sd.Separator)
				slice := reflect.MakeSlice(field.Type(), len(parts), len(parts))
				for j, part := range parts {
					elem := slice.Index(j)
					switch elem.Kind() {
					case reflect.String:
						elem.SetString(part)
					case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
						val, err := strconv.ParseInt(part, 10, 64)
						if err != nil {
							return fmt.Errorf("field %s: %w", structField.Name, err)
						}
						elem.SetInt(val)
					case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
						val, err := strconv.ParseUint(part, 10, 64)
						if err != nil {
							return fmt.Errorf("field %s: %w", structField.Name, err)
						}
						elem.SetUint(val)
					case reflect.Float32, reflect.Float64:
						val, err := strconv.ParseFloat(part, 64)
						if err != nil {
							return fmt.Errorf("field %s: %w", structField.Name, err)
						}
						elem.SetFloat(val)
					case reflect.Bool:
						val, err := strconv.ParseBool(part)
						if err != nil {
							return fmt.Errorf("field %s: %w", structField.Name, err)
						}
						elem.SetBool(val)
					}
				}
				field.Set(slice)
			}
		case reflect.Map:
			if field.Len() == 0 {
				tmpPtr := reflect.New(field.Type())
				if err := json.Unmarshal([]byte(defaultTag), tmpPtr.Interface()); err != nil {
					return fmt.Errorf("field %s: %w", structField.Name, err)
				}
				field.Set(tmpPtr.Elem())
			}
		}
	}
	return nil
}
