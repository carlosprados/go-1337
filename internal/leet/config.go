package leet

import (
	"fmt"
	"os"

	"go.yaml.in/yaml/v3"
)

// variantList accepts either a single string or a list of strings in YAML.
type variantList []string

func (v *variantList) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind == yaml.ScalarNode {
		*v = variantList{n.Value}
		return nil
	}
	var list []string
	if err := n.Decode(&list); err != nil {
		return err
	}
	*v = list
	return nil
}

// ParseOverrides parses a YAML map of letter to variant(s), e.g.
//
//	a: "@"
//	e: ["3", "&"]
func ParseOverrides(data []byte) (map[string][]string, error) {
	var raw map[string]variantList
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	out := make(map[string][]string, len(raw))
	for k, v := range raw {
		out[k] = v
	}
	return out, nil
}

// LoadAlphabet returns the default alphabet with the overrides in the YAML
// file at path applied.
func LoadAlphabet(path string) (*Alphabet, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	overrides, err := ParseOverrides(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	a, err := Default().Override(overrides)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return a, nil
}
