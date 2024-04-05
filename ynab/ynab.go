package ynab

import (
	"fmt"
	"net/http"
)

type YnabTransport struct {
	Transport http.RoundTripper
	Token     string
}

func (y *YnabTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	newReq := req.Clone(req.Context())
	newReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", y.Token))
	return y.Transport.RoundTrip(newReq)
}

func NewYnabClient(token string) *http.Client {
	client := &http.Client{
		Transport: &YnabTransport{
			Transport: http.DefaultTransport,
			Token:     token,
		},
	}
	return client
}
