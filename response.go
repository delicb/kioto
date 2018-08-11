package kioto

import (
	"encoding/json"
	"io"
	"net/http"
)

// Response is thin wrapper around http.Response that provides some
// convenient behavior.
type Response struct {
	*http.Response
	Error error
}

// buildResponse creates new instance of response based on raw HTTP response.
func buildResponse(rawResponse *http.Response, err error) *Response {
	return &Response{
		Response: rawResponse,
		Error:    err,
	}
}

// JSON decodes response body to provided structure from JSON format. Parameter
// has to be pointer to structure into which JSON deserialization will happen.
func (r *Response) JSON(userStruct interface{}) (err error) {
	if r.Error != nil {
		return r.Error
	}

	defer func() {
		closeErr := r.Body.Close()
		// only if there is no other error, set close error
		// otherwise underlying error will be masked
		if err == nil && closeErr != nil {
			err = closeErr
		}
	}()
	jsonDecoder := json.NewDecoder(r.Body)

	err = jsonDecoder.Decode(userStruct)
	if err != nil && err != io.EOF {
		return err
	}
	return nil
}
