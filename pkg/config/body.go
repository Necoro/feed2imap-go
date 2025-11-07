package config

import (
	"fmt"
	"slices"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
)

type Body string

var validBody = []string{"default", "both", "content", "description", "fetch"}

func (b *Body) UnmarshalYAML(node ast.Node) error {
	var val string
	if err := yaml.NodeToValue(node, &val); err != nil {
		return err
	}

	if val == "" {
		val = "default"
	}

	if !slices.Contains(validBody, val) {
		// TODO: change to new validation
		return fmt.Errorf("Invalid value for 'body': %q", val)
	}

	*b = Body(val)
	return nil
}
