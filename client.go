package optic

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
)

var ServiceEndpoint string

func Post[S, R any](tkn string, path string, send *S) (*R, *Exception, error) {
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

	url, err = GetUrl(ServiceEndpoint, path)
	if err != nil {
		return nil, nil, err
	}

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

func Get[S, R any](tkn string, path string, send *S, params ...url.Values) (*R, *Exception, error) {
	var (
		err    error
		exn    *Exception
		url    *url.URL
		data   []byte
		tosend = EMPTY_BYTE_ARRAY
		req    *http.Request
		res    *http.Response
		recd   *R
	)

	url, err = GetUrl(ServiceEndpoint, path)
	if err != nil {
		return nil, nil, err
	}

	if send != nil {
		data, err = json.Marshal(send)
		if err != nil {
			return nil, nil, err
		}
		tosend = bytes.NewBuffer(data)
	}

	if len(params) != 0 {
		url.RawQuery = params[0].Encode()
	}

	req, err = http.NewRequest(http.MethodGet, url.String(), tosend)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Add("Authorization", tkn)

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
