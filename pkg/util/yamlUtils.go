package util

import "gopkg.in/yaml.v3"


func BuildStringNode(key, value, comment string) []*yaml.Node {
	keyNode := &yaml.Node{
		Kind:        yaml.ScalarNode,
		Tag:         "!!str",
		Value:       key,
		LineComment: comment,
	}
	valueNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   "!!str",
		Value: value,
	}
	return []*yaml.Node{keyNode, valueNode}
}

func BuildIntNode(key, value, comment string) []*yaml.Node {
	keyNode := &yaml.Node{
		Kind:        yaml.ScalarNode,
		Tag:         "!!str",
		Value:       key,
		HeadComment: comment,
	}
	valueNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   "!!int",
		Value: value,
	}
	return []*yaml.Node{keyNode, valueNode}
}

func BuildBoolNode(key, value, comment string) []*yaml.Node {
	keyNode := &yaml.Node{
		Kind:        yaml.ScalarNode,
		Tag:         "!!str",
		Value:       key,
		HeadComment: comment,
	}
	valueNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   "!!bool",
		Value: value,
	}
	return []*yaml.Node{keyNode, valueNode}
}

func BuildMapNode(key, comment string) (*yaml.Node, *yaml.Node) {
	return &yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   "!!str",
		Value: key,
		LineComment: comment,
	}, &yaml.Node{Kind: yaml.MappingNode,
		Tag: "!!map",
	}
}

func BuildSequenceNode(key, comment string) (*yaml.Node, *yaml.Node) {
	return &yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   "!!str",
		Value: key,
		HeadComment: comment,
	}, &yaml.Node{Kind: yaml.SequenceNode,
		Tag: "!!seq",
	}
}
