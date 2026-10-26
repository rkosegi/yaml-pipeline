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
	"io"
	"os"
	"strings"

	te "github.com/rkosegi/yaml-pipeline/pkg/pipeline/template_engine"
	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/rkosegi/yaml-toolkit/path"
	"gopkg.in/yaml.v3"
)

var (
	ErrNoDataToSet   = errors.New("no data to set")
	ErrTemplateEmpty = errors.New("template cannot be empty")
	ErrPathEmpty     = errors.New("path cannot be empty")
	ErrFileEmpty     = errors.New("file cannot be empty")
	ErrOutputEmpty   = errors.New("output cannot be empty")
	ErrNotContainer  = errors.New("data element must be container when no path is provided")
	ErrWriterNoInit  = errors.New("underlying writer has not been initialized")

	nwc  = &nopWriteCloser{io.Discard}
	nosd = &StreamDestination{}
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

func (pv *ValOrRef) MarshalYAML() (any, error) {
	if len(pv.Ref) > 0 {
		return map[string]any{
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
		m := make(map[string]any)
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

func anyValFromMap(m map[string]any) *AnyVal {
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
		err = dom.DefaultJsonEncoder(&buff, node.AsAny())
		return buff.Bytes(), err
	case EncodeDataAsYaml:
		err = dom.DefaultYamlEncoder(&buff, node.AsAny())
		return buff.Bytes(), err

	default:
		return nil, fmt.Errorf("unknown encode type '%v'", ed)
	}
}

func (s *SourceOfData) Resolve(ctx ActionContext) (dom.Leaf, error) {
	if s.Tree != nil {
		return dom.LeafNode(s.Tree.Resolve(ctx)), nil
	}
	if s.File != nil {
		data, err := os.ReadFile(s.File.Resolve(ctx))
		if err != nil {
			return nil, err
		}
		return dom.LeafNode(string(data)), nil
	}
	return nil, errors.New("neither 'tree', nor 'file' was specified")
}

type nopWriteCloser struct {
	io.Writer
}

func (nopWriteCloser) Close() error {
	return nil
}

// StreamDestination specifies where to put stream content.
// Content is updated only after stream is closed.
type StreamDestination struct {
	// Tree specifies destination within the global data tree.
	// Append is supported only for dom.Leaf's value of string
	Tree *ValOrRef `json:"tree,omitempty" yaml:"tree,omitempty"`

	// Path to the file on local filesystem. If file can't be opened for writing, operation will fail
	File *ValOrRef `json:"file,omitempty" yaml:"file,omitempty"`

	// When true, then stream content will be appended to the end of existing content.
	Append bool `json:"append,omitempty" yaml:"append,omitempty"`

	wc io.WriteCloser
}

type fileWriter struct {
	fpath  string
	append bool
	file   *os.File
}

func (f *fileWriter) Write(p []byte) (n int, err error) {
	if f.file != nil {
		return f.file.Write(p)
	}
	return 0, ErrWriterNoInit
}

func (f *fileWriter) Close() error {
	if f.file != nil {
		defer func() { f.file = nil }()
		return f.file.Close()
	}
	return nil
}

func (f *fileWriter) init() (err error) {
	flags := os.O_WRONLY | os.O_CREATE
	if f.append {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
	}
	f.file, err = os.OpenFile(f.fpath, flags, 0644)
	return
}

type domLeafNodeWriter struct {
	leafPath path.Path
	rootNode dom.ContainerBuilder
	buf      *bytes.Buffer
	append   bool
}

func (d *domLeafNodeWriter) Write(p []byte) (n int, err error) {
	return d.buf.Write(p)
}

func (d *domLeafNodeWriter) Close() error {
	x := d.buf.String()
	if d.append {
		if en := d.rootNode.Get(d.leafPath); en != nil && en.IsLeaf() {
			if str, ok := en.AsLeaf().Value().(string); ok {
				x = str + x
			}
		}
	}
	d.rootNode.Set(d.leafPath, dom.LeafNode(x))
	return nil
}

func (s *StreamDestination) Write(p []byte) (n int, err error) {
	if s.wc != nil {
		return s.wc.Write(p)
	}
	return 0, ErrWriterNoInit
}

func (s *StreamDestination) Close() (err error) {
	if s.wc != nil {
		err = s.wc.Close()
		s.wc = nil
	}
	return
}

func (s *StreamDestination) Resolve(ctx ActionContext) error {
	if s.Tree != nil {
		s.wc = &domLeafNodeWriter{
			rootNode: ctx.Data(),
			leafPath: pp.MustParse(s.Tree.Resolve(ctx)),
			buf:      new(bytes.Buffer),
			append:   s.Append,
		}
		return nil
	} else if s.File != nil {
		fw := &fileWriter{fpath: s.File.Resolve(ctx), append: s.Append}
		s.wc = fw
		return fw.init()
	}

	s.wc = nwc
	return nil
}

func safeCloneStreamDestination(in *StreamDestination, ctx ActionContext) *StreamDestination {
	if in == nil {
		return nosd
	}
	out := &StreamDestination{
		Tree:   safeCloneValOrRef(in.Tree, ctx),
		File:   safeCloneValOrRef(in.File, ctx),
		Append: in.Append,
	}
	return out
}
