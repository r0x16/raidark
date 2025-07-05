package logger

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"reflect"

	"github.com/r0x16/Raidark/shared/logger/domain"
)

type StdOutLogManager struct {
	logger   *slog.Logger
	logLevel domain.LogLevel
}

var _ domain.LogProvider = &StdOutLogManager{}

func NewStdOutLogManager() *StdOutLogManager {
	manager := &StdOutLogManager{
		logger:   slog.New(slog.NewJSONHandler(os.Stdout, nil)),
		logLevel: domain.Info,
	}
	return manager
}

// Debug implements logger.LogProvider.
func (s *StdOutLogManager) Debug(msg string, data map[string]any) {
	if s.logLevel > domain.Debug {
		return
	}
	s.logger.Debug(msg, s.parseData(data)...)
}

// Info implements logger.LogProvider.
func (s *StdOutLogManager) Info(msg string, data map[string]any) {
	if s.logLevel > domain.Info {
		return
	}
	s.logger.Info(msg, s.parseData(data)...)
}

// Warning implements logger.LogProvider.
func (s *StdOutLogManager) Warning(msg string, data map[string]any) {
	if s.logLevel > domain.Warning {
		return
	}
	s.logger.Warn(msg, s.parseData(data)...)
}

// Error implements logger.LogProvider.
func (s *StdOutLogManager) Error(msg string, data map[string]any) {
	s.logger.Error(msg, s.parseData(data)...)
}

// Critical implements logger.LogProvider.
func (s *StdOutLogManager) Critical(msg string, data map[string]any) {
	s.logger.Error(msg, s.parseData(data)...)
}

// SetLogLevel implements logger.LogProvider.
func (s *StdOutLogManager) SetLogLevel(level domain.LogLevel) {
	s.logLevel = level
}

// Parses data of type map to a slice of slog Attrs.
func (s *StdOutLogManager) parseData(data map[string]any) []any {
	attrs := make([]any, 0, len(data))
	for key, value := range data {
		sanitizedValue := s.sanitizeValue(value)
		attrs = append(attrs, slog.Any(key, sanitizedValue))
	}
	return attrs
}

// sanitizeValue converts complex values that cannot be JSON serialized to simpler representations
func (s *StdOutLogManager) sanitizeValue(value any) any {
	if value == nil {
		return nil
	}

	// Try to marshal to JSON to detect circular references or unsupported types
	if _, err := json.Marshal(value); err != nil {
		// If JSON marshalling fails, return a safe representation
		return s.createSafeRepresentation(value)
	}

	// For complex types, we might still want to limit depth
	return s.limitDepth(value)
}

// createSafeRepresentation creates a safe representation of complex values
func (s *StdOutLogManager) createSafeRepresentation(value any) any {
	if value == nil {
		return nil
	}

	v := reflect.ValueOf(value)
	t := reflect.TypeOf(value)

	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			return fmt.Sprintf("<%s: nil>", t.String())
		}
		return fmt.Sprintf("<%s: %p>", t.String(), value)
	case reflect.Struct:
		return s.extractStructFields(v, t)
	case reflect.Slice, reflect.Array:
		return fmt.Sprintf("<%s: length=%d>", t.String(), v.Len())
	case reflect.Map:
		return fmt.Sprintf("<%s: size=%d>", t.String(), v.Len())
	case reflect.Chan:
		return fmt.Sprintf("<%s: %p>", t.String(), value)
	case reflect.Func:
		return fmt.Sprintf("<%s: %p>", t.String(), value)
	default:
		return fmt.Sprintf("<%s: %v>", t.String(), value)
	}
}

// extractStructFields extracts basic information from struct fields
func (s *StdOutLogManager) extractStructFields(v reflect.Value, t reflect.Type) map[string]any {
	result := make(map[string]any)
	result["_type"] = t.String()

	// Only extract first level fields and limit to avoid circular references
	numFields := v.NumField()
	if numFields > 10 {
		numFields = 10 // Limit number of fields to prevent massive logs
	}

	for i := 0; i < numFields; i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Only include simple types in the first level
		switch fieldValue.Kind() {
		case reflect.String:
			result[field.Name] = fieldValue.String()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			result[field.Name] = fieldValue.Int()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			result[field.Name] = fieldValue.Uint()
		case reflect.Float32, reflect.Float64:
			result[field.Name] = fieldValue.Float()
		case reflect.Bool:
			result[field.Name] = fieldValue.Bool()
		case reflect.Ptr:
			if fieldValue.IsNil() {
				result[field.Name] = nil
			} else {
				result[field.Name] = fmt.Sprintf("<%s: %p>", fieldValue.Type().String(), fieldValue.Interface())
			}
		default:
			result[field.Name] = fmt.Sprintf("<%s>", fieldValue.Type().String())
		}
	}

	return result
}

// limitDepth limits the depth of complex structures to prevent deep nesting
func (s *StdOutLogManager) limitDepth(value any) any {
	if value == nil {
		return nil
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		// For slices/arrays, limit to first few elements and their basic info
		length := v.Len()
		if length > 5 {
			simplified := make([]any, 5)
			for i := 0; i < 5; i++ {
				elem := v.Index(i).Interface()
				simplified[i] = s.simplifyElement(elem)
			}
			return map[string]any{
				"_type":      fmt.Sprintf("%s", v.Type().String()),
				"_length":    length,
				"_elements":  simplified,
				"_truncated": true,
			}
		}
		// For small arrays, still simplify elements
		simplified := make([]any, length)
		for i := 0; i < length; i++ {
			elem := v.Index(i).Interface()
			simplified[i] = s.simplifyElement(elem)
		}
		return simplified
	case reflect.Map:
		return fmt.Sprintf("<%s: size=%d>", v.Type().String(), v.Len())
	default:
		return value
	}
}

// simplifyElement creates a simplified representation of an element
func (s *StdOutLogManager) simplifyElement(elem any) any {
	if elem == nil {
		return nil
	}

	v := reflect.ValueOf(elem)
	t := reflect.TypeOf(elem)

	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			return fmt.Sprintf("<%s: nil>", t.String())
		}
		return fmt.Sprintf("<%s: %p>", t.String(), elem)
	case reflect.Struct:
		return fmt.Sprintf("<%s>", t.String())
	case reflect.Slice, reflect.Array:
		return fmt.Sprintf("<%s: length=%d>", t.String(), v.Len())
	case reflect.Map:
		return fmt.Sprintf("<%s: size=%d>", t.String(), v.Len())
	default:
		// For simple types, return as is if they can be JSON serialized
		if _, err := json.Marshal(elem); err == nil {
			return elem
		}
		return fmt.Sprintf("<%s: %v>", t.String(), elem)
	}
}
