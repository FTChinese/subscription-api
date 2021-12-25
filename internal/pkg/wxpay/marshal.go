package wxpay

import (
	"fmt"
	"github.com/go-pay/gopay"
	"reflect"
	"strings"
)

var boolToYesNo = map[bool]string{
	true:  "Y",
	false: "N",
}

func getTag(tag string) (string, bool) {
	// See if there's omitempty in the tag
	s := strings.Split(tag, ",")

	switch len(s) {
	case 0:
		return "", true

	case 1:
		return s[0], false

	default:
		return s[0], s[1] == "omitempty"
	}
}

func collectStructValue(value reflect.Value, dest gopay.BodyMap) {
	fmt.Printf("Start collecting struct value %s\n", value.Type().Name())

	if value.Kind() != reflect.Struct {
		fmt.Println("Not a struct, stop.")
		return
	}

	typeOfV := value.Type()

	for i := 0; i < typeOfV.NumField(); i++ {
		sf := typeOfV.Field(i)

		fmt.Printf("Processing struct field %s\n", sf.Name)

		if !sf.IsExported() {
			fmt.Printf("%s is not exported, ignore\n", sf.Name)
			continue
		}

		tag, omitempty := getTag(sf.Tag.Get("map"))
		if tag == "-" {
			continue
		}

		vf := value.Field(i)

		if vf.IsZero() && omitempty {
			continue
		}

		switch vf.Kind() {
		case reflect.Bool:
			dest.Set(tag, boolToYesNo[vf.Bool()])

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			dest.Set(tag, vf.Int())

		case reflect.Float32, reflect.Float64:
			dest.Set(tag, vf.Float())

		case reflect.String:
			dest.Set(tag, vf.String())

		case reflect.Struct:
			if sf.Anonymous {
				fmt.Printf("Recursive into embedded struct %s\n", sf.Name)
				collectStructValue(vf, dest)
			}
		}
	}

	fmt.Printf("Finished collecting struct value %s\n", value.Type().Name())
}

// Marshal turns a struct to map[string]interface.
// Reuse json tag.
func Marshal(v interface{}) gopay.BodyMap {
	bm := make(gopay.BodyMap)

	values := reflect.ValueOf(v)

	collectStructValue(values, bm)

	return bm
}
