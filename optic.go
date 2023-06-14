package toss

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

type Exception struct {
	Internal string `json:"internal"`
	Message  string `json:"message"`
	Code     int    `json:"code"`
}

var (
	EMPTY_BYTE_ARRAY = bytes.NewBuffer([]byte{})
)

type Empty struct{}

func FromResponse[T any](res *http.Response) (*T, error) {
	var (
		err error
		tg  T
	)
	defer res.Body.Close()
	if res.StatusCode == http.StatusOK {
		err = json.NewDecoder(res.Body).Decode(&tg)
		if err != nil {
			return nil, err
		}
		return &tg, nil
	}
	return nil, errors.New(fmt.Sprintf("The response failed with status %s", res.Status))
}

func GetUrl(base string, elem ...string) (*url.URL, error) {
	url, err := url.Parse(base)
	if err != nil {
		return nil, err
	}
	extUrl := url.JoinPath(elem...)
	return extUrl, nil
}

func getExceptionFromResponse(res *http.Response) (*Exception, error) {
	var (
		err error
		exn Exception
	)
	defer res.Body.Close()
	err = json.NewDecoder(res.Body).Decode(&exn)
	if err != nil {
		return nil, err
	}
	exn.Code = res.StatusCode
	return &exn, nil
}
