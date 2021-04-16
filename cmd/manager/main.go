/*
Copyright 2019 The OpenEBS Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

// import (
// 	"context"
// 	"flag"
// 	"os"
// 	"runtime"
// 	"time"

// 	"github.com/openebs/node-disk-manager/pkg/apis"
// 	"github.com/openebs/node-disk-manager/pkg/controller"
// 	"github.com/openebs/node-disk-manager/pkg/env"
// 	ndmlogger "github.com/openebs/node-disk-manager/pkg/logs"
// 	"github.com/openebs/node-disk-manager/pkg/setup"
// 	"github.com/openebs/node-disk-manager/pkg/upgrade"
// 	"github.com/openebs/node-disk-manager/pkg/upgrade/v040_041"
// 	"github.com/openebs/node-disk-manager/pkg/upgrade/v041_042"
// 	"github.com/openebs/node-disk-manager/pkg/version"

// 	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
// 	"github.com/operator-framework/operator-sdk/pkg/leader"
// 	"github.com/operator-framework/operator-sdk/pkg/ready"
// 	sdkVersion "github.com/operator-framework/operator-sdk/version"
// 	"k8s.io/klog"
// 	"sigs.k8s.io/controller-runtime/pkg/client"
// 	"sigs.k8s.io/controller-runtime/pkg/client/config"
// 	"sigs.k8s.io/controller-runtime/pkg/manager"
// 	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
// 	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
// )

// //ReconciliationInterval defines the triggering interval for reconciliation operation
// const ReconciliationInterval = 5 * time.Second

// func printVersion() {
// 	klog.Infof("Go Version: %s", runtime.Version())
// 	klog.Infof("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)
// 	klog.Infof("operator-sdk Version: %v", sdkVersion.Version)
// 	klog.Infof("Version Tag: %s", version.GetVersion())
// 	klog.Infof("Git Commit: %s", version.GetGitCommit())
// }

// func main() {
// 	// define klog flags
// 	klog.InitFlags(nil)
// 	flag.Parse()

// 	// The logger instantiated here can be changed to any logger
// 	// implementing the logr.Logger interface. This logger will
// 	// be propagated through the whole operator, generating
// 	// uniform and structured logs.
// 	logf.SetLogger(logf.ZapLogger(false))

// 	// Init logging
// 	ndmlogger.InitLogs()
// 	defer ndmlogger.FlushLogs()

// 	printVersion()

// 	namespace, err := k8sutil.GetWatchNamespace()
// 	if err != nil {
// 		klog.Errorf("Failed to get watch namespace: %v", err)
// 		os.Exit(1)
// 	}

// 	// Get a config to talk to the apiserver
// 	cfg, err := config.GetConfig()
// 	if err != nil {
// 		klog.Errorf("Failed to get config: %v", err)
// 		os.Exit(1)
// 	}

// 	// Become the leader before proceeding
// 	leader.Become(context.TODO(), "node-disk-manager-lock")

// 	r := ready.NewFileReady()
// 	err = r.Set()
// 	if err != nil {
// 		klog.Errorf("Checking for /tmp/operator-sdk-ready failed: %v", err)
// 		os.Exit(1)
// 	}
// 	defer r.Unset()

// 	reconInterval := ReconciliationInterval

// 	// Create a new Cmd to provide shared dependencies and start components
// 	mgr, err := manager.New(cfg, manager.Options{Namespace: namespace, SyncPeriod: &reconInterval, MetricsBindAddress: "0"})
// 	if err != nil {
// 		klog.Errorf("Failed to create a new manager: %v", err)
// 		os.Exit(1)
// 	}

// 	// check if CRDs need to be installed.
// 	// The OPENEBS_IO_INSTALL_CRD env is checked
// 	if env.IsInstallCRDEnabled() {
// 		klog.Info("Installing the components")
// 		// get a new install setup
// 		setupConfig, err := setup.NewInstallSetup(cfg)
// 		if err != nil {
// 			klog.Errorf("Unable to get config for setting up CRDs: %v", err)
// 			os.Exit(1)
// 		}

// 		// install the components
// 		if err = setupConfig.Install(); err != nil {
// 			klog.Errorf("Failed to setup CRDs: %v", err)
// 			os.Exit(1)
// 		}
// 	}

// 	klog.Info("Registering Components")

// 	// Setup Scheme for all resources
// 	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
// 		klog.Errorf("Failed to add APIs to scheme: %v", err)
// 		os.Exit(1)
// 	}

// 	// Upgrade the components if required
// 	k8sClient, err := client.New(cfg, client.Options{})
// 	if err != nil {
// 		klog.Errorf("Failed to get client: %v", err)
// 		os.Exit(1)
// 	}

// 	klog.Info("Check if CR has to be upgraded, and perform upgrade")
// 	err = performUpgrade(k8sClient)
// 	if err != nil {
// 		klog.Errorf("Upgrade failed: %v", err)
// 		os.Exit(1)
// 	}

// 	// Setup all Controllers
// 	if err := controller.AddToManager(mgr); err != nil {
// 		klog.Errorf("Error setting up controller: %v", err)
// 		os.Exit(1)
// 	}

// 	klog.Info("Starting the ndm-operator...")

// 	// Start the Cmd
// 	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
// 		klog.Errorf("Manager exited non-zero: %v", err)
// 		os.Exit(1)
// 	}
// }

// // performUpgrade performs the upgrade operations
// func performUpgrade(client client.Client) error {
// 	v040_v041UpgradeTask := v040_041.NewUpgradeTask("0.4.0", "0.4.1", client)
// 	v041_v042UpgradeTask := v041_042.NewUpgradeTask("0.4.1", "0.4.2", client)
// 	return upgrade.RunUpgrade(v040_v041UpgradeTask, v041_v042UpgradeTask)
// }
