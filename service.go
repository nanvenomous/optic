package optic

import (
	"encoding/json"
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
	Mux         *http.ServeMux
	Middleware  = []func(http.Handler) http.Handler{}
)

func RegisterMiddleware(middleware func(http.Handler) http.Handler) {
	Middleware = append(Middleware, middleware)
}

func SetupService(port, base string) *http.ServeMux {
	Port = port
	if base != "" {
		Base = &url.URL{Path: base}
	}
	cyan("BASE", Base.Path)
	Mux = http.NewServeMux()
	return Mux
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
	// fmt.Println("Serving optic ", optic(), " on port: ", Port)
	return server.ListenAndServe()
}

func sendBytes[S any](w http.ResponseWriter, r *http.Request, code int, send *S) {
	var (
		err error
	)
	bytesToSend, err = json.Marshal(send)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	reflected(code, r.URL.Path)
	w.WriteHeader(code)
	w.Write(bytesToSend)
}

func SendException(w http.ResponseWriter, r *http.Request, exn *Exception) {
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
		pth = getFunctionName(eye)
	} else {
		pth = paths[0]
	}
	violet("PATH", pth)
	ul = Base.JoinPath(pth)
	Mux.HandleFunc(ul.Path, func(w http.ResponseWriter, r *http.Request) {
		body, err = ioutil.ReadAll(r.Body)
		if err != nil {
			SendException(w, r, &Exception{Code: http.StatusBadRequest, Message: ErrorReadingResponseBody, Internal: err.Error()})
			return
		}

		err = json.Unmarshal(body, &rec)
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
	})
}
