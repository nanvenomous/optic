package optic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
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
	path = filepath.Join(path, DEFAULT_BASE_PATH)
	ClientUrl = &url.URL{
		Scheme: schm,
		Host:   host,
		Path:   path,
	}
}

func Reflect[S, R any](tkn string, path string, send *S) (*R, *Exception, error) {
	var (
		err  error
		exn  *Exception
		data []byte
		url  *url.URL
		req  *http.Request
		res  *http.Response
		recd *R
	)

	data, err = json.Marshal(send)
	if err != nil {
		return nil, nil, err
	}

	url = ClientUrl.JoinPath(path)
	req, err = http.NewRequest(http.MethodPost, url.String(), bytes.NewBuffer(data))
	if err != nil {
		return nil, nil, err
	}
	if tkn != "" {
		req.Header.Add("Authorization", tkn)
	}

	res, err = http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, err
	}

	if res.StatusCode != http.StatusOK {
		exn, err = getExceptionFromResponse(res)
		return nil, exn, err
	}

	recd, err = FromResponse[R](res)
	if err != nil {
		return nil, nil, err
	}
	return recd, nil, nil
}
