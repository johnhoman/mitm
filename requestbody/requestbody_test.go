package requestbody

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/gin-gonic/gin"
)

func TestRequestBody(t *testing.T) {

	cases := map[string]struct {
		transformer BodyTransformer
		body        map[string]any
		want        map[string]any
	}{
		"SetsAStringOnTheBody": {
			transformer: BodyTransformerFunc(func(c *gin.Context, body RequestBody) {
				if err := body.SetString("foo.bar", "baz"); err != nil {
					_ = c.AbortWithError(http.StatusInternalServerError, err)
					return
				}
			}),
			body: map[string]any{
				"foo": map[string]any{
					"bar": "",
				},
			},
			want: map[string]any{
				"foo": map[string]any{
					"bar": "baz",
				},
			},
		},
		"AppendsToAList": {
			transformer: BodyTransformerFunc(func(c *gin.Context, body RequestBody) {
				if err := body.AppendMap("foo.bar", map[string]any{"key": "foo-1", "value": "bar-1"}); err != nil {
					_ = c.AbortWithError(http.StatusInternalServerError, err)
					return
				}
			}),
			body: map[string]any{
				"foo": map[string]any{
					"bar": []any{
						map[string]any{"key": "foo-0", "value": "bar-0"},
					},
				},
			},
			want: map[string]any{
				"foo": map[string]any{
					"bar": []any{
						map[string]any{"key": "foo-0", "value": "bar-0"},
						map[string]any{"key": "foo-1", "value": "bar-1"},
					},
				},
			},
		},
		"CreatesAListIfNotExists": {
			transformer: BodyTransformerFunc(func(c *gin.Context, body RequestBody) {
				if err := body.AppendMap("foo.bar", map[string]any{"key": "foo-1", "value": "bar-1"}); err != nil {
					_ = c.AbortWithError(http.StatusInternalServerError, err)
					return
				}
			}),
			body: map[string]any{
				"foo": map[string]any{
					"baz": "bar",
				},
			},
			want: map[string]any{
				"foo": map[string]any{
					"baz": "bar",
					"bar": []any{
						map[string]any{"key": "foo-1", "value": "bar-1"},
					},
				},
			},
		},
	}

	for name, subtest := range cases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			engine := gin.New()
			engine.POST("/echo", TransformBody(subtest.transformer), Echo())

			req := httptest.NewRequest(http.MethodPost, "/echo", JSON(subtest.body))
			engine.ServeHTTP(w, req)
			qt.Assert(t, w.Code, qt.Equals, http.StatusOK)

			var got map[string]any
			qt.Assert(t, json.Unmarshal(w.Body.Bytes(), &got), qt.IsNil)
			qt.Assert(t, got, qt.DeepEquals, subtest.want)
		})
	}
}

func JSON(body map[string]any) io.Reader {
	b, err := json.Marshal(body)
	if err != nil {
		panic(err.(any))
	}
	return bytes.NewReader(b)
}
