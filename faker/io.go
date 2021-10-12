package faker

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
)

func MustMarshalIndent(v interface{}) []byte {
	b, err := json.MarshalIndent(v, "", "\t")

	if err != nil {
		panic(err)
	}

	return b
}

func MustMarshalToReader(v interface{}) io.Reader {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	return bytes.NewReader(b)
}

func MustReadBody(body io.Reader) string {
	b, err := ioutil.ReadAll(body)
	if err != nil {
		panic(err)
	}

	return string(b)
}
