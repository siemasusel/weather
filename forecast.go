package weather

import "context"

type ForecastService interface {
	GetCurrentForecast(ctx context.Context, lat, long float64) (Forecast, error)
	GetAlerts(ctx context.Context, lat, long float64) ([]ForecastAlert, error)
}

type Forecast struct {
	Temperature                float64
	TemperatureUnit            string
	WindSpeed                  string
	WindDirection              string
	ProbabilityOfPrecipitation int
}

type ForecastAlert struct {
	Effective   string
	Expires     string
	Certainty   string
	Urgency     string
	Description string
	Instruction string
}
