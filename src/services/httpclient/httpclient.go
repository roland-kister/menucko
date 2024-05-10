package httpclient

import (
	"io"
	"net/http"
)

const UserAgent = "Mozilla/5.0"

type HTTPClient interface {
	DownloadHTML(url string) (string, error)
}

type DevHTTPClient struct {
	HTMLContent string
}

func (httpClient DevHTTPClient) DownloadHTML(url string) (string, error) {
	return httpClient.HTMLContent, nil
}

type ProdHTTPClient struct{}

func (ProdHTTPClient) DownloadHTML(url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", UserAgent)

	client := http.Client{}

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}

	content, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return string(content), nil
}
