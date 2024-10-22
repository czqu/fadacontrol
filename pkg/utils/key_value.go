package utils

import "reflect"

type KeyValue map[string]interface{}

func (kv KeyValue) Int(name string, defaultValue int) int {
	if v, found := kv[name]; found {
		if castValue, is := v.(int); is {
			return castValue
		}
	}
	return defaultValue
}
func (kv KeyValue) String(name string, defaultValue string) string {
	if v, found := kv[name]; found {
		if castValue, is := v.(string); is {
			return castValue
		}
	}
	return defaultValue
}
func (kv KeyValue) Bool(name string, defaultValue bool) bool {
	if v, found := kv[name]; found {
		if castValue, is := v.(bool); is {
			return castValue
		}
	}
	return defaultValue
}

// StructToMap converts struct to map,ignore tag "-"
func StructToMap(obj interface{}) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	val := reflect.ValueOf(obj).Elem()
	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		tag := field.Tag.Get("json")
		if tag == "-" {
			continue
		}
		data[field.Name] = val.Field(i).Interface()
	}
	return data, nil
}
