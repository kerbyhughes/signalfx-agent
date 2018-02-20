package sources

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/signalfx/signalfx-agent/internal/utils"
	yaml "gopkg.in/yaml.v2"
)

type dynamicValueSpec struct {
	From     *fromPath   `yaml:"#from"`
	Flatten  bool        `yaml:"flatten"`
	Optional bool        `yaml:"optional"`
	Raw      bool        `yaml:"raw"`
	Default  interface{} `yaml:"default"`
}

type fromPath struct {
	sourceName string
	path       string
}

func (fp *fromPath) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var str string
	err := unmarshal(&str)
	if err != nil {
		return err
	}
	if len(str) == 0 {
		return errors.New("#from value is empty")
	}

	parts := strings.SplitN(str, ":", 2)

	if len(parts) == 1 {
		fp.path = parts[0]
	} else {
		fp.sourceName = parts[0]
		fp.path = parts[1]
	}

	return nil
}

func (fp *fromPath) SourceName() string {
	return utils.FirstNonEmpty(fp.sourceName, "file")
}

func (fp *fromPath) Path() string {
	return fp.path
}

func (fp *fromPath) String() string {
	return fp.SourceName() + ":" + fp.Path()
}

// RawDynamicValueSpec is a string that should deserialize to a dynamic value
// path (e.g. {"#from": "/path/to/value"}).
type RawDynamicValueSpec interface{}

func parseRawSpec(r RawDynamicValueSpec) (*dynamicValueSpec, error) {
	text, err := yaml.Marshal(r)
	if err != nil {
		return nil, err
	}

	var dvs dynamicValueSpec
	err = yaml.UnmarshalStrict(text, &dvs)

	if dvs.From == nil {
		// We should never get here for any given user input if the calling
		// code is doing its job.
		return nil, errors.New("#from field is missing")
	}

	if dvs.Raw && dvs.Default != nil {
		return nil, errors.New("cannot have raw remote config value and a default -- default is always parsed as YAML")
	}

	// If we have a default, this makes it always optional.
	if dvs.Default != nil {
		dvs.Optional = true
	}

	return &dvs, err
}
