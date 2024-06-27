package services

import (
	"codstatusbot2.0/logger"
	"codstatusbot2.0/models"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"
)

var url1 = "https://support.activision.com/api/bans/appeal?locale=en"
var url2 = "https://support.activision.com/api/profile?accts=false"

// var url3 = "https://profile.callofduty.com/promotions/redeemCode/"

/*
func ClaimSingleReward(ssoCookie, code string) (string, error) {
	logger.Log.Info("Starting ClaimSingleReward function")
	req, err := http.NewRequest("POST", url3, strings.NewReader(fmt.Sprintf("code=%s", code)))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request to claim reward: %w", err)
	}
	headers := GeneratePostHeaders(ssoCookie)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send HTTP request to claim reward: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}
	if strings.Contains(string(body), "redemption-success") {
		start := strings.Index(string(body), "Just Unlocked:<br><br><div class=\"accent-highlight mw2\">")
		end := strings.Index(string(body), "</div></h4>")
		if start != -1 && end != -1 {
			unlockedItem := strings.TrimSpace(string(body)[start+len("Just Unlocked:<br><br><div class=\"accent-highlight mw2\">") : end])
			return fmt.Sprintf("Successfully claimed reward: %s", unlockedItem), nil
		}
		return "Successfully claimed reward, but couldn't extract details", nil
	}
	logger.Log.Infof("Unexpected response body: %s", string(body))
	return "", fmt.Errorf("failed to claim reward: unexpected response")
}
*/

func VerifySSOCookie(ssoCookie string) bool {
	logger.Log.Infof("Verifying SSO cookie: %s ", ssoCookie)
	req, err := http.NewRequest("GET", url2, nil)
	if err != nil {
		logger.Log.WithError(err).Error("Error creating verification request")
		return false
	}
	headers := GenerateHeaders(ssoCookie)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Log.WithError(err).Error("Error sending verification request")
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		logger.Log.Errorf("Invalid SSOCookie, status code: %d ", resp.StatusCode)
		return false
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Log.WithError(err).Error("Error reading verification response body")
		return false
	}
	if len(body) == 0 {
		logger.Log.Error("Invalid SSOCookie, response body is empty")
		return false
	}
	return true
}

func CheckAccount(ssoCookie string) (models.Status, error) {
	logger.Log.Info("Starting CheckAccount function")
	req, err := http.NewRequest("GET", url1, nil)
	if err != nil {
		return models.StatusUnknown, errors.New("failed to create HTTP request to check account")
	}
	headers := GenerateHeaders(ssoCookie)
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
	// logger.Log.Info("Response Body: ", string(body))
	var data struct {
		Ban []struct {
			Enforcement string `json:"enforcement"`
			Title       string `json:"title"`
			CanAppeal   bool   `json:"canAppeal"`
		} `json:"bans"`
	}
	if string(body) == "" {
		return models.StatusInvalidCookie, nil
	}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err == nil {
		return models.StatusUnknown, errors.New("failed to decode JSON response possible no response was received")
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
	logger.Log.Info("Starting CheckAccountAge function")
	req, err := http.NewRequest("GET", url2, nil)
	if err != nil {
		return 0, 0, 0, errors.New("failed to create HTTP request to check account age")
	}
	headers := GenerateHeaders(ssoCookie)
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
