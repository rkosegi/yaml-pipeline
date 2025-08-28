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
	"strconv"
	"strings"

	"github.com/rkosegi/yaml-toolkit/dom"
	"gopkg.in/yaml.v3"
)

func (ts *TemplateOpSpec) String() string {
	return fmt.Sprintf("Template[Path=%s]", ts.Path)
}

func (ts *TemplateOpSpec) Do(ctx ActionContext) error {
	if len(ts.Template) == 0 {
		return ErrTemplateEmpty
	}
	if ts.Path == nil {
		return ErrPathEmpty
	}
	p := ts.Path.Resolve(ctx)
	if len(p) == 0 {
		return ErrPathEmpty
	}
	ss := ctx.Snapshot()
	val, err := ctx.TemplateEngine().Render(ts.Template, ss)
	if err != nil {
		return err
	}
	if safeBoolDeref(ts.Trim) {
		val = strings.TrimSpace(val)
	}
	if ts.ParseAs == nil {
		ts.ParseAs = ptr(ParseTextAsNone)
	}
	var node dom.Node
	switch *ts.ParseAs {
	case ParseTextAsYaml:
		var yn yaml.Node
		if err = yaml.Unmarshal([]byte(val), &yn); err != nil {
			return err
		}
		node = dom.YamlNodeDecoder()(&yn)
	case ParseTextAsNone:
		node = dom.LeafNode(val)
	case ParseTextAsFloat64:
		var x float64
		if x, err = strconv.ParseFloat(val, 64); err != nil {
			return err
		} else {
			node = dom.LeafNode(x)
		}
	case ParseTextAsInt64:
		var x int64
		if x, err = strconv.ParseInt(val, 10, 64); err != nil {
			return err
		} else {
			node = dom.LeafNode(x)
		}
	default:
		return fmt.Errorf("unknown ParseAs mode: %v", *ts.ParseAs)
	}
	ctx.Data().Set(pp.MustParse(p), node)
	ctx.InvalidateSnapshot()
	return err
}

func (ts *TemplateOpSpec) CloneWith(ctx ActionContext) Action {
	return &TemplateOpSpec{
		Template: ts.Template,
		Trim:     ts.Trim,
		Path:     safeCloneValOrRef(ts.Path, ctx),
	}
}
