package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"

	"github.com/liadmire/sys"
)

func ConfigLoad(fileName string, v interface{}) error {
	filePath := path.Join(sys.SelfDir(), fileName)
	if !sys.FileExists(filePath) {
		return fmt.Errorf("%s config file not exists.", fileName)
	}

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, v)
	if err != nil {
		return err
	}

	return nil
}

func ConfigSave(fileName string, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	filePath := path.Join(sys.SelfDir(), fileName)
	err = ioutil.WriteFile(filePath, data, 0666)
	if err != nil {
		return err
	}

	return nil
}
