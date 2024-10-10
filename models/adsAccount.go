package models

type AdsAccount struct {
	Impressions int     `json:"Impressions"`
	Click       int     `json:"Click"`
	Cost        float64 `json:"Cost"`
	AccountId   string  `json:"AccountId"`
	Name        string  `json:"Name"`
}
