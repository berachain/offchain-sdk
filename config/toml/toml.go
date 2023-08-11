package toml

import (
	"bytes"
	"os"

	"github.com/pelletier/go-toml"
)

// ReadTomlIntoMap reads a TOML file into a map.
func ReadIntoMap[T any](filepath string, target *T) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	return toml.NewDecoder(bytes.NewReader(data)).Decode(target)
}
