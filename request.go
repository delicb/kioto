package kioto

import (
	"context"
	"io"
	"net/http"

	"github.com/delicb/kioto/middlewares/body"

	"github.com/delicb/kioto/cliware"
	"github.com/delicb/kioto/middlewares/headers"
	"github.com/delicb/kioto/middlewares/url"
)

// Request holds information needed to construct HTTP request to send over the wire.
type Request struct {
	Client      *Client
	middlewares *cliware.Chain
	context     context.Context
}

// NewRequest creates and returns instance of Request that will use provided
// client to send HTTP request and use clients middlewares.
func NewRequest(client *Client) *Request {
	return &Request{
		Client:      client,
		middlewares: client.preMiddlewares.ChildChain(),
		context:     context.Background(),
	}
}

// Context returns currently set context.Context of the request.
func (r *Request) Context() context.Context {
	return r.context
}

// WithContext sets new context for this request.
func (r *Request) WithContext(ctx context.Context) *Request {
	r.context = ctx
	return r
}

// Use adds middlewares that will be applied to this request only.
func (r *Request) Use(m ...cliware.Middleware) *Request {
	r.middlewares.Use(m...)
	return r
}

// UseFunc adds middlewares defined as function that will be applied to this request only.
func (r *Request) UseFunc(m func(cliware.Handler) cliware.Handler) *Request {
	r.middlewares.UseFunc(m)
	return r
}

// Send constructs HTTP request by combining all middlewares (request specific and
// defined in Client), sends request and returns response.
func (r *Request) Send() (*Response, error) {
	sender := r.middlewares.Exec(r.Client.postMiddlewares.Exec(cliware.HandlerFunc(r.Client.sendRequest)))
	if r.context == nil {
		r.context = context.Background()
	}
	req := cliware.EmptyRequest().WithContext(r.context)
	resp, err := sender.Handle(req)
	return buildResponse(resp, err), err
}

// Utility methods - shortcuts to using middlewares. These should not map all
// middlewares, only most common used ones for quick use.

// Get sets HTTP method to GET for this request.
func (r *Request) Get() *Request { return r.Method(http.MethodGet) }

// Post sets HTTP method to POST for this request.
func (r *Request) Post() *Request { return r.Method(http.MethodPost) }

// Put sets HTTP method to Put for this request.
func (r *Request) Put() *Request { return r.Method(http.MethodPut) }

// Delete sets HTTP method to DELETE for this request.
func (r *Request) Delete() *Request { return r.Method(http.MethodDelete) }

// Patch sets HTTP method to PATCH for this request.
func (r *Request) Patch() *Request { return r.Method(http.MethodPatch) }

// Head sets HTTP method to HEAD for this request.
func (r *Request) Head() *Request { return r.Method(http.MethodHead) }

// Options sets HTTP method to OPTIONS for this request.
func (r *Request) Options() *Request { return r.Method(http.MethodOptions) }

// Connect sets HTTP method to CONNECT for this request.
func (r *Request) Connect() *Request { return r.Method(http.MethodConnect) }

// Trace sets HTTP method to Trace for this request.
func (r *Request) Trace() *Request { return r.Method(http.MethodTrace) }

// Method sets HTTP method (verb) for this request.
func (r *Request) Method(method string) *Request {
	r.Use(headers.Method(method))
	return r
}

// URL sets url to be used for this request.
func (r *Request) URL(u string) *Request {
	r.Use(url.URL(u))
	return r
}

// Header adds provided headers with provided value to this request.
func (r *Request) Header(key, val string) *Request {
	r.Use(headers.Add(key, val))
	return r
}

// Body sets HTTP body for this request.
func (r *Request) Body(reader io.Reader) *Request {
	r.Use(body.Reader(reader))
	return r
}
