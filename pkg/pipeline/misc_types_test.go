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
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestValOrRefMarshalYaml(t *testing.T) {
	t.Run("Marshal ref", func(t *testing.T) {
		var (
			out  bytes.Buffer
			node yaml.Node
		)
		assert.NoError(t, yaml.NewEncoder(&out).Encode(&ValOrRef{Ref: "path.to.elem"}))
		assert.NoError(t, yaml.NewDecoder(&out).Decode(&node))
		assert.Equal(t, yaml.MappingNode, node.Content[0].Kind)
	})
	t.Run("Marshal direct value", func(t *testing.T) {
		var (
			out  bytes.Buffer
			node yaml.Node
		)
		assert.NoError(t, yaml.NewEncoder(&out).Encode(&ValOrRef{Val: "abc"}))
		assert.NoError(t, yaml.NewDecoder(&out).Decode(&node))
		assert.Equal(t, yaml.ScalarNode, node.Content[0].Kind)
	})
}
