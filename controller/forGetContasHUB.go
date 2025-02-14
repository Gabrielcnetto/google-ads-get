package controller

import (
	"encoding/csv"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func ProcessAccountsAlpesHub(c *gin.Context) {
	file, err := os.Open("dadoscontas.csv")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao abrir o arquivo CSV"})
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Read()

	results := make(map[string]float64)
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		alpesHubID := record[1]
		conversions := PostLeadReportHandler(c, alpesHubID)

		results[alpesHubID] = float64(conversions) // Convertendo para float64
	}

	c.JSON(http.StatusOK, gin.H{"results": results})
}
