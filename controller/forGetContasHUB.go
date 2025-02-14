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
	alpesHubID := accountId
	authURL := "http://localhost:7070/AuthModelos/" + alpesHubID

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Permitir até 10 redirecionamentos antes de falhar
			if len(via) > 10 {
				return fmt.Errorf("muitos redirecionamentos")
			}
			return nil
		},
	}

	// Variável para verificar se o fluxo foi finalizado
	done := false

	// Loop para seguir os redirecionamentos até chegar no status 200
	for !done {
		resp, err := client.Get(authURL)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro na requisição: %v", err)})
			return
		}
		defer resp.Body.Close()

		// Se já chegou no 200, finalizar
		if resp.StatusCode == 200 {
			done = true
		} else if resp.StatusCode != http.StatusFound {
			// Se o status não for 302 ou 200, retornar erro
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro inesperado: status %d", resp.StatusCode)})
			return
		} else {
			// Se for 302, seguir o redirecionamento
			authURL = resp.Header.Get("Location")
			if authURL == "" {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Redirecionamento sem Location"})
				return
			}
		}
	}

	// Após todos os redirecionamentos, responder sucesso
	c.JSON(http.StatusOK, gin.H{"message": "Fluxo de autorização concluído", "status": "200 OK"})
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
