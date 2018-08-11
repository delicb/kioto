package kioto

import (
	"fmt"
	"runtime"
	"time"

	"github.com/delicb/kioto/cliware"
)

// ClientOption defines function type for modifying how doer instance behaves
type ClientOption func(opts *clientOptions)

// clientOptions holds available options for modification on how Client behaves.
type clientOptions struct {
	disableRetry    bool
	middlewares     []cliware.Middleware
	potsMiddlewares []cliware.Middleware
	httpClient      HTTPDoer
	userAgent       UserAgent
	timeout         time.Duration
}

// DisableRetry causes that HTTP requests will not be retried if they failed.
// By default, HTTP requests are retried if classifier determines that error
// is recoverable. If this is used, RoundTripper on provided doer will be used.
// If not, RoundTripper will be wrapped with one capable of doing retries.
func DisableRetry() ClientOption {
	return func(opts *clientOptions) {
		opts.disableRetry = true
	}
}

// Middlewares sets default list of middlewares to be used for each request made
// with this doer.
func Middlewares(middlewares ...cliware.Middleware) ClientOption {
	return func(opts *clientOptions) {
		opts.middlewares = middlewares
	}
}

// PostMiddlewares sets list of middlewares to be used for each request made
// with this doer, but AFTER all middlewares specific to request itself
// are applied. For example, this can be useful for cleanup or error parsing
// that needs to be applied for all requests.
func PostMiddlewares(middlewares ...cliware.Middleware) ClientOption {
	return func(opts *clientOptions) {
		opts.potsMiddlewares = middlewares
	}
}

// HTTPClient sets instance of Doer to use. Any *http.Client implements this
// interface (e.g. http.DefaultClient). If not provided, new instance
// of *http.Client will be created.
func HTTPClient(sender HTTPDoer) ClientOption {
	return func(opts *clientOptions) {
		opts.httpClient = sender
	}
}

// ProductInfo sets information about product that is using this client.
// This information (together with library info and Go runtime) will be
// included in User-Agent string.
func ProductInfo(name, version string) ClientOption {
	return func(opts *clientOptions) {
		opts.userAgent = UserAgent{
			Product: name,
			Version: version,
		}
	}
}

// Timeout sets maximal duration of request preMiddlewares failing. This only has
// effect if HTTPClient option is not used. If it used, timeout should be set
// on doer directly.
// If not set, default is 0, which means that there is no timeout.
// Timeout can be set per request as well, by setting context with timeout
// or deadline. Lower value will be applied, so this should be highest value for
// all requests made with this doer.
func Timeout(timeout time.Duration) ClientOption {
	return func(opts *clientOptions) {
		opts.timeout = timeout
	}
}

// UserAgent holds information about product and version using this library
// to be used as part of UserAgent string.
type UserAgent struct {
	Product string
	Version string
}

// IsEmpty checks if UserAgent is set.
func (ua UserAgent) IsEmpty() bool {
	return ua.Product == "" && ua.Version == ""
}

// String implements Stringer interface and generates UserAgent stirng.
func (ua UserAgent) String() string {

	return fmt.Sprintf("%s/%s %s/%s %s/%s",
		ua.Product, ua.Version,
		LibraryName, LibraryVersion,
		"go", runtime.Version(),
	)
}
