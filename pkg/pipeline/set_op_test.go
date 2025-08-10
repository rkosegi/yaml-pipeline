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

func TestExecuteSetOp(t *testing.T) {
	var (
		ss  SetOpSpec
		gd  dom.ContainerBuilder
		err error
	)
	ss = SetOpSpec{
		Data: map[string]interface{}{
			"sub1": 123,
		},
	}
	gd = dom.ContainerNode()
	assert.NoError(t, New(WithData(gd)).Execute(&ss))
	assert.Equal(t, 123, gd.Lookup("sub1").AsLeaf().Value())

	ss = SetOpSpec{
		Data: map[string]interface{}{
			"sub1": 123,
		},
		Path: ptr("sub0"),
	}
	gd = dom.ContainerNode()
	assert.NoError(t, New(WithData(gd)).Execute(&ss))
	assert.Equal(t, 123, gd.Lookup("sub0.sub1").AsLeaf().Value())
	assert.Contains(t, ss.String(), "sub0")

	ss = SetOpSpec{}
	err = New(WithData(gd)).Execute(&ss)
	assert.Error(t, err)
	assert.Equal(t, ErrNoDataToSet, err)
}

func TestSetOpInvalidSetStrategy(t *testing.T) {
	assert.Error(t, New().Execute(&SetOpSpec{
		Data:     map[string]interface{}{},
		Strategy: setStrategyPointer("unknown"),
	}))
}

func TestSetOpMergeRoot(t *testing.T) {
	var (
		gd dom.ContainerBuilder
		ss SetOpSpec
	)

	ss = SetOpSpec{
		Data: map[string]interface{}{
			"sub1": 123,
		},
		Strategy: setStrategyPointer(SetStrategyMerge),
	}
	gd = dom.ContainerNode()
	gd.AddValue("sub2", dom.LeafNode(1))
	assert.NoError(t, New(WithData(gd)).Execute(&ss))
	assert.Equal(t, 123, gd.Lookup("sub1").AsLeaf().Value())
	assert.Equal(t, 2, len(gd.Children()))

	gd = dom.ContainerNode()
	gd.AddValueAt("sub2.sub3a", dom.LeafNode(2))
	ss = SetOpSpec{
		Data: map[string]interface{}{
			"sub2": map[string]interface{}{
				"sub3b": 123,
			},
		},
		Strategy: setStrategyPointer(SetStrategyMerge),
	}
	assert.NoError(t, New(WithData(gd)).Execute(&ss))
	assert.Equal(t, 2, len(gd.Lookup("sub2").AsContainer().Children()))
	assert.Equal(t, 2, gd.Lookup("sub2.sub3a").AsLeaf().Value())
	assert.Equal(t, 123, gd.Lookup("sub2.sub3b").AsLeaf().Value())
}

func TestSetOpMergeSubPath(t *testing.T) {
	var (
		gd dom.ContainerBuilder
		ss SetOpSpec
	)

	ss = SetOpSpec{
		Data: map[string]interface{}{
			"sub20": 123,
		},
		Strategy: setStrategyPointer(SetStrategyMerge),
		Path:     ptr("sub10"),
	}
	gd = dom.ContainerNode()

	assert.NoError(t, New(WithData(gd)).Execute(&ss))
	assert.Equal(t, 123, gd.Lookup("sub10.sub20").AsLeaf().Value())

	gd = dom.ContainerNode()
	gd.AddValueAt("sub10.sub20.sub30", dom.LeafNode(2))
	ss = SetOpSpec{
		Data: map[string]interface{}{
			"sub20": map[string]interface{}{
				"sub3b": 123,
			},
		},
		Path:     ptr("sub10"),
		Strategy: setStrategyPointer(SetStrategyMerge),
	}
	assert.NoError(t, New(WithData(gd)).Execute(&ss))
	assert.Equal(t, 2, len(gd.Lookup("sub10.sub20").AsContainer().Children()))
	assert.Equal(t, 2, gd.Lookup("sub10.sub20.sub30").AsLeaf().Value())
	assert.Equal(t, 123, gd.Lookup("sub10.sub20.sub3b").AsLeaf().Value())
}
