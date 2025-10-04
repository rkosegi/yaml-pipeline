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

func TestSwitchOp(t *testing.T) {
	var (
		ss SwitchOpSpec
		gd dom.ContainerBuilder
	)
	ss = SwitchOpSpec{
		Cases: map[string]ActionSpec{
			"Alice": {
				Operations: OpSpec{
					Set: &SetOpSpec{
						Data: map[string]interface{}{
							"result": "Alice is winner",
						},
					},
				},
			},
			"Bob": {
				Operations: OpSpec{
					Set: &SetOpSpec{
						Data: map[string]interface{}{
							"result": "Bob is winner",
						},
					},
				},
			},
		},
		Default: &ActionSpec{
			Operations: OpSpec{
				Set: &SetOpSpec{
					Data: map[string]interface{}{
						"result": "No winner",
					},
				},
			},
		},
		Expr: ValOrRef{Val: "{{ .name }}"},
	}

	t.Log(ss)

	t.Run("should match one of cases", func(t *testing.T) {
		gd = dom.ContainerNode()
		gd.AddValue("name", dom.LeafNode("Bob"))
		assert.NoError(t, New(WithData(gd)).Execute(&ss))
		assert.Equal(t, "Bob is winner", gd.Child("result").AsLeaf().Value())
	})

	t.Run("should fallback to default as no value matched", func(t *testing.T) {
		gd = dom.ContainerNode()
		gd.AddValue("name", dom.LeafNode("Charlie"))
		assert.NoError(t, New(WithData(gd)).Execute(&ss))
		assert.Equal(t, "No winner", gd.Child("result").AsLeaf().Value())
	})

	t.Run("should fail if no cases and no default is present", func(t *testing.T) {
		spec := &SwitchOpSpec{}
		assert.Error(t, spec.Do(mockEmptyActCtx()))
	})
}

func TestSwitchOpCloneWith(t *testing.T) {

	var ss *SwitchOpSpec

	t.Run("default action should be cloned if defined", func(t *testing.T) {
		ss = &SwitchOpSpec{Expr: ValOrRef{Val: ".forEach"}, Default: &ActionSpec{
			Operations: OpSpec{
				Log: &LogOpSpec{},
			},
		}}
		out := ss.CloneWith(mockEmptyActCtx()).(*SwitchOpSpec)
		assert.NotNil(t, out.Default)
	})

	t.Run("should clone without default", func(t *testing.T) {
		ss = &SwitchOpSpec{Expr: ValOrRef{Val: ".forEach"}}
		out := ss.CloneWith(mockEmptyActCtx()).(*SwitchOpSpec)
		assert.Nil(t, out.Default)
	})

}
