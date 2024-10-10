package controller

import (
	"bufio"
	"context"
	"fmt"
	"google-ads-get/models"
	"google-ads-get/sheets"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shenzhencenter/google-ads-pb/services"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func AuthGetAcessTokenMultipleAccounts(c *gin.Context) {
	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	config = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  "http://localhost:7070/oauth2callbackMultipleAccounts", // Modifique conforme o necessário
		Scopes: []string{
			"https://www.googleapis.com/auth/adwords",
			"https://www.googleapis.com/auth/spreadsheets"}, // Escopo para acessar a API do Google Ads
	}
	// Obter o ID da conta da URL (parâmetro de rota, e não query string)
	//accountID := c.Param("customerID")

	// Gerar a URL de login do Google para autorização, incluindo o ID da conta no "state"
	url := config.AuthCodeURL("", oauth2.AccessTypeOffline)

	// Redirecionar o usuário para a URL de login
	c.Redirect(http.StatusFound, url)
}
func OAuth2CallbackMultipleAccounts(c *gin.Context) {
	// Pegar o código de autorização retornado pelo Google
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Faltando o código de autorização"})
		return
	}

	// Trocar o código de autorização por um Access Token
	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		log.Fatalf("Falha ao trocar o código por token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao obter o Access Token"})
		return
	}

	// Redirecionar para a rota que faz a requisição à API do Google Ads
	c.Redirect(http.StatusFound, "/MultipleAccounts?access_token="+token.AccessToken)
}
func GetGoogleAdsDataForMultipleAccounts(c *gin.Context) {
	fileName := "accountIds.txt"
	accountIDs := []string{}

	// Abrir o arquivo e ler os IDs de contas
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println("Erro ao abrir o arquivo:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read account IDs file"})
		return
	}
	defer file.Close()

	// Ler o arquivo linha por linha
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		accountID := scanner.Text()
		// Adicionar cada linha (ID) ao slice accountIDs
		accountIDs = append(accountIDs, accountID)
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Erro ao ler o arquivo:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read from file"})
		return
	}

	// Pegando o access token da query
	accessToken := c.Query("access_token")
	developerToken := "MvSisVf6otSXPLwUvGbUaw" // Coloque seu developer token correto aqui
	loginCustomerID := "9155619419"            // ID da MCC

	// Verificar se as variáveis necessárias estão presentes
	if accessToken == "" || developerToken == "" || loginCustomerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required environment variables"})
		return
	}

	// Criar um novo contexto e adicionar os headers necessários
	ctx := context.Background()
	headers := metadata.Pairs(
		"authorization", "Bearer "+accessToken,
		"developer-token", developerToken,
		"login-customer-id", loginCustomerID, // Aqui, use o ID da MCC
	)
	ctx = metadata.NewOutgoingContext(ctx, headers)

	// Credenciais para a conexão gRPC
	cred := grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(nil, ""))

	// Conectar ao servidor da API do Google Ads via gRPC
	conn, err := grpc.Dial("googleads.googleapis.com:443", cred)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Criar um cliente para o Google Ads
	client := services.NewGoogleAdsServiceClient(conn)

	// Obter a data atual para criar a query
	now := time.Now()
	firstDay := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	lastDay := firstDay.AddDate(0, 1, -1)
	startDate := firstDay.Format("2006-01-02")
	endDate := lastDay.Format("2006-01-02")

	// Criar um slice para armazenar todas as campanhas de todos os IDs
	var allCampaigns []models.AdsAccount

	// Iterar sobre a lista de IDs de contas
	for _, customerID := range accountIDs {
		// Fazer a query para a conta atual
		querySearch := "SELECT customer.id, customer.descriptive_name, metrics.impressions, metrics.clicks, metrics.cost_micros FROM customer WHERE segments.date BETWEEN '" + startDate + "' AND '" + endDate + "'"

		// Criar a requisição para a API do Google Ads
		req := &services.SearchGoogleAdsRequest{
			CustomerId: customerID, // ID da conta individual
			Query:      querySearch,
		}

		// Enviar a requisição
		res, err := client.Search(ctx, req)
		if err != nil {
			st, ok := status.FromError(err)
			if ok {
				// Retornar informações detalhadas do erro gRPC
				c.JSON(http.StatusInternalServerError, gin.H{
					"error_code":    st.Code(),
					"error_message": st.Message(),
					"details":       st.Details(),
				})
			} else {
				// Retornar erro genérico se não for um erro gRPC
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}

		// Processar os resultados da resposta e adicionar ao slice de campanhas
		for _, row := range res.Results {
			accountName := ""
			if row.Customer.DescriptiveName != nil {
				accountName = *row.Customer.DescriptiveName
			}

			allCampaigns = append(allCampaigns, models.AdsAccount{
				Impressions: int(*row.Metrics.Impressions),
				Click:       int(*row.Metrics.Clicks),
				Cost:        float64(*row.Metrics.CostMicros) / 1e6,
				AccountId:   customerID,  // Armazena o ID da conta atual
				Name:        accountName, // Nome da conta
			})
		}
	}
	err = sheets.WriteToGoogleSheets(ctx, accessToken, allCampaigns)
	if err != nil {
		st, _ := status.FromError(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error_code": st.Code(),
			"error_message": st.Message(),
			"details":       st.Details(),
		})
		return
	}

	// Retornar os dados combinados de todas as campanhas
	c.JSON(http.StatusOK, gin.H{
		"message":   "Dados inseridos com sucesso na planilha!",
		"campaigns": allCampaigns,
	})
}
