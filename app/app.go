// Copyright 2015 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

// Sample bookshelf is a fully-featured app demonstrating several Google Cloud APIs, including Datastore, Cloud SQL, Cloud Storage.
// See https://cloud.google.com/go/getting-started/tutorial-app
package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"google.golang.org/appengine"
)

var (
	// See template.go
	defaultTmpl    = parseTemplate("default.html")
	resultsTmpl    = parseTemplate("results.html")
	serviceAccount = os.Getenv("service_account")
	privateKey     = os.Getenv("private_key")
	automlModel    = os.Getenv("automl_model")
)

func main() {
	registerHandlers()
	appengine.Main()
}

func registerHandlers() {
	// Use gorilla/mux for rich routing.
	// See http://www.gorillatoolkit.org/pkg/mux
	r := mux.NewRouter()

	r.Methods("GET").Path("/").Handler(appHandler(homeHandler))
	r.Methods("GET").Path("/results").Handler(http.RedirectHandler("/", http.StatusFound))
	r.Methods("POST").Path("/results").Handler(appHandler(uploadHandler))

	// Respond to App Engine and Compute Engine health checks.
	// Indicate the server is healthy.
	r.Methods("GET").Path("/_ah/health").HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("ok"))
		})

	// [START request_logging]
	// Delegate all of the HTTP routing and serving to the gorilla/mux router.
	// Log all requests using the standard Apache format.
	http.Handle("/", handlers.CombinedLoggingHandler(os.Stderr, r))
	// [END request_logging]
}

// homeHandler displays the homescreen for the harness
func homeHandler(w http.ResponseWriter, r *http.Request) *appError {
	return defaultTmpl.Execute(w, r, nil)
}

type ResultsRequest struct {
	Filename      string
	BinaryContent []byte
}

func uploadHandler(w http.ResponseWriter, r *http.Request) *appError {
	//uploadHandler

	r.ParseMultipartForm(32 << 20)
	file, _, err := r.FormFile("imageToProcess")
	if err != nil {
		return appErrorf(err, "Could not read file upload")
	}
	defer file.Close()

	bs, err := ioutil.ReadAll(file)
	if err != nil {
		return appErrorf(err, "Could not read file upload to MS")
	}

	client := getClient(
		serviceAccount,
		privateKey)

	base64Image := base64.StdEncoding.EncodeToString(bs)
	afr := SendBinaryRequest(client, automlModel, base64Image)

	model := new(ResultModel)
	model.ImageContent = base64Image

	model.LabelAnnotations = afr.Responses[0].LabelAnnotations
	model.CustomLabelAnnotations = afr.Responses[0].CustomLabelAnnotations

	return resultsTmpl.Execute(w, r, model)
}

type ResultModel struct {
	LabelAnnotations []*LabelAnnotations `json:"labelAnnotations,omitempty"`

	CustomLabelAnnotations []*CustomLabelAnnotations `json:"customlabelAnnotations,omitempty"`

	ImageContent string
}

// http://blog.golang.org/error-handling-and-go
type appHandler func(http.ResponseWriter, *http.Request) *appError

type appError struct {
	Error   error
	Message string
	Code    int
}

func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if e := fn(w, r); e != nil { // e is *appError, not os.Error.
		log.Printf("Handler error: status code: %d, message: %s, underlying err: %#v",
			e.Code, e.Message, e.Error)

		http.Error(w, e.Message, e.Code)
	}
}

func appErrorf(err error, format string, v ...interface{}) *appError {
	return &appError{
		Error:   err,
		Message: fmt.Sprintf(format, v...),
		Code:    500,
	}
}
