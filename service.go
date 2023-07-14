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

// Glass is the core internal component of a Mirror
// and optic.Glass recieves the struct R from the client
// it then does the processing server side to either return the desired S or a HTTPError
type Glass[S, R any] func(*R, *http.Request) (*S, HTTPError)

var (
	port                             = "4444"
	base                             = &url.URL{Path: defaultBasePath}
	mux                              *http.ServeMux
	allMiddleware                    = []func(http.Handler) http.Handler{}
	decodeHTTPError, encodeHTTPError HTTPError
)

// RegisterMiddleware optic will store all your middleware and add it before calling optic.Serve
// note you must call optic.Serve manuall and you must have passed a http.ServerMux on call to SetupService
func RegisterMiddleware(middleware func(http.Handler) http.Handler) {
	allMiddleware = append(allMiddleware, middleware)
}

// SetupService takes all the necessary data to register Mirrors and run an optic service
// you also need to provide the errors that will be sent over the network if serialization fails
// you can optionally pass a http.ServeMux
func SetupService(localPort, basePath string, encodeErr, decodeErr HTTPError, httpMux *http.ServeMux) {
	// TODO: more validation of input params (i.e. localPort number)
	// also if httpMux is optional we need to fail hard on registering middleware or remove that feature alltogether
	encodeHTTPError = encodeErr
	decodeHTTPError = decodeErr
	port = localPort
	if basePath != "" {
		base = &url.URL{Path: basePath}
	}
	cyan("BASE", base.Path)
	if httpMux != nil {
		mux = httpMux
	}
}

// Serve calls mux.ListenAndServe
// If you registered net/http middleware it will be applied first
// This is a convenience function you can serve manually with http.ListenAndServe or mux.ListenAndServe
func Serve() error {
	var (
		handler http.Handler
		server  *http.Server
	)
	handler = mux
	for _, mdwr := range allMiddleware {
		handler = mdwr(handler)
	}

	server = &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}

	serving(port)
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
		w.WriteHeader(encodeHTTPError.GetCode())
		err = encoder.Encode(encodeHTTPError)
		if err != nil {
			fmt.Println("[OPTIC] unable to send error body", err)
		}
		return
	}
	reflected(code, r.URL.Path)
}

// SendHTTPError exposed so you can manually send an optic HttpError in normal net/http HandleFunc
// note you will need to return from the handler after calling this method
func SendHTTPError(w http.ResponseWriter, r *http.Request, httpErr HTTPError) {
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

// Mirror registers a Glass handler at the desired path
// e.g. optic.Mirror(subtract, "/RunSubtraction/")
// once the handler is registered use optic.Glance to make the request to the Mirror
func Mirror[R, S any](glass Glass[S, R], paths ...string) {
	var (
		pth        string
		ul         *url.URL
		handleFunc func(http.ResponseWriter, *http.Request)
	)

	if len(paths) == 0 { // not passing a path will simply use the function name
		pth = getFunctionName(glass)
	} else {
		pth = paths[0]
	}
	violet("PATH", pth)
	ul = base.JoinPath(pth)

	handleFunc = func(w http.ResponseWriter, r *http.Request) {
		var (
			err     error
			rec     R
			send    *S
			httpErr HTTPError
			decoder *json.Decoder
		)

		decoder = json.NewDecoder(r.Body)
		err = decoder.Decode(&rec)
		if err != nil {
			SendHTTPError(w, r, decodeHTTPError)
			return
		}

		send, httpErr = glass(&rec, r)
		httpErrIsNil := reflect.ValueOf(httpErr).Kind() == reflect.Ptr && reflect.ValueOf(httpErr).IsNil()
		if !httpErrIsNil {
			SendHTTPError(w, r, httpErr)
			return
		}

		sendBytes(w, r, http.StatusOK, send)
	}

	if mux != nil {
		mux.HandleFunc(ul.Path, handleFunc)
	} else {
		http.HandleFunc(ul.Path, handleFunc)
	}
}
