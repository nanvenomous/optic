// this basic example shows how to set up optic routes and call them with the provided client
package main

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/nanvenomous/optic"
)

const (
	port           = "4040"
	host           = "127.0.0.1"
	userOpticRoute = "/api/optic/"
)

type solution struct {
	Answer int
}

type subtraction struct {
	First  int
	Second int
}

type division struct {
	Top    int
	Bottom int
}

type userHTTPError struct {
	Message string
	Code    int
}

// GetCode method returns the http.StatusCode for the HTTPError
// necessary to send the http code in a response
func (e *userHTTPError) GetCode() int {
	return e.Code
}

func subtract(recieved *subtraction, r *http.Request) (*solution, optic.HTTPError) {
	// Do something with a header (like check Authorization, or get a cookie)
	fmt.Println(r.Header.Get("Authorization"))

	return &solution{Answer: recieved.First - recieved.Second}, nil
}

func divide(recieved *division, _ *http.Request) (*solution, optic.HTTPError) {
	if recieved.Bottom == 0 { // return an error
		return nil, &userHTTPError{Code: http.StatusUnprocessableEntity, Message: "Impossible to divide by Zero"}
	}
	return &solution{Answer: recieved.Top / recieved.Bottom}, nil
}

func setupService() {
	var (
		err       error
		encodeErr = &userHTTPError{Code: http.StatusInternalServerError, Message: "Failed to encode your response."}
		decodeErr = &userHTTPError{Code: http.StatusNotAcceptable, Message: "Failed to decode your request body."}
		mux       *http.ServeMux
	)
	mux = http.NewServeMux()
	optic.SetupService(port, userOpticRoute, encodeErr, decodeErr, mux)

	// An optical mirror simply recieves information and sends information back
	optic.Mirror(subtract, "/RunSubtraction/")
	optic.Mirror(divide) // by default optic will use function name as route

	// Add other routes not handled by optic, as you would with any net/http service
	mux.HandleFunc("/health-check/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	// Add any net/http middleware
	optic.RegisterMiddleware(exampleMiddleware)

	// for this example, run the service in the background
	go func() {
		err = optic.Serve() // run the service
		if err != nil {
			panic(err)
		}
	}()
}

func main() {
	setupService()

	waitServiceUp() // only for this example

	optic.SetupClient(host, port, userOpticRoute, false)
	// Make requests
	var (
		err     error          // internal error
		httpErr *userHTTPError // service exception
		sln     solution       // output
	)
	httpErr, err = optic.Glance[userHTTPError]("/RunSubtraction/", &subtraction{First: 1, Second: 2}, &sln)
	fmt.Println(err, httpErr) // <nil> <nil>
	fmt.Println(sln.Answer)   // -1

	httpErr, err = optic.Glance[userHTTPError]("/divide/", &division{Top: 1, Bottom: 0}, &sln)
	fmt.Println(err, httpErr) // <nil> &{ Impossible to divide by Zero 422}
	fmt.Println(sln)          // <nil>

	//                                                     send                          receive
	httpErr, err = optic.Glance[userHTTPError]("/divide/", &division{Top: 4, Bottom: 2}, &sln)
	fmt.Println(err, httpErr) // <nil> <nil>
	fmt.Println(sln.Answer)   // 2
}

func exampleMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// add some middleware (like CORS for example)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		next.ServeHTTP(w, r)
	})
}

// this method is to wait for the service to initialize
func waitServiceUp() error {
	var (
		err error
		req *http.Request
	)

	req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s:%s/health-check/", host, port), bytes.NewBuffer([]byte{}))
	if err != nil {
		return err
	}
	for {
		time.Sleep(time.Millisecond * 50)
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		} else if res.StatusCode == http.StatusOK {
			break
		}
	}
	return nil
}
