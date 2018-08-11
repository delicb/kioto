package kioto

import (
	"fmt"
	"log"
	"net/http"

	"github.com/delicb/kioto/middlewares/headers"

	"github.com/delicb/kioto/middlewares/errors"
)

type HTTPBinResponse struct {
	Headers struct {
		UserAgent string `json:"User-Agent"`
	} `json:"headers"`
}

func Example() {
	respBody := new(HTTPBinResponse)

	client := New(
		HTTPClient(http.DefaultClient),
		Middlewares(errors.Errors()),
	)
	client.Use(headers.Set("User-Agent", "kioto-example"))
	// send request
	resp, err := client.Request().Get().URL("https://httpbin.org/get").Send()

	// check errors
	// because of errors middleware included in doer, every status codes
	// 400+ will be turned into errors.
	if err != nil {
		panic(err)
	}

	if err := resp.JSON(respBody); err != nil {
		log.Fatalf("Got error response from httpbin: %v", err)
	}
	fmt.Println(respBody.Headers.UserAgent)
	// output: kioto-example
}
