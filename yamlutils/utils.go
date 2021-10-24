package yamlutils

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

func Load(path string, out interface{}) error {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(content, out); err != nil {
		return err
	}
	return nil
}

func Save(path string, in interface{}) error {
	data, err := dump(in)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, 0644)
}


func dump(in interface{}) (out []byte, err error) {
        data, err := yaml.Marshal(in)
        if err != nil {
                return nil, err
        }
		return data, nil
}