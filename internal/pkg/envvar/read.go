package envvar

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const tagName = "env"

// Read fills the target with the env var
func Read(target interface{}) error {
	rv := reflect.ValueOf(target)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("invalid type %v passed", reflect.TypeOf(target))
	}

	v := rv.Elem()
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("not a struct pointer")
	}

	refType := v.Type()

	envVarKeys := make([]string, refType.NumField())

	for i := 0; i < refType.NumField(); i++ {
		refTypeField := refType.Field(i)
		envVarKey, ok := refTypeField.Tag.Lookup(tagName)
		if !ok {
			return fmt.Errorf("env tag not set for field %s", refTypeField.Name)
		}
		envVarKeys[i] = envVarKey
	}

	for i, key := range envVarKeys {
		if key == "-" {
			continue
		}

		value, ok := os.LookupEnv(key)
		if !ok {
			return fmt.Errorf("%s not present", key)
		}
		value = strings.TrimSpace(value)

		fieldValue := v.Field(i)
		if !fieldValue.IsValid() || !fieldValue.CanSet() {
			return fmt.Errorf("field %s is not valid or cannot be set", refType.Field(i).Name)
		}

		// try parsing the "type" first
		fieldType := refType.Field(i)
		switch fieldType.Type {
		case reflect.TypeOf(time.Nanosecond):
			t, err := time.ParseDuration(value)
			if err != nil {
				return err
			}
			fieldValue.Set(reflect.ValueOf(t))
			continue
		}

		// then try the built-in "kinds"
		switch fieldValue.Kind() {
		case reflect.String:
			fieldValue.SetString(value)
		case reflect.Int:
			num, err := strconv.Atoi(value)
			if err != nil {
				return err
			}
			if fieldValue.OverflowInt(int64(num)) {
				return fmt.Errorf("int %d overflows for field %s", num, refType.Field(i).Name)
			}
			fieldValue.SetInt(int64(num))
		case reflect.Float64:
			f, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return err
			}
			if fieldValue.OverflowFloat(f) {
				return fmt.Errorf("float64 %v overflows for field %s", f, refType.Field(i).Name)
			}
			fieldValue.SetFloat(f)
		default:
			return fmt.Errorf("unsupported type %s", fieldValue.Type().Name())
		}
	}
	return nil
}
