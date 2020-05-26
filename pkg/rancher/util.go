package rancher

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const (
	maxHTTPRedirect = 5
)

func DoGet(url, username, password, cacert string, insecure bool) ([]byte, error) {
	if url == "" {
		return nil, fmt.Errorf("Doing get: URL is nil")
	}

	client := &http.Client{
		Timeout: time.Duration(10 * time.Second),
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= maxHTTPRedirect {
				return fmt.Errorf("Stopped after %d redirects", maxHTTPRedirect)
			}
			if len(username) > 0 && len(password) > 0 {
				req.SetBasicAuth(username, password)
			}
			return nil
		},
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure},
		Proxy:           http.ProxyFromEnvironment,
	}

	if cacert != "" {
		// Get the SystemCertPool, continue with an empty pool on error
		rootCAs, _ := x509.SystemCertPool()
		if rootCAs == nil {
			rootCAs = x509.NewCertPool()
		}

		// Append our cert to the system pool
		if ok := rootCAs.AppendCertsFromPEM([]byte(cacert)); !ok {
			// log.Println("No certs appended, using system certs only")
		}
		transport.TLSClientConfig.RootCAs = rootCAs
	}
	client.Transport = transport

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("Doing get: %v", err)
	}
	if len(username) > 0 && len(password) > 0 {
		req.SetBasicAuth(username, password)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Doing get: %v", err)
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)

}

func NormalizeURL(url string) string {
	if url == "" || url == "https://" || url == "http://" {
		return ""
	}

	url = strings.TrimSuffix(url, "/")

	if !strings.HasSuffix(url, "/v3") {
		url = url + "/v3"
	}

	return url
}