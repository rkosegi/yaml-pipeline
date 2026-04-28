/*
Copyright 2026 Richard Kosegi

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package pipeline

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rkosegi/yaml-toolkit/common"
	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/stretchr/testify/assert"
)

type failingRoundTripper struct{}

func (failingRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("transport failed")
}

func TestExecuteHttpOp(t *testing.T) {
	var (
		is HttpOpSpec
		gd dom.ContainerBuilder
	)

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodOptions:
			w.WriteHeader(http.StatusInternalServerError)
			return
		case http.MethodPost:
			{
				if strings.Contains(r.URL.RawQuery, "fail") {
					w.Header().Set("Content-Type", "application/octet-stream")
					_, _ = w.Write([]byte(`FAIL`))
				} else if strings.Contains(r.URL.RawQuery, "wrong-status") {
					w.Header().Set("Content-Type", "application/octet-stream")
					w.WriteHeader(http.StatusInternalServerError)
				} else {
					w.Header().Set("Content-Type", "application/json")
					_, _ = w.Write([]byte(`{"out":123}`))
				}
			}
		}
	}))
	defer s.Close()
	t.Run("Garbage method", func(t *testing.T) {
		gd = dom.ContainerNode()
		is = HttpOpSpec{
			Method: "garbage method",
			Url:    s.URL,
		}
		assert.Error(t, New(WithData(gd)).Execute(&is))
	})
	t.Run("OPTIONS", func(t *testing.T) {
		gd = dom.ContainerNode()
		is = HttpOpSpec{
			Method: HttpMethodOPTIONS,
			BodyFrom: &SourceOfData{
				Tree: &ValOrRef{Val: "nonExistent"},
			},
			Url: s.URL,
		}
		assert.NoError(t, New(WithData(gd)).Execute(&is))
	})
	t.Run("GET with failing transport", func(t *testing.T) {
		is = HttpOpSpec{
			Method: HttpMethodGET,
			Url:    s.URL,
		}
		httpRp = &failingRoundTripper{}
		defer func() { httpRp = http.DefaultTransport }()
		assert.Error(t, New().Execute(&is))
	})
	t.Run("POST", func(t *testing.T) {
		gd = dom.ContainerNode()
		gd.AddValue("inputBody", dom.ContainerNode().AddValue("test", dom.LeafNode(1)))
		is = HttpOpSpec{
			Method: HttpMethodPOST,
			Url:    s.URL,
			Headers: &HttpHeaders{
				"accept": []string{"application/json"},
			},
			BodyFrom: &SourceOfData{
				Tree: &ValOrRef{Val: "inputBody"},
			},
			ParseAs: new(ParseContentModeJson),
			StoreTo: &ValOrRef{
				Val: "Output",
			},
			ValidStatusCodes: new([]int{http.StatusOK}),
		}
		t.Run("Default", func(t *testing.T) {
			f := filepath.Join(t.TempDir(), "writeTo")
			is.WriteTo = new(&ValOrRef{Val: f})
			defer func() { is.WriteTo = nil }()
			assert.NoError(t, New(WithData(gd)).Execute(&is))
			assert.Equal(t, float64(123), gd.Get(pp.MustParse("Output.body.out")).AsLeaf().Value())
			data, err := os.ReadFile(f)
			assert.NoError(t, err)
			assert.NotNil(t, data)
		})
		t.Run("bodyFrom, but incomplete config", func(t *testing.T) {
			is.BodyFrom = &SourceOfData{}
			defer func() { is.BodyFrom = nil }()
			assert.Error(t, New(WithData(gd)).Execute(&is))
		})
		t.Run("bodyFrom, but unknown encoding", func(t *testing.T) {
			is.BodyFrom = &SourceOfData{Tree: &ValOrRef{Val: "nonExistent"}}
			is.EncodeAs = new(EncodeDataAs("garbage"))
			defer func() { is.BodyFrom = nil; is.EncodeAs = nil }()
			assert.Error(t, New(WithData(gd)).Execute(&is))
		})
		t.Run("Write To invalid", func(t *testing.T) {
			is.WriteTo = new(&ValOrRef{Val: "/I/hope/this/path/does/not/exist"})
			defer func() { is.WriteTo = nil }()
			assert.Error(t, New(WithData(gd)).Execute(&is))
		})
		t.Run("Wrong response code", func(t *testing.T) {
			is.Url = is.Url + "?wrong-status"
			defer func() { is.Url = s.URL }()
			assert.Error(t, New(WithData(gd)).Execute(&is))
		})
		t.Run("Wrong response body", func(t *testing.T) {
			is.Url = is.Url + "?fail"
			defer func() { is.Url = s.URL }()
			assert.Error(t, New(WithData(gd)).Execute(&is))
		})
		t.Run("No ParseAs", func(t *testing.T) {
			is.ParseAs = nil
			assert.NoError(t, New(WithData(gd)).Execute(&is))
			assert.Equal(t, `{"out":123}`, gd.Get(pp.MustParse("Output.body")).AsLeaf().Value())
		})
	})
}

func TestHttpOpCloneWith(t *testing.T) {
	eo := &HttpOpSpec{
		Method: HttpMethodGET,
	}
	d := dom.ContainerNode()
	eo = eo.CloneWith(newMockActBuilder().data(d).build()).(*HttpOpSpec)
	assert.Equal(t, HttpMethodGET, eo.Method)
}

func TestHttpOpProcessResponse(t *testing.T) {
	resp := &http.Response{Body: io.NopCloser(common.FailingReader())}
	ac := newMockActBuilder().build()
	ho := &HttpOpSpec{}
	err := ho.processResponse(ac, resp)
	assert.Error(t, err)
}
