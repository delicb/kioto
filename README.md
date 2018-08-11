# kioto

Extensible HTTP client with fluent syntax suitable for both quick use and writing
client libraries for large APIs.

Main idea behind `kioto` is to use composable client middlewares to create HTTP
requests. It is intended as a way to make HTTP client implementations more readable,
code more reusable and easier to use. 

# Install
Run `go get github.com/delicb/kioto` in terminal.


# Examples
```go
package kioto_example

import (
	"fmt"
	"log"
	"net/http"

	"github.com/delicb/kioto/middlewares/headers"

	"github.com/delicb/kioto/middlewares/errors"
)

// HTTPBinResponse will hold response returned from httpbin.org server
type HTTPBinResponse struct {
	Headers struct {
		UserAgent string `json:"User-Agent"`
	} `json:"headers"`
}

func Example() {
	respBody := new(HTTPBinResponse)

	// create new kioto client that will be using http.DefaultClient to send
	// actual requests and handle errors via provided middleware.
	client := New(
		HTTPClient(http.DefaultClient),
		Middlewares(errors.Errors()),
	)
	// set custom header
	client.Use(headers.Set("User-Agent", "kioto-example"))

	// send request
	resp, err := client.Request().Get().URL("https://httpbin.org/get").Send()

	// check errors
	// because of errors middleware included in doer, every status codes
	// 400+ will be turned into errors.
	if err != nil {
		panic(err)
	}

	// check body
	if err := resp.JSON(respBody); err != nil {
		log.Fatalf("Got error response from httpbin: %v", err)
	}
	fmt.Println(respBody.Headers.UserAgent)
	// output: kioto-example
}

```

# Contributing
`kioto` is open to contributions in form of issues, pull requests, discussions, 
new ideas. 

## Maintainers
* Bojan Delic
