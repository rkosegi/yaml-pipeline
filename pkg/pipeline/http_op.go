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
	"maps"
	"mime/multipart"
	"net/http"
	"os"
	"slices"

	"github.com/rkosegi/yaml-toolkit/dom"
)

// TODO use service from registry instead
var httpRp = http.DefaultTransport
var defValidStatusCodes = []int{200, 201, 202, 204}

func (h *HttpOpSpec) canHaveReqBody(m HttpMethod) bool {
	switch m {
	case HttpMethodPOST, HttpMethodPUT, HttpMethodPATCH:
		return true
	}
	return false
}

func (h *HttpOpSpec) isStatusValid(resp *http.Response) bool {
	var vsc = &defValidStatusCodes
	if h.ValidStatusCodes != nil {
		vsc = h.ValidStatusCodes
	}
	return slices.Contains(*vsc, resp.StatusCode)
}

func (h *HttpOpSpec) Do(ctx ActionContext) (err error) {
	ss := ctx.Snapshot()
	method := ctx.TemplateEngine().RenderLenient(string(h.Method), ss)
	u := ctx.TemplateEngine().RenderLenient(h.Url, ss)
	var (
		req     *http.Request
		reqBody io.Reader = nil
	)
	var ct string
	if reqBody, ct, err = h.getReqBodyReader(ctx); err != nil {
		return err
	}
	if req, err = http.NewRequestWithContext(ctx.Context(), method, u, reqBody); err != nil {
		return err
	}
	req.Header.Set("content-type", ct)
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
	return h.processResponse(ctx, resp)
}

func (h *HttpOpSpec) getFormReader(ctx ActionContext) (io.Reader, string, error) {
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	defer func(writer *multipart.Writer) {
		_ = writer.Close()
	}(w)

	for _, k := range slices.Sorted(maps.Keys(*h.Form)) {
		hf := (*h.Form)[k]
		if err := hf.AddToMultipartForm(ctx, k, w); err != nil {
			return nil, "", err
		}
	}

	return body, w.FormDataContentType(), nil
}

func (h *HttpOpSpec) getReqBodyFromReader(ctx ActionContext) (io.Reader, string, error) {
	var err error
	if h.EncodeAs == nil {
		h.EncodeAs = new(EncodeDataAsBinary)
	}
	var node dom.Leaf
	if node, err = h.BodyFrom.Resolve(ctx); err != nil {
		return nil, h.EncodeAs.ContentType(), err
	}
	var data []byte
	if data, err = h.EncodeAs.Encode(node); err != nil {
		return nil, h.EncodeAs.ContentType(), err
	}
	return bytes.NewReader(data), h.EncodeAs.ContentType(), nil
}

func (h *HttpOpSpec) getReqBodyReader(ctx ActionContext) (io.Reader, string, error) {
	var (
		reqBody io.Reader = nil
		err     error
	)
	ct := "application/octet-stream"
	if h.canHaveReqBody(h.Method) {
		if h.BodyFrom != nil {
			return h.getReqBodyFromReader(ctx)
		}
		if h.Form != nil {
			h.Method = HttpMethodPOST
			return h.getFormReader(ctx)
		}
	}
	return reqBody, ct, err
}

func (h *HttpOpSpec) processResponse(ctx ActionContext, resp *http.Response) (err error) {
	ctx.Logger().Log(fmt.Sprintf("Got response %v", resp.StatusCode))

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
			ctx.Logger().Log(fmt.Sprintf("setting result data of type '%v' to '%v'",
				node.Desc(), p))
			defer ctx.InvalidateSnapshot()
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
