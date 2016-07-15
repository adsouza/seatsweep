// Copyright 2016 Kevin Bowrin All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/oxtoacart/bpool"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	// EnvPrefix is the prefix for the environment variables.
	EnvPrefix string = "SEATSWEEP_"
)

var (

	// Command Line Flags for the Server
	address   = flag.String("address", ":8877", "Address for the server to bind on.")
	debug     = flag.Bool("debug", false, "Run the server in debug/development mode.")
	staticDir = flag.String("staticdir", "static", "Directory where the static files are stored. "+
		"If not absolute, will be joined to the current working directory.")
	templatesDir = flag.String("templatesdir", "templates", "Directory where the template files are stored. "+
		"If not absolute, will be joined to the current working directory.")

	// A shared BufferPool for executing templates.
	// This is used to catch errors before sending the result to the client.
	bufpool *bpool.SizedBufferPool

	// The html templates we will use when rending web pages for the client.
	htmlTemplates map[string]*template.Template
)

func init() {
	flag.Usage = func() {
		fmt.Println("SeatSweep\nVersion 0.0.3")
		flag.PrintDefaults()
		fmt.Println("  The possible environment variables:")
		flag.VisitAll(func(f *flag.Flag) {
			fmt.Printf("  %v%v\n", EnvPrefix, strings.ToUpper(f.Name))
		})
	}

	bufpool = bpool.NewSizedBufferPool(256, 2000)
	htmlTemplates = make(map[string]*template.Template)
}

func redirectHTTPSUsingXForwardedHost(w http.ResponseWriter, r *http.Request) {
	redirectURL := *r.URL
	redirectURL.Scheme = "https"
	redirectURL.Host = r.Header.Get("X-Forwarded-Host")

	w.Header().Set("Strict-Transport-Security", "max-age=31536000; preload")

	http.Redirect(w, r, redirectURL.String(), http.StatusMovedPermanently)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	executeTemplateOrError(w, "home.html")
}

func mapHandler(w http.ResponseWriter, r *http.Request) {
	executeTemplateOrError(w, "map.html")
}

func executeTemplateOrError(w http.ResponseWriter, name string) {

    homeTemplate, ok := htmlTemplates[name]
	if !ok {
		errorText := fmt.Sprintf("Template Error - Can't find template %v", name)
		http.Error(w, errorText, http.StatusInternalServerError)
		return
	}

	buf := bufpool.Get()
	defer bufpool.Put(buf)

	err := homeTemplate.Execute(buf, nil)
	if err != nil {
		errorText := fmt.Sprintf("Template Error - Can't execute template %v", name)
		http.Error(w, errorText, http.StatusInternalServerError)
		return
	}

	// Set the header and write the buffer to the http.ResponseWriter
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, err = buf.WriteTo(w)
	if err != nil {
		log.Printf("Error writing to ResponseWriter: %v", err)
	}
}

func main() {
	// Process the flags.
	flag.Parse()

	// If any flags have not been set, see if there are
	// environment variables that set them.
	overrideUnsetFlagsFromEnvironmentVariables()

	// Find the static and templates directories
	staticDirAbs, err := filepath.Abs(*staticDir)
	if err != nil {
		log.Fatalf("Error building absolute path to static directory: %v", err)
	}
	templateDirAbs, err := filepath.Abs(*templatesDir)
	if err != nil {
		log.Fatalf("Error building absolute path to tempalte directory: %v", err)
	}

	log.Printf("Using static directory: %v\n", staticDirAbs)
	log.Printf("Using template directory: %v\n", templateDirAbs)

	err = processTemplates(templateDirAbs)
	if err != nil {
		log.Fatalf("Template processing error: %v\n", err)
	}

	r := mux.NewRouter().StrictSlash(true)

	// Redirect to HTTPS if not in debug mode
	if !*debug {
		r.Headers("X-Forwarded-Proto", "http").MatcherFunc(func(r *http.Request, rm *mux.RouteMatch) bool {
			return r.Header.Get("X-Forwarded-Host") != ""
		}).HandlerFunc(redirectHTTPSUsingXForwardedHost)
	}

	r.HandleFunc("/", homeHandler)
	r.HandleFunc("/map", mapHandler)
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(staticDirAbs))))

	log.Fatalf("FATAL: %v", http.ListenAndServe(*address, r))
}

// If any flags are not set, use environment variables to set them.
func overrideUnsetFlagsFromEnvironmentVariables() {

	// A map of pointers to unset flags.
	listOfUnsetFlags := make(map[*flag.Flag]bool)

	// flag.Visit calls a function on "only those flags that have been set."
	// flag.VisitAll calls a function on "all flags, even those not set."
	// No way to ask for "only unset flags". So, we add all, then
	// delete the set flags.

	// First, visit all the flags, and add them to our map.
	flag.VisitAll(func(f *flag.Flag) { listOfUnsetFlags[f] = true })

	// Then delete the set flags.
	flag.Visit(func(f *flag.Flag) { delete(listOfUnsetFlags, f) })

	// Loop through our list of unset flags.
	// We don't care about the values in our map, only the keys.
	for k := range listOfUnsetFlags {

		// Build the corresponding environment variable name for each flag.
		uppercaseName := strings.ToUpper(k.Name)
		environmentVariableName := fmt.Sprintf("%v%v", EnvPrefix, uppercaseName)

		// Look for the environment variable name.
		// If found, set the flag to that value.
		// If there's a problem setting the flag value,
		// there's a serious problem we can't recover from.
		environmentVariableValue := os.Getenv(environmentVariableName)
		if environmentVariableValue != "" {
			err := k.Value.Set(environmentVariableValue)
			if err != nil {
				log.Fatalf("FATAL: Unable to set configuration option %v from environment variable %v, "+
					"which has a value of \"%v\"",
					k.Name, environmentVariableName, environmentVariableValue)
			}
		}
	}
}

// Process the template files so they are ready to access and execute by handlers. 
func processTemplates(absTemplateDir string) (err error) {

	templateFiles, err := filepath.Glob(filepath.Join(absTemplateDir, "*.html"))
	if err != nil {
		return err
	}

	// The base.html file needs to be present.
	baseTemplateIndex := -1
	baseTemplateAbs := ""
	for i, templateFile := range templateFiles {
		if filepath.Base(templateFile) == "base.html" {
			baseTemplateIndex = i
			baseTemplateAbs = templateFile
			break
		}
	}
	if baseTemplateIndex == -1 {
		return errors.New("Can't find base.html")
	}

	// Delete the base template from the slice
	templateFiles = append(templateFiles[:baseTemplateIndex], templateFiles[baseTemplateIndex+1:]...)

	for _, templateFile := range templateFiles {
		htmlTemplates[filepath.Base(templateFile)] = template.Must(template.ParseFiles(baseTemplateAbs, templateFile))
	}

	return nil
}
