package agent

import (
	"context"
	"fmt"
	"github.com/eldius/document-feeder/internal/config"
	"github.com/tmc/langchaingo/tools"
	"net/url"
	"strconv"
	"strings"
)

var (
	_ tools.Tool = &WeatherAgent{}

	openMeteoDailyMetrics = []string{
		"temperature_2m_max",
		"temperature_2m_min",
		"apparent_temperature_max",
		"apparent_temperature_min",
		"sunrise",
		"sunset",
		"daylight_duration",
		"sunshine_duration",
		"uv_index_max",
		"uv_index_clear_sky_max",
	}

	openMeteoHourlyMetrics = []string{
		"temperature_2m",
		"relative_humidity_2m",
		"apparent_temperature",
		"precipitation_probability",
		"rain",
		"visibility",
		"cloud_cover",
	}

	openMeteoCurrentMetrics = []string{
		"temperature_2m",
		"relative_humidity_2m",
		"apparent_temperature",
		"is_day",
		"precipitation",
		"rain",
		"cloud_cover",
		"wind_speed_10m",
		"wind_direction_10m",
	}
)

const (
	openMeteoBaseURL = "https://api.weather.com/forecast"
)

type WeatherParameters struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lon"`
	Timezone  string  `json:"timezone"`
}

func getParameters() (WeatherParameters, error) {
	var params WeatherParameters
	if err := config.GetFetchConfigStruct("agents.weather", &params); err != nil {
		return WeatherParameters{}, fmt.Errorf("getting weather parameters: %w", err)
	}

	return params, nil
}

// buildWeatherURL creates a parameterized URL for weather API calls
func buildWeatherURL(baseURL string, params WeatherParameters, additionalParams map[string]string) (*url.URL, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parsing base URL: %w", err)
	}

	// Create url.Values instance for query parameters
	queryParams := url.Values{}

	// Add weather parameters
	queryParams.Set("lat", strconv.FormatFloat(params.Latitude, 'f', -1, 64))
	queryParams.Set("lon", strconv.FormatFloat(params.Longitude, 'f', -1, 64))
	queryParams.Set("timezone", params.Timezone)

	// Add any additional parameters
	for key, value := range additionalParams {
		queryParams.Set(key, value)
	}

	// Set the query parameters on the URL
	u.RawQuery = queryParams.Encode()

	return u, nil
}

type WeatherAgent struct {
	Parameters WeatherParameters
}

func NewWeatherAgent() (*WeatherAgent, error) {
	params, err := getParameters()
	if err != nil {
		return nil, fmt.Errorf("getting weather parameters: %w", err)
	}
	return &WeatherAgent{Parameters: params}, nil
}

func (w WeatherAgent) Name() string {
	return "weather_forecast"
}

func (w WeatherAgent) Description() string {
	return "Provides weather forecast for my location"
}

func (w WeatherAgent) Call(ctx context.Context, input string) (string, error) {
	additionalParams := map[string]string{
		"units":   "metric",
		"lang":    "pt",
		"daily":   strings.Join(openMeteoDailyMetrics, ","),
		"hourly":  strings.Join(openMeteoHourlyMetrics, ","),
		"current": strings.Join(openMeteoCurrentMetrics, ","),
	}

	weatherURL, err := buildWeatherURL(openMeteoBaseURL, w.Parameters, additionalParams)
	if err != nil {
		return "", fmt.Errorf("building weather URL: %w", err)
	}

	// Aqui você pode usar weatherURL.String() para fazer a requisição HTTP
	fmt.Printf("Weather API URL: %s\n", weatherURL.String())

	//TODO implement actual HTTP request
	panic("implement me")
}
