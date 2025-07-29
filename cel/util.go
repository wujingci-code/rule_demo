package cel

import "reflect"

func structToMap(s interface{}) map[string]interface{} {
	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()

	out := make(map[string]interface{}, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		name := field.Tag.Get("json")
		if name == "" {
			name = field.Name
		}
		out[name] = v.Field(i).Interface()
	}
	return out
}
