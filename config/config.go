package config

import (
	"github.com/goccy/go-yaml"
	"github.com/pkg/errors"
	"io/ioutil"
)

type Config struct {
	EnableDevMode     bool   `yaml:"enableDevMode"`
	ControllerId      string `yaml:"controllerId"`
	ControllerKey     string `yaml:"controllerKey"` // todo: add secure keystore
	MpcManagerAddress string `yaml:"mpcManagerAddress"`
	MpcServerUrl      string `yaml:"mpcServerUrl"`
	EthRpcUrl         string `yaml:"ethRpcUrl"`
	EthWsUrl          string `yaml:"ethWsUrl"`
	CChainIssueUrl    string `yaml:"cChainIssueUrl"`
	PChainIssueUrl    string `yaml:"pChainIssueUrl"`

	NetworkConfig  `yaml:"networkConfig"`
	DatabaseConfig `yaml:"databaseConfig"`
}

type NetworkConfig struct {
	NetworkId uint32 `yaml:"networkId"`
	ChainId   int64  `yaml:"chainId"`
	CChainId  string `yaml:"cChainId"`
	AvaxId    string `yaml:"avaxId"`

	ImportFee  uint64 `yaml:"importFee"`
	ExportFee  uint64 `yaml:"exportFee"`
	GasPerByte uint64 `yaml:"gasPerByte"`
	GasPerSig  uint64 `yaml:"gasPerSig"`
	GasFixed   uint64 `yaml:"gasFixed"`
}

type DatabaseConfig struct {
	BadgerDbPath string `yaml:"badgerDbPath"`
}

func ParseFile(filename string) *Config {
	return verifyConfig(parseContent(readFile(filename)))
}

func readFile(filename string) []byte {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(errors.Wrapf(err, "failed to read file %q", filename))
	}
	return content
}

func parseContent(content []byte) *Config {
	var c Config
	err := yaml.Unmarshal(content, &c)
	if err != nil {
		panic(errors.Wrap(err, "failed to parse NetworkConfig"))
	}
	return &c
}

func verifyConfig(c *Config) *Config {
	// todo
	return c
}
