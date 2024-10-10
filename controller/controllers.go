package controller

import (
	"context"
	"google-ads-get/models"
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

func InitialPage(c *gin.Context) {

}

// ID DO CLIENTE teste:
// 1001930688294-4ikpvo0nks0kqv2cphtmmr8sr7fprgdt.apps.googleusercontent.com

// CHAVE SECRETA teste:
// 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

//CHAVE DO GOOGLE ADS teste:
//pwqUPX4NQRv7Osp-rqGjww

var config *oauth2.Config

func Init() {
	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	config = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  "http://localhost:7070/oauth2callback",              // Modifique conforme o necessário
		Scopes:       []string{"https://www.googleapis.com/auth/adwords"}, // Escopo para acessar a API do Google Ads
	}
}
func AuthGetAcessToken(c *gin.Context) {
	// Obter o ID da conta da URL (parâmetro de rota, e não query string)
	accountID := c.Param("customerID")

	// Gerar a URL de login do Google para autorização, incluindo o ID da conta no "state"
	url := config.AuthCodeURL(accountID, oauth2.AccessTypeOffline)

	// Redirecionar o usuário para a URL de login
	c.Redirect(http.StatusFound, url)
}

func OAuth2Callback(c *gin.Context) {
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

	// Obter o accountID do state
	accountID := c.Query("state")

	// Redirecionar para a rota que faz a requisição à API do Google Ads
	c.Redirect(http.StatusFound, "/Mcc/"+accountID+"?access_token="+token.AccessToken)
}

// Dividindo as campanhas
func GetGoogleAdsDataAutorizations(c *gin.Context) {
	// Pegando os tokens do ambiente
	accessToken := c.Query("access_token")
	developerToken := "MvSisVf6otSXPLwUvGbUaw"
	customerID := c.Params.ByName("customerID") // ID da conta específica (filha)
	loginCustomerID := "9155619419"             // Substitua pelo ID da MCC

	if customerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CustomerID inválido"})
		return
	}
	// Verificar se as variáveis de ambiente estão definidas
	if accessToken == "" || developerToken == "" || customerID == "" || loginCustomerID == "" {
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
	now := time.Now()

	// Primeiro dia do mês atual
	firstDay := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	// Último dia do mês atual
	lastDay := firstDay.AddDate(0, 1, -1)

	// Formatando as datas como string no formato 'YYYY-MM-DD'
	startDate := firstDay.Format("2006-01-02")
	endDate := lastDay.Format("2006-01-02")
	// Fazer uma requisição para obter campanhas
	//querySearch := "SELECT campaign.id, campaign.name, metrics.impressions, metrics.clicks, metrics.cost_micros FROM campaign WHERE campaign.status = 'ENABLED' AND segments.date BETWEEN '2024-10-01' AND '2024-10-31'"
	//querySearch := "SELECT customer.id, customer.descriptive_name, metrics.impressions, metrics.clicks, metrics.cost_micros FROM customer WHERE segments.date BETWEEN '2024-10-01' AND '2024-10-31'"
	querySearch := "SELECT customer.id, customer.descriptive_name, metrics.impressions, metrics.clicks, metrics.cost_micros FROM customer WHERE segments.date BETWEEN '" + startDate + "' AND '" + endDate + "'"

	req := &services.SearchGoogleAdsRequest{
		CustomerId: customerID, // ID da conta individual dentro da MCC
		Query:      querySearch,
	}

	// Enviar a requisição e capturar a resposta
	res, err := client.Search(ctx, req)
	if err != nil {
		// Captura detalhes do erro usando status.FromError
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

	// Processar os resultados e enviar a resposta
	var campaigns []models.AdsAccount
	for _, row := range res.Results {
		accountName := ""
		if row.Customer.DescriptiveName != nil {
			accountName = *row.Customer.DescriptiveName
		}

		campaigns = append(campaigns, models.AdsAccount{
			Impressions: int(*row.Metrics.Impressions),
			Click:       int(*row.Metrics.Clicks),
			Cost:        float64(*row.Metrics.CostMicros) / 1e6,
			AccountId:   c.Params.ByName("customerID"),
			Name:        accountName, // Nome da conta
		})
	}

	// Retornar os dados das campanhas
	c.JSON(http.StatusOK, gin.H{
		"campaigns": campaigns,
	})
}
