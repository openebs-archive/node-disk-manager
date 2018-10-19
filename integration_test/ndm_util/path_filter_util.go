package ndmutil

import (
	"github.com/ghodss/yaml"
	sysutil "github.com/openebs/CITF/utils/system"
	"io/ioutil"
	"strings"
)

func ReplaceAndApplyConfig(configMap ConfigMap) error {
	yamlBytes, err := ioutil.ReadFile(GetNDMOperatorFilePath())
	if err != nil {
		return err
	}

	yamlString := string(yamlBytes)
	configString, err := yaml.Marshal(configMap)
	stringConfig := strings.Replace(string(configString), "\n", "\n    ", -1)
	s1 := strings.Split(yamlString, "node-disk-manager.config: |")[0]
	s2 := strings.SplitN(yamlString, "---", 2)[1]
	yamlString = s1 + "\n  node-disk-manager.config: | \n    " + stringConfig + "\n--- \n" + s2
	return sysutil.RunCommandWithGivenStdin("kubectl apply -f -", yamlString)
}
