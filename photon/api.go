package photon

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/siemasusel/weather"

	"github.com/pkg/errors"
)

const apiURLStr = "https://photon.komoot.io/api/"

type PhotonAPI struct {
	client *http.Client
}

func NewPhotonAPI(client *http.Client) *PhotonAPI {
	return &PhotonAPI{
		client: client,
	}
}

func (p *PhotonAPI) GetCoordinatesForUSCity(ctx context.Context, city string) (weather.Coordinates, error) {
	apiURL, err := url.Parse(apiURLStr)
	if err != nil {
		return weather.Coordinates{}, errors.Wrapf(err, "unable to parse api url '%s'", apiURLStr)
	}

	values := apiURL.Query()
	values.Add("q", city)

	apiURL.RawQuery = values.Encode()

	resp, err := p.client.Get(apiURL.String())
	if err != nil {
		return weather.Coordinates{}, errors.Wrap(err, "unable to request photon api")
	}

	defer resp.Body.Close()

	result, err := parseResponse(resp)
	if err != nil {
		return weather.Coordinates{}, err
	}

	return getUSCountryCoordinates(result)
}

func parseResponse(resp *http.Response) (response, error) {
	if resp.StatusCode != http.StatusOK {
		return response{}, errors.Errorf("photon api response with non-200 status code %d", resp.StatusCode)
	}

	var result response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return response{}, errors.Wrapf(err, "unable to parse photon response")
	}

	return result, nil
}

func getUSCountryCoordinates(result response) (weather.Coordinates, error) {
	for _, f := range result.Features {
		if f.Properties.CountryCode == "US" && len(f.Geometry.Coordinates) == 2 {
			return weather.Coordinates{
				Lon: f.Geometry.Coordinates[0],
				Lat: f.Geometry.Coordinates[1],
			}, nil
		}
	}

	return weather.Coordinates{}, errors.New("could not find US coordinates for your city in photon api")
}

type response struct {
	Features []featureResponse `json:"features"`
}

type featureResponse struct {
	Geometry   geometryFeatureResponse   `json:"geometry"`
	Properties propertiesFeatureResponse `json:"properties"`
}

type geometryFeatureResponse struct {
	Coordinates []float64 `json:"coordinates"`
}

type propertiesFeatureResponse struct {
	CountryCode string `json:"countrycode"`
	State       string `json:"state"`
}
