package controller

import (
	"encoding/csv"
	"fmt"
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

// google ads

func ProcessAccountsAlpesGoogleAds(c *gin.Context, accountId string) {
	// ID fixo para a conta (exemplo)
	alpesHubID := accountId

	// URL do endpoint que queremos chamar
	authURL := "http://localhost:7070/AuthModelos/" + alpesHubID

	// Fazer a requisição GET
	respAuth, err := http.Get(authURL)
	if err != nil {
		// Se ocorrer um erro, retornar erro 500
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao chamar o endpoint de autorização"})
		return
	}
	defer respAuth.Body.Close()

	// Verificar se o status da resposta é OK (200)
	if respAuth.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro: status %d", respAuth.StatusCode)})
		return
	}

	// Se a resposta for bem-sucedida, retornar sucesso
	c.JSON(http.StatusOK, gin.H{"message": "Fluxo de autorização concluído", "status": respAuth.Status})
}
func ProcessAccountsFromCSV(c *gin.Context) {
	// Abrir o arquivo CSV
	file, err := os.Open("dadoscontas.csv")
	if err != nil {
		c.JSON(500, gin.H{"error": "Erro ao abrir o arquivo CSV"})
		return
	}
	defer file.Close()

	// Criar um leitor CSV
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		c.JSON(500, gin.H{"error": "Erro ao ler o arquivo CSV"})
		return
	}

	// Percorrer as linhas do CSV
	for _, record := range records {
		if len(record) > 0 { // Garantir que a linha não está vazia
			accountId := record[0] // Pegar o primeiro campo da linha

			// Chamar a função para processar a conta
			ProcessAccountsAlpesGoogleAds(c, accountId)
		}
	}

	// Resposta final após o loop
	c.JSON(200, gin.H{"message": "Processamento concluído para todas as contas"})
}
