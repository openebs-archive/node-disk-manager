package main

import (
	"flag"

	"k8s.io/klog"
	"k8s.io/klog"
)

func main() {
	flag.Set("alsologtostderr", "true")
	flag.Parse()

	klogFlags := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(klogFlags)

	// Sync the glog and klog flags.
	flag.CommandLine.VisitAll(func(f1 *flag.Flag) {
		f2 := klogFlags.Lookup(f1.Name)
		if f2 != nil {
			value := f1.Value.String()
			f2.Value.Set(value)
		}
	})

	klog.Info("hello from glog!")
	klog.Info("nice to meet you, I'm klog")
	klog.Flush()
	klog.Flush()
}
