package controller

import (
	"github.com/gin-gonic/gin"
)

func ProcessAccountsAlpesHub(c *gin.Context, alpesHubID string) {
	conversions := PostLeadReportHandler(c, alpesHubID)

	results := map[string]float64{
		alpesHubID: float64(conversions),
	}

	print("results", results)
}
