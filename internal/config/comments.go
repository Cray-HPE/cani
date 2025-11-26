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
