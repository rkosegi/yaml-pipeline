/*
Copyright 2025 Richard Kosegi

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
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"

	"github.com/rkosegi/yaml-toolkit/dom"
)

func (h *HttpOpSpec) canHaveReqBody(m string) bool {
	switch HttpMethod(m) {
	case HttpMethodPOST, HttpMethodPUT:
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
	if h.BodyFrom != nil && h.canHaveReqBody(method) {
		// TODO read and encode request body
	}
	// TODO use NewRequestWithContext, but that will require global context.Context() within the ActionContext
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
	if resp, err = http.DefaultClient.Do(req); err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if !h.isStatusValid(resp) {
		return fmt.Errorf("invalid status %v", resp.Status)
	}

	if !safeBoolDeref(h.SkipBody) {
		var content []byte
		if content, err = io.ReadAll(resp.Body); err != nil {
			return err
		}
		if h.StoreTo != nil {
			var node dom.Node
			if node, err = h.ParseAs.toValue(content); err != nil {
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
