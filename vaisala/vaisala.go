// vaisala.go
// Get Data from vaisala
package vaisala

import (
  "time"
  "fmt"
  "net/http"
  "io/ioutil"
  "encoding/xml"
  "github.com/golang/glog"
  "github.com/Lunkov/lib-mc"
)

type WorkerInfo struct {
  mc.WorkerInfo
}

func NewWorker() *WorkerInfo {
  w := new(WorkerInfo)
  w.API = "vaisala.weather"
  return w
}

type deviceInfo struct {
  Name         string `xml:"name"`
  Serial       string `xml:"serial"`
  Type         string `xml:"type"`
  Description  string `xml:"description"`
  Location     string `xml:"location"`
  Latitude     string `xml:"lat"`
  Longitude    string `xml:"lon"`
}

type measurementInfo struct {
  Id          string  `xml:"id,attr"`
  Timestamp   string  `xml:"timestamp"`
  Type        string  `xml:"type"`
  Value       float64 `xml:"value"`
}

type dataInfo struct {
  Device          deviceInfo       `xml:"device"`
  Measurements  []measurementInfo  `xml:"measurements>meas"`
}

func (w *WorkerInfo) httpGet(t0 string, t1 string) (dataInfo) {
  var data dataInfo
  url := fmt.Sprintf("%s&k=%s&t0=%s&t1=%s", w.ClientData.Url, w.ClientData.Token, t0, t1)
  resp, err := http.Get(url)
  if err != nil {
    w.ClientData.Status.Ok = false
    w.ClientData.Status.LastError = fmt.Sprintf("Vaisala(%s): %s", url, err)
    glog.Errorf("ERR: %s\n", w.ClientData.Status.LastError)
    return data
  }
  defer resp.Body.Close()
  body, err := ioutil.ReadAll(resp.Body)
  if glog.V(9) {
    glog.Infof("DBG: Vaisala: %v", string(body))
  }
  if err := xml.Unmarshal(body, &data); err != nil {
    w.ClientData.Status.Ok = false
    w.ClientData.Status.LastError = fmt.Sprintf("Vaisala: %s", err)
    glog.Errorf("ERR: %s\n", w.ClientData.Status.LastError)
    return data
  }
  w.ClientData.Status.Ok = true
  return data
}

func (w *WorkerInfo) GetData() {
  if glog.V(2) {
    glog.Infof("Vaisala started")
  }
  // Шаблон времени
  layoutTime := "2006-01-02 15:04:05"
  tt := time.Now()
  _, offset := tt.Zone()
  offset_minutes := offset/60
  
  tp0 := tt.Add(time.Duration(-offset_minutes - 25) * time.Minute)
  tp1 := tt.Add(time.Duration(-offset_minutes + 15) * time.Minute)
  t0 := tp0.Format(time.RFC3339)
  t1 := tp1.Format(time.RFC3339)
  data := w.httpGet(t0, t1)
  if !w.ClientData.Status.Ok {
    return
  }

  /*
   * Давление: миллибары в миллиметры ртутного столба
   * Концентрация веществ: ppm в мг/м3
   * https://www.gazoanalizators.ru/converter.html
   * "CO": 1.16197,
   * "NO2": 1.9085, 
   * "SO2": 2.65722,
   * "O3": 1.99116}
   * ! http://www.kipkomplekt.ru/calculator.php !
   * "CO": 1.176, 
   * "NO2": 1.932, 
   * "SO2": 2.6901,
   * "O3": 2.016} 
   */
  var params_factor = map[string]float64{"Air Pres.": 0.750062,
                                         "CO": 1.176, // 1.16197
                                         "NO2": 1.932, // 1.9085
                                         "SO2": 2.6901, // 2.65722
                                         "O3": 2.016} // 1.99116
  /*
  «nw» — северо-западное.
  «n» — северное.
  «ne» — северо-восточное.
  «e» — восточное.
  «se» — юго-восточное.
  «s» — южное.
  «sw» — юго-западное.
  «w» — западное.
  */
  var dm []mc.DeviceMetric
  w.ClientData.Status.CntDevices = 1
  for _, v := range data.Measurements {
    if metricCODE, ok := w.ClientData.ParamsCode[v.Type]; ok {
      dt, _ := time.Parse(layoutTime, v.Timestamp)
      dt = dt.Add(time.Duration(-offset_minutes) * time.Minute)
      val := v.Value
      if factor, ok := params_factor[v.Type]; ok {
        val *= factor
      }
      dm = append(dm, mc.DeviceMetric{Device_ID: w.ClientData.Device_ID, Metric_CODE: metricCODE, DT: dt, Value: val})
      w.ClientData.Status.CntMetrics ++
    }
  }
  w.SendMetrics(&dm)
  if glog.V(2) {
    glog.Infof("Vaisala finished")
  }
  w.ClientData.Status.Ok = true
}

