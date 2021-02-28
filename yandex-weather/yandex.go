package yandex

import (
  "time"
  "fmt"
  "net/http"
  "io/ioutil"
  "encoding/json"

  "github.com/golang/glog"
  "github.com/Lunkov/lib-mc"
)

type WorkerInfo struct {
  mc.WorkerInfo
}

func NewWorker() *WorkerInfo {
  w := new(WorkerInfo)
  w.API = "yandex.weather"
  return w
}

// https://yandex.com/dev/weather/doc/dg/concepts/pricing.html#about
// https://yandex.com/dev/weather/doc/dg/concepts/forecast-info.html#forecast-info
type WeatherInfo struct {
  Temp           float32    `json:"temp"`
  Condition      string     `json:"condition"`
  Wind_Speed     float32    `json:"wind_speed"`
  Wind_Direction float32    `json:"wind_dir"`
  PressureMM     float32    `json:"pressure_mm"`
  Humidity       float32    `json:"humidity"`
}

type dataInfo struct {
  Fact      WeatherInfo `json:"fact"`
}

func (w *WorkerInfo) GetData() {
  if glog.V(2) {
    glog.Infof("LOG: Yandex Weather started")
  }
  var data dataInfo
  
  url := fmt.Sprintf("https://api.weather.yandex.ru/v2/informers?lat=%f&lon=%f&lang=en_US", w.ClientData.Latitude, w.ClientData.Longitude)
  req, _ := http.NewRequest("GET", url, nil)
  req.Header.Add("X-Yandex-API-Key", w.ClientData.Token)
  res, err := http.DefaultClient.Do(req)
  if err != nil {
    w.ClientData.Status.Ok = false
    w.ClientData.Status.LastError = fmt.Sprintf("Yandex Weather: %s", err)
    glog.Errorf("ERR: %s\n", w.ClientData.Status.LastError)
    return
  }
  
  defer res.Body.Close()
  if res.StatusCode != http.StatusOK {
    w.ClientData.Status.Ok = false
    w.ClientData.Status.LastError = fmt.Sprintf("Yandex Weather: StatusCode: %d", res.StatusCode)
    glog.Errorf("ERR: %s\n", w.ClientData.Status.LastError)
    return
  }
  
  body, _ := ioutil.ReadAll(res.Body)
  if err = json.Unmarshal(body, &data); err != nil {
    w.ClientData.Status.Ok = false
    w.ClientData.Status.LastError = fmt.Sprintf("Open Weather Map: %s", err)
    glog.Errorf("ERR: %s\n", w.ClientData.Status.LastError)
    return
  }
  
  var dm []mc.DeviceMetric
  dt := time.Now()
  
  dm = append(dm, mc.DeviceMetric{Device_ID: w.ClientData.Device_ID, Metric_CODE: "Weather.Station.Air.Temperature", DT: dt, Value: float64(data.Fact.Temp)})
  dm = append(dm, mc.DeviceMetric{Device_ID: w.ClientData.Device_ID, Metric_CODE: "Weather.Station.Air.Wind.Speed", DT: dt, Value: float64(data.Fact.Wind_Speed)})
  dm = append(dm, mc.DeviceMetric{Device_ID: w.ClientData.Device_ID, Metric_CODE: "Weather.Station.Air.Wind.Direction", DT: dt, Value: float64(data.Fact.Wind_Direction)})
  dm = append(dm, mc.DeviceMetric{Device_ID: w.ClientData.Device_ID, Metric_CODE: "Weather.Station.Air.Humidity", DT: dt, Value: float64(data.Fact.Humidity)})
  dm = append(dm, mc.DeviceMetric{Device_ID: w.ClientData.Device_ID, Metric_CODE: "Weather.Station.Air.Pressure", DT: dt, Value: float64(data.Fact.PressureMM)})

  w.ClientData.Status.CntDevices = 1
  w.ClientData.Status.Ok = true
  
  w.SendMetrics(&dm)
  if glog.V(2) {
    glog.Infof("LOG: Yandex Weather finished")
  }
  w.ClientData.Status.Ok = true
}
