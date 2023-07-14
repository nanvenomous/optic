//go:build js
// +build js

package optic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	fetch "marwan.io/wasm-fetch"
)

var (
	clientURL *url.URL
)

// SetupClient takes all parameters to setup the net/http client
// prereq to calling optic.Glance
func SetupClient(host, port, path string, secure bool) {
	schm := "http"
	if secure {
		schm = "https"
	}
	if port != "" {
		host = fmt.Sprintf("%s:%s", host, port)
	}
	clientURL = &url.URL{
		Scheme: schm,
		Host:   host,
		Path:   path,
	}
}

// Glance makes a request to the Mirror at the same path (Glance in the Mirror)
// you send a go struct and you recieve a go struct back
// the type param is the error struct you expect to receive
func Glance[E any](path string, send any, recieve any, headers ...http.Header) (*E, error) {
	var (
		err     error
		httpErr E
		data    []byte
		url     *url.URL
		res     *fetch.Response
	)

	data, err = json.Marshal(send)
	if err != nil {
		return nil, err
	}

	url = clientURL.JoinPath(path)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*6)
	defer cancel()
	res, err = fetch.Fetch(url.String(), &fetch.Opts{
		Body:   bytes.NewBuffer(data),
		Method: fetch.MethodPost,
		Signal: ctx,
	})

	if res.Status != http.StatusOK {
		httpErr, err = getHTTPErrorFromFetchResponse[E](res)
		return &httpErr, err
	}

	err = fromFetchResponse(res, recieve)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func getHTTPErrorFromFetchResponse[E any](res *fetch.Response) (E, error) {
	var (
		err     error
		httpErr E
	)
	// defer res.Body.Close()
	err = json.NewDecoder(bytes.NewBuffer(res.Body)).Decode(&httpErr)
	if err != nil {
		return httpErr, err
	}
	return httpErr, nil
}

func fromFetchResponse(res *fetch.Response, recieved any) error {
	var (
		err error
	)
	// defer res.Body.Close()
	if res.Status == http.StatusOK {
		err = json.NewDecoder(bytes.NewBuffer(res.Body)).Decode(recieved)
		if err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("The response failed with status %s", res.Status)
}
