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

func TestLoopOpCloneWith(t *testing.T) {
	op := &LoopOpSpec{
		Init: &ActionSpec{
			Operations: OpSpec{
				Log: &LogOpSpec{
					Message: "Ola",
				},
			},
		},
		PostAction: &ActionSpec{
			Operations: OpSpec{
				Abort: &AbortOpSpec{
					Message: "Hi",
				},
			},
		},
		Test: "{{ false }}",
	}
	op = op.CloneWith(mockEmptyActCtx()).(*LoopOpSpec)
	assert.Equal(t, "{{ false }}", op.Test)
	assert.Equal(t, "Ola", op.Init.Operations.Log.Message)
	assert.Equal(t, "Hi", op.PostAction.Operations.Abort.Message)
}

func TestLoopOpSimple(t *testing.T) {
	op := LoopOpSpec{
		Init: &ActionSpec{
			Operations: OpSpec{
				Set: &SetOpSpec{
					Data: map[string]interface{}{
						"i": 0,
					},
				},
			},
		},
		Action: ActionSpec{
			Operations: OpSpec{
				Log: &LogOpSpec{
					Message: "Iteration {{ .i }}",
				},
			},
		},
		Test: "{{ lt (.i | int) 10 }}",
		PostAction: &ActionSpec{
			Operations: OpSpec{
				Template: &TemplateOpSpec{
					Template: "{{ add .i  1 }}",
					Path:     &ValOrRef{Val: "i"},
				},
			},
		},
	}

	d := dom.ContainerNode()
	ac := newMockActBuilder().data(d).build()
	err := op.Do(ac)
	assert.NoError(t, err)
	assert.Equal(t, "10", d.Child("i").AsLeaf().Value())
}

func TestLoopOpNegative(t *testing.T) {
	var (
		err error
		op  *LoopOpSpec
	)
	op = &LoopOpSpec{
		Init: &ActionSpec{
			Operations: OpSpec{
				Abort: &AbortOpSpec{},
			},
		},
	}
	err = op.Do(mockEmptyActCtx())
	assert.Error(t, err)

	op = &LoopOpSpec{
		Test: "{{ true }}",
		Action: ActionSpec{
			Operations: OpSpec{
				Abort: &AbortOpSpec{},
			},
		},
	}
	err = op.Do(mockEmptyActCtx())
	assert.Error(t, err)

	op = &LoopOpSpec{
		PostAction: &ActionSpec{
			Operations: OpSpec{
				Abort: &AbortOpSpec{},
			},
		},
		Test:   "{{ true }}",
		Action: ActionSpec{},
	}
	err = op.Do(mockEmptyActCtx())
	assert.Error(t, err)

	op = &LoopOpSpec{
		Test:   "{{ NotAFunction }}",
		Action: ActionSpec{},
	}
	err = op.Do(mockEmptyActCtx())
	assert.Error(t, err)

}
