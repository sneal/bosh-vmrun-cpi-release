package stemcell

import (
	"gopkg.in/yaml.v3"
)

type StemcellManifest struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

func NewStemcellManifest(data []byte) (*StemcellManifest, error) {
	newManifest := &StemcellManifest{}
	err := yaml.Unmarshal(data, newManifest)

	if err != nil {
		return nil, err
	}

	return newManifest, nil
}
