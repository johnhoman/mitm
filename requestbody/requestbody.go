package requestbody

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/crossplane/crossplane-runtime/pkg/fieldpath"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type RequestBody map[string]any

func (r RequestBody) SetString(path string, value string) error {
	paved := fieldpath.Pave(r)
	if err := paved.SetString(path, value); err != nil {
		return err
	}
	return nil
}

func (r RequestBody) AppendMap(path string, value map[string]any) error {
	paved := fieldpath.Pave(r)
	v, err := paved.GetValue(path)
	if err != nil {
		if !fieldpath.IsNotFound(err) {
			return err
		}
		v = make([]any, 0)
	}
	sl, ok := v.([]any)
	if !ok {
		return errors.Errorf("Cannot append to type %T, want type %T", v, []any{})
	}
	return paved.SetValue(path, append(sl, value))
}

type BodyTransformer interface {
	Transform(c *gin.Context, body RequestBody)
}

type BodyTransformerFunc func(c *gin.Context, body RequestBody)

func (f BodyTransformerFunc) Transform(c *gin.Context, body RequestBody) {
	f(c, body)
}

type BodyTransformerChain []BodyTransformer

func (t BodyTransformerChain) Transform(c *gin.Context, body RequestBody) {
	for _, tr := range t {
		tr.Transform(c, body)
		if c.IsAborted() {
			return
		}
	}
}

// TransformBody is a Request Body transformer middleware. It takes
// a transformer and applies it to a decoded request body, then resets
// the body on the request
func TransformBody(f BodyTransformer) gin.HandlerFunc {
	// The interface for this is kind of rough to implement with a struct
	// type, so maybe not needed to be configurable
	return func(c *gin.Context) {
		body := RequestBody{}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		f.Transform(c, body)
		if c.IsAborted() {
			return
		}
		b, err := json.Marshal(body)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		c.Request.ContentLength = int64(len(b))
		c.Request.Header.Set("Content-Length", strconv.Itoa(len(b)))
		c.Request.Body = io.NopCloser(bytes.NewReader(b))
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
