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
	"regexp"
	"testing"

	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/stretchr/testify/assert"
)

func createRePtr(in string) *regexp.Regexp {
	x := regexp.MustCompile(in)
	return x
}

func TestEnvOpDo(t *testing.T) {
	envGetter = func() []string {
		return []string{
			"MOCK1=val1",
			"MOCK2=val2",
			"XYZ=123",
		}
	}

	eo := EnvOp{
		Path:    "Sub",
		Include: createRePtr(`MOCK\d+`),
		Exclude: createRePtr("XYZ"),
	}
	d := dom.ContainerNode()
	err := eo.Do(newMockActBuilder().data(d).build())
	assert.NoError(t, err)
	assert.Equal(t, "val1", d.Lookup("Sub.Env.MOCK1").AsLeaf().Value())
	assert.Equal(t, "val2", d.Lookup("Sub.Env.MOCK2").AsLeaf().Value())
	assert.Contains(t, eo.String(), "Sub")
}

func TestEnvOpCloneWith(t *testing.T) {
	eo := &EnvOp{
		Path: "{{ .NewPath }}",
	}
	d := dom.ContainerNode()
	d.AddValue("NewPath", dom.LeafNode("root"))
	eo = eo.CloneWith(newMockActBuilder().data(d).build()).(*EnvOp)
	assert.Equal(t, "root", eo.Path)
}
