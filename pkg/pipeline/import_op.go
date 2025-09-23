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
	"encoding/base64"
	"fmt"

	"github.com/antchfx/htmlquery"
	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/rkosegi/yaml-toolkit/fluent"
	"github.com/rkosegi/yaml-toolkit/props"
	"golang.org/x/net/html"
)

const (
	ImportOpXmlAttributeNode = "Attrs"
	ImportOpXmlValueNode     = "Value"
)

type (
	ImportOpXmlLayoutFn func(dom.ContainerBuilder, *html.Node)
)

var (
	importOpXmlLayoutFnMap = map[XmlLayout]ImportOpXmlLayoutFn{
		XmlLayoutDefault: convertHtmlNode2Dom,
	}
	importOpXmlDefOpts = &XmlImportOptions{Layout: ptr(XmlLayoutDefault), Query: &ValOrRef{Val: "/html"}}
)

func (pfm ParseFileMode) toValue(content []byte) (dom.Node, error) {
	switch pfm {
	case ParseFileModeBinary:
		return dom.LeafNode(base64.StdEncoding.EncodeToString(content)), nil
	case ParseFileModeText:
		return dom.LeafNode(string(content)), nil
	case ParseFileModeYaml:
		return dom.DecodeReader(bytes.NewReader(content), dom.DefaultYamlDecoder)
	case ParseFileModeJson:
		return dom.DecodeReader(bytes.NewReader(content), dom.DefaultJsonDecoder)
	case ParseFileModeProperties:
		return dom.DecodeReader(bytes.NewReader(content), props.DecoderFn)
	default:
		return nil, fmt.Errorf("invalid ParseFileMode: %v", pfm)
	}
}

func (ia *ImportOpSpec) String() string {
	return fmt.Sprintf("Import[file=%s,path=%s,mode=%s]", ia.File, ia.Path, ia.Mode)
}

func (ia *ImportOpSpec) loadXml(content string, ctx ActionContext) (dom.Container, error) {
	var (
		err      error
		buff     bytes.Buffer
		srcNode  *html.Node
		layout   Html2DomLayout
		layoutFn ImportOpXmlLayoutFn
		ok       bool
	)
	opts := fluent.NewConfigHelper[XmlImportOptions]().
		Add(importOpXmlDefOpts).
		Add(ia.Xml).Result()

	qry := opts.Query.Resolve(ctx)
	if len(qry) == 0 {
		qry = "/html"
	}

	if layoutFn, ok = importOpXmlLayoutFnMap[*opts.Layout]; !ok {
		return nil, fmt.Errorf("unknown layout %s", layout)
	}
	_, _ = buff.WriteString(content)
	srcNode, _ = htmlquery.Parse(&buff)
	srcNode, err = htmlquery.Query(srcNode, qry)
	if err != nil {
		return nil, err
	}
	if srcNode == nil {
		return nil, fmt.Errorf("cannot find node using %s", qry)
	}
	cb := dom.ContainerNode()
	layoutFn(cb, srcNode)
	return cb, err
}

func (ia *ImportOpSpec) Do(ctx ActionContext) error {
	p := ctx.TemplateEngine().RenderLenient(ia.Path, ctx.Snapshot())
	file := ctx.TemplateEngine().RenderLenient(ia.File, ctx.Snapshot())
	ctx.Logger().Log(fmt.Sprintf("Importing file %s using mode %s into %s", file, ia.Mode, p))
	var (
		val dom.Node
		err error
	)
	if ia.Mode == ParseFileModeXml {
		val, err = parseFile(file, ParseFileModeText)
	} else {
		val, err = parseFile(file, ia.Mode)
	}
	if err != nil {
		return errWithInfo(err, "import/parseFile")
	}
	if ia.Mode == ParseFileModeXml {
		if val, err = ia.loadXml(val.AsLeaf().Value().(string), ctx); err != nil {
			return errWithInfo(err, "import/loadXml")
		}
	}
	if len(p) > 0 {
		ctx.Data().Set(pp.MustParse(p), val)
		ctx.InvalidateSnapshot()
	} else {
		if !val.IsContainer() {
			return ErrNotContainer
		} else {
			for k, v := range val.AsContainer().Children() {
				ctx.Data().Set(pp.MustParse(k), v)
				ctx.InvalidateSnapshot()
			}
		}
	}
	return nil
}

func (ia *ImportOpSpec) CloneWith(ctx ActionContext) Action {
	return &ImportOpSpec{
		Mode: ia.Mode,
		Path: ctx.TemplateEngine().RenderLenient(ia.Path, ctx.Snapshot()),
		File: ctx.TemplateEngine().RenderLenient(ia.File, ctx.Snapshot()),
	}
}
