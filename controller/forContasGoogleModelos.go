package controller

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	// Variáveis globais para armazenar o token de acesso
	Token  *oauth2.Token
	Tokens = make(map[string]*oauth2.Token) // Mapeia o customerID para o token
)

func initializeOAuthConfig() {
	clientID := "459162752034-80q8hukn6eu45nt4fi0sic5ac51vc3ks.apps.googleusercontent.com"
	clientSecret := "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"
	config = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  "http://localhost:7070/oauth2callbackMultipleGetModelos",
		Scopes: []string{
			"https://www.googleapis.com/auth/adwords",
			"https://www.googleapis.com/auth/spreadsheets"},
	}
}

func AuthGetAcessTokenMultipleGetModelosGETFOR(c *gin.Context) (*oauth2.Token, error) {
	// Gerar a URL de login do Google para autorização
	url := config.AuthCodeURL("", oauth2.AccessTypeOffline)
	c.Redirect(http.StatusFound, url)

	// Após o redirecionamento, é necessário capturar o código de autorização e trocar por um token
	code := c.DefaultQuery("code", "")
	if code == "" {
		return nil, fmt.Errorf("código de autorização não encontrado")
	}

	// Trocar o código de autorização por um Access Token
	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("falha ao trocar o código por token: %v", err)
	}

	return token, nil
}

func OAuth2CallbackMultipleAccountsGetModelosGETFOR(c *gin.Context) {
	// Pegar o código de autorização retornado pelo Google
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Faltando o código de autorização"})
		return
	}

	// Trocar o código de autorização por um Access Token
	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		log.Printf("Falha ao trocar o código por token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao obter o Access Token"})
		return
	}

	// Armazenar o token globalmente ou por customerID
	customerID := c.DefaultQuery("customer_id", "")
	if customerID != "" {
		Tokens[customerID] = token
	} else {
		Token = token
	}

	// Agora você pode usar o token conforme necessário
	log.Printf("Token salvo: %s\n", token.AccessToken)
	c.JSON(http.StatusOK, gin.H{"message": "Token armazenado com sucesso"})
}

func CentralizaGetFuncoesGoogleAds(c *gin.Context) {
	// Obter o token antes de prosseguir
	token, err := AuthGetAcessTokenMultipleGetModelosGETFOR(c)
	if err != nil {
		c.JSON(500, gin.H{"error": "Erro ao obter o token"})
		return
	}

	// Agora que temos o token, podemos continuar com o processamento
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
			accountId := record[0]                                   // Pegar o primeiro campo da linha
			GetTopAndWorstAdGroupsForModelosFOR(c, accountId, token) // Passar o token para a próxima função
		}
	}

	// Resposta final após o loop
	c.JSON(200, gin.H{"message": "Processamento concluído para todas as contas"})
}
