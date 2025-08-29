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

func TestForeachCloneWith(t *testing.T) {
	op := ForEachOpSpec{
		Item: &ValOrRefSlice{&ValOrRef{Val: "a"}, &ValOrRef{Val: "b"}, &ValOrRef{Val: "c"}},
		Action: ActionSpec{
			Operations: OpSpec{},
		},
	}
	a := op.CloneWith(mockEmptyActCtx()).(*ForEachOpSpec)
	assert.NotNil(t, a)
	assert.Equal(t, 3, len(*a.Item))
}

func TestForeachStringItem(t *testing.T) {
	op := ForEachOpSpec{
		Item: &ValOrRefSlice{&ValOrRef{Val: "a"}, &ValOrRef{Val: "b"}, &ValOrRef{Val: "c"}},
		Action: ActionSpec{
			Operations: OpSpec{
				Set: &SetOpSpec{
					Path: ptr("{{ .forEach }}"),
					Data: map[string]interface{}{
						"X": "abc",
					},
				},
				Env: &EnvOpSpec{},
				Export: &ExportOpSpec{
					File:   &ValOrRef{Val: "/tmp/a-{{ .forEach }}.yaml"},
					Format: OutputFormatYaml,
				},
				Exec: &ExecOpSpec{
					Program: "sh",
					Args:    &[]string{"-c", "rm -f /tmp/a-{{ .forEach }}.yaml"},
				},
				Log: &LogOpSpec{
					Message: "Hi {{ .forEach }}",
				},
				TemplateFile: &TemplateFileOpSpec{
					File:   "../../testdata/simple.template",
					Output: "/tmp/abc.out",
				},
				Loop: &LoopOpSpec{
					Test: "false",
					Action: ActionSpec{
						Operations: OpSpec{Log: &LogOpSpec{
							Message: "Ola!",
						}},
					},
				},
			},
		},
	}
	d := dom.ContainerNode()
	err := op.Do(newMockActBuilder().data(d).build())
	assert.NoError(t, err)
	assert.Equal(t, "abc", d.Get(pp.MustParse("a.X")).AsLeaf().Value())
	assert.Equal(t, "abc", d.Get(pp.MustParse("b.X")).AsLeaf().Value())
	assert.Equal(t, "abc", d.Get(pp.MustParse("c.X")).AsLeaf().Value())
}

func TestForeachStringItemChildError(t *testing.T) {
	op := ForEachOpSpec{
		Item: &ValOrRefSlice{&ValOrRef{Val: "a"}, &ValOrRef{Val: "b"}, &ValOrRef{Val: "c"}},
		Action: ActionSpec{
			Operations: OpSpec{
				Set: &SetOpSpec{
					Path: ptr("{{ .forEach }}"),
				},
			},
		},
	}
	d := dom.ContainerNode()
	err := op.Do(newMockActBuilder().data(d).build())
	assert.Error(t, err)
}

func TestForeachQuery(t *testing.T) {
	type testcase struct {
		qry        string
		tmpl       string
		variable   string
		path       string
		validateFn func(data dom.ContainerBuilder)
	}
	data := dom.DecodeAnyToNode(map[string]interface{}{
		"leaf": "X",
		"sub": map[string]interface{}{
			"leaf1": "Y",
		},
		"items": []interface{}{"a", "b", "c"},
	}).(dom.ContainerBuilder)
	for _, qry := range []string{
		"leaf", "sub", "items",
	} {
		op := &ForEachOpSpec{
			Query: &ValOrRef{Val: qry},
			Action: ActionSpec{
				Operations: OpSpec{
					Abort: &AbortOpSpec{},
				},
			},
		}
		assert.Error(t, op.Do(newMockActBuilder().data(data).build()))
	}
	for _, tc := range []testcase{
		{
			qry:      "leaf",
			tmpl:     "{{ .forEach }}",
			variable: "forEach",
			path:     "Result",
			validateFn: func(d dom.ContainerBuilder) {
				assert.Equal(t, "X", d.Child("Result").AsLeaf().Value())
			},
		},
		{
			validateFn: func(d dom.ContainerBuilder) {
				assert.Equal(t, "Y", d.Child("Result").AsLeaf().Value())
			},
			qry:      "sub",
			tmpl:     "{{ get .sub .forEach }}",
			variable: "forEach",
			path:     "Result",
		},
		{
			validateFn: func(d dom.ContainerBuilder) {
				assert.Equal(t, "c", d.Child("Result").AsLeaf().Value())
			},
			qry:      "items",
			tmpl:     "{{ .XYZ }}",
			variable: "XYZ",
			path:     "Result",
		},
	} {
		t.Run(tc.qry, func(t *testing.T) {
			op := &ForEachOpSpec{
				Variable: ptr(tc.variable),
				Query:    &ValOrRef{Val: tc.qry},
				Action: ActionSpec{
					Operations: OpSpec{
						Template: &TemplateOpSpec{
							Template: tc.tmpl,
							Path:     &ValOrRef{Val: tc.path},
						},
					},
				},
			}
			assert.NoError(t, op.Do(newMockActBuilder().data(data).build()))
			assert.NotNil(t, op.String())
			tc.validateFn(data)
		})
	}
}

func TestForeachGlob(t *testing.T) {
	op := ForEachOpSpec{
		Glob: &ValOrRef{Val: "../../testdata/doc?.yaml"},
		Action: ActionSpec{
			Operations: OpSpec{
				Import: &ImportOpSpec{
					File: "{{ .forEach }}",
					Path: "import.files.{{ b64enc (osBase .forEach) }}",
					Mode: ParseFileModeYaml,
				},
			},
		},
	}
	d := dom.ContainerNode()
	err := op.Do(newMockActBuilder().data(d).build())
	assert.NoError(t, err)
	assert.Equal(t, 2, len(d.Get(pp.MustParse("import.files")).AsContainer().Children()))
}

func TestForeachActionSpec(t *testing.T) {
	var (
		err error
		op  *ForEachOpSpec
	)
	op = &ForEachOpSpec{
		Item: &ValOrRefSlice{&ValOrRef{Val: "a"}, &ValOrRef{Val: "b"}, &ValOrRef{Val: "c"}},
		Action: ActionSpec{
			Children: map[string]ActionSpec{
				"sub": {
					Operations: OpSpec{
						Log: &LogOpSpec{
							Message: "Hi {{ .forEach }}",
						},
					},
				},
			},
		},
	}
	err = op.Do(newMockActBuilder().testLogger(t).build())
	assert.NoError(t, err)

	op = &ForEachOpSpec{
		Item: &ValOrRefSlice{&ValOrRef{Val: "a"}, &ValOrRef{Val: "b"}, &ValOrRef{Val: "c"}},
		Action: ActionSpec{
			Children: map[string]ActionSpec{
				"sub": {
					Operations: OpSpec{
						Template: &TemplateOpSpec{
							Path:     &ValOrRef{Val: "X"},
							Template: "{{ add .X 1 }}",
						},
					},
				},
			},
		},
	}
	d := dom.ContainerNode()
	d.AddValue("X", dom.LeafNode(100))
	err = op.Do(newMockActBuilder().data(d).build())
	assert.NoError(t, err)
	assert.Equal(t, "103", d.Child("X").AsLeaf().Value())
}

func TestForeachGlobChildError(t *testing.T) {
	op := ForEachOpSpec{
		Glob: &ValOrRef{Val: "../../testdata/doc?.yaml"},
		Action: ActionSpec{
			Operations: OpSpec{
				Set: &SetOpSpec{
					Path: ptr("{{ .forEach }}"),
				},
			},
		},
	}
	err := op.Do(newMockActBuilder().testLogger(t).build())
	assert.Error(t, err)
}

func TestForeachGlobInvalid(t *testing.T) {
	op := ForEachOpSpec{
		Glob: &ValOrRef{Val: "[]]"},
		Action: ActionSpec{
			Operations: OpSpec{
				Import: &ImportOpSpec{
					File: "{{ .forEach }}",
					Path: "import.files.{{ b64enc (osBase .forEach) }}",
					Mode: ParseFileModeYaml,
				},
			},
		},
	}
	d := dom.ContainerNode()
	err := op.Do(newMockActBuilder().data(d).build())
	assert.Error(t, err)
}
