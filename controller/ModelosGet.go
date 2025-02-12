package controller

import (
	"context"
	"fmt"
	"google-ads-get/models"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shenzhencenter/google-ads-pb/services"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func AuthGetAcessTokenMultipleGetModelos(c *gin.Context) {
	clientID := "459162752034-80q8hukn6eu45nt4fi0sic5ac51vc3ks.apps.googleusercontent.com"
	clientSecret := "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"
	config = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  "http://localhost:7070/oauth2callbackMultipleGetModelos", // Modifique conforme o necessário
		Scopes: []string{
			"https://www.googleapis.com/auth/adwords",
			"https://www.googleapis.com/auth/spreadsheets"}, // Escopo para acessar a API do Google Ads
	}

	// Obter o ID da conta da URL (parâmetro de rota, e não query string)
	customerID := c.Param("customerID")

	// Gerar a URL de login do Google para autorização, incluindo o ID da conta no "state"
	url := config.AuthCodeURL(customerID, oauth2.AccessTypeOffline)

	// Redirecionar o usuário para a URL de login
	c.Redirect(http.StatusFound, url)
}
func OAuth2CallbackMultipleAccountsGetModelos(c *gin.Context) {
	// Pegar o código de autorização retornado pelo Google
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Faltando o código de autorização"})
		return
	}

	// Capturar o customerID do parâmetro "state"
	customerID := c.Query("state")
	if customerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Faltando o ID do cliente (state)"})
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
	urlToGet := "/MultipleAccountsModelos/" + customerID + "/?access_token=" + token.AccessToken
	c.Redirect(http.StatusFound, urlToGet)
}

func getPreviousMonthRange() (string, string, string) {
	now := time.Now()
	// Subtrai 1 mês da data atual
	firstDayOfPreviousMonth := now.AddDate(0, -1, 0)
	// Calcula o primeiro dia do mês anterior
	firstDay := time.Date(firstDayOfPreviousMonth.Year(), firstDayOfPreviousMonth.Month(), 1, 0, 0, 0, 0, time.UTC)
	// Calcula o último dia do mês anterior
	lastDay := firstDay.AddDate(0, 1, -1)

	// Nome do mês em português
	monthName := firstDay.Format("January 2006")
	// Traduzindo o mês para português
	monthNames := map[string]string{
		"January":   "Janeiro",
		"February":  "Fevereiro",
		"March":     "Março",
		"April":     "Abril",
		"May":       "Maio",
		"June":      "Junho",
		"July":      "Julho",
		"August":    "Agosto",
		"September": "Setembro",
		"October":   "Outubro",
		"November":  "Novembro",
		"December":  "Dezembro",
	}
	monthName = monthNames[firstDay.Format("January")]

	// Retorna os dois dias como strings no formato yyyy-mm-dd e o nome do mês em português
	return firstDay.Format("2006-01-02"), lastDay.Format("2006-01-02"), monthName
}

// Função para buscar os dados do Google Ads com base no mês anterior
func GetTopAndWorstAdGroupsForModelos(c *gin.Context, accountID string) string {
	accessToken := c.Query("access_token")
	if accessToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Faltando o access_token"})
		return "error"
	}

	// Definir as credenciais do Google Ads API
	developerToken := "MvSisVf6otSXPLwUvGbUaw" // Coloque seu developer token correto aqui
	loginCustomerID := "9155619419"            // ID da MCC

	// Criar um novo contexto e adicionar os headers necessários
	ctx := context.Background()
	headers := metadata.Pairs(
		"authorization", "Bearer "+accessToken,
		"developer-token", developerToken,
		"login-customer-id", loginCustomerID, // ID da MCC
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

	// Calcular o primeiro e o último dia do mês anterior
	startDate, endDate, monthName := getPreviousMonthRange()

	// Consulta para buscar o nome da conta
	queryAccountName := "SELECT customer.descriptive_name FROM customer"
	reqName := &services.SearchGoogleAdsRequest{
		CustomerId: accountID,
		Query:      queryAccountName,
	}
	resName, err := client.Search(ctx, reqName)
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
		return "error"
	}

	var accountName string
	if len(resName.Results) > 0 {
		accountName = *resName.Results[0].Customer.DescriptiveName
	} else {
		accountName = "Nome não encontrado"
	}

	// Organizar os resultados
	querySearch := fmt.Sprintf(
		"SELECT campaign.id, campaign.name, ad_group.id, ad_group.name, metrics.impressions "+
			"FROM ad_group WHERE campaign.name LIKE '%%modelos%%' "+
			"AND segments.date BETWEEN '%s' AND '%s'",
		startDate, endDate)

	// Criar a requisição para a API do Google Ads
	req := &services.SearchGoogleAdsRequest{
		CustomerId: accountID, // ID da conta individual
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
		return "error"
	}

	// Verificar se a conta tem campanhas "modelos"
	if len(res.Results) == 0 {
		// Se não houver resultados, informar que a conta não tem campanha "modelos"
		accountResult := fmt.Sprintf("Na conta %s, não foi encontrada a campanha 'modelos'.", accountName)

		// Chama a função para salvar os dados na planilha
		if err := WriteToGoogleSheetsLastModelos(ctx, accessToken, []string{accountName}, []string{accountResult}, monthName); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao escrever os dados na planilha"})
			return "error"
		}

		// Retornar o resultado da conta específica
		c.JSON(http.StatusOK, gin.H{
			"message": "Dados de grupos de anúncios recuperados com sucesso!",
			"result":  accountResult,
		})
		return accountResult
	} else {
		// Organizar os resultados por impressões
		var adGroups []models.AdGroup
		for _, row := range res.Results {
			adGroupName := *row.AdGroup.Name
			impressions := int(*row.Metrics.Impressions)
			adGroups = append(adGroups, models.AdGroup{
				Name:        adGroupName,
				Impressions: impressions,
			})
		}

		// Ordenar os grupos de anúncios por impressões (do maior para o menor)
		sort.Slice(adGroups, func(i, j int) bool {
			return adGroups[i].Impressions > adGroups[j].Impressions
		})

		// Se houver mais de 3 grupos de anúncios, selecionar os 2 melhores e 2 piores
		var top2 []models.AdGroup
		var worst2 []models.AdGroup

		if len(adGroups) > 3 {
			top2 = adGroups[:2]                 // Os 2 melhores
			worst2 = adGroups[len(adGroups)-2:] // Os 2 piores
		} else {
			// Se houver menos de 3 grupos de anúncios
			accountResult := fmt.Sprintf("Na conta %s, a campanha 'modelos' tem poucos grupos de anúncios.", accountName)

			// Chama a função para salvar os dados na planilha
			if err := WriteToGoogleSheetsLastModelos(ctx, accessToken, []string{accountName}, []string{accountResult}, monthName); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao escrever os dados na planilha"})
				return "error"
			}

			// Retornar o resultado da conta específica
			c.JSON(http.StatusOK, gin.H{
				"message": "Dados de grupos de anúncios recuperados com sucesso!",
				"result":  accountResult,
			})
			return accountResult
		}

		// Caso não haja 2 elementos, preenchemos com "Grupo insuficiente"
		if len(top2) < 2 {
			top2 = append(top2, models.AdGroup{Name: "Grupo insuficiente"})
		}
		if len(worst2) < 2 {
			worst2 = append(worst2, models.AdGroup{Name: "Grupo insuficiente"})
		}

		// Criar o texto a ser enviado
		accountResult := fmt.Sprintf("No mês %s, Tivemos um total de X leads gerados com um CPA Médio de R$XX.00. A Campanha que mais gerou Leads foi a Campanha X. Na Campanha Modelos, os modelos que mais tiveram buscas foram o %s e %s, e os menos buscados %s e %s. Com isso, é válido estudar ações que maximizem ainda mais o resultados dos modelos mais Buscados, fazendo assim ações que vão de acordo com o interesse do público.",
			monthName, top2[0].Name, top2[1].Name, worst2[0].Name, worst2[1].Name)

		// Chama a função para salvar os dados na planilha
		if err := WriteToGoogleSheetsLastModelos(ctx, accessToken, []string{accountName}, []string{accountResult}, monthName); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao escrever os dados na planilha"})
			return "error"
		}

		return accountResult
	}
}

// Função para escrever os dados na planilha
func WriteToGoogleSheetsLastModelos(ctx context.Context, accessToken string, accountNames []string, accountResults []string, monthName string) error {
	// Criar o cliente HTTP com o accessToken já fornecido
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: accessToken,
	}))

	// Criar o serviço do Google Sheets com o cliente autenticado
	sheetsService, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("erro ao criar serviço do Google Sheets: %w", err)
	}

	// Transformar os dados das campanhas em um formato para a planilha
	var sheetData [][]interface{}
	// Cabeçalho
	sheetData = append(sheetData, []interface{}{"Nome da Conta", "Texto", "Mês"})

	// Adicionar os dados das campanhas
	for i, result := range accountResults {
		// Criar o ID como o índice da conta (por exemplo, 1, 2, 3...)
		sheetData = append(sheetData, []interface{}{
			accountNames[i], // Nome da Conta
			result,          // Texto
			monthName,       // Mês
		})
	}

	// Definir o intervalo da planilha onde os dados serão inseridos
	spreadsheetId := "1rFYqz1HtqMA6ttFHhp5e_2uLjd3eL9yieLM4IB1zidg"
	writeRange := "comentarios!A1:D" // Atualizando para 4 colunas

	// Criar o corpo da requisição com os dados
	valueRange := &sheets.ValueRange{
		Values: sheetData,
	}

	// Fazer a requisição para inserir os dados na planilha
	_, err = sheetsService.Spreadsheets.Values.Append(spreadsheetId, writeRange, valueRange).ValueInputOption("RAW").Do()
	if err != nil {
		return fmt.Errorf("erro ao escrever dados na planilha: %w", err)
	}

	return nil
}
