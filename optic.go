// Package optic a generic net/http extension that makes exchanging data in go really fun
package optic

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	defaultBasePath = "/"
)

// HTTPError optic needs a http.StatusCode to properly send the response
// otherwise HTTPError can be any go struct
type HTTPError interface {
	GetCode() int
}

// Empty is useful when you want to Glance with no input struct
type Empty struct{}

// FromResponse decodes the net/http Response.Body into the desired struct
// exporting because FromResponse is useful on it's own in normal net/http handlers
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
	return fmt.Errorf("The response failed with status %s", res.Status)
}

func getHTTPErrorFromResponse[E any](res *http.Response) (E, error) {
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
