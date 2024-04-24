package utils

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"

	"gopkg.in/yaml.v3"
)

var prefix = "ryax.tech/"

func Annotate(filePath string, deadline string, duration string, cores int) error {

	fileContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	decoder := yaml.NewDecoder(bytes.NewBuffer(fileContent))
	for {
		data := make(map[interface{}]interface{})

		err := decoder.Decode(&data)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if data["kind"] == "Pod" {
			annotations := make(map[interface{}]interface{})
			annotations[prefix+"deadline"] = deadline
			annotations[prefix+"duration"] = duration
			annotations[prefix+"resources.cores"] = strconv.Itoa(cores)
			annotationsInYaml := data["metadata"].(map[string]interface{})
			annotationsInYaml["annotations"] = annotations
		}
		strData, err := yaml.Marshal(&data)
		if err != nil {
			return err
		}
		fmt.Println(string(strData))
		fmt.Println("---")
	}
	return nil
}
