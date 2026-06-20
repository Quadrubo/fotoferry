package config

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

// parseGroup reads indexed env var groups (MAPPING__0__*, MAPPING__1__*) into a
// slice of T, stopping at the first index whose sentinel key is empty.
func parseGroup[T any](v *viper.Viper, prefix, sentinel string) ([]T, error) {
	var items []T
	rt := reflect.TypeOf((*T)(nil)).Elem()

	for i := 0; ; i++ {
		p := fmt.Sprintf("%s__%d__", prefix, i)
		if v.GetString(p+sentinel) == "" {
			break
		}
		item := reflect.New(rt).Elem()
		if err := fillStruct(v, p, item); err != nil {
			return nil, err
		}
		items = append(items, item.Interface().(T))
	}
	return items, nil
}

func fillStruct(v *viper.Viper, prefix string, rv reflect.Value) error {
	rt := rv.Type()
	for j := 0; j < rt.NumField(); j++ {
		f := rt.Field(j)
		key := f.Tag.Get("env")
		if key == "" {
			continue
		}
		val := v.GetString(prefix + key)
		if val == "" {
			val = f.Tag.Get("default")
		}
		if err := setField(rv.Field(j), prefix+key, val); err != nil {
			return err
		}
	}
	return nil
}

func setField(field reflect.Value, name, val string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(val)
	case reflect.Bool:
		if val == "" {
			return nil
		}
		b, err := strconv.ParseBool(strings.TrimSpace(val))
		if err != nil {
			return fmt.Errorf("config: %s: invalid boolean value %q", name, val)
		}
		field.SetBool(b)
	case reflect.Slice:
		field.Set(reflect.ValueOf(splitList(val)))
	default:
		panic(fmt.Sprintf("config: %s: unsupported field kind %s", name, field.Kind()))
	}
	return nil
}

func splitList(val string) []string {
	var parts []string
	for _, s := range strings.Split(val, ",") {
		if s = strings.TrimSpace(s); s != "" {
			parts = append(parts, s)
		}
	}
	return parts
}
