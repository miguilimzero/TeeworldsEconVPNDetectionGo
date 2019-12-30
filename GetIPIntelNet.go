package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"runtime/debug"
	"strconv"
)

// GetIPIntelNet :
type GetIPIntelNet struct {
	Client    *http.Client
	Limiter   *RateLimiter
	Email     string
	Threshold float64
}

// Name : Get API Name
func (giin GetIPIntelNet) Name() string {
	return "https://getipintel.net"
}

type getIPIntelResponseDataStatus struct {
	Status string `json:"status"`
}

type getIPIntelResponseDataSuccess struct {
	Status      string `json:"status"`
	Result      string `json:"result"`
	QueryIP     string `json:"queryIP"`
	QueryFlags  string `json:"queryFlags"`
	QueryFormat string `json:"queryFormat"`
	Contact     string `json:"contact"`
}

type getIPIntelResponseDataError struct {
	Status      string `json:"status"`
	Result      string `json:"result"`
	Message     string `json:"message"`
	QueryIP     string `json:"queryIP"`
	QueryFlags  string `json:"queryFlags"`
	QueryFormat string `json:"queryFormat"`
	Contact     string `json:"contact"`
}

// Fetch :
func (giin GetIPIntelNet) Fetch(IP string) (string, error) {
	u, _ := url.Parse("http://check.getipintel.net/check.php")

	// build url query
	params := url.Values{}
	params.Add("ip", IP)
	params.Add("contact", giin.Email)
	params.Add("format", "json")

	u.RawQuery = params.Encode()

	request, _ := http.NewRequest("GET", u.String(), nil)
	response, err := giin.Client.Do(request)

	if err != nil {
		debug.PrintStack()
		return "", err
	}

	// status
	statusCode := response.StatusCode
	// body
	bytes, _ := ioutil.ReadAll(response.Body)

	if statusCode == 200 {

		status := getIPIntelResponseDataStatus{}
		err = json.Unmarshal(bytes, &status)
		if err != nil {
			return "", err
		}

		if status.Status == "success" {
			successJSON := getIPIntelResponseDataSuccess{}
			err := json.Unmarshal(bytes, &successJSON)

			if err != nil {
				return "", errors.New("failed to unmarshal SUCCESS response message")
			}

			return successJSON.Result, nil
		} else if status.Status == "error" {
			errorJSON := getIPIntelResponseDataError{}
			err := json.Unmarshal(bytes, &errorJSON)

			if err != nil {
				return "", errors.New("failed to unmarshal error response message")
			}
			return "", errors.New(errorJSON.Message)
		}
	}

	return "", errors.New("Unknown response from api: " + string(bytes))

}

// IsVpn :
func (giin GetIPIntelNet) IsVpn(IP string) (bool, error) {
	if !giin.Limiter.Allow() {
		return false, errors.New("API GetIPIntel reached the daily limit")
	}

	body, err := giin.Fetch(IP)
	if err != nil {
		log.Println(err.Error())
		return false, errors.New("failed to fetch data")
	}

	vpnProbability, err := strconv.ParseFloat(body, 64)

	if err != nil {
		log.Println("Could not convert '", body, "' to float64")
		return false, errors.New("Failed to convert retrieved value to float64")
	}

	if 0.0 <= vpnProbability && vpnProbability <= 1.0 && vpnProbability >= giin.Threshold {
		return true, nil
	}
	return false, nil
}
