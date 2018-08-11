package kioto_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/mccutchen/go-httpbin/httpbin"
	"github.com/stretchr/testify/suite"
	"github.com/delicb/kioto"
	"github.com/delicb/kioto/cliware"
	"github.com/delicb/kioto/middlewares/auth"
	"github.com/delicb/kioto/middlewares/responsebody"
	"github.com/delicb/kioto/middlewares/retry"
)

type integrationSuite struct {
	suite.Suite
	server *httptest.Server
}

func (i *integrationSuite) SetupSuite() {
	runIntegration := strings.ToLower(os.Getenv("KIOTO_INTEGRATION_TEST"))
	var shouldRun bool
	for _, val := range []string{"true", "on", "1"} {
		if val == runIntegration {
			shouldRun = true
		}
	}
	if !shouldRun {
		i.T().Skip()
	}

	i.server = httptest.NewServer(httpbin.NewHTTPBin().Handler())
}

func (i *integrationSuite) TearDownSuite() {
	i.server.Close()
}

func (i *integrationSuite) url(path string) string {
	return i.server.URL + "/" + path
}

func (i *integrationSuite) TestGet() {
	client := kioto.New()
	resp, err := client.Request().Get().URL(i.server.URL).Send()
	i.Require().NoError(err)
	i.Equal(200, resp.StatusCode)
}

func (i *integrationSuite) TestPost() {
	client := kioto.New()
	resp, err := client.Request().Post().URL(i.url("post")).Send()
	i.Require().NoError(err)
	i.Equal(200, resp.StatusCode)
}

func (i *integrationSuite) TestTimeout() {
	client := kioto.New(kioto.HTTPClient(&http.Client{
		Timeout: 1 * time.Second,
	}))
	_, err := client.Request().Get().URL(i.url("delay/2")).Send()
	i.Require().Error(err)
	i.Contains(strings.ToLower(err.Error()), "timeout")
}

func (i *integrationSuite) TestContextTimeout() {
	client := kioto.New()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	_, err := client.Request().Get().WithContext(ctx).URL(i.url("delay/2")).Send()
	i.Require().Error(err)
	i.Contains(strings.ToLower(err.Error()), "context deadline exceeded")
}

func (i *integrationSuite) TestJSON() {
	client := kioto.New()
	var response = make(map[string]interface{})
	resp, err := client.Request().
		Get().
		URL(i.url("user-agent")).
		Use(responsebody.JSON(&response)).
		Send()
	i.Require().NoError(err)
	i.Equal(200, resp.StatusCode)
	i.Contains(response, "user-agent")
}

func (i *integrationSuite) TestRetry() {
	client := kioto.New()
	tryTimes := 3
	tried := 0
	resp, err := client.Request().
		Get().
		URL(i.url("status/500")).
		Use(retry.SetClassifier(func(r *http.Response, err error) bool {
			tried++
			return r.StatusCode >= 500
		})).
		Use(retry.Times(tryTimes)).
		Send()
	i.True(tried >= tryTimes, "retry was not attempted expected number of times")
	i.Equal(500, resp.StatusCode)
	i.NoError(err)
}

func (i *integrationSuite) TestDisableRetry() {
	client := kioto.New(kioto.DisableRetry())
	tried := 0
	resp, err := client.Request().
		Get().
		URL(i.url("status/500")).
		Use(retry.SetClassifier(func(r *http.Response, err error) bool {
			tried++
			// just simulate to retry everything, it should not happen
			// anyway because of DisableRetry option
			return true
		})).
		Send()
	i.True(tried == 0, "Retry attempted and was not expected to")
	i.Equal(500, resp.StatusCode)
	i.NoError(err)

}

func (i *integrationSuite) TestAuth() {
	client := kioto.New()
	resp, err := client.Request().
		Get().
		URL(i.url("basic-auth/foo/bar")).
		Use(auth.Basic("foo", "bar")).
		Send()
	i.NoError(err)
	i.Equal(200, resp.StatusCode)
}

func (i *integrationSuite) TestErrors() {
	client := kioto.New()
	resp, err := client.Request().
		Get().
		URL(i.url("status/500")).
		Use(cliware.ResponseProcessor(func(r *http.Response, e error) error {
			if e != nil {
				return e
			}
			if r.StatusCode >= 500 {
				return errors.New("server error")
			}
			return nil
		})).
		Send()
	i.Error(err)
	i.Contains(err.Error(), "server error")
	i.Equal(500, resp.StatusCode)
}

func (i *integrationSuite) TestClientLevelErrors() {
	client := kioto.New()
	client.UsePost(cliware.ResponseProcessor(func(resp *http.Response, err error) error {
		if err != nil {
			return err
		}
		if resp.StatusCode >= 500 {
			return errors.New("server error")
		}
		return nil
	}))
	resp, err := client.Request().
		Get().
		URL(i.url("status/500")).
		Send()
	i.Require().Error(err)
	i.Contains(err.Error(), "server error")
	i.Equal(500, resp.StatusCode)
}

type errorRoundTripper struct {
	errToReturn error
}

func (rt *errorRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return nil, rt.errToReturn
}

func (i *integrationSuite) TestCustomHTTPClient() {
	httpClient := &http.Client{
		Transport: &errorRoundTripper{errors.New("round trip error")},
	}
	client := kioto.New(kioto.HTTPClient(httpClient))
	_, err := client.Request().Get().URL("foobar").Send()
	i.Error(err)
	i.Regexp("round trip error", err.Error())
}

func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(integrationSuite))
}
