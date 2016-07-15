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
	req.Header.Add("X-Forwarded-Host", "test2.com")
	if err != nil {
		t.Fatal(err)
	}

	redirectHTTPSUsingXForwardedHost(tester, req)

	if tester.Code != http.StatusMovedPermanently {
		t.Error("HTTP request wasn't redirected to HTTPS.")
	}
	if tester.Header().Get("Strict-Transport-Security") == "" {
		t.Error("HSTS header isn't set.")
	}
	if tester.Header().Get("Location") != "https://test2.com" {
		t.Error("X-Forwarded-Host header wasn't used.")
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
