// Copyright 2016 Kevin Bowrin All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
    "fmt"
    "log"
    "net/http"
    "os"
    "flag"
    "strings"
    "github.com/gorilla/mux"
    "path/filepath"
)

const (
	// EnvPrefix is the prefix for the environment variables.
	EnvPrefix string = "SEATSWEEP_"
)

var (
	address = flag.String("address", ":8877", "Address for the server to bind on.")
	debug = flag.Bool("debug", false, "Run the server in debug/development mode.")
	staticDir = flag.String("staticdir", "static", "Directory where the static files are stored." +
	                                               "If not absolute, will be joined to the current working directory.")
)

func init() {
	flag.Usage = func() {
		fmt.Println("SeatSweep\nVersion 0.0.2")
		flag.PrintDefaults()
		fmt.Println("  The possible environment variables:")
		flag.VisitAll(func(f *flag.Flag) {
			fmt.Printf("  %v%v\n", EnvPrefix, strings.ToUpper(f.Name))
		})
	}
}

func redirectHTTPS(w http.ResponseWriter, r *http.Request) {

	redirectURL := *r.URL
	redirectURL.Scheme = "https"
	redirectURL.Host = r.Header.Get("X-Forwarded-Host")	

	w.Header().Set("Strict-Transport-Security", "max-age=31536000; preload")

    http.Redirect(w, r, redirectURL.String(), http.StatusMovedPermanently)    
}

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hi there!")
}

func main() {  

	// Process the flags.
	flag.Parse()

	// If any flags have not been set, see if there are
	// environment variables that set them.
	overrideUnsetFlagsFromEnvironmentVariables()

	r := mux.NewRouter().StrictSlash(true)

	// Redirect to HTTPS if not in debug mode
	if ! *debug {
       	r.Headers("X-Forwarded-Proto", "http").MatcherFunc(func(r *http.Request, rm *mux.RouteMatch) bool {
            return r.Header.Get("X-Forwarded-Host") != ""
        }).HandlerFunc(redirectHTTPS)
	}

    r.HandleFunc("/", handler)

    absStaticDir, err := filepath.Abs(*staticDir)
    if err != nil {
    	log.Fatalf("Error building absolute path to static directory:\n%v", err)
    }    

    r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(absStaticDir))))

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

