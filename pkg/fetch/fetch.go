package fetch

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

var httpClient = &http.Client{}

type Fetch struct {
	method string
	url    string
	body   io.Reader
	Header http.Header
	Errors []error
}

func NewFetch() *Fetch {
	return &Fetch{
		body:   nil,
		Header: http.Header{},
	}
}

func (f *Fetch) Get(url string) *Fetch {
	f.method = "GET"
	f.url = url

	return f
}

func (f *Fetch) Post(url string) *Fetch {
	f.method = "POST"
	f.url = url

	return f
}

func (f *Fetch) Put(url string) *Fetch {
	f.method = "PUT"
	f.url = url

	return f
}

func (f *Fetch) SetAuth(key string) *Fetch {
	f.Header.Add("Authorization", "Bearer "+key)

	return f
}

func (f *Fetch) Send(body io.Reader) *Fetch {
	f.body = body
	return f
}

func (f *Fetch) SendJSON(v interface{}) *Fetch {
	d, err := json.Marshal(v)
	if err != nil {
		f.Errors = append(f.Errors, err)

		return f
	}

	f.Header.Add("Content-Type", ContentJSON)
	f.body = bytes.NewReader(d)

	return f
}

func (f *Fetch) End() (*http.Response, []error) {
	if f.Errors != nil {
		return nil, f.Errors
	}
	req, err := http.NewRequest(f.method, f.url, f.body)
	if err != nil {
		f.Errors = append(f.Errors, err)
		return nil, f.Errors
	}

	req.Header = f.Header

	resp, err := httpClient.Do(req)
	if err != nil {
		f.Errors = append(f.Errors, err)
		return nil, f.Errors
	}

	return resp, nil
}

func (f *Fetch) EndJSON(v interface{}) []error {
	resp, errs := f.End()
	if errs != nil {
		return f.Errors
	}

	dec := json.NewDecoder(resp.Body)

	err := dec.Decode(v)
	if err != nil {
		f.Errors = append(f.Errors, err)
		return f.Errors
	}

	return nil
}
