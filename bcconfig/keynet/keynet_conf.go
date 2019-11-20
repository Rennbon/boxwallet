package keynet

import (
	"github.com/Rennbon/boxwallet/bcconfig"
	"github.com/Rennbon/boxwallet/bccore"
	"github.com/mitchellh/mapstructure"
)

const (
	keyStorageConfigKey = "keynet"
)

type Config struct {
	Net bccore.Net
}

func DecodeConfig(cfg bcconfig.Provider) (c Config, err error) {
	m := cfg.GetStringMap(keyStorageConfigKey)
	err = mapstructure.WeakDecode(m, &c)
	return
}
