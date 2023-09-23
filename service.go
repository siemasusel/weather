package weather

import (
	"context"

	"github.com/pkg/errors"
)

type Service struct {
	coordService    CoordinatesService
	forecastService ForecastService
}

func NewService(coordService CoordinatesService, forecastService ForecastService) *Service {
	return &Service{
		coordService:    coordService,
		forecastService: forecastService,
	}
}

func (s *Service) GetCoordinates(ctx context.Context, city string) (Coordinates, error) {
	return s.coordService.GetCoordinatesForUSCity(ctx, city)
}

func (s *Service) GetForecastAndAlerts(ctx context.Context, coords Coordinates) (Forecast, []ForecastAlert, error) {
	forecast, err := s.forecastService.GetCurrentForecast(ctx, coords.Lat, coords.Lon)
	if err != nil {
		return Forecast{}, nil, errors.Wrap(err, "unable to get forecast information")
	}

	alerts, err := s.forecastService.GetAlerts(ctx, coords.Lat, coords.Lon)
	if err != nil {
		return Forecast{}, nil, errors.Wrap(err, "unable to get alerts information")
	}

	return forecast, alerts, nil
}
