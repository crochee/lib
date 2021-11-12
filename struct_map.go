package lirity

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/crochee/lirity/variable"
)

func Struct2Map(s interface{}) map[string]interface{} {
	return Struct2MapWithTag(s, "json")
}

func Struct2MapWithTag(s interface{}, tagName string) map[string]interface{} {
	t := reflect.TypeOf(s)
	value := reflect.ValueOf(s)

	if value.Kind() == reflect.Ptr && value.Elem().Kind() == reflect.Struct {
		t = t.Elem()
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		return nil
	}
	m := make(map[string]interface{})

	for i := 0; i < t.NumField(); i++ {
		fv := value.Field(i)
		ft := t.Field(i)

		if !fv.CanInterface() {
			continue
		}

		if ft.PkgPath != "" { // unexported
			continue
		}
		name, option := parseTag(ft.Tag.Get(tagName))
		if name == "-" {
			continue // ignore "-"
		}
		if name == "" {
			name = ft.Name // use field name
		}
		if option == "omitempty" && isEmpty(&fv) {
			continue // skip empty field
		}
		// ft.Anonymous means embedded field
		if ft.Anonymous {
			if fv.Kind() == reflect.Ptr && fv.IsNil() {
				continue // nil
			}
			if (fv.Kind() == reflect.Struct) ||
				(fv.Kind() == reflect.Ptr && fv.Elem().Kind() == reflect.Struct) {
				// embedded struct
				embedded := Struct2MapWithTag(fv.Interface(), tagName)
				for embName, embValue := range embedded {
					m[embName] = embValue
				}
			}
			continue
		}

		if option == "string" {
			temp := num2String(fv)
			if temp != nil {
				m[name] = temp
				continue
			}
		}

		m[name] = fv.Interface()
	}

	return m
}

func num2String(fv reflect.Value) interface{} {
	kind := fv.Kind()
	if kind == reflect.Int || kind == reflect.Int8 || kind == reflect.Int16 || kind == reflect.Int32 || kind == reflect.Int64 {
		return strconv.FormatInt(fv.Int(), variable.DecimalSystem)
	} else if kind == reflect.Uint || kind == reflect.Uint8 || kind == reflect.Uint16 || kind == reflect.Uint32 || kind == reflect.Uint64 {
		return strconv.FormatUint(fv.Uint(), variable.DecimalSystem)
	} else if kind == reflect.Float32 || kind == reflect.Float64 {
		return strconv.FormatFloat(fv.Float(), 'f', 2, 64)
	}
	// TODO support more types
	return nil
}

func isEmpty(value *reflect.Value) bool {
	k := value.Kind()
	if k == reflect.Bool {
		return !value.Bool()
	} else if reflect.Int < k && k < reflect.Int64 {
		return value.Int() == 0
	} else if reflect.Uint < k && k < reflect.Uintptr {
		return value.Uint() == 0
	} else if k == reflect.Float32 || k == reflect.Float64 {
		return value.Float() == 0
	} else if k == reflect.Array || k == reflect.Map || k == reflect.Slice || k == reflect.String {
		return value.Len() == 0
	} else if k == reflect.Interface || k == reflect.Ptr {
		return value.IsNil()
	}
	return false
}

func parseTag(tag string) (tag0, tag1 string) {
	tags := strings.Split(tag, ",")

	if len(tags) == 0 {
		return
	}

	if len(tags) == 1 {
		tag0 = tags[0]
		return
	}
	tag0, tag1 = tags[0], tags[1]
	return
}
