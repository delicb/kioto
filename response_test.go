package kioto

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

type responseSuite struct {
	suite.Suite
}

func (t *responseSuite) TestJSONSuccess() {
	type d struct {
		A string `json:"a"`
	}

	rawResponse := &http.Response{
		Body: ioutil.NopCloser(strings.NewReader(`{"a": "foo"}`)),
	}
	resp := buildResponse(rawResponse, nil)
	data := new(d)
	err := resp.JSON(data)
	t.NoError(err)
	t.Equal("foo", data.A)
}

func (t *responseSuite) TestJSONHTTPError() {
	err := errors.New("some error")
	resp := buildResponse(nil, err)
	data := make(map[string]interface{})
	gotErr := resp.JSON(&data)
	t.Equal(err, gotErr, "got unexpected error")
}

func (t *responseSuite) TestJSONBadSyntax() {
	rawResponse := &http.Response{
		Body: ioutil.NopCloser(strings.NewReader(`{`)),
	}
	resp := buildResponse(rawResponse, nil)
	data := make(map[string]interface{})
	err := resp.JSON(&data)
	t.Error(err)
}

func TestResponseSuite(t *testing.T) {
	suite.Run(t, new(responseSuite))
}
