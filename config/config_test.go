package config

import (
	"fmt"
	"github.com/avalido/mpc-controller/logger"
	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParseConfig(t *testing.T) {
	logger.DevMode = true

	filename := "./config1.yaml"
	config := ParseConfig(filename)
	require.NotNil(t, config)
}

func TestMarshalConfig(t *testing.T) {
	var c config
	bytes, err := yaml.Marshal(c)
	require.Nil(t, err)
	fmt.Println(string(bytes))
}
