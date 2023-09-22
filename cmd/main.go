package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/siemasusel/weather"
	"github.com/siemasusel/weather/nws"
	"github.com/siemasusel/weather/photon"

	"golang.org/x/exp/slog"
)

const refreshRate = 10 * time.Second

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	httpClient := &http.Client{
		Timeout: 15 * time.Second,
	}
	photonAPI := photon.NewPhotonAPI(httpClient)
	nwsAPI := nws.NewNWSAPI(httpClient)

	service := weather.NewService(photonAPI, nwsAPI)

	if len(os.Args) < 2 {
		slog.ErrorContext(ctx, "Missing city argument.")
		os.Exit(1)
	}
	city := os.Args[1]

	coordInfo, err := service.GetCoordinatesInfo(ctx, city)
	if err != nil {
		slog.ErrorContext(ctx, "Unable to get get coordinates information", "city", city, "err", err.Error())
		os.Exit(1)
	}

	displayForecastInformation(ctx, coordInfo, service)
	for {
		select {
		case <-ctx.Done():
			slog.Info("Application shutting down")
			return
		case <-time.After(refreshRate):
			displayForecastInformation(ctx, coordInfo, service)
		}
	}
}

func displayForecastInformation(ctx context.Context, coordInfo weather.Coordinates, service *weather.Service) {
	forecast, alerts, err := service.GetForecastAndAlerts(ctx, coordInfo)
	if err != nil {
		slog.Error("Unable to get forecast or alerts information.", "err", err.Error())
		return
	}

	printForecast(forecast)
	printAlerts(alerts)
}

func printForecast(forecast weather.Forecast) {
	slog.Info("New forecast information.",
		"temperature", fmt.Sprintf("%.2f %s", forecast.Temperature, forecast.TemperatureUnit),
		"wind_speed", forecast.WindSpeed,
		"wind_direction", forecast.WindDirection,
		"probability_of_precipitation", forecast.ProbabilityOfPrecipitation,
	)
}

func printAlerts(alerts []weather.ForecastAlert) {
	for _, alert := range alerts {
		slog.Warn("Alert",
			"effective", alert.Effective,
			"expires", alert.Expires,
			"certainty", alert.Certainty,
			"urgency", alert.Urgency,
			"description", alert.Description,
			"instruction", alert.Instruction,
		)
	}
}
