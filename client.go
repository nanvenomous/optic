package optic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

var (
	ServiceEndpoint string
	ClientUrl       *url.URL
)

func SetupClient(host, port, path string, secure bool) {
	schm := "http"
	if secure {
		schm = "https"
	}
	if port != "" {
		host = fmt.Sprintf("%s:%s", host, port)
	}
	ClientUrl = &url.URL{
		Scheme: schm,
		Host:   host,
		Path:   path,
	}
}

func Glance[E any](path string, send any, recieve any, headers ...http.Header) (*E, error) {
	var (
		err     error
		httpErr E
		data    []byte
		url     *url.URL
		req     *http.Request
		res     *http.Response
	)

	data, err = json.Marshal(send)
	if err != nil {
		return nil, err
	}

	url = ClientUrl.JoinPath(path)
	req, err = http.NewRequest(http.MethodPost, url.String(), bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	if len(headers) > 0 {
		req.Header = headers[0]
	}
	req.Header.Add("credentials", "include")

	res, err = http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		httpErr, err = getHttpErrorFromResponse[E](res)
		return &httpErr, err
	}

	err = FromResponse(res, recieve)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
