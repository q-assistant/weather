package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const OpenWeatherMapApiUrl = "https://api.openweathermap.org/data/2.5"
const OpenWeatherMapPeriodDaily = "daily"
const OpenWeatherMapPeriodHourly = "hourly"
const OpenWeatherMapName = "open-weather-map"

type OpenWeatherMap struct {
	key      string
	secret   string
	location *Location
	client   *http.Client
}

type OpenWeatherMapWeatherForecastResponse struct {
	TimeZone string                `json:"timezone"`
	Current  *OpenWeatherMapMain   `json:"current"`
	Minutely []*OpenWeatherMapMain `json:"minutely"`
	Hourly   []*OpenWeatherMapMain `json:"hourly"`
}

type OpenWeatherMapWeatherResponse struct {
	DateTime int64                    `json:"dt"`
	Weather  []*OpenWeatherMapWeather `json:"weather"`
	Main     *OpenWeatherMapMain      `json:"main"`
	Wind     struct {
		Speed float64 `json:"speed"`
		Deg   int     `json:"deg"`
	} `json:"wind"`
	Clouds struct {
		All int `json:"all"`
	} `json:"clouds"`
}

type OpenWeatherMapWeather struct {
	ID          int    `json:"id"`
	Main        string `json:"main"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

type OpenWeatherMapMain struct {
	Dt        int64                    `json:"dt"`
	Temp      float64                  `json:"temp"`
	FeelsLike float64                  `json:"feels_like"`
	TempMin   float64                  `json:"temp_min"`
	TempMax   float64                  `json:"temp_max"`
	Pressure  int                      `json:"pressure"`
	Humidity  float64                  `json:"humidity"`
	Weather   []*OpenWeatherMapWeather `json:"weather"`
}

type Config struct {
	Key      string
	Secret   string
	Location *Location
}

type Location struct {
	Lat  float64
	Lon  float64
	Name string
}

func NewOpenWeatherMap(cnf *Config) *OpenWeatherMap {
	return &OpenWeatherMap{
		key:      cnf.Key,
		secret:   cnf.Secret,
		location: cnf.Location,
		client:   http.DefaultClient,
	}
}

func (owm *OpenWeatherMap) GetCurrent() (*Weather, error) {
	req, err := owm.client.Get(fmt.Sprintf("%s/weather?lat=%.3f&lon=%.3f&appid=%s&units=metric", OpenWeatherMapApiUrl, owm.location.Lat, owm.location.Lat, owm.secret))
	if err != nil {
		return nil, err
	}

	defer req.Body.Close()

	b, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	var out *OpenWeatherMapWeatherResponse
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}

	return &Weather{
		Type:        out.Weather[0].Main,
		DateTime:    time.Unix(out.DateTime, 0),
		Description: out.Weather[0].Description,
		Temp:        out.Main.Temp,
		TempFeeling: out.Main.FeelsLike,
		TempMin:     out.Main.TempMin,
		TempMax:     out.Main.TempMax,
		Humidity:    out.Main.Humidity,
	}, nil
}

func (owm *OpenWeatherMap) GetForecast() ([]*Weather, error) {
	fmt.Println(owm.location.Lat, owm.location.Lat)
	url := fmt.Sprintf("%s/onecall?lat=%.3f&lon=%.3f&appid=%s&units=metric", OpenWeatherMapApiUrl, owm.location.Lat, owm.location.Lon, owm.secret)
	fmt.Println(url)
	req, err := owm.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("%s: %s", OpenWeatherMapName, err.Error())
	}

	defer req.Body.Close()

	if req.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("%s: invalid api key", OpenWeatherMapName)
	}

	b, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("%s: %s", OpenWeatherMapName, err.Error())
	}

	var out *OpenWeatherMapWeatherForecastResponse
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, fmt.Errorf("%s: %s", OpenWeatherMapName, err.Error())
	}

	list := make([]*Weather, len(out.Hourly))
	for i, w := range out.Hourly {
		list[i] = &Weather{
			Type:        w.Weather[0].Main,
			DateTime:    time.Unix(w.Dt, 0),
			TimeZone:    out.TimeZone,
			Temp:        w.Temp,
			TempFeeling: w.FeelsLike,
			TempMin:     w.TempMin,
			TempMax:     w.TempMax,
			Humidity:    w.Humidity,
		}
	}

	return list, nil
}
