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
	"os"
	"strings"

	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/rkosegi/yaml-toolkit/props"
)

func (e *ExportOpSpec) String() string {
	var (
		sb    strings.Builder
		parts []string
	)
	sb.WriteString("Export[")
	if e.File != nil {
		parts = append(parts, fmt.Sprintf("file=%v", e.File))
	}
	parts = append(parts, fmt.Sprintf("format=%v", e.Format))
	if e.Path != nil {
		parts = append(parts, fmt.Sprintf("path=%v", e.Path))
	}
	sb.WriteString(strings.Join(parts, ","))
	sb.WriteByte(']')
	return sb.String()
}

func (e *ExportOpSpec) Do(ctx ActionContext) (err error) {
	var (
		d      dom.Node
		enc    dom.EncoderFunc
		defVal dom.Node
	)
	defVal = dom.ContainerNode()
	d = ctx.Data()
	if e.Path != nil {
		d = ctx.Data().Lookup(e.Path.Resolve(ctx))
	}
	switch e.Format {
	case OutputFormatYaml:
		enc = dom.DefaultYamlEncoder
	case OutputFormatJson:
		enc = dom.DefaultJsonEncoder
	case OutputFormatProperties:
		enc = props.EncoderFn
	case OutputFormatText:
		enc = func(w io.Writer, v interface{}) error {
			_, err = fmt.Fprintf(w, "%v", v)
			return err
		}
		defVal = dom.LeafNode("")

	default:
		return fmt.Errorf("unknown output format: %s", e.Format)
	}
	if d == nil {
		d = defVal
	}
	fp := e.File.Resolve(ctx)
	ctx.Logger().Log("opening file", fp)
	f, err := os.OpenFile(fp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()
	if e.Format == OutputFormatText {
		if !d.IsLeaf() {
			return fmt.Errorf("unsupported node for 'text' format: %v", d)
		}
		return enc(f, d.AsLeaf().Value())
	}
	return enc(f, dom.DefaultNodeEncoderFn(d.AsContainer()))
}

func (e *ExportOpSpec) CloneWith(ctx ActionContext) Action {
	ss := ctx.Snapshot()
	return &ExportOpSpec{
		File:   safeCloneValOrRef(e.File, ctx),
		Path:   safeCloneValOrRef(e.Path, ctx),
		Format: OutputFormat(ctx.TemplateEngine().RenderLenient(string(e.Format), ss)),
	}
}
