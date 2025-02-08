package models

type AdGroup struct {
	Name        string `json:"name"`
	Impressions int    `json:"impressions"`
}

type ModelSearchSummary struct {
	CampaignName string    `json:"campaign_name"`
	TopModels    []AdGroup `json:"top_models"`
	BottomModels []AdGroup `json:"bottom_models"`
}
