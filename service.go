package optic

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"runtime"
	"strings"
)

type Eye[S, R any] func(*R, *http.Request) (*S, HttpError)

var (
	Port                             = "4444"
	Base                             = &url.URL{Path: default_base_path}
	Mux                              *http.ServeMux
	allMiddleware                    = []func(http.Handler) http.Handler{}
	decodeHttpError, encodeHttpError HttpError
)

func RegisterMiddleware(middleware func(http.Handler) http.Handler) {
	allMiddleware = append(allMiddleware, middleware)
}

func SetupService(port, base string, encodeErr, decodeErr HttpError, mux *http.ServeMux) {
	// TODO: more validation of input params (i.e. port number)
	// also if mux is optional we need to fail hard on registering middleware or remove that feature alltogether
	encodeHttpError = encodeErr
	decodeHttpError = decodeErr
	Port = port
	if base != "" {
		Base = &url.URL{Path: base}
	}
	cyan("BASE", Base.Path)
	if mux != nil {
		Mux = mux
	}
}

func Serve() error {
	var (
		handler http.Handler
		server  *http.Server
	)
	handler = Mux
	for _, mdwr := range allMiddleware {
		handler = mdwr(handler)
	}

	server = &http.Server{
		Addr:    ":" + Port,
		Handler: handler,
	}

	serving(Port)
	return server.ListenAndServe()
}

func sendBytes[S any](w http.ResponseWriter, r *http.Request, code int, send S) {
	var (
		err     error
		encoder *json.Encoder
	)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	encoder = json.NewEncoder(w)
	err = encoder.Encode(send)
	if err != nil {
		w.WriteHeader(encodeHttpError.GetCode())
		err = encoder.Encode(encodeHttpError)
		if err != nil {
			fmt.Println("[OPTIC] unable to send error body", err)
		}
		return
	}
	reflected(code, r.URL.Path)
}

// SendHttpError exposed so you can manually send an optic HttpError in normal net/http HandleFunc
// note you will need to return from the handler after calling this method
func SendHttpError(w http.ResponseWriter, r *http.Request, httpErr HttpError) {
	sendBytes(w, r, httpErr.GetCode(), httpErr)
}

func getFunctionName(i interface{}) string {
	var (
		pth      string
		splitPth []string
	)
	pth = runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
	splitPth = strings.Split(pth, ".")
	pth = "/" + splitPth[len(splitPth)-1] + "/"
	return pth
}

func Mirror[R, S any](eye Eye[S, R], paths ...string) {
	var (
		pth        string
		ul         *url.URL
		handleFunc func(http.ResponseWriter, *http.Request)
	)

	if len(paths) == 0 { // not passing a path will simply use the function name
		pth = getFunctionName(eye)
	} else {
		pth = paths[0]
	}
	violet("PATH", pth)
	ul = Base.JoinPath(pth)

	handleFunc = func(w http.ResponseWriter, r *http.Request) {
		var (
			err     error
			rec     R
			send    *S
			httpErr HttpError
			decoder *json.Decoder
		)

		decoder = json.NewDecoder(r.Body)
		err = decoder.Decode(&rec)
		if err != nil {
			SendHttpError(w, r, decodeHttpError)
			return
		}

		send, httpErr = eye(&rec, r)
		if httpErr != nil {
			SendHttpError(w, r, httpErr)
			return
		}

		sendBytes(w, r, http.StatusOK, send)
	}

	if Mux != nil {
		Mux.HandleFunc(ul.Path, handleFunc)
	} else {
		http.HandleFunc(ul.Path, handleFunc)
	}
}
