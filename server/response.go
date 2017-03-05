package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

type response struct {
	code int
	head http.Header
	body []byte
}

func newResponse(res *http.Response) (r response, err error) {
	r = response{
		code: res.StatusCode,
		head: res.Header,
	}
	r.body, err = ioutil.ReadAll(res.Body)
	if err == nil {
		res.Body = ioutil.NopCloser(bytes.NewReader(r.body))
	}
	return r, err
}

func (r *response) Send(w http.ResponseWriter) {
	for key, value := range r.head {
		w.Header()[key] = value
	}
	w.WriteHeader(r.code)
	w.Write(r.body)
}
