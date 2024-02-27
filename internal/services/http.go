package services

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/silenta-salmans/sbchecker/internal"
	"github.com/silenta-salmans/sbchecker/internal/logger"
	"github.com/silenta-salmans/sbchecker/models"
)

var URL = "https://support.activision.com/api/bans/appeal?locale=en"

func VerifySSOCookie(ssoCookie string) (int, error) {
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return 0, err
	}

	headers := internal.GenerateHeaders(ssoCookie)

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Log.WithError(err).Error("Error reading response body")
		return 0, err
	}

	if string(body) == "" {
		return 0, nil
	}

	return resp.StatusCode, nil
}

func CheckAccount(ssoCookie string) (models.Status, error) {
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return models.StatusUnknown, err
	}

	headers := internal.GenerateHeaders(ssoCookie)

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return models.StatusUnknown, err
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
		return models.StatusUnknown, err
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
