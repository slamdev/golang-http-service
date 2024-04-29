package integration

import (
	"fmt"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"golang-http-service/api/petstore"
)

func CreatePetStoreAPIClient(url string) (petstore.ClientWithResponsesInterface, error) {
	transport := otelhttp.NewTransport(http.DefaultTransport)
	httpClient := &http.Client{Timeout: time.Minute, Transport: transport}
	apiClient, err := petstore.NewClientWithResponses(url, petstore.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("failed to create harbor api client: %w", err)
	}
	return apiClient, nil
}
