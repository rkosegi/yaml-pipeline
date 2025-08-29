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
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/antchfx/htmlquery"
	"github.com/rkosegi/yaml-toolkit/dom"
	"golang.org/x/net/html"
)

type (
	LayoutFn func(dom.ContainerBuilder, *html.Node)
)

var (
	layoutFnMap = map[Html2DomLayout]LayoutFn{
		Html2DomLayoutDefault: convertHtmlNode2Dom,
	}
)

const (
	AttributeNode = "Attrs"
	ValueNode     = "Value"
)

func (x *Html2DomOpSpec) String() string {
	return fmt.Sprintf("Html2Dom[from=%s,to=%s, query=%v]", x.From, x.To, x.Query)
}

func (x *Html2DomOpSpec) Do(ctx ActionContext) error {
	ss := ctx.Snapshot()
	from := ctx.TemplateEngine().RenderLenient(x.From, ss)
	to := ctx.TemplateEngine().RenderLenient(x.To, ss)
	var qry string
	if x.Query != nil {
		qry = x.Query.Resolve(ctx)
	}
	if len(from) == 0 {
		return errors.New("'from' is empty")
	}
	if len(to) == 0 {
		return errors.New("'to' is empty")
	}
	fromNode := ctx.Data().Get(pp.MustParse(from))
	if fromNode == nil || !fromNode.IsLeaf() {
		return fmt.Errorf("cannot find leaf node at '%s'", from)
	}
	htmlData := fromNode.AsLeaf().Value().(string)
	var (
		err      error
		buff     bytes.Buffer
		srcNode  *html.Node
		layout   Html2DomLayout
		layoutFn LayoutFn
		ok       bool
	)
	layout = Html2DomLayoutDefault
	if x.Layout != nil {
		layout = *x.Layout
	}
	if layoutFn, ok = layoutFnMap[layout]; !ok {
		return fmt.Errorf("unknown layout %s", layout)
	}
	_, _ = buff.WriteString(htmlData)
	// TODO: how can parse return an error?
	srcNode, _ = htmlquery.Parse(&buff)
	if len(qry) == 0 {
		qry = "/html"
	}
	srcNode, err = htmlquery.Query(srcNode, qry)
	if err != nil {
		return err
	}
	if srcNode == nil {
		return fmt.Errorf("cannot find node at %s", from)
	}
	cb := dom.ContainerNode()
	layoutFn(cb, srcNode)
	ctx.Data().AddValueAt(to, cb)
	ctx.InvalidateSnapshot()
	return nil
}

func convertHtmlNode2Dom(cb dom.ContainerBuilder, node *html.Node) {
	switch node.Type {
	case html.ElementNode:
		c := dom.ContainerNode()
		if existing := cb.Child(node.Data); existing != nil {
			if existing.IsList() {
				existing.(dom.ListBuilder).Append(c)
			} else {
				l := dom.ListNode(existing, c)
				cb.AddValue(node.Data, l)
			}
		} else {
			cb.AddValue(node.Data, c)
		}
		if len(node.Attr) > 0 {
			ac := c.AddContainer(AttributeNode)
			for _, attr := range node.Attr {
				ac.AddValue(attr.Key, dom.LeafNode(attr.Val))
			}
		}
		for child := range node.ChildNodes() {
			convertHtmlNode2Dom(c, child)
		}

	case html.TextNode:
		if val := strings.TrimSpace(node.Data); val != "" {
			cb.AddValue(ValueNode, dom.LeafNode(node.Data))
		}
	}
}

func (x *Html2DomOpSpec) CloneWith(ctx ActionContext) Action {
	ss := ctx.Snapshot()
	return &Html2DomOpSpec{
		From:   ctx.TemplateEngine().RenderLenient(x.From, ss),
		Query:  x.Query,
		To:     ctx.TemplateEngine().RenderLenient(x.To, ss),
		Layout: x.Layout,
	}
}
