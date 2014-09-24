package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/mitchellh/go-homedir"
	"log"
	"net/http"
	"os"
)

func logHandler(path *string) {
	f, err := os.OpenFile(*path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		log.Fatalf("Error while log output: %v", err)
	}

	log.SetOutput(f)
}

type handlerError struct {
	Error   error
	Message string
	Code    int
}

// a custom type that we can use for handling errors and formatting responses
type handler func(w http.ResponseWriter, r *http.Request) (interface{}, *handlerError)

// attach the standard ServeHTTP method to our handler so the http library can call it
func (fn handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// here we could do some prep work before calling the handler if we wanted to

	// call the actual handler
	response, err := fn(w, r)

	// check for errors
	if err != nil {
		log.Printf("ERROR: %v\n", err.Error)
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Message), err.Code)
		return
	}
	if response == nil {
		log.Printf("ERROR: response from method is nil\n")
		http.Error(w, "Internal server error. Check the logs.", http.StatusInternalServerError)
		return
	}

	// turn the response into JSON
	bytes, e := json.Marshal(response)
	if e != nil {
		http.Error(w, "Error marshalling JSON", http.StatusInternalServerError)
		return
	}

	// send the response and log
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
	fmt.Println(r)
	log.Printf("%s %s %s %d", r.RemoteAddr, r.Method, r.URL, 200)
}

// Views handlers
func home(w http.ResponseWriter, r *http.Request) (interface{}, *handlerError) {
	return "ploplop", nil
}

// Main program
func main() {
	homeDir, _ := homedir.Dir()

	// flag arguments
	var port *int = flag.Int("p", 8080, "Serving port")
	var _ *string = flag.String("d", homeDir, "Videos root directory")
	var logpath *string = flag.String("l", "gogostream.log", "Logging path")

	// setting up logger
	logHandler(logpath)

	// setting up routes
	router := mux.NewRouter()
	router.Handle("/", handler(home)).Methods("GET")
	http.Handle("/", router)

	log.Printf("Streaming on port %d\n", *port)
	addr := fmt.Sprintf("127.0.0.1:%d", *port)

	err := http.ListenAndServe(addr, nil)
	log.Println(err.Error())
}
