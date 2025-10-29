package structx

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// SetEnvDefault Applies default values to the struct, with default values sourced from the tag.
func SetEnvDefault(s any) error {
	return (&DefaultSetter{TagName: "envDefault", Separator: ","}).Set(s)
}

// DefaultSetter manages the application of default values to struct fields
type DefaultSetter struct {
	TagName   string // Tag name for storing default values (e.g., "default")
	Separator string // Separator for slice elements in default value tags (e.g., ",")
}

// Set applies default values to struct fields recursively
func (sd *DefaultSetter) Set(s interface{}) error {
	v := reflect.ValueOf(s)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return fmt.Errorf("expected non-nil pointer to struct, got %T", s)
	}

	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("pointer must point to struct, got %T", s)
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		structField := t.Field(i)

		defaultTag := structField.Tag.Get(sd.TagName)
		if !field.CanSet() || (defaultTag == "" && !isComplexType(field)) {
			continue
		}

		if err := sd.applyDefaultValue(field, structField, defaultTag); err != nil {
			return fmt.Errorf("field %s: %w", structField.Name, err)
		}
	}

	return nil
}

// applyDefaultValue dispatches to the correct type-specific function based on the field's kind
func (sd *DefaultSetter) applyDefaultValue(field reflect.Value, structField reflect.StructField, defaultTag string) error {
	switch field.Kind() {
	case reflect.String:
		return sd.setString(field, defaultTag)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return sd.setInt(field, defaultTag)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return sd.setUint(field, defaultTag)
	case reflect.Float32, reflect.Float64:
		return sd.setFloat(field, defaultTag)
	case reflect.Bool:
		return sd.setBool(field, defaultTag)
	case reflect.Struct:
		return sd.setStruct(field)
	case reflect.Ptr:
		return sd.setPtr(field)
	case reflect.Slice:
		return sd.setSlice(field, defaultTag, structField.Name)
	case reflect.Map:
		return sd.setMap(field, defaultTag)
	default:
		return fmt.Errorf("unsupported field type: %s", field.Kind())
	}
}

// setString handles default value assignment for string fields
func (sd *DefaultSetter) setString(field reflect.Value, defaultTag string) error {
	if field.String() == "" {
		field.SetString(defaultTag)
	}
	return nil
}

// setInt handles default value assignment for signed integer fields
func (sd *DefaultSetter) setInt(field reflect.Value, defaultTag string) error {
	if field.Int() == 0 && defaultTag != "" {
		val, err := strconv.ParseInt(defaultTag, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid int value '%s': %w", defaultTag, err)
		}
		field.SetInt(val)
	}
	return nil
}

// setUint handles default value assignment for unsigned integer fields
func (sd *DefaultSetter) setUint(field reflect.Value, defaultTag string) error {
	if field.Uint() == 0 && defaultTag != "" {
		val, err := strconv.ParseUint(defaultTag, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid uint value '%s': %w", defaultTag, err)
		}
		field.SetUint(val)
	}
	return nil
}

// setFloat handles default value assignment for floating-point fields
func (sd *DefaultSetter) setFloat(field reflect.Value, defaultTag string) error {
	if field.Float() == 0 && defaultTag != "" {
		val, err := strconv.ParseFloat(defaultTag, 64)
		if err != nil {
			return fmt.Errorf("invalid float value '%s': %w", defaultTag, err)
		}
		field.SetFloat(val)
	}
	return nil
}

// setBool handles default value assignment for boolean fields
func (sd *DefaultSetter) setBool(field reflect.Value, defaultTag string) error {
	if !field.Bool() && defaultTag != "" {
		val, err := strconv.ParseBool(defaultTag)
		if err != nil {
			return fmt.Errorf("invalid bool value '%s': %w", defaultTag, err)
		}
		field.SetBool(val)
	}
	return nil
}

// setStruct handles default value assignment for struct fields (nested structures)
func (sd *DefaultSetter) setStruct(field reflect.Value) error {
	return sd.Set(field.Addr().Interface())
}

// setPtr handles default value assignment for pointer fields
func (sd *DefaultSetter) setPtr(field reflect.Value) error {
	if field.IsNil() {
		newField := reflect.New(field.Type().Elem())
		field.Set(newField)
	}
	return sd.Set(field.Interface())
}

// setSlice handles default value assignment for slice fields
func (sd *DefaultSetter) setSlice(field reflect.Value, defaultTag, fieldName string) error {
	if field.Len() == 0 && defaultTag != "" {
		parts := strings.Split(defaultTag, sd.Separator)
		slice := reflect.MakeSlice(field.Type(), len(parts), len(parts))

		for j, part := range parts {
			elem := slice.Index(j)
			if err := sd.setSliceElement(elem, part); err != nil {
				return err
			}
		}
		field.Set(slice)
	}
	return nil
}

// setSliceElement handles default value assignment for individual slice elements
func (sd *DefaultSetter) setSliceElement(elem reflect.Value, part string) error {
	switch elem.Kind() {
	case reflect.String:
		elem.SetString(part)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		val, err := strconv.ParseInt(part, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid int slice element '%s': %w", part, err)
		}
		elem.SetInt(val)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val, err := strconv.ParseUint(part, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid uint slice element '%s': %w", part, err)
		}
		elem.SetUint(val)
	case reflect.Float32, reflect.Float64:
		val, err := strconv.ParseFloat(part, 64)
		if err != nil {
			return fmt.Errorf("invalid float slice element '%s': %w", part, err)
		}
		elem.SetFloat(val)
	case reflect.Bool:
		val, err := strconv.ParseBool(part)
		if err != nil {
			return fmt.Errorf("invalid bool slice element '%s': %w", part, err)
		}
		elem.SetBool(val)
	default:
		return fmt.Errorf("unsupported slice element type: %s", elem.Kind())
	}
	return nil
}

// setMap handles default value assignment for map fields
func (sd *DefaultSetter) setMap(field reflect.Value, defaultTag string) error {
	// If the field is an empty map and defaultTag is not empty
	if field.Len() == 0 && defaultTag != "" {
		// Create a new empty map
		mapValue := reflect.MakeMap(field.Type())

		// Try to parse defaultTag as a JSON string
		if isJSON(defaultTag) {
			// Try unmarshalling the defaultTag as a JSON object into the map
			tmpPtr := reflect.New(field.Type())
			if err := json.Unmarshal([]byte(defaultTag), tmpPtr.Interface()); err == nil {
				// If parsing is successful, set the value to the field
				field.Set(tmpPtr.Elem())
				return nil
			}
		}

		// If it's not a JSON string, parse the defaultTag as key-value pairs
		// env=dev,version=1.0,debug=true
		entries := strings.Split(defaultTag, sd.Separator)
		for _, entry := range entries {
			// Split each key-value pair
			parts := strings.SplitN(entry, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid map entry format: %s", entry)
			}

			// Parse the key and value
			keyStr := parts[0]
			valueStr := parts[1]

			// Convert the key and value to the correct types based on the map's key and value types
			key := reflect.ValueOf(keyStr).Convert(field.Type().Key())
			value := reflect.ValueOf(valueStr).Convert(field.Type().Elem())

			// Set the key-value pair in the map
			mapValue.SetMapIndex(key, value)
		}

		// Set the populated map back to the field
		field.Set(mapValue)
	}
	return nil
}

// isJSON checks if a string is in valid JSON format
func isJSON(str string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(str), &js) == nil
}

// isComplexType checks if the field is of type struct or pointer, for which defaults should be set
func isComplexType(field reflect.Value) bool {
	return field.Kind() == reflect.Struct || field.Kind() == reflect.Ptr
}
