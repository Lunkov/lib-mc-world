package owm

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
  w.API = "open.weather.map"
  return w
}

// https://rapidapi.com/community/api/open-weather-map 
type CoordInfo struct {
  Latitude           float64     `json:"lat"`
  Longitude          float64     `json:"lon"`
}

type WeatherInfo struct {
  Main           string     `json:"main"`
  Description    string     `json:"description"`
}

type MainInfo struct {
  Temp           float32    `json:"temp"`
  Pressure       float32    `json:"pressure"`
  Humidity       float32    `json:"humidity"`
}

type WindInfo struct {
  Speed        float32    `json:"speed"`
  Direction    float32    `json:"deg"`
}

type RainInfo struct {
  H1          float32    `json:"1h"`
}

type SysInfo struct {
  //Sunrise      time.Time    `json:"sunrise"`
  //Sunset       time.Time    `json:"sunset"`
}

type CloudsInfo struct {
  All          float32    `json:"all"`
}

type dataInfo struct {
  //Dt        time.Time       `json:"dt"`
  Coord     CoordInfo       `json:"coord"`
  Weather   []WeatherInfo   `json:"weather"`
  Base      string          `json:"base"`
  Main      MainInfo        `json:"main"`
  Wind      WindInfo        `json:"wind"`
  Rain      RainInfo        `json:"rain"`
  Clouds    CloudsInfo      `json:"clouds"`
  Sys       SysInfo         `json:"sys"`
  Name      string          `json:"name"`
  Cod       int             `json:"cod"`
}

func (w *WorkerInfo) GetData() {
  if glog.V(2) {
    glog.Infof("LOG: Open Weather Map started")
  }
  var data dataInfo

  url := fmt.Sprintf("https://community-open-weather-map.p.rapidapi.com/weather?lat=%f&lon=%f&id=%s&lang=en&units=metric&mode=json", w.ClientData.Latitude, w.ClientData.Longitude, w.ClientData.Login)

  req, _ := http.NewRequest("GET", url, nil)

  req.Header.Add("x-rapidapi-host", "community-open-weather-map.p.rapidapi.com")
  req.Header.Add("x-rapidapi-key", w.ClientData.Token)

  res, _ := http.DefaultClient.Do(req)

  defer res.Body.Close()
  body, _ := ioutil.ReadAll(res.Body)
  if err := json.Unmarshal(body, &data); err != nil {
    w.ClientData.Status.Ok = false
    w.ClientData.Status.LastError = fmt.Sprintf("Open Weather Map: %s", err)
    glog.Errorf("ERR: %s\n", w.ClientData.Status.LastError)
    return
  }
  dt := time.Now()
  
  var dm []mc.DeviceMetric
  
  dm = append(dm, mc.DeviceMetric{Device_ID: w.ClientData.Device_ID, Metric_CODE: "Weather.Station.Air.Temperature", DT: dt, Value: float64(data.Main.Temp)})
  dm = append(dm, mc.DeviceMetric{Device_ID: w.ClientData.Device_ID, Metric_CODE: "Weather.Station.Air.Wind.Speed", DT: dt, Value: float64(data.Wind.Speed)})
  dm = append(dm, mc.DeviceMetric{Device_ID: w.ClientData.Device_ID, Metric_CODE: "Weather.Station.Air.Wind.Direction", DT: dt, Value: float64(data.Wind.Direction)})
  dm = append(dm, mc.DeviceMetric{Device_ID: w.ClientData.Device_ID, Metric_CODE: "Weather.Station.Air.Humidity", DT: dt, Value: float64(data.Main.Humidity)})
  dm = append(dm, mc.DeviceMetric{Device_ID: w.ClientData.Device_ID, Metric_CODE: "Weather.Station.Air.Pressure", DT: dt, Value: float64(data.Main.Pressure)})

  w.ClientData.Status.CntDevices = 1
  w.ClientData.Status.CntMetrics = 5
  w.ClientData.Status.Ok = true
  
  w.SendMetrics(&dm)
  if glog.V(2) {
    glog.Infof("LOG: Open Weather Map finished")
  }
  w.ClientData.Status.Ok = true
}
