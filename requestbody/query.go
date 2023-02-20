package requestbody

import (
    "fmt"
    "net/url"

    "github.com/gin-gonic/gin"
)

type RequestQuery struct { values url.Values }
func (r RequestQuery) SetString(key, s string) {
    fmt.Println(r.values.Get(key))
    r.values.Del(key)
    fmt.Println(r.values.Get(key))
    r.values.Set(key, s)
    fmt.Println(r.values.Get(key))
}
func (r RequestQuery) GetString(key string) string { return r.values.Get(key) }
func (r RequestQuery) Encode() string { return r.values.Encode() }

type QueryTransformer interface {
    Transform(c *gin.Context, q RequestQuery)
}

type QueryTransformerFunc func(c *gin.Context, q RequestQuery)

func (f QueryTransformerFunc) Transform(c *gin.Context, q RequestQuery) {
    f(c, q)
}

type QueryTransformerChain []QueryTransformer

func (t QueryTransformerChain) Transform(c *gin.Context, q RequestQuery) {
    for _, tr := range t {
        tr.Transform(c, q)
        if c.IsAborted() {
            return
        }
    }
}

// TransformQuery is a Request Body transformer middleware. It takes
// a transformer and applies it to a decoded request body, then resets
// the body on the request
func TransformQuery(f QueryTransformer) gin.HandlerFunc {
    // The interface for this is kind of rough to implement with a struct
    // type, so maybe not needed to be configurable
    return func(c *gin.Context) {
        query := RequestQuery{values: c.Request.URL.Query()}
        f.Transform(c, query)
        if c.IsAborted() {
            return
        }
        c.Request.URL.RawQuery = query.Encode()
    }
}
