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

var (
	ErrorReadingResponseBody      = "There was an error reading the request data. Are you sure you sent a valid request body?"
	ErrorUnmarshalingResponseBody = "There was an error unmarshaling the request data into the proper format."
)

type Eye[S, R any] func(*R, *http.Request) (*S, *Exception)

var (
	Port       = "4444"
	Base       = &url.URL{Path: DEFAULT_BASE_PATH}
	Mux        *http.ServeMux
	Middleware = []func(http.Handler) http.Handler{}
)

func RegisterMiddleware(middleware func(http.Handler) http.Handler) {
	Middleware = append(Middleware, middleware)
}

func SetupService(port, base string, mux *http.ServeMux) {
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
	for _, mdwr := range Middleware {
		handler = mdwr(handler)
	}

	server = &http.Server{
		Addr:    ":" + Port,
		Handler: handler,
	}

	serving(Port)
	return server.ListenAndServe()
}

func sendBytes[S any](w http.ResponseWriter, r *http.Request, code int, send *S) {
	var (
		err     error
		encoder *json.Encoder
	)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	encoder = json.NewEncoder(w)
	err = encoder.Encode(send)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	reflected(code, r.URL.Path)
}

func SendException(w http.ResponseWriter, r *http.Request, exn *Exception) {
	fmt.Println("[MESSAGE] ", exn.Message)
	fmt.Println("[INTERNAL] ", exn.Internal)
	sendBytes(w, r, exn.Code, exn)
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
			exn     *Exception
			decoder *json.Decoder
		)

		decoder = json.NewDecoder(r.Body)
		err = decoder.Decode(&rec)
		if err != nil {
			SendException(w, r, &Exception{Code: http.StatusNotAcceptable, Message: ErrorUnmarshalingResponseBody, Internal: err.Error()})
			return
		}

		send, exn = eye(&rec, r)
		if exn != nil {
			SendException(w, r, exn)
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
