package services

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"codstatusbot2.0/logger"
	"codstatusbot2.0/models"
)

var url1 = "https://support.activision.com/api/bans/appeal?locale=en"
var url2 = "https://support.activision.com/api/profile?accts=false"

func VerifySSOCookie(ssoCookie string) (int, error) {
	logger.Log.Infof("Verifying SSO cookie: %s", ssoCookie)
	req, err := http.NewRequest("GET", url1, nil)
	if err != nil {
		return 0, errors.New("failed to create HTTP request to verify SSO cookie")
	}
	headers := GenerateHeaders(ssoCookie)
	logger.Log.Info("Creating headers")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, errors.New("failed to send HTTP request to verify SSO cookie")
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, errors.New("failed to read response body from verify SSO cookie request")
	}

	if string(body) == "" {
		return 0, nil
	}
	return resp.StatusCode, nil
}

func CheckAccount(ssoCookie string) (models.Status, error) {
	logger.Log.Info("Starting CheckAccount function")
	req, err := http.NewRequest("GET", url1, nil)
	if err != nil {
		return models.StatusUnknown, errors.New("failed to create HTTP request to check account")
	}
	headers := GenerateHeaders(ssoCookie)
	logger.Log.Info("Creating headers")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return models.StatusUnknown, errors.New("failed to send HTTP request to check account")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return models.StatusUnknown, errors.New("failed to read response body from check account request")
	}
	if string(body) == "" {
		return models.StatusInvalidCookie, nil
	}

	var data struct {
		Error   string `json:"error"`
		Success string `json:"success"`
		Ban     []struct {
			Enforcement string `json:"enforcement"`
			Title       string `json:"title"`
			CanAppeal   bool   `json:"canAppeal"`
		} `json:"bans"`
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return models.StatusUnknown, errors.New("failed to decode JSON response from check account request")
	}
	if data.Error != "" || data.Success != "true" {
		return models.StatusUnknown, errors.New("error checking account status: " + data.Error)
	}
	if len(data.Ban) == 0 {
		return models.StatusGood, nil
	}
	for _, ban := range data.Ban {
		if ban.Enforcement == "PERMANENT" {
			return models.StatusPermaban, nil
		} else if ban.Enforcement == "UNDER_REVIEW" {
			return models.StatusShadowban, nil
		} else {
			return models.StatusGood, nil
		}
	}
	return models.StatusUnknown, nil
}

func CheckAccountAge(ssoCookie string) (int, int, int, error) {
	logger.Log.Info("Starting CheckAccountAge function")
	req, err := http.NewRequest("GET", url2, nil)
	if err != nil {
		return 0, 0, 0, errors.New("failed to create HTTP request to check account age")
	}
	headers := GenerateHeaders(ssoCookie)
	logger.Log.Info("Creating headers")
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
		return 0, 0, 0, errors.New("failed to decode JSON response from check account age request")
	}

	created, err := time.Parse(time.RFC3339, data.Created)
	if err != nil {
		return 0, 0, 0, errors.New("failed to parse created date in check account age request")
	}

	duration := time.Since(created)
	years := int(duration.Hours() / 24 / 365)
	months := int(duration.Hours()/24/30) % 12
	days := int(duration.Hours()/24) % 365 % 30

	return years, months, days, nil
}
