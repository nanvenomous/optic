
# (o)ptic

A generic web extension to [net/http](https://pkg.go.dev/net/http)

Optic focuses on simplifying the data exchange portion of your service, primarily for cases from a go service to a go client (WASM app, cli, ...)

## Quick example

Define the entities
```go
type Division struct {
	Top    int
	Bottom int
}
type Solution struct {
	Answer int
}
```

Setup the service route
```go
func Divide(recieved *Division, r *http.Request) (send *Solution, exn *optic.Exception) {
	if recieved.Bottom == 0 { // return an error
		return nil, &optic.Exception{Code: http.StatusUnprocessableEntity, Message: "Impossible to divide by Zero"}
	}
	return &Solution{Answer: recieved.Top / recieved.Bottom}, nil
}

var (
    err error
    mux *http.ServeMux
)
mux = optic.SetupService(PORT, OPTIC_ROUTE)
// An optic mirror simply recieves information and sends information back
optic.Mirror(Divide) // by default optic will use function name as route
// run the service
err = optic.Serve()
if err != nil {
    panic(err)
}
```

Setup the client and make a request
```go
var (
    err error            // internal error
    exn *optic.Exception // service exception
    sln *Solution        // output
)
optic.SetupClient(HOST, PORT, OPTIC_ROUTE, false)
//                            Put in  , Get out                    Route
sln, exn, err = optic.Reflect[Division, Solution]("Bearer Token", "/Divide/", &Division{Top: 4, Bottom: 2})
fmt.Println(err, exn)   // <nil> <nil>
fmt.Println(sln.Answer) // 2
```


Optic is drop in compatible with [net/http](https://pkg.go.dev/net/http)

It returns a `*http.ServerMux` for middleware or other routes
```go
var (
    err error
    mux *http.ServeMux
)
mux = optic.SetupService(PORT, OPTIC_ROUTE)
// Add other routes not handled by optic, as you would with any net/http service
mux.HandleFunc("/health-check/", func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
})

func exampleMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// add some middleware (like CORS for example)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		next.ServeHTTP(w, r)
	})
}

// optic can register middleware for you
optic.RegisterMiddleware(exampleMiddleware)
// or you can do it yourself
var (
    handler http.Handler
)
handler = exampleMiddleware(mux)
```

For the full example in code see [./examples/main.go](https://github.com/nanvenomous/optic/blob/mainline/example/main.go) 

Run the example like so:
![run example](.rsrc/run-example.gif)
