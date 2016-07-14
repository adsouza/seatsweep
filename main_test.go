// Copyright 2016 Kevin Bowrin All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// If debug isn't set, a HTTP request should redirect to HTTPS.
func TestHTTPToHTTPSRedirectWhenNotDebug(t *testing.T) {

	tester := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "http://test.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	redirectHTTPS(tester, req)

	if tester.Code != http.StatusMovedPermanently {
		t.Error("HTTP request wasn't redirected to HTTPS.")
	}
	if tester.HeaderMap.Get("Strict-Transport-Security") == "" {
		t.Error("HSTS header isn't set.")
	}
	if !strings.HasPrefix(tester.HeaderMap.Get("Location"), "https") {
		t.Log(tester.HeaderMap)
		t.Error("Scheme is incorrect.")
	}

}

// See if setting an env var overrides an unset flag.
func TestEnvironmentVariableOverrideByFlag(t *testing.T) {
	os.Setenv(EnvPrefix+"ADDRESS", ":8080")
	overrideUnsetFlagsFromEnvironmentVariables()
	if *address != ":8080" {
		t.Error("Setting an environment variable did not override an unset flag.")
	}
}
