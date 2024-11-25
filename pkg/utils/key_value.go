package utils

import (
	"encoding/json"
	"errors"
	"fadacontrol/internal/base/version"
	"golang.org/x/net/context"
	"google.golang.org/protobuf/reflect/protoreflect"
	"reflect"
	"strconv"
	"strings"
	"time"
)

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
func GetRemoteConfig(key string, region version.ProductRegion, defaultValue interface{}) (interface{}, error) {
	client, err := NewClientBuilder().SetTimeout(5 * time.Second).Build()
	if err != nil {
		return defaultValue, err
	}
	url := "https://update.czqu.net/"
	url = url + version.ProductName + "/" + region.String() + "/" + "config.json"
	resp, err := client.Get(url)
	if err != nil {
		return defaultValue, err
	}
	config := map[string]interface{}{}
	err = json.Unmarshal([]byte(resp), &config)
	if err != nil {
		return defaultValue, err
	}
	value, found := config[key]
	if !found {
		return defaultValue, errors.New("key not found")
	}
	return value, nil
}
func SplitWindowsAccount(account string) (domain, username string) {
	parts := strings.SplitN(account, `\`, 2)
	if len(parts) != 2 {
		return "", account
	}
	return parts[0], parts[1]
}
func GetValueFromContext[T any](ctx context.Context, key string, defaultValue T) T {

	value := ctx.Value(key)
	if value == nil {
		return defaultValue
	}

	castedValue, ok := value.(T)
	if !ok {
		return defaultValue
	}

	return castedValue
}
func ProtoNumberToString(num protoreflect.EnumNumber) string {
	return strconv.FormatInt(int64(num), 10)
}
func Int64ToString(num int64) string {
	return strconv.FormatInt(num, 10)
}
