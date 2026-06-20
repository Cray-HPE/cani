/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2026 Hewlett Packard Enterprise Development LP
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a
 *  copy of this software and associated documentation files (the "Software"),
 *  to deal in the Software without restriction, including without limitation
 *  the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the
 *  Software is furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included
 *  in all copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 *  THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 *  OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 *  ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 *  OTHER DEALINGS IN THE SOFTWARE.
 *
 */
package config

import (
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

// FieldComment holds the comment metadata for a single YAML field.
type FieldComment struct {
	HeadComment string
	LineComment string
	FootComment string
}

// extractComments extracts comment metadata from struct tags (head_comment,
// line_comment, foot_comment) and returns a map keyed by YAML field name.
func extractComments(structType interface{}) map[string]FieldComment {
	v := reflect.ValueOf(structType)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()
	if t.Kind() != reflect.Struct {
		return nil
	}

	out := make(map[string]FieldComment, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("yaml")
		yamlName := strings.SplitN(tag, ",", 2)[0]
		if yamlName == "" || yamlName == "-" {
			continue
		}
		fc := FieldComment{
			HeadComment: field.Tag.Get("head_comment"),
			LineComment: field.Tag.Get("line_comment"),
			FootComment: field.Tag.Get("foot_comment"),
		}
		if fc.HeadComment != "" || fc.LineComment != "" || fc.FootComment != "" {
			out[yamlName] = fc
		}
	}
	return out
}

// applyComments updates the YAML key and value nodes inside a MappingNode
// so that comments from struct tags are idempotently applied on every save.
func applyComments(mapNode *yaml.Node, comments map[string]FieldComment) {
	if mapNode == nil || mapNode.Kind != yaml.MappingNode || len(comments) == 0 {
		return
	}
	for i := 0; i+1 < len(mapNode.Content); i += 2 {
		keyNode := mapNode.Content[i]
		valNode := mapNode.Content[i+1]
		fc, ok := comments[keyNode.Value]
		if !ok {
			continue
		}
		if fc.HeadComment != "" {
			keyNode.HeadComment = fc.HeadComment
		}
		if fc.LineComment != "" {
			// Only apply line comments to scalars or empty inline nodes.
			// yaml.v3 misplaces line comments on multi-line mapping/sequence
			// nodes, causing them to float to the next sibling key.
			if valNode.Kind == yaml.ScalarNode || len(valNode.Content) == 0 {
				valNode.LineComment = fc.LineComment
			}
		}
		if fc.FootComment != "" {
			valNode.FootComment = fc.FootComment
		}
	}
}
