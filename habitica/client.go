package habitica

import (
	"fmt"
	"net/http"
)

type HabiticaTransport struct {
	Transport http.RoundTripper
	UserId    string
	ApiKey    string
}

func (h *HabiticaTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	newReq := req.Clone(req.Context())
	newReq.Header.Set("x-api-user", h.UserId)
	newReq.Header.Set("x-api-key", h.ApiKey)
	newReq.Header.Set("x-client", fmt.Sprintf("%s-vitus", h.UserId))
	return h.Transport.RoundTrip(newReq)

}

func NewHabiticaClient(userId, apiKey string) *http.Client {
	client := &http.Client{
		Transport: &HabiticaTransport{
			Transport: http.DefaultTransport,
			UserId:    userId,
			ApiKey:    apiKey,
		},
	}
	return client
}
