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
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"

	"github.com/rkosegi/yaml-toolkit/dom"
)

// TODO use service from registry instead
var httpRp = http.DefaultTransport

func (h *HttpOpSpec) canHaveReqBody(m HttpMethod) bool {
	switch m {
	case HttpMethodPOST, HttpMethodPUT, HttpMethodPATCH:
		return true
	}
	return false
}

func (h *HttpOpSpec) isStatusValid(resp *http.Response) bool {
	if h.ValidStatusCodes != nil && len(*h.ValidStatusCodes) > 0 {
		return slices.Contains(*h.ValidStatusCodes, resp.StatusCode)
	}
	return true
}

func (h *HttpOpSpec) Do(ctx ActionContext) (err error) {
	ss := ctx.Snapshot()
	method := ctx.TemplateEngine().RenderLenient(string(h.Method), ss)
	u := ctx.TemplateEngine().RenderLenient(h.Url, ss)
	var (
		req     *http.Request
		reqBody io.Reader = nil
	)
	if reqBody, err = h.getReqBodyReader(ctx); err != nil {
		return err
	}
	// TODO use NewRequestWithContext, but that will require global context.Context() within the ActionContext
	// see https://github.com/rkosegi/yaml-pipeline/issues/303
	if req, err = http.NewRequest(method, u, reqBody); err != nil {
		return err
	}
	if h.Headers != nil {
		for k, vals := range *h.Headers {
			for _, v := range vals {
				req.Header.Add(
					ctx.TemplateEngine().RenderLenient(k, ss),
					ctx.TemplateEngine().RenderLenient(v, ss),
				)
			}
		}
	}
	ctx.Logger().Log(fmt.Sprintf("Requesting data from %s using %s method", u, method))
	var resp *http.Response
	if resp, err = httpRp.RoundTrip(req); err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if err = h.processResponse(ctx, resp); err != nil {
		return err
	}

	return nil
}

func (h *HttpOpSpec) getReqBodyReader(ctx ActionContext) (io.Reader, error) {
	var (
		reqBody io.Reader = nil
		err     error
	)
	if h.BodyFrom != nil && h.canHaveReqBody(h.Method) {
		if h.EncodeAs == nil {
			h.EncodeAs = new(EncodeDataAsBinary)
		}
		var node dom.Leaf
		if node, err = h.BodyFrom.Resolve(ctx); err != nil {
			return nil, err
		}
		var data []byte
		if data, err = h.EncodeAs.Encode(node); err != nil {
			return nil, err
		}
		reqBody = bytes.NewReader(data)
	}
	return reqBody, err
}

func (h *HttpOpSpec) processResponse(ctx ActionContext, resp *http.Response) (err error) {
	if !h.isStatusValid(resp) {
		return fmt.Errorf("invalid status %v", resp.Status)
	}

	if !safeBoolDeref(h.SkipBody) {
		var content []byte
		if content, err = io.ReadAll(resp.Body); err != nil {
			return err
		}
		var pa = h.ParseAs
		if h.StoreTo != nil {
			if pa == nil {
				ctx.Logger().Log("tag::warn", "parseAs was not specified, will try to parse response body as a text")
				pa = new(ParseContentModeText)
			}
			var node dom.Node
			if node, err = pa.toValue(content); err != nil {
				return err
			}
			p := (*h.StoreTo).Resolve(ctx)
			o := dom.ContainerNode()
			o.AddValue("status", dom.LeafNode(resp.StatusCode))
			o.AddValue("body", node)
			o.AddValue("headers", httpHeaders2Dom(resp.Header))
			ctx.Logger().Log(fmt.Sprintf("setting result data of %v at %v", node.Desc(), p))
			ctx.Data().Set(pp.MustParse(p), o)
		}
		if h.WriteTo != nil {
			o := (*h.WriteTo).Resolve(ctx)
			ctx.Logger().Log(fmt.Sprintf("writing to %v", o))
			if err = os.WriteFile(o, content, 0644); err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *HttpOpSpec) String() string {
	return fmt.Sprintf("Http[method=%v,URL=%s]", h.Method, h.Url)
}

func (h *HttpOpSpec) CloneWith(_ ActionContext) Action {
	return &HttpOpSpec{
		Method:           h.Method,
		Url:              h.Url,
		ValidStatusCodes: safeCopyIntSlice(h.ValidStatusCodes),
		StoreTo:          h.StoreTo,
	}
}
