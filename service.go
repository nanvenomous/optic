package optic

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"runtime"
	"strings"
)

var (
	ErrorReadingResponseBody      = "There was an error reading the request data. Are you sure you sent a valid request body?"
	ErrorUnmarshalingResponseBody = "There was an error unmarshaling the request data into the proper format."
)

type Eye[R, S any] func(*R, *http.Request) (*S, *Exception)

var (
	bytesToSend []byte
	Port        = "4444"
	Base        = &url.URL{Path: DEFAULT_BASE_PATH}
)

func SetupService(port, base string) {
	Port = port
	if base != "" {
		Base = &url.URL{Path: base}
		Base = Base.JoinPath(DEFAULT_BASE_PATH)
	}
}

func Serve() error {
	fmt.Println("Serving on port: ", Port)
	return http.ListenAndServe(":"+Port, nil)
}

func sendBytes[S any](w http.ResponseWriter, code int, send *S) {
	var (
		err error
	)
	bytesToSend, err = json.Marshal(send)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(code)
	w.Write(bytesToSend)
}

func SendException(w http.ResponseWriter, exn *Exception) {
	sendBytes(w, exn.Code, exn)
}

func GetFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func Mirror[R, S any](eye Eye[R, S], paths ...string) {
	var (
		err  error
		body []byte
		rec  R
		send *S
		exn  *Exception
		pth  string
		ul   *url.URL
	)

	if len(paths) == 0 { // not passing a path will simply use the function name
		pth = GetFunctionName(eye)
		pths := strings.Split(pth, ".")
		pth = "/" + pths[len(pths)-1] + "/"
	} else {
		pth = paths[0]
	}
	fmt.Println("[SERVICE ENDPOINT] ", pth)
	ul = Base.JoinPath(pth)
	http.HandleFunc(ul.Path, func(w http.ResponseWriter, r *http.Request) {
		body, err = ioutil.ReadAll(r.Body)
		if err != nil {
			SendException(w, &Exception{Code: http.StatusBadRequest, Message: ErrorReadingResponseBody, Internal: err.Error()})
			return
		}

		err = json.Unmarshal(body, &rec)
		if err != nil {
			SendException(w, &Exception{Code: http.StatusNotAcceptable, Message: ErrorUnmarshalingResponseBody, Internal: err.Error()})
			return
		}

		send, exn = eye(&rec, r)
		if exn != nil {
			SendException(w, exn)
			return
		}

		sendBytes(w, http.StatusOK, send)
	})
}

// type HeaderMethod[S any] func(w http.ResponseWriter, r *http.Request) (*S, *Exception)

type EmitterLogic[S any] func(*http.Request) (*S, *Exception)

func Emitter[S any](path string, el EmitterLogic[S]) {
	var (
		send *S
		exn  *Exception
	)

	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		send, exn = el(r)
		if exn != nil {
			SendException(w, exn)
			return
		}
		sendBytes(w, http.StatusOK, send)
	})
}
