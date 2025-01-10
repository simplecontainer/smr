package definitions

import (
	"encoding/json"
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

func ReadFile(filePath string) ([]byte, error) {
	var jsonData []byte = nil

	if filePath != "" {
		YAML, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("YAML file not found:   #%v ", err)
		}

		var body interface{}
		if err = yaml.Unmarshal(YAML, &body); err != nil {
			return nil, err
		}

		body = convert(body)

		if jsonData, err = json.Marshal(body); err != nil {
			return nil, err
		}
	}

	return jsonData, nil
}

func convert(i interface{}) interface{} {
	switch x := i.(type) {
	case map[interface{}]interface{}:
		m2 := map[string]interface{}{}
		for k, v := range x {
			m2[k.(string)] = convert(v)
		}
		return m2
	case []interface{}:
		for i, v := range x {
			x[i] = convert(v)
		}
	}
	return i
}
