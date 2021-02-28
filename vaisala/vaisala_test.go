// vaisala.go
// Get Data from vaisala
package vaisala

import (
  "testing"
  "github.com/stretchr/testify/assert"
  
  "time"
  "flag"
  
  "github.com/golang/glog"
  "github.com/Lunkov/lib-mc"
)

func TestCheckVaisala(t *testing.T) {
	flag.Set("alsologtostderr", "true")
	flag.Set("log_dir", ".")
	flag.Set("v", "9")
	flag.Parse()

	glog.Info("Logging configured")
  
  ow := NewWorker()
  
  mc.WorkerRegister(ow)

  go mc.Init("./etc.tests/")
  time.Sleep(2 * time.Second)

  assert.Equal(t, "vaisala.weather", ow.GetAPI())
  assert.Equal(t, false, ow.ClientData.Status.Ok)

  r1 := mc.GetWorkersResults()
  res, _ := (r1["vaisala.weather"][""]).(mc.Result)
  assert.Equal(t, int64(1), res.Status.CntDevices)
  assert.Equal(t, int64(0), res.Status.CntMetrics)
}
