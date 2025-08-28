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
	"testing"

	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/stretchr/testify/assert"
)

func TestExecuteTemplateOp(t *testing.T) {
	var (
		err error
		ts  TemplateOpSpec
		gd  dom.ContainerBuilder
	)

	gd = dom.ContainerNode()
	gd.AddValueAt("root.leaf1", dom.LeafNode(123456))
	ts = TemplateOpSpec{
		Template: `{{ (mul .root.leaf1 2) | quote }}`,
		Path:     &ValOrRef{Val: "result.x1"},
		Trim:     ptr(true),
	}
	assert.NoError(t, New(WithData(gd)).Execute(&ts))
	assert.Equal(t, "\"246912\"", gd.Lookup("result.x1").AsLeaf().Value())
	assert.Contains(t, ts.String(), "result.x1")

	// empty template error
	ts = TemplateOpSpec{}
	err = New(WithData(gd)).Execute(&ts)
	assert.Error(t, err)
	assert.Equal(t, ErrTemplateEmpty, err)

	// empty path error
	ts = TemplateOpSpec{
		Template: `TEST`,
	}
	err = New(WithData(gd)).Execute(&ts)
	assert.Error(t, err)
	assert.Equal(t, ErrPathEmpty, err)

	// empty path error (other form)
	ts = TemplateOpSpec{
		Template: `TEST`,
		Path:     &ValOrRef{Ref: "a.b.c.d"},
	}
	err = New(WithData(gd)).Execute(&ts)
	assert.Error(t, err)
	assert.Equal(t, ErrPathEmpty, err)

	ts = TemplateOpSpec{
		Template: `{{}}{{`,
		Path:     &ValOrRef{Val: "result"},
	}
	assert.Error(t, New(WithData(gd)).Execute(&ts))

	ts = TemplateOpSpec{
		Template: `{{ invalid_func }}`,
		Path:     &ValOrRef{Val: "result"},
	}
	assert.Error(t, New(WithData(gd)).Execute(&ts))

	t.Run("render value to path with dot", func(t *testing.T) {
		ts = TemplateOpSpec{
			Template: `OK`,
			Path:     &ValOrRef{Val: "sub.path\\.with\\.dot"},
		}
		assert.NoError(t, New(WithData(gd)).Execute(&ts))
		assert.Equal(t, "OK", gd.Child("sub").AsContainer().Child("path.with.dot").AsLeaf().Value())
	})
}

func TestExecuteTemplateOpAsYaml(t *testing.T) {
	var (
		err error
		ts  TemplateOpSpec
		gd  dom.ContainerBuilder
	)

	// 1, render yaml source manually
	gd = dom.ContainerNode()
	ts = TemplateOpSpec{
		Template: `
items:
{{- range (split "," "a,b,c") }}
{{ printf "- %s" . }}
{{- end }}`,
		Path:    &ValOrRef{Val: "Out"},
		ParseAs: ptr(ParseTextAsYaml),
	}
	err = New(WithData(gd)).Execute(&ts)
	assert.NoError(t, err)
	assert.Equal(t, 3, gd.Lookup("Out.items").AsList().Size())

	// 2, render using template function
	gd = dom.ContainerNode()
	ts = TemplateOpSpec{
		Template: `
items:
{{ (split "," "a,b,c") | list | toYaml }}
`,
		Path:    &ValOrRef{Val: "Out"},
		ParseAs: ptr(ParseTextAsYaml),
	}
	err = New(WithData(gd)).Execute(&ts)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(gd.Lookup("Out.items").AsList().Items()[0].AsContainer().Children()))

	// 3, render invalid
	gd = dom.ContainerNode()
	ts = TemplateOpSpec{
		Template: `*** this is not a YAML ***`,
		Path:     &ValOrRef{Val: "Out"},
		ParseAs:  ptr(ParseTextAsYaml),
	}
	err = New(WithData(gd)).Execute(&ts)
	assert.Error(t, err)
}

func TestExecuteTemplateOpAsInvalid(t *testing.T) {
	assert.Error(t, New().Execute(&TemplateOpSpec{
		Template: `---\nOla: Hi`,
		Path:     &ValOrRef{Val: "Out"},
		ParseAs:  ptr(ParseTextAs("invalid")),
	}))
}

func TestExecuteTemplateOpAsFloat64(t *testing.T) {
	var (
		gd  dom.ContainerBuilder
		ts  *TemplateOpSpec
		err error
	)
	gd = dom.ContainerNode()
	ts = &TemplateOpSpec{
		Template: `{{ maxf 1.5 3 4.5 }}`,
		Path:     &ValOrRef{Val: "Out"},
		ParseAs:  ptr(ParseTextAsFloat64),
	}
	err = New(WithData(gd)).Execute(ts)
	assert.NoError(t, err)
	assert.Equal(t, 4.5, gd.Lookup("Out").AsLeaf().Value())

	gd.AddValueAt("X", dom.LeafNode("Ou"))
	ts = &TemplateOpSpec{
		Template: `XYZ`,
		Path:     &ValOrRef{Val: "Out"},
		ParseAs:  ptr(ParseTextAsFloat64),
	}
	err = New(WithData(gd)).Execute(ts)
	assert.Error(t, err)
}

func TestExecuteTemplateOpAsInt64(t *testing.T) {
	var (
		gd  dom.ContainerBuilder
		ts  *TemplateOpSpec
		err error
	)
	gd = dom.ContainerNode()
	ts = &TemplateOpSpec{
		Template: `{{ max 1 3 5 }}`,
		Path:     &ValOrRef{Val: "Out"},
		ParseAs:  ptr(ParseTextAsInt64),
	}
	err = New(WithData(gd)).Execute(ts)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), gd.Lookup("Out").AsLeaf().Value())

	gd.AddValueAt("X", dom.LeafNode("Ou"))
	ts = &TemplateOpSpec{
		Template: `XYZ`,
		Path:     &ValOrRef{Val: "Out"},
		ParseAs:  ptr(ParseTextAsInt64),
	}
	err = New(WithData(gd)).Execute(ts)
	assert.Error(t, err)
}
