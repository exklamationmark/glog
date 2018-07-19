package glog

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"

	"github.com/pkg/errors"
)

func init() {
	// register default HTTP handler (same place as pprof)
	http.HandleFunc("/debug/log/settings", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			ShowSettings(w, r)
		case http.MethodPost:
			ChangeSettings(w, r)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
}

// settings define the schema of API's request & response body.
type settings struct {
	ErrThreshold string `json:"stderrthreshold"`
	Verbosity    int    `json:"v"`
}

// ShowSettings prints the current logger's stderrThreshold and verbosity.
func ShowSettings(w http.ResponseWriter, r *http.Request) {
	enc := json.NewEncoder(w)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	enc.Encode(settings{
		ErrThreshold: severityName[logging.stderrThreshold],
		Verbosity:    int(logging.verbosity),
	})
}

// ChangeSettings updates both stderrThreshold and verbosity.
func ChangeSettings(w http.ResponseWriter, r *http.Request) {
	dec := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var req settings
	if err := dec.Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"error":"%v"}`, errors.Wrapf(err, "cannot decode request's body"))
		return
	}

	severity, ok := severityByName(req.ErrThreshold)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"error":"%#v"}`, errors.Errorf("unknown error threshold: %s", req.ErrThreshold))
		return
	}
	if req.Verbosity < 0 || req.Verbosity > math.MaxInt32 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"error":"%#v"}`, errors.Errorf("invalid verbosity level: %d", req.Verbosity))
		return
	}

	logging.stderrThreshold.set(severity)
	logging.setVState(Level(req.Verbosity), logging.vmodule.filter, false)
}
