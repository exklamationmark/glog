package glog

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func defaultLoggingT() loggingT {
	l := loggingT{}
	l.stderrThreshold = errorLog
	l.setVState(0, nil, false)
	go l.flushDaemon() // this will leak loggingT structs, but only in test

	return l
}

func TestGetCurrentSettings(t *testing.T) {
	defer func() {
		logging = defaultLoggingT()
	}()

	r := httptest.NewRequest(http.MethodGet, "/debug/log", nil)
	w := httptest.NewRecorder()
	ShowSettings(w, r)
	w.Flush()

	resp := w.Result()
	dec := json.NewDecoder(resp.Body)
	defer resp.Body.Close()
	var body settings
	if want, got := http.StatusOK, resp.StatusCode; want != got {
		t.Errorf("diffrent response's status code; want= %v; got= %v", want, got)
	}
	if err := dec.Decode(&body); err != nil {
		t.Errorf("expected no error; got= %v", err)
	}
	if want, got := severityName[errorLog], body.ErrThreshold; want != got {
		t.Errorf("wrong err threshold; want= %v; got= %v", want, got)
	}
	if want, got := 0, body.Verbosity; want != got {
		t.Errorf("wrong verbosity level; want= %v; got= %v", want, got)
	}
}

func TestChangeSettings(t *testing.T) {
	defer func() {
		logging = defaultLoggingT()
	}()

	var testCases = []struct {
		testName string

		req               string
		expectedStatus    int
		expectedSeverity  severity
		expectedVerbosity Level
		expectedResponse  string
	}{
		{
			testName:          "normal request",
			req:               `{"stderrthreshold":"info","v":99}`,
			expectedStatus:    http.StatusOK,
			expectedSeverity:  infoLog,
			expectedVerbosity: Level(99),
			expectedResponse:  "",
		},
		{
			testName:          "invalid err threshold (missing)",
			req:               `{"v":99}`,
			expectedStatus:    http.StatusBadRequest,
			expectedSeverity:  errorLog,
			expectedVerbosity: 0,
			expectedResponse:  `{"error":"unknown error threshold: "}`,
		},
		{
			testName:          "invalid verbosity level (negative)",
			req:               `{"stderrthreshold":"info","v":-1}`,
			expectedStatus:    http.StatusBadRequest,
			expectedSeverity:  errorLog,
			expectedVerbosity: 0,
			expectedResponse:  `{"error":"invalid verbosity level: -1"}`,
		},
		{
			testName:          "invalid request body",
			req:               `{"}`,
			expectedStatus:    http.StatusBadRequest,
			expectedSeverity:  errorLog,
			expectedVerbosity: 0,
			expectedResponse:  `{"error":"cannot decode request's body: unexpected EOF"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			logging = defaultLoggingT()

			var buf bytes.Buffer
			buf.WriteString(tc.req)
			r := httptest.NewRequest(http.MethodPost, "/debug/log/settings", &buf)
			w := httptest.NewRecorder()
			ChangeSettings(w, r)
			w.Flush()

			resp := w.Result()
			defer resp.Body.Close()

			if want, got := tc.expectedStatus, resp.StatusCode; want != got {
				t.Errorf("diffrent response's status code; want= %v; got= %v", want, got)
			}
			if want, got := tc.expectedSeverity, logging.stderrThreshold.get(); want != got {
				t.Errorf("different error threshold; want= %v;  got= %v", want, got)
			}
			if want, got := tc.expectedVerbosity, logging.verbosity.get(); want != got {
				t.Errorf("different verbosity level; want= %v;  got= %v", want, got)
			}

			if len(tc.expectedResponse) > 0 {
				body, _ := ioutil.ReadAll(resp.Body)
				if want, got := tc.expectedResponse, string(body); want != got {
					t.Errorf("different error response;\n  want= %v\n   got= %v", want, got)
				}
			}
		})
	}
}
