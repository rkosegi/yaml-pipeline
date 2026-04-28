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

func TestCallUnregistered(t *testing.T) {
	assert.Error(t, mockEmptyActCtx().Executor().Execute(&CallOpSpec{
		Name: "invalid",
	}))
}

func TestCallOp(t *testing.T) {
	var (
		ctx ActionContext
		err error
	)
	t.Run("args from - valid", func(t *testing.T) {
		ctx = mockEmptyActCtx()

		ctx.Data().Set(pp.MustParse("input.myargs.arg1"), dom.LeafNode(1))
		ctx.Data().Set(pp.MustParse("input.myargs.arg2"), dom.LeafNode("X"))
		err = ctx.Executor().Execute(&ActionSpec{
			Children: map[string]ActionSpec{
				"def": {
					ActionMeta: ActionMeta{
						Order: new(1),
					},
					Operations: OpSpec{
						Define: &DefineOpSpec{
							Name: "myprog",
							Action: ActionSpec{
								Operations: OpSpec{
									Set: &SetOpSpec{
										Data: map[string]interface{}{
											"AAA": 123,
											"BBB": "{{ .args.arg2 }}",
										},
										Render: new(true),
									},
								},
							},
						},
					},
				},
				"run": {
					ActionMeta: ActionMeta{
						Order: new(2),
					},
					Operations: OpSpec{
						Call: &CallOpSpec{
							ArgsFrom: &ValOrRef{
								Val: "input.myargs",
							},
							Name: "myprog",
						},
					},
				},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, 123, ctx.Data().Child("AAA").AsLeaf().Value())
		assert.Equal(t, "X", ctx.Data().Child("BBB").AsLeaf().Value())
	})
	t.Run("args from - invalid args from - leaf", func(t *testing.T) {
		ctx = mockEmptyActCtx()

		ctx.Data().Set(pp.MustParse("input.myargs.arg10"), dom.LeafNode(10))
		err = ctx.Executor().Execute(&ActionSpec{
			Children: map[string]ActionSpec{
				"def": {
					ActionMeta: ActionMeta{
						Order: new(1),
					},
					Operations: OpSpec{
						Define: &DefineOpSpec{
							Name: "myprog",
						},
					},
				},
				"run": {
					ActionMeta: ActionMeta{
						Order: new(2),
					},
					Operations: OpSpec{
						Call: &CallOpSpec{
							ArgsFrom: &ValOrRef{
								Val: "input.myargs.args10",
							},
							Name: "myprog",
						},
					},
				},
			},
		})
		assert.Error(t, err)
	})
}
