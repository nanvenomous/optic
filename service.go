package optic

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

var (
	ErrorReadingResponseBody      = "There was an error reading the request data. Are you sure you sent a valid request body?"
	ErrorUnmarshalingResponseBody = "There was an error unmarshaling the request data into the proper format."
)

type Eye[R, S any] func(recieve *R) (*S, *Exception)

var (
	bytesToSend []byte
)

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

func sendException(w http.ResponseWriter, exn *Exception) {
	sendBytes(w, exn.Code, exn)
}

// Emit optic

func Reflect[R, S any](path string, eye Eye[R, S]) {
	var (
		err  error
		body []byte
		rec  R
		send *S
		exn  *Exception
	)

	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		body, err = ioutil.ReadAll(r.Body)
		if err != nil {
			sendException(w, &Exception{Code: http.StatusBadRequest, Message: ErrorReadingResponseBody, Internal: err.Error()})
			return
		}

		err = json.Unmarshal(body, &rec)
		if err != nil {
			sendException(w, &Exception{Code: http.StatusNotAcceptable, Message: ErrorUnmarshalingResponseBody, Internal: err.Error()})
			return
		}

		send, exn = eye(&rec)
		if exn != nil {
			sendException(w, exn)
			return
		}

		sendBytes(w, http.StatusOK, send)
	})
}
