package controller

import (
	"io"
	"io/ioutil"
	"strings"

	"github.com/tidwall/gjson"
)

// GetJSONString get a string field from http request body.
// Return empty string even if the passed in data does not contain the required key.
func GetJSONString(data io.ReadCloser, path string) (string, error) {
	b, err := ioutil.ReadAll(data)
	defer data.Close()

	if err != nil {
		return "", err
	}

	result := gjson.GetBytes(b, path)

	if !result.Exists() {
		return "", nil
	}

	value := strings.TrimSpace(result.String())

	return value, nil
}
