package structsx

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// EnvDefaultsSet 给结构体应用默认值，默认值来源于 tag。
func EnvDefaultsSet(s any) error {
	return (&DefaultSetter{TagName: "envDefault", Separator: ","}).Set(s)
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
