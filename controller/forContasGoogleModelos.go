package controller

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	// Variáveis globais para armazenar o token de acesso

	tokens = make(map[string]*oauth2.Token) // Mapeia o customerID para o token
)

func AuthGetAcessTokenMultipleGetModelosGETFOR(c *gin.Context) {
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

	// Gerar a URL de login do Google para autorização
	// O 'state' pode ser qualquer identificador dinâmico, aqui estamos deixando em branco
	url := config.AuthCodeURL("", oauth2.AccessTypeOffline)

	// Redirecionar o usuário para a URL de login
	c.Redirect(http.StatusFound, url)
}

func OAuth2CallbackMultipleAccountsGetModelosGETFOR(c *gin.Context) {
	// Pegar o código de autorização retornado pelo Google
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Faltando o código de autorização"})
		return
	}

	// Capturar o customerID do parâmetro "state"
	// O "state" pode ser deixado em branco no fluxo, mas poderia ser um identificador único se necessário
	customerID := c.Query("state")
	if customerID == "" {
		// Se o customerID não foi passado, podemos gerar um identificador único ou associá-lo ao usuário de outra maneira
		// Aqui vamos gerar um valor fictício como exemplo
		customerID = "defaultCustomerID"
	}

	// Trocar o código de autorização por um Access Token
	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		log.Printf("Falha ao trocar o código por token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao obter o Access Token"})
		return
	}

	// Armazenar o token no mapa de tokens
	tokens[customerID] = token

	// Agora você pode usar o token conforme necessário, por exemplo:
	log.Printf("Token salvo para %s: %s\n", customerID, token.AccessToken)

	// Redirecionar para a próxima parte do fluxo, sem a necessidade de passar o customerID
	//urlToGet := "/MultipleAccountsModelos/?access_token=" + token.AccessToken

	// Redirecionar o usuário
	print("O Token final é:", tokens)
	// c.Redirect(http.StatusFound, urlToGet)
}
