package fetch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

var httpClient = &http.Client{}

type BasicAuth struct {
	Username string
	Password string
}

func (a BasicAuth) IsZero() bool {
	return a.Username == "" || a.Password == ""
}

type Fetch struct {
	method    string
	url       string
	Query     url.Values
	body      io.Reader
	Header    http.Header
	Errors    []error
	basicAuth BasicAuth
}

func New() *Fetch {
	return &Fetch{
		method: "GET",
		url:    "",
		Query:  url.Values{},
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

func (f *Fetch) SetParam(key, value string) *Fetch {
	f.Query.Set(key, value)

	return f
}

func (f *Fetch) SetParamMap(kv map[string]string) *Fetch {
	for k, v := range kv {
		f.Query.Set(k, v)
	}

	return f
}

func (f *Fetch) AddQueryParam(key, value string) *Fetch {
	f.Query.Add(key, value)
	return f
}

func (f *Fetch) SetQuery(q url.Values) *Fetch {
	f.Query = q

	return f
}

func (f *Fetch) SetBearerAuth(key string) *Fetch {
	f.Header.Set("Authorization", "Bearer "+key)

	return f
}

func (f *Fetch) SetBasicAuth(username, password string) *Fetch {
	f.basicAuth = BasicAuth{
		Username: strings.TrimSpace(username),
		Password: strings.TrimSpace(password),
	}

	return f
}

func (f *Fetch) AcceptLang(v string) *Fetch {
	f.Header.Set("Accept-Language", v)

	return f
}

func (f *Fetch) SetHeader(k, v string) *Fetch {
	f.Header.Set(k, v)

	return f
}

func (f *Fetch) SetHeaderMap(h map[string]string) *Fetch {
	for k, v := range h {
		f.Header.Set(k, v)
	}

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

	if len(f.Query) != 0 {
		f.url = fmt.Sprintf("%s?%s", f.url, f.Query.Encode())
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

func (f *Fetch) EndBytes() (*http.Response, []byte, []error) {
	resp, errs := f.End()
	if errs != nil {
		return resp, nil, f.Errors
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		f.Errors = append(f.Errors, err)
		return resp, nil, f.Errors
	}

	return resp, b, nil
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
