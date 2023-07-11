package optic

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

const (
	default_base_path = "/"
)

type HttpError interface {
	GetCode() int
}

// Empty is useful when you want to Glance with no input
type Empty struct{}

func FromResponse(res *http.Response, recieved any) error {
	var (
		err error
	)
	defer res.Body.Close()
	if res.StatusCode == http.StatusOK {
		err = json.NewDecoder(res.Body).Decode(recieved)
		if err != nil {
			return err
		}
		return nil
	}
	return errors.New(fmt.Sprintf("The response failed with status %s", res.Status))
}

func getHttpErrorFromResponse[E any](res *http.Response) (E, error) {
	var (
		err     error
		httpErr E
	)
	defer res.Body.Close()
	err = json.NewDecoder(res.Body).Decode(&httpErr)
	if err != nil {
		return httpErr, err
	}
	return httpErr, nil
}
