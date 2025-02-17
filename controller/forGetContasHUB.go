package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func ProcessAccountsAlpesHub(c *gin.Context, alpesHubID string) {
	conversions := PostLeadReportHandler(c, alpesHubID)

	results := map[string]float64{
		alpesHubID: float64(conversions),
	}

	c.JSON(http.StatusOK, gin.H{"results": results})
}
