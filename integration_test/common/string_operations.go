package common

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/golang/glog"
	"gopkg.in/yaml.v2"
)

// AllUnicodeWhiteSpaces is a string which has all the white space characters in it.
// We can use it in strings.Trim, strings.Split etc.
const AllUnicodeWhiteSpaces = "\t\n\v\f\r \x85\xA0"

// TrimWhitespaces takes a string then it trims all the whitespace from the same and return
func TrimWhitespaces(str string) string {
	return strings.Trim(str, AllUnicodeWhiteSpaces)
	// Or we can effectively write the following too
	// return strings.TrimFunc(str, unicode.IsSpace)
}

// PrettyString returns the prettified string of the interface supplied. (If it can)
func PrettyString(in interface{}) string {
	jsonStr, err := json.MarshalIndent(in, "", "    ")
	if err != nil {
		if Debug {
			err := fmt.Errorf("Unable to marshal, Error: %+v", err)
			if err != nil {
				fmt.Printf("Unable to marshal, Error: %+v\n", err)
			}
		}
		return fmt.Sprintf("%+v", in)
	}
	return string(jsonStr)
}

// ReplaceHexCodesWithValue finds any string slice which is equivalent to hexadecimal value of
// a character then it replaces that with its value.
// If any error occured while resolving it leaves that part as it is.
// e.g. "\\x20" in a string will be replaced with value of "\x20" i.e. space
func ReplaceHexCodesWithValue(s string) string {
	regexString := "\\\\[xX][0-9a-fA-F]{2}"
	pattern, err := regexp.Compile(regexString)
	if err != nil {
		glog.Fatalf("Error while compiling regex: %q. Error: %+v", regexString, err)
	}

	return pattern.ReplaceAllStringFunc(s, func(s string) string {
		bytes, err := hex.DecodeString(s[2:])
		if err != nil {
			glog.Errorf("Error occured while resolving %q. Error: %+v", s, err)
			return s
		}
		return string(bytes)
	})
}

// ConvertMapI2MapS walks the given dynamic object recursively, and
// converts maps with interface{} key type to maps with string key type.
// This function comes handy if you want to marshal a dynamic object into
// JSON where maps with interface{} key type are not allowed.
//
// Recursion is implemented into values of the following types:
//   -map[interface{}]interface{}
//   -map[string]interface{}
//   -[]interface{}
//
// When converting map[interface{}]interface{} to map[string]interface{},
// fmt.Sprint() with default formatting is used to convert the key to a string key.
//
// Source: https://github.com/icza/dyno/blob/6009b3da28e195fd676c79e5bcbee68bcda793e3/dyno.go#L515
func ConvertMapI2MapS(v interface{}) interface{} {
	switch x := v.(type) {
	case map[interface{}]interface{}:
		m := map[string]interface{}{}
		for k, v2 := range x {
			switch k2 := k.(type) {
			case string: // Fast check if it's already a string
				m[k2] = ConvertMapI2MapS(v2)
			default:
				m[fmt.Sprint(k)] = ConvertMapI2MapS(v2)
			}
		}
		v = m

	case []interface{}:
		for i, v2 := range x {
			x[i] = ConvertMapI2MapS(v2)
		}

	case map[string]interface{}:
		for k, v2 := range x {
			x[k] = ConvertMapI2MapS(v2)
		}
	}

	return v
}

// ConvertYAMLtoJSON converts yaml bytes into json bytes
func ConvertYAMLtoJSON(yamlBytes []byte) ([]byte, error) {
	var body interface{}
	if err := yaml.Unmarshal(yamlBytes, &body); err != nil {
		return []byte{}, err
	}

	body = ConvertMapI2MapS(body)

	b, err := json.MarshalIndent(body, "", "    ")
	if err != nil {
		return []byte{}, err
	}

	return b, nil
}

// ConvertJSONtoYAML converts json bytes into yaml bytes
func ConvertJSONtoYAML(jsonBytes []byte) ([]byte, error) {
	var body interface{}
	if err := json.Unmarshal(jsonBytes, &body); err != nil {
		return []byte{}, err
	}

	body = ConvertMapI2MapS(body)

	b, err := yaml.Marshal(body)
	if err != nil {
		return []byte{}, err
	}

	return b, nil
}
