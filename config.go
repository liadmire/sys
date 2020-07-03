package sys

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
)

func ConfigLoad(fileName string, v interface{}) error {
	filePath := path.Join(SelfDir(), fileName)
	if !FileExists(filePath) {
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

	filePath := path.Join(SelfDir(), fileName)
	err = ioutil.WriteFile(filePath, data, 0666)
	if err != nil {
		return err
	}

	return nil
}
