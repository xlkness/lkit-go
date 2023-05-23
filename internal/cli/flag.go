package cli

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
)

func extractArgPairs2Flag(flag interface{}, args []*argPair) error {
	if flag == nil {
		return nil
	}

	var stTo = reflect.TypeOf(flag)
	var stVo = reflect.ValueOf(flag)
	switch stTo.Kind() {
	case reflect.Ptr:
		stTo = stTo.Elem()
		stVo = stVo.Elem()
	// case reflect.Struct:
	// 	break
	default:
		return fmt.Errorf("invalid flags parse struct(%+v), must be pointer or struct", flag)
	}

	for i := 0; i < stTo.NumField(); i++ {
		field := stTo.Field(i)

		key, find := field.Tag.Lookup("name")
		if !find {
			continue
		}

		//desc := field.Tag.Get("desc")

		setValue, find := os.LookupEnv(key)
		if !find {
			setValue, find = field.Tag.Lookup("default")
			if !find {
				setValue = fmt.Sprintf("not_found_env_%v", key)
			}
		}

		for _, arg := range args {
			if arg.flag == key {
				setValue = arg.value
				break
			}
		}

		switch field.Type.Kind() {
		case reflect.String:
			stVo.Field(i).SetString(setValue)
		case reflect.Int, reflect.Int64:
			value, err := strconv.ParseInt(setValue, 10, 64)
			if err != nil {
				return fmt.Errorf("get field %v atoi error:%v", setValue, err)
			}
			stVo.Field(i).SetInt(int64(value))
		case reflect.Bool:
			stVo.Field(i).SetBool(setValue == "true")
		default:
			return fmt.Errorf("parse flag kind invalid,must be string/int/int64/bool, not %+v", field.Type.Kind())
		}
	}

	return nil
}
