package health

import (
	"errors"
	"net/http"
	"testing"
)

type fake struct{}

func (*fake) Close() error               { return nil }
func (*fake) Read(p []byte) (int, error) { return len(p), nil }

func TestHealth(t *testing.T) {
	var last, code int
	var msg string
	var err error

	// Helper overrides
	exit = func(i int) { last = i }
	write = func(m string) (int, error) { msg = m; return len(m), nil }
	get = func(string) (*http.Response, error) {
		return &http.Response{StatusCode: code, Status: "y", Body: &fake{}}, err
	}

	assert := func(exitCode, c int, h, m, meta string, e error) {
		err, code, *hc = e, c, h
		Check()
		if msg != m || last != exitCode {
			t.Errorf("Expected: (%v, %d)\nReceived: (%v, %d)\nTest: %s", m, exitCode, msg, last, meta)
		}
	}

	assert(0, 0, "", "", "Empty Flags", nil)
	assert(1, 0, "x", "err:bad", "Error Response", errors.New("bad"))
	assert(1, 418, "x", "status:y", "Bad Status", nil)
	assert(0, 200, "x", "OK", "Good", nil)
}
