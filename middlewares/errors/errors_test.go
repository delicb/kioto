package errors_test

import (
	"bytes"
	sterrors "errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/delicb/kioto/cliware"
	"github.com/delicb/kioto/middlewares/errors"
)

func TestErrors(t *testing.T) {
	for _, data := range []struct {
		Response      *http.Response
		OriginalError error
		Error         *errors.HTTPError
	}{
		{
			Response: &http.Response{
				StatusCode: 200,
			},
			OriginalError: nil,
			Error:         nil,
		},
		{
			Response: &http.Response{
				StatusCode: 302,
			},
			OriginalError: nil,
			Error:         nil,
		},
		{
			Response: &http.Response{
				StatusCode: 400,
				Request: &http.Request{
					Method: "POST",
					URL: &url.URL{
						Host: "delic.rs",
						Path: "/foobar",
					},
				},
			},
			OriginalError: nil,
			Error: &errors.HTTPError{
				StatusCode: 400,
				Method:     "POST",
			},
		},
		{
			Response: &http.Response{
				StatusCode: 401,
				Body:       ioutil.NopCloser(bytes.NewBuffer([]byte("body"))),
				Request: &http.Request{
					Method: "GET",
					URL: &url.URL{
						Host: "golang.com",
						Path: "/somepath",
					},
				},
			},
			OriginalError: nil,
			Error: &errors.HTTPError{
				StatusCode: 401,
				Method:     "GET",
				Body:       []byte("body"),
			},
		},
		{
			Response:      nil,
			OriginalError: sterrors.New("custom error"),
			Error:         nil,
		},
	} {
		m := errors.Errors()
		req := cliware.EmptyRequest()
		handler := createHandler(data.Response, data.OriginalError)
		_, err := m.Exec(handler).Handle(req)

		if data.Error == nil && data.OriginalError == nil {
			assert.NoError(t, err)
			continue
		}

		if data.OriginalError != nil {
			assert.Equal(t, data.OriginalError.Error(), err.Error())
			continue
		}

		if httpErr, ok := err.(*errors.HTTPError); ok {
			require.Equal(t, httpErr.StatusCode, data.Error.StatusCode)
			require.Equal(t, httpErr.Method, data.Error.Method)
			require.Equal(t, httpErr.Body, data.Error.Body)
		} else {
			t.Errorf("Wrong error type. Expected HTTPError, got: %T", err)
		}
	}
}

func TestHTTPError_Error(t *testing.T) {
	for _, data := range []struct {
		Error    *errors.HTTPError
		Expected string
	}{
		{
			Error: &errors.HTTPError{
				Method: "POST",
			},
			Expected: ".*POST.*",
		},
		{
			Error: &errors.HTTPError{
				Name: "401 Forbidden",
			},
			Expected: ".*401.*",
		},
		{
			Error: &errors.HTTPError{
				RequestURL: "some_url",
			},
			Expected: ".*some_url.*",
		},
	} {
		errStr := data.Error.Error()
		match, err := regexp.Match(data.Expected, []byte(errStr))
		if err != nil {
			t.Error("Regex match failed: ", err)
		}
		if !match {
			t.Errorf("Wrong error string. Got: %s, did not match regexp: %s", errStr, data.Expected)
		}
	}
}

func createHandler(wantedResponse *http.Response, wantError error) cliware.Handler {
	return cliware.HandlerFunc(func(req *http.Request) (resp *http.Response, err error) {
		return wantedResponse, wantError
	})
}
