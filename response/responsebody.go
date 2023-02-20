package responsebody

import (
	"bufio"
	"encoding/json"
	"io"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/johnhoman/mitm/requestbody"
)

type BodyTransformer interface {
	Transform(c *gin.Context, body requestbody.RequestBody)
}

type BodyTransformerFunc func(c *gin.Context, body requestbody.RequestBody)

func (f BodyTransformerFunc) Transform(c *gin.Context, body requestbody.RequestBody) {
	f(c, body)
}

type BodyTransformerChain []BodyTransformer

func (t BodyTransformerChain) Transform(c *gin.Context, body requestbody.RequestBody) {
	for _, tr := range t {
		tr.Transform(c, body)
		if c.IsAborted() {
			return
		}
	}
}

// TransformResponseBody is a Request Body transformer middleware. It takes
// a transformer and applies it to a decoded request body, then resets
// the body on the request
func TransformResponseBody(f BodyTransformer) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Call the remained of the chain
		c.Next()

		body := requestbody.RequestBody{}
		if c.Writer.Written() {
			// This is the problem. I can't unwrite this
			return
		}
	}
}

func Echo() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer deferClose(c.Request.Body)
		var m map[string]any
		if err := json.NewDecoder(c.Request.Body).Decode(&m); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		c.JSON(http.StatusOK, m)
	}
}

func deferClose(closer io.ReadCloser) { _ = closer.Close() }
