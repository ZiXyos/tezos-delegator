package conf

import "github.com/zixyos/goloader/config"

type DelegatorConfig struct {
	Service struct {
		Name    string `toml:"name"`
		Version string `toml:"version"`
	} `toml:"service"`

	HTTP struct {
		Port         int `toml:"port"`
		ReadTimeout  int `toml:"read_timeout"`
		WriteTimeout int `toml:"write_timeout"`
	} `toml:"http"`

	Storage struct {
		Database struct {
			Host     string `toml:"host"`
			Port     int    `toml:"port"`
			Username string `toml:"username"`
			Password string `toml:"password"`
			Database string `toml:"database"`
		} `toml:"database"`
	} `toml:"storage"`

	Logging struct {
		Level  string `toml:"level"`
		Format string `toml:"format"`
	} `toml:"logging"`
}

func LoadConfig() (*DelegatorConfig, error) {
	var dConfig DelegatorConfig

	err := config.Load(&dConfig, config.WithFs(FileFS))
	if err != nil {
		return nil, err
	}

	return &dConfig, nil
}
