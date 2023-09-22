package nws

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/siemasusel/weather"

	"github.com/pkg/errors"
)

const apiURL = "https://api.weather.gov"

type NWSAPI struct {
	client *http.Client
}

func NewNWSAPI(client *http.Client) *NWSAPI {
	return &NWSAPI{
		client: client,
	}
}

func (n *NWSAPI) GetCurrentForecast(ctx context.Context, lat, lon float64) (weather.Forecast, error) {
	point, err := n.getPoint(ctx, lat, lon)
	if err != nil {
		return weather.Forecast{}, errors.Wrap(err, "unable to get point from api")
	}

	resp, err := n.callAPI(ctx, point.ForecastHourly)
	if err != nil {
		return weather.Forecast{}, errors.Wrap(err, "unable to get forecast from api")
	}

	var forecast forecastResponse
	if err = json.NewDecoder(resp.Body).Decode(&forecast); err != nil {
		return weather.Forecast{}, errors.Wrapf(err, "unable to decode points")
	}

	if len(forecast.Periods) == 0 {
		return weather.Forecast{}, errors.Wrapf(err, "unable to get forecast information from api")
	}

	currentPeriod := forecast.Periods[0]

	return weather.Forecast{
		Temperature:                currentPeriod.Temperature,
		TemperatureUnit:            currentPeriod.TemperatureUnit,
		WindSpeed:                  currentPeriod.WindSpeed,
		WindDirection:              currentPeriod.WindDirection,
		ProbabilityOfPrecipitation: currentPeriod.ProbabilityOfPrecipitation.Value,
	}, nil
}

func (n *NWSAPI) GetAlerts(ctx context.Context, lat, lon float64) ([]weather.ForecastAlert, error) {
	endpoint := fmt.Sprintf("%s/alerts/active?point=%s,%s", apiURL, formatFloat(lat), formatFloat(lon))
	resp, err := n.callAPI(ctx, endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get alerts from api")
	}
	defer resp.Body.Close()

	var alerts alertResponse
	if err = json.NewDecoder(resp.Body).Decode(&alerts); err != nil {
		return nil, errors.Wrapf(err, "unable to decode alerts")
	}

	return parseAlerts(alerts), nil
}

func parseAlerts(alerts alertResponse) []weather.ForecastAlert {
	alertsModel := make([]weather.ForecastAlert, 0)
	for _, f := range alerts.Features {
		alertsModel = append(alertsModel, weather.ForecastAlert(f))
	}

	return alertsModel
}

func formatFloat(f float64) string {
	return strings.TrimRight(fmt.Sprintf("%.4f", f), "0")
}

func (n *NWSAPI) getPoint(ctx context.Context, lat float64, lon float64) (pointResponse, error) {
	endpoint := fmt.Sprintf("%s/points/%s,%s", apiURL, formatFloat(lat), formatFloat(lon))

	resp, err := n.callAPI(ctx, endpoint)
	if err != nil {
		return pointResponse{}, err
	}
	defer resp.Body.Close()

	var point pointResponse
	if err = json.NewDecoder(resp.Body).Decode(&point); err != nil {
		return pointResponse{}, errors.Wrapf(err, "unable to decode points")
	}

	return point, nil
}

func (n *NWSAPI) callAPI(ctx context.Context, endpoint string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/ld+json")

	resp, err := n.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		return nil, errors.Errorf("api.weather.gov response with non-200 status code %d and message %s", resp.StatusCode, string(body))
	}

	return resp, nil
}

type forecastResponse struct {
	Updated   string                   `json:"updated"`
	Units     string                   `json:"units"`
	Elevation forecastElevation        `json:"elevation"`
	Periods   []forecastResponsePeriod `json:"periods"`
}

type forecastElevation struct {
	Value float64 `json:"value"`
	Units string  `json:"unitCode"`
}

type forecastResponsePeriod struct {
	Number          int32   `json:"number"`
	Name            string  `json:"name"`
	StartTime       string  `json:"startTime"`
	EndTime         string  `json:"endTime"`
	Temperature     float64 `json:"temperature"`
	TemperatureUnit string  `json:"temperatureUnit"`
	WindSpeed       string  `json:"windSpeed"`
	WindDirection   string  `json:"windDirection"`
	Summary         string  `json:"shortForecast"`
	Details         string  `json:"detailedForecast"`

	ProbabilityOfPrecipitation probabilityOfPrecipitation `json:"probabilityOfPrecipitation"`
}

type probabilityOfPrecipitation struct {
	UnitCode string `json:"unitCode"`
	Value    int    `json:"value"`
}

type pointResponse struct {
	ID                  string `json:"@id"`
	CWA                 string `json:"cwa"`
	Office              string `json:"forecastOffice"`
	GridX               int64  `json:"gridX"`
	GridY               int64  `json:"gridY"`
	Forecast            string `json:"forecast"`
	ForecastHourly      string `json:"forecastHourly"`
	ObservationStations string `json:"observationStations"`
	ForecastGridData    string `json:"forecastGridData"`
	Timezone            string `json:"timeZone"`
	RadarStation        string `json:"radarStation"`
}

type alertResponse struct {
	Features []alertProperties `json:"@graph"`
}

type alertProperties struct {
	Effective   string `json:"effective"`
	Expires     string `json:"expires"`
	Certainty   string `json:"certainty"`
	Urgency     string `json:"urgency"`
	Description string `json:"description"`
	Instruction string `json:"instruction"`
}
