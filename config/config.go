// config/config.go
package config

import (
    "io/ioutil"
    "raft"

    "gopkg.in/yaml.v2"
)

type Config struct {
    RPCAddr    string       `yaml:"rpcAddr"`
    HTTPAddr   string       `yaml:"httpAddr"`
    RaftConfig *raft.Config `yaml:"raft"`
}

func LoadConfig(filePath string) (*Config, error) {
    data, err := ioutil.ReadFile(filePath)
    if err != nil {
        return nil, err
    }

    var config Config
    err = yaml.Unmarshal(data, &config)
    if err != nil {
        return nil, err
    }

    return &config, nil
}
