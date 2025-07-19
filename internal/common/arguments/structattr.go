package arguments

import (
	"reflect"
)

// SlowAttributeArray converts an input value into an array of field names and values, for use in slogger methods.
// Note: This function is not optimized for performance and should not be used in performance-critical contexts.
// This function uses reflection and performs recursion on nested objects
func SlowAttributeArray[T any](tag string, x T) []any {
	attrs := make([]any, 0)
	val := reflect.ValueOf(x)
	if val.Kind() != reflect.Struct {
		attrs = append(attrs, tag, x)
		return attrs
	}
	return append(attrs, structRecursion(tag, val)...)
}

func structRecursion(tag string, val reflect.Value) []any {
	attrs := make([]any, 0)
	if val.NumField() == 0 {
		attrs = append(attrs, tag, struct{}{})
		return attrs
	}
	fieldTagPrefix := tag
	if len(fieldTagPrefix) > 0 {
		fieldTagPrefix = fieldTagPrefix + "."
	}
	for i := 0; i < val.NumField(); i++ {
		// Skip unexported fields
		if !val.Type().Field(i).IsExported() {
			continue
		}

		field := val.Field(i)

		// Handle interface{} fields
		if field.Kind() == reflect.Interface && !field.IsNil() && field.Elem().Kind() == reflect.Struct {
			// Extract the underlying value from the interface
			field = field.Elem()
		}
		// Handle pointer fields
		if field.Kind() == reflect.Ptr && !field.IsNil() && field.Elem().Kind() == reflect.Struct {
			// Extract the underlying value from the pointer
			field = field.Elem()
		}

		fieldTag := fieldTagPrefix + val.Type().Field(i).Name
		fieldValue := field.Interface()

		if field.Kind() == reflect.Struct {
			attrs = append(attrs, structRecursion(fieldTag, field)...)
		} else {
			attrs = append(attrs, fieldTag, fieldValue)
		}
	}
	if len(attrs) == 0 {
		attrs = append(attrs, tag, struct{}{})
	}
	return attrs
}
