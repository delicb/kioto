package kioto

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultClient(t *testing.T) {
	client := New()
	if httpClient, ok := client.doer.(*http.Client); !ok {
		t.Fatalf("Default http clinet of wrong type: %T", client.doer)
	} else {
		assert.Equal(t, time.Duration(0), httpClient.Timeout, "timeout not default")
	}

	assert.Len(t, client.preMiddlewares.Middlewares(), 0, "unexpected middlewares found")
	assert.Len(t, client.postMiddlewares.Middlewares(), 0, "unexpected middlewares found")
}

func TestDefaultClientUserAgent(t *testing.T) {
	baseClient := trackingClient(200, nil)
	client := New(HTTPClient(baseClient), ProductInfo("foobar", "0.1"))
	assert.Len(t, client.preMiddlewares.Middlewares(), 1, "expected at least one middleware")

	_, _ = client.Request().Get().Send()
	assert.Regexp(t, "foobar/0.1 .*", baseClient.lastRequest.Header.Get("User-Agent"))
}
