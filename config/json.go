package config

type jsonConfig struct {
	filePath string
	isCrypt  bool
	data     interface{}
}

func NewJSONConfig(filePath string, isCrypt bool, data interface{}) *jsonConfig {
	return &jsonConfig{
		filePath: filePath,
		isCrypt:  isCrypt,
		data:     data,
	}
}

func (jc *jsonConfig) Load() error {
	return nil
}

func (jc *jsonConfig) Save() error {
	return nil
}
