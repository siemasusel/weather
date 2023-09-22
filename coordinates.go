package weather

import (
	"context"
)

type CoordinatesService interface {
	GetCoordinatesForUSCity(ctx context.Context, city string) (Coordinates, error)
}

type Coordinates struct {
	Lat float64
	Lon float64
}
