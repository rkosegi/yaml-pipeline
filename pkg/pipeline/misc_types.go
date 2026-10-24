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
	"errors"
	"fmt"
	"strings"
	"time"

	te "github.com/rkosegi/yaml-pipeline/pkg/pipeline/template_engine"
	"github.com/rkosegi/yaml-toolkit/dom"
	"gopkg.in/yaml.v3"
)

var (
	ErrNoDataToSet   = errors.New("no data to set")
	ErrTemplateEmpty = errors.New("template cannot be empty")
	ErrPathEmpty     = errors.New("path cannot be empty")
	ErrFileEmpty     = errors.New("file cannot be empty")
	ErrOutputEmpty   = errors.New("output cannot be empty")
	ErrNotContainer  = errors.New("data element must be container when no path is provided")
)

func errWithInfo(err error, info string) error {
	return fmt.Errorf("%s: %w", info, err)
}

// AnyVal can represent any DOM value (leaf, list, container)
type AnyVal struct {
	v dom.Node
}

func (pv *AnyVal) UnmarshalYAML(node *yaml.Node) error {
	pv.v = dom.YamlNodeDecoder()(node)
	return nil
}

// Value get actual value
func (pv *AnyVal) Value() dom.Node {
	return pv.v
}

// ValOrRef is either immediate leaf value or reference to a dom.Leaf in data tree at given path.
// If Path references non-existent node, or node pointed to is not a dom.Leaf, empty value is returned
type ValOrRef struct {
	isRef bool
	// Ref is resolved reference, if any
	Ref string `yaml:"ref,omitempty" json:"ref,omitempty"`
	// Val is value of dom.Leaf pointed to by ref after Resolve(ctx) has been called, or immediate value
	// if Ref is empty
	Val string `yaml:"val,omitempty" json:"val,omitempty"`
}

func (pv *ValOrRef) CloneWith(ctx ActionContext) *ValOrRef {
	return &ValOrRef{
		isRef: pv.isRef,
		Ref:   ctx.TemplateEngine().RenderLenient(pv.Ref, ctx.Snapshot()),
		Val:   ctx.TemplateEngine().RenderLenient(pv.Val, ctx.Snapshot()),
	}
}

func (pv *ValOrRef) MarshalYAML() (interface{}, error) {
	if len(pv.Ref) > 0 {
		return map[string]interface{}{
			"ref": pv.Ref,
		}, nil
	}
	return pv.Val, nil
}

func (pv *ValOrRef) check() error {
	if len(pv.Val) == 0 && len(pv.Ref) == 0 {
		return errors.New("missing 'val' or 'ref' field")
	}
	return nil
}

func (pv *ValOrRef) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.MappingNode:
		m := make(map[string]interface{})
		_ = node.Decode(&m)
		if x, ok := m["ref"]; !ok {
			return errors.New("missing 'ref' field")
		} else {
			pv.isRef = true
			pv.Ref = x.(string)
		}
		return pv.check()

	case yaml.ScalarNode:
		pv.Val = node.Value
		return pv.check()
	}

	return fmt.Errorf("invalid value: '%v'", node)
}

func (pv *ValOrRef) Resolve(ctx ActionContext) string {
	if pv == nil {
		return ""
	}
	ss := ctx.Snapshot()
	if pv.isRef {
		if n := ctx.Data().Get(pp.MustParse(pv.Ref)); n != nil && n.IsLeaf() {
			v := fmt.Sprintf("%v", n.AsLeaf().Value())
			return ctx.TemplateEngine().RenderLenient(v, ss)
		}
		return ""
	}
	return ctx.TemplateEngine().RenderLenient(pv.Val, ss)
}

func (pv *ValOrRef) String() string {
	var (
		sb    strings.Builder
		parts []string
	)
	sb.WriteByte('[')
	if len(pv.Ref) > 0 {
		parts = append(parts, fmt.Sprintf("Ref=%v", pv.Ref))
	}
	if len(pv.Val) > 0 {
		parts = append(parts, fmt.Sprintf("Val=%v", pv.Val))
	}
	sb.WriteString(strings.Join(parts, ","))
	sb.WriteByte(']')
	return sb.String()
}

type ValOrRefSlice []*ValOrRef

func (pv *ValOrRefSlice) String() string {
	var strs []string
	for _, val := range *pv {
		strs = append(strs, val.String())
	}
	return "[" + strings.Join(strs, ",") + "]"
}

func anyValFromMap(m map[string]interface{}) *AnyVal {
	return &AnyVal{v: dom.DecodeAnyToNode(m)}
}

// ChildActions is map of named actions that are executed as a part of parent action
type ChildActions map[string]ActionSpec

// StrKeysStrValues is key-value map with strings.
type StrKeysStrValues map[string]string

// AsAnyValuesMap converts this map to StrKeysAnyValues suitable for template related operations.
func (skv StrKeysStrValues) AsAnyValuesMap() StrKeysAnyValues {
	out := make(StrKeysAnyValues, len(skv))
	for k, v := range skv {
		out[k] = v
	}
	return out
}

// RenderValues renders values of this map using provided TemplateEngine and data.
func (skv StrKeysStrValues) RenderValues(teng te.TemplateEngine, data StrKeysAnyValues) StrKeysStrValues {
	out := make(map[string]string, len(skv))
	for k, v := range skv {
		out[k] = teng.RenderLenient(v, data)
	}
	return out
}

// StrKeysAnyValues is a map with string keys and values with any type
type StrKeysAnyValues map[string]any

type RetryConfig struct {
	// Retry policy to use
	Policy RetryPolicy `json:"policy" yaml:"policy"`
	// Interval between retries when policy is "fixed"
	FixedInterval *time.Duration `json:"interval,omitempty" yaml:"interval,omitempty"`
	// Backoff policy configuration
	Backoff *RetryBackoffSpec `json:"backoff,omitempty" yaml:"backoff,omitempty"`
}

func (rc *RetryConfig) ShouldRetry(ctx ActionContext) bool {
	return false
}

func (rc *RetryConfig) Wait(ctx ActionContext) {

}

func ensureLeafNode(node dom.Node) (dom.Leaf, error) {
	if !node.IsLeaf() {
		return nil, fmt.Errorf("expected dom.Leaf() node, got %v", node.Desc())
	}
	return node.AsLeaf(), nil
}

func encodeStringLikeNode(node dom.Node) ([]byte, error) {
	leaf, err := ensureLeafNode(node)
	if err != nil {
		return nil, err
	}
	switch x := leaf.Value().(type) {
	case string:
		return []byte(x), nil
	case []byte:
		return x, nil
	default:
		return nil, fmt.Errorf("unsupported value type %T: %v", leaf.Value(), leaf.Value())
	}
}

func (ed EncodeDataAs) Encode(node dom.Node) ([]byte, error) {
	if node == nil {
		return nil, nil
	}
	var (
		buff bytes.Buffer
		err  error
	)
	switch ed {
	case EncodeDataAsBinary:
		return encodeStringLikeNode(node)
	case EncodeDataAsJson:
		err = dom.DefaultJsonEncoder(&buff, node)
		return buff.Bytes(), err
	case EncodeDataAsYaml:
		err = dom.DefaultYamlEncoder(&buff, node)
		return buff.Bytes(), err

	default:
		return nil, fmt.Errorf("unknown encode type '%v'", ed)
	}
}
