package main

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/nanvenomous/optic"
)

const (
	PORT        = "4040"
	HOST        = "127.0.0.1"
	OPTIC_ROUTE = "/api/optic/"
)

type Solution struct {
	Answer int
}

type Subtraction struct {
	First  int
	Second int
}

type Division struct {
	Top    int
	Bottom int
}

func Subtract(recieved *Subtraction, r *http.Request) (*Solution, *optic.Exception) {
	// Do something with a header (like check Authorization)
	fmt.Println(r.Header.Get("Authorization"))

	return &Solution{Answer: recieved.First - recieved.Second}, nil
}

func Divide(recieved *Division, r *http.Request) (*Solution, *optic.Exception) {
	if recieved.Bottom == 0 { // return an error
		return nil, &optic.Exception{Code: http.StatusUnprocessableEntity, Message: "Impossible to divide by Zero"}
	}
	return &Solution{Answer: recieved.Top / recieved.Bottom}, nil
}

func setupService() {
	var (
		err error
		mux *http.ServeMux
	)
	mux = http.NewServeMux()
	optic.SetupService(PORT, OPTIC_ROUTE, mux)

	// An optical mirror simply recieves information and sends information back
	optic.Mirror(Subtract, "/RunSubtraction/")
	optic.Mirror(Divide) // by default optic will use function name as route

	// Add other routes not handled by optic, as you would with any net/http service
	mux.HandleFunc("/health-check/", func(w http.ResponseWriter, r *http.Request) {
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

	optic.SetupClient(HOST, PORT, OPTIC_ROUTE, false)
	// Make requests
	var (
		err error            // internal error
		exn *optic.Exception // service exception
		sln Solution         // output
	)
	exn, err = optic.Glance("/RunSubtraction/", &Subtraction{First: 1, Second: 2}, &sln)
	fmt.Println(err, exn)   // <nil> <nil>
	fmt.Println(sln.Answer) // -1

	exn, err = optic.Glance("/Divide/", &Division{Top: 1, Bottom: 0}, &sln)
	fmt.Println(err, exn) // <nil> &{ Impossible to divide by Zero 422}
	fmt.Println(sln)      // <nil>

	//                                  send                          receive
	exn, err = optic.Glance("/Divide/", &Division{Top: 4, Bottom: 2}, &sln)
	fmt.Println(err, exn)   // <nil> <nil>
	fmt.Println(sln.Answer) // 2
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

	req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s:%s/health-check/", HOST, PORT), bytes.NewBuffer([]byte{}))
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
