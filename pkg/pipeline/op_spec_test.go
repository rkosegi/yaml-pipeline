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

func TestOpSpecCloneWith(t *testing.T) {
	o := OpSpec{
		Set: &SetOpSpec{
			Data: map[string]interface{}{
				"a": 1,
			},
			Path: ptr("{{ .Path }}"),
		},
		Patch: &PatchOpSpec{
			Path: "{{ .Path3 }}",
		},
		ForEach: &ForEachOpSpec{
			Item: &ValOrRefSlice{&ValOrRef{Val: "left"}, &ValOrRef{Val: "right"}},
			Action: ActionSpec{
				Operations: OpSpec{},
			},
		},
		Template: &TemplateOpSpec{
			Path: &ValOrRef{Val: "{{ .Path }}"},
		},
		TemplateFile: &TemplateFileOpSpec{
			File:   "{{ .Path }}",
			Output: "{{ .Path3 }}",
		},
		Import: &ImportOpSpec{
			Path: "{{ .Path }}",
			Mode: ParseFileModeYaml,
		},
		Call: &CallOpSpec{
			Name: "invalid",
		},
		Define: &DefineOpSpec{
			Name: "def1",
			Action: ActionSpec{
				Operations: OpSpec{
					Log: &LogOpSpec{
						Message: "hello",
					},
				},
			},
		},
		Env: &EnvOpSpec{
			Path: ptr("{{ .Path }}"),
		},
		Exec: &ExecOpSpec{
			Program: "{{ .Shell }}",
		},
		Ext: &ExtOpSpec{
			Function: "noop",
		},
		Export: &ExportOpSpec{
			File:   &ValOrRef{Val: "/tmp/file.yaml"},
			Path:   &ValOrRef{Val: "{{ .Path }}"},
			Format: OutputFormatYaml,
		},
		Log: &LogOpSpec{
			Message: "Path: {{ .Path }}",
		},
		Loop: &LoopOpSpec{
			Test: "false",
			Action: ActionSpec{
				Operations: OpSpec{Log: &LogOpSpec{
					Message: "Ola!",
				}},
			},
		},
		Abort: &AbortOpSpec{
			Message: "abort",
		},
	}

	a := o.CloneWith(newMockActBuilder().data(dom.DecodeAnyToNode(map[string]interface{}{
		"Path":  "root.sub2",
		"Path3": "/root/sub3",
		"Shell": "/bin/bash",
	}).(dom.ContainerBuilder)).build()).(OpSpec)
	t.Log(a.String())
	assert.Equal(t, "root.sub2", *a.Set.Path)
	assert.Equal(t, "root.sub2", a.Import.Path)
	assert.Equal(t, "/root/sub3", a.Patch.Path)
	assert.Equal(t, "root.sub2", a.Template.Path.Val)
	assert.Equal(t, "root.sub2", a.Export.Path.Val)
	assert.Equal(t, "root.sub2", *a.Env.Path)
	assert.Equal(t, "/bin/bash", a.Exec.Program)
	assert.Equal(t, "hello", a.Define.Action.Operations.Log.Message)
	assert.Equal(t, "invalid", a.Call.Name)
}
