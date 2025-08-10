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

func TestAbortOpDo(t *testing.T) {
	eo := &AbortOpSpec{
		Message: "conditions not met",
	}
	assert.Error(t, eo.Do(mockEmptyActCtx()))
}

func TestAbortOpCloneWith(t *testing.T) {
	eo := &AbortOpSpec{
		Message: "Unsupported format: {{ .Format }}",
	}
	d := dom.ContainerNode()
	d.AddValue("Format", dom.LeafNode("toml"))
	eo = eo.CloneWith(newMockActBuilder().data(d).build()).(*AbortOpSpec)
	assert.Equal(t, "Unsupported format: toml", eo.Message)
}

func TestAbortPipeline(t *testing.T) {
	var (
		err error
	)
	p := ActionSpec{
		ActionMeta: ActionMeta{
			When: strPointer("{{ eq .ENV \"prod\"}}"),
		},
		Operations: OpSpec{
			Abort: &AbortOpSpec{
				Message: "Pipeline should not run in production",
			},
		},
	}
	d := dom.ContainerNode()
	d.AddValue("ENV", dom.LeafNode("prod"))
	err = newTestExec(d).Execute(p)
	assert.Error(t, err)
	d.AddValue("ENV", dom.LeafNode("dev"))
	err = newTestExec(d).Execute(p)
	assert.NoError(t, err)
}
