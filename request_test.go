package kioto

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/delicb/kioto/cliware"
)

type requestSuite struct {
	suite.Suite
	trackingClient *trackClient
	client         *Client
}

func (t *requestSuite) SetupTest() {
	t.trackingClient = trackingClient(200, nil)
	t.client = New(HTTPClient(t.trackingClient))
}

func (t *requestSuite) TearDownTest() {
	t.client = nil
}

func (t *requestSuite) TestNewRequest() {
	req := NewRequest(t.client)
	t.Equal(req.Client, t.client)
}

func (t *requestSuite) TestContext() {
	req := NewRequest(t.client)
	type ctxKey string
	key := ctxKey("a")
	ctx := context.WithValue(context.Background(), key, "b")
	req.WithContext(ctx)
	t.Equal(ctx, req.Context())
	_, err := req.Send()
	t.NoError(err)

	ctxValue := t.trackingClient.lastRequest.Context().Value(key)
	t.Equal(ctxValue, "b")
}

func (t *requestSuite) TestUse() {
	req := NewRequest(t.client)
	tm := trackingMiddleware()
	req.Use(tm)
	_, err := req.Send()
	t.NoError(err)
	t.True(tm.Called())
}

func (t *requestSuite) TestUseFunc() {
	req := NewRequest(t.client)
	var called bool
	req.UseFunc(func(next cliware.Handler) cliware.Handler {
		return cliware.HandlerFunc(func(request *http.Request) (*http.Response, error) {
			called = true
			return next.Handle(request)
		})
	})
	_, err := req.Send()
	t.NoError(err)
	t.True(called, "function middleware not called")
}

func (t *requestSuite) TestMethod() {
	req := NewRequest(t.client)
	req.Method("FOO")
	_, err := req.Send()
	t.NoError(err)
	t.Equal("FOO", t.trackingClient.lastRequest.Method)
}

func (t *requestSuite) TestMethods() {
	type data struct {
		method       func(r *Request)
		expectedVerb string
	}
	testData := []data{
		{
			method:       func(rr *Request) { rr.Get() },
			expectedVerb: http.MethodGet,
		},
		{
			method:       func(rr *Request) { rr.Post() },
			expectedVerb: http.MethodPost,
		},
		{
			method:       func(rr *Request) { rr.Put() },
			expectedVerb: http.MethodPut,
		},
		{
			method:       func(rr *Request) { rr.Delete() },
			expectedVerb: http.MethodDelete,
		},
		{
			method:       func(rr *Request) { rr.Patch() },
			expectedVerb: http.MethodPatch,
		},
		{
			method:       func(rr *Request) { rr.Head() },
			expectedVerb: http.MethodHead,
		},
		{
			method:       func(rr *Request) { rr.Options() },
			expectedVerb: http.MethodOptions,
		},
		{
			method:       func(rr *Request) { rr.Connect() },
			expectedVerb: http.MethodConnect,
		},
		{
			method:       func(rr *Request) { rr.Trace() },
			expectedVerb: http.MethodTrace,
		},
	}
	for _, d := range testData {
		d := d
		t.T().Run(d.expectedVerb, func(tt *testing.T) {
			req := NewRequest(t.client)
			d.method(req)
			_, err := req.Send()
			t.NoError(err)
			t.Equal(d.expectedVerb, t.trackingClient.lastRequest.Method)

		})

	}
}

func (t *requestSuite) TestURL() {
	req := NewRequest(t.client)
	url := "http://example.com/foobar"
	req.URL(url)
	_, err := req.Send()
	t.NoError(err)
	t.Equal(url, t.trackingClient.lastRequest.URL.String())
}

func (t *requestSuite) TestHeader() {
	req := NewRequest(t.client)

	req.Header("foo", "bar")
	_, err := req.Send()
	t.NoError(err)
	t.Equal("bar", t.trackingClient.lastRequest.Header.Get("foo"))
}

func (t *requestSuite) TestBody() {
	req := NewRequest(t.client)
	reader := strings.NewReader("some content")
	req.Body(reader)
	_, err := req.Send()
	t.NoError(err)
	content, err := ioutil.ReadAll(t.trackingClient.lastRequest.Body)
	t.NoError(err)
	t.Equal("some content", string(content), "body content did not match")
}

func TestRequestSuite(t *testing.T) {
	suite.Run(t, new(requestSuite))
}
