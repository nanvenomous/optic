package toss

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

	// return func(c *gin.Context) {
	// 	var (
	// 		err    error
	// 		exn    *toss.Exception
	// 		authId string
	// 		req    R
	// 		send   *S
	// 	)
	// 	authId, err = keycloak.UserFromToken(database.CTX, c.GetHeader("Authorization"))
	// 	// authId, err = security.ValidateTokenFromRequestHeader(c)
	// 	if err != nil {
	// 		exception.Handler(c, err, exception.BadTokenRequestError)
	// 		return
	// 	}

	// 	if err = c.ShouldBindJSON(&req); err != nil {
	// 		exception.Handler(c, err, exception.UnprocessableRequestEntityError)
	// 		return
	// 	}
	// 	send, exn = logic(authId, &req)
	// 	if exn != nil {
	// 		exception.Handler(c, exn.Internal, exn.Message)
	// 		return
	// 	}

	// 	c.JSON(http.StatusOK, send)
	// }
}
