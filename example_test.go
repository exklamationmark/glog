package glog_test

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"code.in.spdigital.io/sp-digital/energy-framework/pkg/glog"
)

func Example() {
	flag.Parse()
	defer glog.Flush()

	go func() {
		http.ListenAndServe(":8088", nil)
	}()

	for range time.NewTicker(time.Second).C {
		glog.Info("Info")
		glog.Warning("Warning")
		glog.Error("Error")
		glog.V(0).Info("v 0")
		glog.V(1).Info("v 1")
		glog.V(2).Info("v 2")
		glog.V(3).Info("v 3")
	}
	fmt.Println("exit")
}
