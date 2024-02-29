package config

import (
	"fmt"
	"slices"

	"gopkg.in/yaml.v3"
)

type Body string

var validBody = []string{"default", "both", "content", "description", "fetch"}

func (b *Body) UnmarshalYAML(node *yaml.Node) error {
	var val string
	if err := node.Decode(&val); err != nil {
		return err
	}

	if val == "" {
		val = "default"
	}

	if !slices.Contains(validBody, val) {
		return TypeError("line %d: Invalid value for 'body': %q", node.Line, val)
	}

	*b = Body(val)
	return nil
}

func TypeError(format string, v ...any) *yaml.TypeError {
	return &yaml.TypeError{Errors: []string{fmt.Sprintf(format, v...)}}
}
