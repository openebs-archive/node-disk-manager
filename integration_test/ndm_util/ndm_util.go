package ndmutil

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"reflect"
	"runtime"
	"strings"
	"time"

	"io/ioutil"

	"context"
	"errors"
	"sync"

	"github.com/golang/glog"
	logutil "github.com/openebs/CITF/utils/log"
	strutil "github.com/openebs/CITF/utils/string"
	sysutil "github.com/openebs/CITF/utils/system"
	cr "github.com/openebs/node-disk-manager/integration_test/common_resource"
	"github.com/openebs/node-disk-manager/integration_test/k8s_util"
	core_v1 "k8s.io/api/core/v1"
	v1beta1 "k8s.io/api/extensions/v1beta1"
)

const (
	// NdmYAML is the name of Configuration file for node-disk-manager
	NdmYAML = "node-disk-manager.yaml"
	// NdmOperatorYAML is the name of the Complete Configuration file for node-disk-manager
	// with service-account, cluster-role and cluster-role-binding
	NdmOperatorYAML = "ndm-operator.yaml"
	// NdmTestYAMLPath is the directory name where we shall keep temporary configuration file
	NdmTestYAMLPath = "/tmp/"
	// NdmTestYAMLName is the name of temporary configuration file
	NdmTestYAMLName = "NDM_Test_" + NdmYAML
	// NdmNamespace is the namespace of the node-disk-manager
	NdmNamespace = core_v1.NamespaceDefault
	// NdmTag is the default image tag used by the build scripts
	NdmTag = "ci"
)

var (
	// MaxTry is the number of tries that the functions in this package make. (Not all functions)
	MaxTry int = 5
	// WaitTimeUnit is the unit time duration that is mostly used by the functions in this package
	// to wait for after a try (Not always applicable)
	WaitTimeUnit time.Duration = 1 * time.Second

	logger logutil.Logger
)

// GetNDMDir returns the path to the node-disk-manager repository
func GetNDMDir() string {
	// Assumptions: GOPATH has only one path (no colons)
	return path.Join(os.Getenv("GOPATH"), "src/github.com/openebs/node-disk-manager/")
}

// GetNDMBinDir returns the path of bin folder of node-disk-manager repository
func GetNDMBinDir() string {
	return path.Join(GetNDMDir(), "bin/")
}

// GetNDMTestConfigurationDir returns value of the constant NdmTestYAMLPath
// which is the directory name where we shall keep temporary configuration file
func GetNDMTestConfigurationDir() string {
	return NdmTestYAMLPath
}

// GetNDMTestConfigurationFileName returns value of the constant NdmTestYAMLName
// which is the name of temporary configuration file
func GetNDMTestConfigurationFileName() string {
	return NdmTestYAMLName
}

// GetNDMConfigurationFileName returns value of the constant NdmYAML
// which is the name of Configuration file for node-disk-manager
func GetNDMConfigurationFileName() string {
	return NdmYAML
}

// GetNDMConfigurationFilePath returns the path of Configuration file for node-disk-manager
// i.e. GetNDMDir(), "samples" and GetNDMConfigurationFileName() joined together
func GetNDMConfigurationFilePath() string {
	return path.Join(GetNDMDir(), "samples", GetNDMConfigurationFileName())
}

// GetNDMOperatorFileName returns value of the constant NdmOperatorYAML
// which is the name of the Complete Configuration file for node-disk-manager
// with service-account, cluster-role and cluster-role-binding
func GetNDMOperatorFileName() string {
	return NdmOperatorYAML
}

// GetNDMOperatorFilePath returns the path of Complete Configuration file for node-disk-manager
// i.e. GetNDMDir() and GetNDMOperatorFileName() joined together
func GetNDMOperatorFilePath() string {
	return path.Join(GetNDMDir(), GetNDMOperatorFileName())
}

// GetNDMTestConfigurationFilePath returns the path of Test Configuration file for node-disk-manager
// i.e. GetNDMTestConfigurationDir() and GetNDMTestConfigurationFileName() joined together
func GetNDMTestConfigurationFilePath() string {
	return path.Join(GetNDMTestConfigurationDir(), GetNDMTestConfigurationFileName())
}

// GetDockerImageName returns the docker image name of node-disk-manager
// that will build when we build the project
func GetDockerImageName() string {
	return "openebs/node-disk-manager-" + sysutil.GetenvFallback("XC_ARCH", runtime.GOARCH)
}

// GetDockerImageTag returns the docker tag of the node-disk-manager docker image
// that wiil be used when we build the project
func GetDockerImageTag() string {
	if tag, ok := os.LookupEnv("TAG"); ok {
		return strings.TrimSpace(tag)
	}

	return NdmTag
}

// GetNDMNamespace returns value of the constant NdmNamespace
// which is the namespace of the node-disk-manager
func GetNDMNamespace() string {
	return NdmNamespace
}

// TODO: Check the pod current status like we do in `kubectl describe`
// Example: Check if all volumes are mounted correctly
// func ValidateNDMPod(ndmPod v1.Pod) bool {}

// ValidateNdmLog checks the supplied log and checks for any error in the log.
// :param string log: log of node-disk-manager-xxx Pod
// :return: bool: `true` if log contains no error otherwise return `false`.
func ValidateNdmLog(log string) bool {
	// Definition of this function should grow as program grows
	if strings.Contains(strings.ToLower(log), "started the controller") {
		return true
	}
	return false
}

// GetNDMLogAndValidateFor extracts log of supplied pod of node-disk-manager and then validate
// return the validation status and error occurred during process
func GetNDMLogAndValidateFor(ndmPod *core_v1.Pod) (bool, error) {
	// log if debug is enabled
	logger.PrintfDebugMessage("checking logs for pod %q in namespace %q", ndmPod.Name, ndmPod.Namespace)

	// loop until we get the log
	var log string
	var err error
	for {
		// Getting the log
		log, err = cr.CitfInstance.K8S.GetLog(ndmPod.Name, ndmPod.Namespace)
		if err != nil {
			return false, err
		}
		// exit loop only when we get something in the log
		if len(log) != 0 {
			break
		} else { // otherwise we wait for 1 second before we check again. This way we will have better probability to have a good number of log
			time.Sleep(time.Second)
		}
	}

	// Validating the log
	if !ValidateNdmLog(log) {
		return false, nil
	}

	return true, nil
}

// GetNDMLogAndValidate extracts log of node-disk-manager and then validate
// return the validation status and error occurred during process
func GetNDMLogAndValidate() (bool, error) {
	// Getting the log
	ndmPods, err := k8sutil.GetNdmPods()
	if err != nil {
		return false, err
	}

	// check log of every ndm pod, if any one fails return false
	for _, ndmPod := range ndmPods {
		if validated, err := GetNDMLogAndValidateFor(&ndmPod); err != nil {
			return false, err
		} else if !validated {
			return false, nil
		}
	}

	return true, nil
}

// YAMLPrepare reads and parse the configuration file into v1beta1.DaemonSet and changes some fields
// so that it can be applied to create node-disk-manager daemonset from recently built image.
// Then it returns that v1beta1.DaemonSet Structure
func YAMLPrepare() (v1beta1.DaemonSet, error) {
	// Prepare the yaml
	yamlBytes, err := ioutil.ReadFile(GetNDMConfigurationFilePath())
	if err != nil {
		return v1beta1.DaemonSet{}, err
	}

	// Get the DaemonSet Struct
	ds, err := cr.CitfInstance.K8S.GetDaemonsetStructFromYamlBytes(yamlBytes)
	if err != nil {
		return v1beta1.DaemonSet{}, err
	}

	fmt.Println("Image name:", GetDockerImageName()+":"+GetDockerImageTag())

	// Assumption: In following two statements it is assumed that
	// this pod has only one container

	// Set image name
	ds.Spec.Template.Spec.Containers[0].Image = GetDockerImageName() + ":" + GetDockerImageTag()

	// set imagePullPolicy
	ds.Spec.Template.Spec.Containers[0].ImagePullPolicy = "IfNotPresent"

	// If namespace is not mentioned in YAML then its namespace should be default namespace
	if ds.Namespace == "" {
		ds.Namespace = core_v1.NamespaceDefault
	}

	return ds, nil
}

// YAMLPrepareAndWriteInTempPath reads and parse the configuration file and changes some fields
// so that it can be applied to create node-disk-manager daemonset from recently built image.
// Then it saves that configuration to temp directory which will be cleaned by calling Clean()
func YAMLPrepareAndWriteInTempPath() error {
	dsManifest, err := YAMLPrepare()
	if err != nil {
		return err
	}

	jsonBytes, err := json.Marshal(dsManifest)
	if err != nil {
		return err
	}

	yamlBytes, err := strutil.ConvertJSONtoYAML(jsonBytes)
	if err != nil {
		return err
	}
	ioutil.WriteFile(GetNDMTestConfigurationFilePath(), yamlBytes, 0644)
	return nil
}

// PrepareAndApplyYAML prepares and applies the node-disk-manager configuration
func PrepareAndApplyYAML() error {
	dsManifest, err := YAMLPrepare()
	if err != nil {
		return err
	}

	fmt.Println(strutil.PrettyString(dsManifest))

	dsManifest, err = cr.CitfInstance.K8S.ApplyDSFromManifestStruct(dsManifest)
	fmt.Println("After applying...")
	fmt.Println(strutil.PrettyString(dsManifest))
	return err
}

// ReplaceImageInYAMLAndApply prepares NDM Operator YAML by string replacement and
// applies the same using kubectl
func ReplaceImageInYAMLAndApply() error {
	yamlBytes, err := ioutil.ReadFile(GetNDMOperatorFilePath())
	if err != nil {
		return err
	}

	yamlString := string(yamlBytes)
	yamlString = strings.Replace(yamlString, "amd64:ci", sysutil.GetenvFallback("XC_ARCH", runtime.GOARCH)+":"+GetDockerImageTag(), 1)
	yamlString = strings.Replace(yamlString, "imagePullPolicy: Always", "imagePullPolicy: IfNotPresent", 1)

	fmt.Println("Image name:", GetDockerImageName()+":"+GetDockerImageTag())
	logger.PrintlnDebugMessage("String to apply:", yamlString)

	return sysutil.RunCommandWithGivenStdin("kubectl apply -f -", yamlString)
}

// GetLsblkOutputOnHost runs `lsblk -bro name,size,type,mountpoint` on the host
// and parses the output in a map then returns the map
func GetLsblkOutputOnHost() (map[string]map[string]string, error) {
	// NOTE: lsblk in Ubuntu-Trusty does not have column serial
	// lsblkOutputStr, err := ExecCommand("lsblk -brdno name,size,model,serial")
	lsblkOutputStr, err := sysutil.ExecCommand("lsblk -brdno name,size,model")
	if err != nil {
		return nil, err
	}

	lsblkOutput := map[string]map[string]string{}
	// Assumption: In output of `lsblk -brdno name,size,model`
	// none of `name`, `size`, `model` have only white-spaces in them.
	attrs := []string{"Name", "Size", "Model"}

	scanner := bufio.NewScanner(strings.NewReader(strings.TrimSpace(lsblkOutputStr)))
	for scanner.Scan() {
		oneLsblkOutput := map[string]string{}
		for i, value := range strings.Fields(scanner.Text()) {
			oneLsblkOutput[attrs[i]] = strings.TrimSpace(value)
		}
		// key will be ndm's path i.e. /dev/<name>
		lsblkOutput["/dev/"+oneLsblkOutput["Name"]] = oneLsblkOutput
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lsblkOutput, nil
}

// GetNDMDeviceListOutputFromThePod runs `ndm device list` in the node-disk-manager pod
// and parses the output in a map then returns the map
func GetNDMDeviceListOutputFromThePod(ndmPod *core_v1.Pod) (map[string]map[string]string, error) {
	// Assumption: either only one container is there or if kubectl binary is there then default container gives correct result
	ndmOutputStr, err := cr.CitfInstance.K8S.ExecToPod("ndm device list", "", ndmPod.Name, ndmPod.Namespace)
	if err != nil {
		return nil, err
	}

	ndmOutput := map[string]map[string]string{}

	for _, diskDetail := range strings.Split(strings.TrimSpace(ndmOutputStr), "\n\n") {
		oneNdmOutput := map[string]string{}
		scanner := bufio.NewScanner(strings.NewReader(strings.TrimSpace(diskDetail)))
		for scanner.Scan() {
			keyValue := strings.SplitN(scanner.Text(), ": ", 2)
			var key, value string
			key = strings.TrimSpace(keyValue[0])
			if len(keyValue) > 1 {
				value = strings.TrimSpace(keyValue[1])
			} else {
				value = ""
			}

			oneNdmOutput[key] = value
		}
		if err := scanner.Err(); err != nil {
			return nil, err
		}

		ndmOutput[oneNdmOutput["Path"]] = oneNdmOutput
	}

	return ndmOutput, nil
}

// MatchDisksOutsideAndInsideFor takes output of `lsblk -brdno name,size,model` on the host
// and output of `ndm device list` inside the pod supplied. Then it matches the two
func MatchDisksOutsideAndInsideFor(ndmPod *core_v1.Pod) (bool, error) {
	lsblkOutput, err := GetLsblkOutputOnHost()
	if err != nil {
		return false, err
	}

	logger.PrintfDebugMessage("`lsblk` output: %q", lsblkOutput)

	ndmOutput, err := GetNDMDeviceListOutputFromThePod(ndmPod)
	if err != nil {
		return false, err
	}

	logger.PrintfDebugMessage("`ndm device list` output: %q\n", ndmOutput)

	// For all devices listed in ndm device list
	for devPath, devDetailsNdm := range ndmOutput {
		devDetailsLsblk, ok := lsblkOutput[devPath]
		diskInfo := struct {
			diskPresent bool
			diskStatus  string
		}{ok, devDetailsNdm["Status"]}

		switch {
		case diskInfo.diskPresent && diskInfo.diskStatus == "Active":
			// Check properties of required disk from lsblk with ndm
			for k, v := range devDetailsLsblk {
				// In some machines, we are currently facing some problem in getting the vendor name of device
				// So ignore this for now
				if k == "Vendor" {
					continue
				}

				// Name in lsblk output is just device name which is there in devPath of ndmOutput
				// And it is not stored under "Name" key in ndmOutput so no need to check this.
				if k == "Name" {
					continue
				}

				// Matching the value of the attribute (i.e. k) in lsblk output (i.e. v) with
				// output of same attribute in ndm
				// lsblk truncates the Output so we need to check for the prefix
				if !(strings.HasPrefix(devDetailsNdm[k], v)) &&
					// now check if after replacing space with underscore matches
					!(strings.HasPrefix(devDetailsNdm[k], strings.Replace(v, " ", "_", -1))) {
					// If normal strings does not matches then try to match it by
					// resolving hex codes and trimming again
					var hexReplacedValue string
					hexReplacedValue, err = strutil.ReplaceHexCodesWithValue(v)
					// if error occurred in decoding
					if err != nil {
						return false, err
					}
					hexReplacedValue = strings.TrimSpace(hexReplacedValue)
					if !(strings.HasPrefix(devDetailsNdm[k], hexReplacedValue)) &&
						// If even that fails then it tries to replace space with underscore
						// as ndm sometimes gets values this way
						!(strings.HasPrefix(devDetailsNdm[k], strings.Replace(hexReplacedValue, " ", "_", -1))) {
						return false, nil
					}
				}
			}
		case diskInfo.diskPresent && diskInfo.diskStatus != "Active":
			return false, nil
		case !diskInfo.diskPresent && diskInfo.diskStatus == "Active":
			return false, nil
		case !diskInfo.diskPresent && diskInfo.diskStatus != "Active":
			continue
		}
	}

	return true, nil
}

// MatchDisksOutsideAndInside takes output of `lsblk -brdno name,size,model` on the host
// and output of `ndm device list` inside all the ndm pod. Then it matches the two
func MatchDisksOutsideAndInside() (bool, error) {
	ndmPods, err := k8sutil.GetNdmPods()
	if err != nil {
		return false, err
	}
	for _, ndmPod := range ndmPods {
		if matched, err := MatchDisksOutsideAndInsideFor(&ndmPod); err != nil {
			return false, err
		} else if !matched {
			return false, nil
		}
	}
	return true, nil
}

// WaitTillDefaultNSisReady busy waits until default namespace is ready
// or number of try exceeds MaxTry (at least once).
// Each try is make after waiting for time period of WaitTimeUnit
func WaitTillDefaultNSisReady() {
	// Ensuring minimum value for the arguments

	// Making sure that it tries at least once to apply the YAML
	maxTry := MaxTry
	if maxTry < 1 {
		maxTry = 1
	}
	waitTimeUnit := WaitTimeUnit
	if waitTimeUnit < 0*time.Second {
		waitTimeUnit = 0 * time.Second
	}

	ndmNS := core_v1.Namespace{}
	for i := 0; i < maxTry; i++ {
		namespaces, err := cr.CitfInstance.K8S.GetAllNamespacesCoreV1NamespaceArray()
		if err == nil {
			for _, namespace := range namespaces {
				if namespace.Name == GetNDMNamespace() {
					ndmNS = namespace
					break
				}
			}
		} else {
			fmt.Printf("Try - %d: Error getting namespaces. Error: %+v\n", i, err)
			time.Sleep(waitTimeUnit)
			continue
		}

		// If namespace is not even created
		if reflect.DeepEqual(ndmNS, core_v1.Namespace{}) {
			fmt.Printf("Try - %d: Waiting as Namespace %q has not been created yet.\n", i, GetNDMNamespace())
			time.Sleep(waitTimeUnit)
			continue
		}

		if cr.CitfInstance.K8S.IsNSinGoodPhase(ndmNS) {
			break
		}
		fmt.Printf("Try - %d: Waiting as Namespace %q is in %q phase\n", i, ndmNS.Name, ndmNS.Status.Phase)
	}

	// Final Check
	if reflect.DeepEqual(ndmNS, core_v1.Namespace{}) {
		glog.Fatalf("Namespace %q didn't came up in %v", GetNDMNamespace(), time.Duration(MaxTry)*waitTimeUnit)
	} else if !cr.CitfInstance.K8S.IsNSinGoodPhase(ndmNS) {
		glog.Fatalf("Namespace %q is still in bad phase: %q after %v", GetNDMNamespace(), ndmNS.Status.Phase, time.Duration(MaxTry)*waitTimeUnit)
	}
	fmt.Printf("Namespace %q is ready\n", ndmNS.Name)
}

// getNDMPodsForEachNode gets NDM pods then it gets nodes, if there are same number of nodes and pods
// return ndmPods other wise the error
// this is function will be used in default block of `WaitTillNDMisUpOrTimeout`
func getNDMPodsForEachNode() (ndmPods []core_v1.Pod, err error) {
	ndmPods, err = k8sutil.GetNdmPods()
	if err != nil {
		// form error string
		errString := fmt.Sprintf("error occurred in getting ndm pods: %+v", err)
		// update err
		err = errors.New(errString)
		// print error string
		fmt.Println(errString)
		// sleep and continue
		time.Sleep(time.Second)
	}

	var nodes []core_v1.Node
	nodes, err = cr.CitfInstance.K8S.GetNodes()
	if err != nil {
		err = fmt.Errorf("failed to get the number of nodes: %+v", err)
	}

	// if not as many ndm-pods are there as there are nodes
	// Assumption: node-disk-manager is a daemonset
	if len(ndmPods) != len(nodes) {
		err = fmt.Errorf("%d node-disk-manager pods found for %d node(s)", len(ndmPods), len(nodes))
	}
	return
}

// blockForNDMPod blocks until supplied pod is up or context is cancelled
// This function is sub-task of `WaitTillNDMisUpOrTimeout`
func blockForNDMPod(ctx context.Context, cancel context.CancelFunc, ndmPod core_v1.Pod, wg *sync.WaitGroup, err *error) {
	defer wg.Done()

	*err = cr.CitfInstance.K8S.BlockUntilPodIsUpWithContext(ctx, &ndmPod)
	if *err != nil {
		*err = fmt.Errorf("error while waiting for pod %q: %+v", ndmPod.Name, *err)
		cancel()
	} else {
		fmt.Printf("pod %q is up.\n", ndmPod.Name)
	}
}

// WaitTillNDMisUpOrTimeout busy waits until all ndm pods are up or timeout hits.
// Each try is make after waiting for WaitTimeUnit number of seconds.
func WaitTillNDMisUpOrTimeout(timeout time.Duration) (err error) {
	startTime := time.Now()

	logger.PrintlnDebugMessage("Waiting for NDM pods to be up...")
	var ndmPods []core_v1.Pod
	timeoutChan := time.After(timeout)
GetNdmPods:
	for {
		select {
		case <-timeoutChan:
			logger.PrintlnDebugMessage("timeout for checking NDM Pods to be up")
			// return the last error
			return
		default:
			ndmPods, err = getNDMPodsForEachNode()
			if err != nil {
				continue
			}
			// if err is nil break the loop
			break GetNdmPods
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout-time.Since(startTime))
	defer cancel()

	ndmErrors := make([]error, len(ndmPods))
	var wg sync.WaitGroup
	for i, ndmPod := range ndmPods {
		// create a go-routine for every ndm-pod so that we can return as soon as we hit any error in any of them
		wg.Add(1)
		// Sending a copy of `ndmPod` as it will be used inside this goroutine even when original one changes bacause of loop
		go blockForNDMPod(ctx, cancel, ndmPod, &wg, &ndmErrors[i])
	}
	wg.Wait()

	// return first non-nil error if any
	for _, err := range ndmErrors {
		if err != nil {
			return err
		}
	}
	// otherwise return nil
	return nil
}

// Clean is intended to clean the residue of the testing.
// It should be run at the very end of the test.
// CAUTION: it calls `cr.CitfInstance.Environment.Teardown()`
// which stops all Docker Containers in your machine.
func Clean() {
	// Check minikube status and delete if minikube is running
	fmt.Println("Checking minikube status...")
	minikubeStatus, err := cr.CitfInstance.Environment.Status()
	if err != nil {
		fmt.Printf("Error occurred while checking status of minikube. Error: %+v\n", err)
	}
	if state, ok := minikubeStatus["minikube"]; ok && (state == "Running" || state == "Stopped") {
		fmt.Println("Deleting minikube...")
		err = cr.CitfInstance.Environment.Teardown()
		if err != nil {
			fmt.Printf("Error while deleting minikube. Error: %+v\n", err)
		}
	} else {
		fmt.Println("Machine not present.")
	}

	// Remove all docker containers
	fmt.Println("Removing docker containers...")
	err = cr.CitfInstance.Docker.Teardown()
	if err != nil {
		fmt.Printf("Error occurred when deleting containers. Error: %+v\n", err)
	}

	fmt.Printf("Removing %q...\n", GetNDMTestConfigurationFileName())
	if _, err = os.Stat(GetNDMTestConfigurationFilePath()); os.IsNotExist(err) {
		fmt.Printf("%q not present\n", GetNDMTestConfigurationFileName())
	} else {
		err = os.Remove(GetNDMTestConfigurationFilePath())
		if err != nil {
			fmt.Printf("Error occurred while removing NDM's temporary configuration file. Error: %+v\n", err)
		} else {
			fmt.Printf("%q removed.\n", GetNDMTestConfigurationFilePath())
		}
	}
}

// SetIncludePath is used to set the include section in
// path filter of NDMConfig
func (c *ConfigMap) SetIncludePath(deviceList ...string) {
	for index, element := range c.FilterConfigs {
		if element.Key == "path-filter" {
			c.FilterConfigs[index].Include = strings.Join(deviceList, ",")
			c.FilterConfigs[index].Exclude = ""
		}
	}
}

// SetExcludePath is used to set the exclude section in
// path filter of NDMConfig
func (c *ConfigMap) SetExcludePath(deviceList ...string) {
	for index, element := range c.FilterConfigs {
		if element.Key == "path-filter" {
			c.FilterConfigs[index].Exclude = strings.Join(deviceList, ",")
			c.FilterConfigs[index].Include = ""
		}
	}
}

// SetPathFilter is used to change the state of
// path filter in NDMConfig
func (c *ConfigMap) SetPathFilter(state string) {
	for index, element := range c.FilterConfigs {
		if element.Key == "path-filter" {
			c.FilterConfigs[index].State = state
		}
	}
}

// MatchNDMDeviceList is used to match the NDM devices and device
// paths specified as string. Both include and exclude can be
// matched using by changing the pathType bool. pathType `true`
// means exclude filter will be checked and pathType `false`
// means include filter will be checked.
func MatchNDMDeviceList(pathType bool, devicePaths ...string) (bool, error) {
	ndmPods, err := k8sutil.GetNdmPods()
	if err != nil {
		return false, err
	}

	for _, ndmPod := range ndmPods {
		deviceList, err := GetNDMDeviceListOutputFromThePod(&ndmPod)
		if err != nil {
			return false, err
		}
		for _, devicePath := range devicePaths {
			if _, ok := deviceList[devicePath]; ok && pathType {
				return false, nil
			}
		}
	}
	return true, nil
}
