package responsebody_test

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"bytes"

	"github.com/delicb/kioto/cliware"
	"github.com/delicb/kioto/middlewares/responsebody"
)

func TestJSON(t *testing.T) {
	for _, data := range []struct {
		RawData  string
		Expected map[string]interface{}
		Error    error
	}{
		{
			RawData: `{"foo": "bar"}`,
			Expected: map[string]interface{}{
				"foo": "bar",
			},
			Error: nil,
		},
		{
			RawData:  `{"foo": "bar"}`,
			Expected: map[string]interface{}{},
			Error:    errors.New("some error"),
		},
	} {
		data := data
		var body map[string]interface{}
		req := cliware.EmptyRequest()
		handler := func(req *http.Request) (*http.Response, error) {
			r := &http.Response{
				Body: ioutil.NopCloser(strings.NewReader(data.RawData)),
			}
			return r, data.Error
		}

		_, err := responsebody.JSON(&body).Exec(cliware.HandlerFunc(handler)).Handle(req)
		if data.Error != nil {
			require.Error(t, err)
			require.Equal(t, err, data.Error)
			continue
		}
		require.NoError(t, err)
		require.Equal(t, data.Expected, body)
	}
}

func TestString(t *testing.T) {
	for _, data := range []struct {
		Data  string
		Error error
	}{
		{
			Data:  "foo bar",
			Error: nil,
		},
		{
			Data:  "foo bar",
			Error: errors.New("custom error"),
		},
	} {
		data := data
		var body string
		req := cliware.EmptyRequest()
		handler := func(req *http.Request) (*http.Response, error) {
			r := &http.Response{
				Body: ioutil.NopCloser(strings.NewReader(data.Data)),
			}
			return r, data.Error
		}
		_, err := responsebody.String(&body).Exec(cliware.HandlerFunc(handler)).Handle(req)
		if data.Error != nil {
			require.Error(t, err)
			require.Equal(t, data.Error, err)
			continue
		}
		require.Equal(t, data.Data, body)
	}
}

func TestWriter(t *testing.T) {
	for _, data := range []struct {
		Data  string
		Error error
	}{
		{
			Data:  "foo bar",
			Error: nil,
		},
		{
			Data:  "foo bar",
			Error: errors.New("my error"),
		},
	} {
		data := data
		buf := &bytes.Buffer{}
		req := cliware.EmptyRequest()
		handler := func(req *http.Request) (*http.Response, error) {
			r := &http.Response{
				Body: ioutil.NopCloser(strings.NewReader(data.Data)),
			}
			return r, data.Error
		}
		_, err := responsebody.Writer(buf).Exec(cliware.HandlerFunc(handler)).Handle(req)

		if data.Error != nil {
			require.Error(t, err)
			require.Equal(t, err, data.Error)
			continue
		}

		require.NoError(t, err)
		require.Equal(t, buf.String(), data.Data)
	}
}
