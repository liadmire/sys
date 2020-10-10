package config

type jsonConfig struct {
	filePath string
	isCrypt  bool
	data     interface{}
}

func NewJSONConfig(filePath string, isCrypt b)

func (jc *jsonConfig) Load() error {
	return nil
}

func (jc *jsonConfig) Save() error {
	return nil
}
