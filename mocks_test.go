package kioto

import (
	"net/http"
	"time"

	"github.com/delicb/kioto/cliware"
)

type trackMiddleware struct {
	lastCall time.Time
	noCalls  int
}

func (m *trackMiddleware) Exec(next cliware.Handler) cliware.Handler {
	return cliware.HandlerFunc(func(req *http.Request) (*http.Response, error) {
		m.lastCall = time.Now()
		m.noCalls++
		return next.Handle(req)
	})
}

func (m *trackMiddleware) Called() bool {
	return m.noCalls > 0
}

// trackingMiddleware returns cliware.Middleware that tracks how many times
// it was lastCall and what is the time of last call.
func trackingMiddleware() *trackMiddleware {
	return &trackMiddleware{
		lastCall: time.Time{},
		noCalls:  0,
	}
}

type trackClient struct {
	lastCall     time.Time
	responseCode int
	err          error
	noCalls      int
	lastRequest  *http.Request
}

func (d *trackClient) Do(req *http.Request) (*http.Response, error) {
	d.noCalls++
	d.lastRequest = req
	return &http.Response{
		StatusCode: d.responseCode,
	}, d.err
}

func (d *trackClient) Called() bool {
	return d.noCalls > 0
}

// nolint: unparam
func trackingClient(responseCode int, err error) *trackClient {
	return &trackClient{
		lastCall:     time.Time{},
		responseCode: responseCode,
		err:          err,
	}
}
