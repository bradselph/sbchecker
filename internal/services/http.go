package services

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"sbchecker/internal"
	"sbchecker/internal/logger"
	"sbchecker/models"
)

var URL1 = "https://support.activision.com/api/bans/appeal?locale=en"
var URL2 = "https://support.activision.com/api/profile?accts=false"

func VerifySSOCookie(ssoCookie string) (int, error) {
	req, err := http.NewRequest("GET", URL1, nil)
	if err != nil {
		return 0, errors.New("failed to create HTTP request to verify SSO cookie")
	}
	headers := internal.GenerateHeaders(ssoCookie)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, errors.New("failed to send HTTP request to verify SSO cookie")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Log.WithError(err).Error("Error reading response body from verify SSO cookie request")
		return 0, errors.New("failed to read response body from verify SSO cookie request")
	}
	if string(body) == "" {
		return 0, nil
	}
	return resp.StatusCode, nil
}
func CheckAccount(ssoCookie string) (models.Status, error) {
	req, err := http.NewRequest("GET", URL1, nil)
	if err != nil {
		return models.StatusUnknown, errors.New("failed to create HTTP request to check account")
	}
	headers := internal.GenerateHeaders(ssoCookie)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return models.StatusUnknown, errors.New("failed to send HTTP request to check account")
	}
	defer resp.Body.Close()
	var data struct {
		Ban []struct {
			Enforcement string `json:"enforcement"`
			Title       string `json:"title"`
			CanAppeal   bool   `json:"canAppeal"`
		} `json:"bans"`
	}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return models.StatusUnknown, errors.New("failed to decode JSON response from check account request")
	}
	if len(data.Ban) == 0 {
		return models.StatusGood, nil
	} else {
		for _, ban := range data.Ban {
			if ban.Enforcement == "PERMANENT" {
				return models.StatusPermaban, nil
			} else if ban.Enforcement == "UNDER_REVIEW" {
				return models.StatusShadowban, nil
			} else {
				return models.StatusGood, nil
			}
		}
	}
	return models.StatusUnknown, nil
}
func CheckAccountAge(ssoCookie string) (int, int, int, error) {
	req, err := http.NewRequest("GET", URL2, nil)
	if err != nil {
		return 0, 0, 0, errors.New("failed to create HTTP request to check account age")
	}
	headers := internal.GenerateHeaders(ssoCookie)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, 0, 0, errors.New("failed to send HTTP request to check account age")
	}
	defer resp.Body.Close()
	var data struct {
		Created string `json:"created"`
	}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		logger.Log.WithError(err).Error("Error decoding JSON response from check account age request")
		return 0, 0, 0, errors.New("failed to decode JSON response from check account age request")
	}
	created, err := time.Parse(time.RFC3339, data.Created)
	if err != nil {
		logger.Log.WithError(err).Error("Error parsing created date in check account age request")
		return 0, 0, 0, errors.New("failed to parse created date in check account age request")
	}
	duration := time.Since(created)
	years := int(duration.Hours() / 24 / 365)
	months := int(duration.Hours()/24/30) % 12
	days := int(duration.Hours()/24) % 365 % 30
	return years, months, days, nil
}
