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

	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/rkosegi/yaml-toolkit/props"
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

func (ia *ImportOpSpec) Do(ctx ActionContext) error {
	file := ctx.TemplateEngine().RenderLenient(ia.File, ctx.Snapshot())
	ctx.Logger().Log(fmt.Sprintf("Importing file %s using mode %s", file, ia.Mode))
	val, err := parseFile(file, ia.Mode)
	if err != nil {
		return err
	}
	p := ctx.TemplateEngine().RenderLenient(ia.Path, ctx.Snapshot())
	if len(p) > 0 {
		ctx.Data().AddValueAt(p, val)
		ctx.InvalidateSnapshot()
	} else {
		if !val.IsContainer() {
			return ErrNotContainer
		} else {
			for k, v := range val.AsContainer().Children() {
				ctx.Data().AddValueAt(k, v)
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
