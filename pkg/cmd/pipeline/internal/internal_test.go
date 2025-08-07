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

package internal

import (
	"testing"

	"github.com/magiconair/properties/assert"
	"github.com/rkosegi/yaml-toolkit/dom"
)

func TestApplyValues(t *testing.T) {
	gd := dom.ContainerNode()
	ApplyValues(gd, []string{"a.b", "x.y[2]=hello"})
	assert.Equal(t, "", gd.Child("a").AsContainer().
		Child("b").
		AsLeaf().Value())
	assert.Equal(t, "hello", gd.Child("x").AsContainer().
		Child("y").
		AsList().Items()[2].
		AsLeaf().Value())
}

func TestApplyVarsToDom(t *testing.T) {
	gd := dom.ContainerNode()
	ApplyVarsToDom(map[string]interface{}{
		"x": 1,
		"y": "AAAA",
	}, "myvars", gd)
	assert.Equal(t, 1, gd.Child("myvars").AsContainer().Child("x").AsLeaf().Value())
	assert.Equal(t, "AAAA", gd.Child("myvars").AsContainer().Child("y").AsLeaf().Value())

	// should not panic due to nil for map
	ApplyVarsToDom(nil, "anything", gd)
}

func TestGetLogTag(t *testing.T) {
	var (
		tag   string
		found bool
	)
	tag, found = GetLogTag("", "")
	assert.Equal(t, false, found)
	assert.Equal(t, "", tag)

	tag, found = GetLogTag()
	assert.Equal(t, false, found)
	assert.Equal(t, "", tag)

	tag, found = GetLogTag("tag::skip", "")
	assert.Equal(t, true, found)
	assert.Equal(t, "skip", tag)
}
