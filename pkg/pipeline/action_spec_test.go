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

func TestActionSpecShouldPropagateError(t *testing.T) {
	var as *ActionSpec
	as = &ActionSpec{}
	assert.Nil(t, as.ErrorPropagation)
	as = &ActionSpec{
		ErrorPropagation: new(ErrorPropagationPolicy("unknown")),
	}
	assert.True(t, as.shouldPropagateError())
	as = &ActionSpec{
		ErrorPropagation: new(ErrorPropagationPolicyIgnore),
	}
	assert.False(t, as.shouldPropagateError())

	as = &ActionSpec{
		ErrorPropagation: new(ErrorPropagationPolicyIgnore),
		Operations: OpSpec{
			Abort: &AbortOpSpec{Message: "abort"},
		},
	}
	assert.NoError(t, as.Do(newMockActBuilder().build()))
}

func TestActionSpecLifecycle(t *testing.T) {
	var (
		as *ActionSpec
		gd dom.ContainerBuilder
	)

	t.Run("success", func(t *testing.T) {
		gd = dom.ContainerNode()
		gd.AddValue("success", dom.LeafNode("no"))
		as = &ActionSpec{
			Lifecycle: &ActionLifecycle{
				OnSuccess: &ActionSpec{
					Operations: OpSpec{
						Set: &SetOpSpec{
							Data: map[string]any{
								"success": "yes",
							},
						},
					},
				},
			},
		}
		assert.NoError(t, as.Do(newMockActBuilder().data(gd).build()))
		assert.Equal(t, "yes", gd.Get(pp.MustParse("success")).AsLeaf().Value())
	})
	t.Run("failure", func(t *testing.T) {
		gd = dom.ContainerNode()
		gd.AddValue("failure", dom.LeafNode("no"))
		gd.AddValue("complete", dom.LeafNode("no"))
		as = &ActionSpec{
			Operations: OpSpec{
				Abort: &AbortOpSpec{Message: "abort"},
			},
			Lifecycle: &ActionLifecycle{
				OnComplete: &ActionSpec{
					Operations: OpSpec{
						Set: &SetOpSpec{
							Data: map[string]any{
								"complete": "yes",
							},
						},
					},
				},
				OnFailure: &ActionSpec{
					Operations: OpSpec{
						Set: &SetOpSpec{
							Data: map[string]any{
								"failure": "yes",
							},
						},
					},
				},
			},
		}
		assert.Error(t, as.Do(newMockActBuilder().data(gd).build()))
		assert.Equal(t, "yes", gd.Get(pp.MustParse("failure")).AsLeaf().Value())
		assert.Equal(t, "yes", gd.Get(pp.MustParse("complete")).AsLeaf().Value())
	})
}
