package kioto

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/delicb/kioto/cliware"
	"github.com/delicb/kioto/middlewares/url"
)

type middlewareSuite struct {
	suite.Suite
}

func (t *middlewareSuite) TestSingleMiddleware() {
	m := trackingMiddleware()
	baseClient := trackingClient(200, nil)
	client := New(HTTPClient(baseClient), Middlewares(m))
	resp, err := client.Request().Get().Use(url.BaseURL("https://httpbin.org"), url.Path("/")).Send()

	t.NoError(err)
	t.True(m.Called(), "middleware not executed")
	t.True(baseClient.Called(), "client not executed")
	t.Equal(200, resp.StatusCode, "wrong status code")
}

func (t *middlewareSuite) TestClientAndRequestMiddleware() {
	clientMiddleware := trackingMiddleware()
	requestMiddleware := trackingMiddleware()
	baseClient := trackingClient(200, nil)
	client := New(HTTPClient(baseClient), Middlewares(clientMiddleware))
	_, err := client.Request().Get().Use(requestMiddleware).Send()

	t.NoError(err)
	t.True(clientMiddleware.Called(), "client middleware not called")
	t.True(requestMiddleware.Called(), "request middleware not called")
	t.True(requestMiddleware.lastCall.After(clientMiddleware.lastCall), "Request middleware should have been lastCall after client middleware")
	t.True(baseClient.Called(), "client not called")
}

func (t *middlewareSuite) TestClientPostMiddleware() {
	clientPreMiddleware := trackingMiddleware()
	requestMiddleware := trackingMiddleware()
	clientPostMiddleware := trackingMiddleware()
	baseClient := trackingClient(200, nil)
	client := New(HTTPClient(baseClient), Middlewares(clientPreMiddleware), PostMiddlewares(clientPostMiddleware))
	_, err := client.Request().Get().Use(requestMiddleware).Send()

	t.NoError(err)
	t.True(clientPreMiddleware.Called(), "client pre middleware not lastCall")
	t.True(requestMiddleware.Called(), "request middleware not lastCall")
	t.True(clientPostMiddleware.Called(), "client post middleware not lastCall")
	t.True(requestMiddleware.lastCall.After(clientPreMiddleware.lastCall), "Request middleware should have been after client pre middleware")
	t.True(clientPostMiddleware.lastCall.After(requestMiddleware.lastCall), "Request middleware should have been lastCall before client post middleware")
	t.True(baseClient.Called(), "base client not called")
}

func (t *middlewareSuite) TestSharedMiddleware() {
	m := trackingMiddleware()
	client := New(HTTPClient(trackingClient(200, nil)), Middlewares(m))
	_, err := client.Request().Get().Send()

	t.NoError(err)
	t.Equal(1, m.noCalls, "request not counted")

	// send same request with same client to see how count changes
	_, err = client.Request().Get().Send()

	t.NoError(err)
	t.Equal(2, m.noCalls, "request not counted")
}

func (t *middlewareSuite) TestStopPropagationOnError() {
	m := trackingMiddleware()
	baseClient := trackingClient(200, nil)
	client := New(HTTPClient(baseClient), Middlewares(m))

	var detectedError bool
	client.UseFunc(func(next cliware.Handler) cliware.Handler {
		return cliware.HandlerFunc(func(req *http.Request) (*http.Response, error) {
			resp, err := next.Handle(req)
			if err != nil {
				detectedError = true
			}
			return resp, err
		})
	})

	req := client.Request().Get().UseFunc(func(next cliware.Handler) cliware.Handler {
		return cliware.HandlerFunc(func(req *http.Request) (*http.Response, error) {
			// not calling next handler, just fail here
			return nil, errors.New("some error")
		})
	})

	_, err := req.Send()

	t.Error(err, "expected error")
	t.True(detectedError, "error not detected")
	t.False(baseClient.Called(), "base client should have not been called")
}

func TestMiddlewareSuite(t *testing.T) {
	suite.Run(t, new(middlewareSuite))
}
