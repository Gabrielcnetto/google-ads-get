package controller

import (
	"context"
	"fmt"
	"google-ads-get/models"
	"log"
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/shenzhencenter/google-ads-pb/services"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

func GetTopAndWorstAdGroupsForModelosFOR(c *gin.Context, accountID string, token string) string {
	log.Println("Token passado por construtor:", token) // Melhor usar log.Println para debugging

	if token == "" {
		log.Println("Token não encontrado! Retornando erro.")
		return ""
	}
	log.Println("Entrei nesta função")

	// Definir as credenciais do Google Ads API
	developerToken := "MvSisVf6otSXPLwUvGbUaw" // Substituir pelo token correto
	loginCustomerID := "9155619419"            // ID da MCC

	// Criar um novo contexto com os headers corretos
	ctx := context.Background()
	headers := metadata.Pairs(
		"authorization", "Bearer "+token, // Aqui agora usamos o token correto passado
		"developer-token", developerToken,
		"login-customer-id", loginCustomerID,
	)
	ctx = metadata.NewOutgoingContext(ctx, headers)

	// Configuração de conexão gRPC com credenciais seguras
	conn, err := grpc.Dial("googleads.googleapis.com:443", grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(nil, "")))
	if err != nil {
		log.Fatalf("Falha ao conectar ao Google Ads API: %v", err)
	}
	defer conn.Close()

	// Criar cliente Google Ads
	client := services.NewGoogleAdsServiceClient(conn)

	// Obtém o período do mês anterior
	startDate, endDate, monthName := getPreviousMonthRange()

	// Consulta para buscar o nome da conta
	queryAccountName := "SELECT customer.descriptive_name FROM customer"
	reqName := &services.SearchGoogleAdsRequest{
		CustomerId: accountID,
		Query:      queryAccountName,
	}
	resName, err := client.Search(ctx, reqName)
	if err != nil {
		log.Println("Erro ao buscar nome da conta:", err)
		return "error"
	}

	accountName := "Nome não encontrado"
	if len(resName.Results) > 0 {
		accountName = *resName.Results[0].Customer.DescriptiveName
	}

	// Consulta para buscar os grupos de anúncios
	querySearch := fmt.Sprintf(
		"SELECT campaign.id, campaign.name, ad_group.id, ad_group.name, metrics.impressions "+
			"FROM ad_group WHERE campaign.name LIKE '%%modelos%%' "+
			"AND segments.date BETWEEN '%s' AND '%s'",
		startDate, endDate)

	req := &services.SearchGoogleAdsRequest{
		CustomerId: accountID,
		Query:      querySearch,
	}

	res, err := client.Search(ctx, req)
	if err != nil {
		log.Println("Erro ao buscar grupos de anúncios:", err)
		return "error"
	}

	if len(res.Results) == 0 {
		accountResult := fmt.Sprintf("Na conta %s, não foi encontrada a campanha 'modelos'.", accountName)
		log.Println(accountResult)

		// Salvar na planilha
		if err := WriteToGoogleSheetsLastModelos(ctx, token, []string{accountName}, []string{accountResult}, monthName); err != nil {
			log.Println("Erro ao escrever na planilha:", err)
			return "error"
		}

		return accountResult
	}

	// Processar os dados de grupos de anúncios
	var adGroups []models.AdGroup
	for _, row := range res.Results {
		adGroupName := *row.AdGroup.Name
		impressions := int(*row.Metrics.Impressions)
		adGroups = append(adGroups, models.AdGroup{
			Name:        adGroupName,
			Impressions: impressions,
		})
	}

	// Consultar o custo total para o mês
	queryCost := fmt.Sprintf(
		"SELECT customer.id, customer.descriptive_name, metrics.cost_micros "+
			"FROM customer WHERE segments.date BETWEEN '%s' AND '%s'",
		startDate, endDate)

	reqCost := &services.SearchGoogleAdsRequest{
		CustomerId: accountID,
		Query:      queryCost,
	}

	resCost, err := client.Search(ctx, reqCost)
	if err != nil {
		log.Println("Erro ao buscar custo:", err)
		return "error"
	}

	// Cálculo do custo total
	var totalCost float64
	if len(resCost.Results) > 0 {
		totalCost = float64(*resCost.Results[0].Metrics.CostMicros) / 1e6 // Convertendo micros para a unidade correta
	}

	// Ordenar por impressões
	sort.Slice(adGroups, func(i, j int) bool {
		return adGroups[i].Impressions > adGroups[j].Impressions
	})

	var top2, worst2 []models.AdGroup
	if len(adGroups) > 3 {
		top2 = adGroups[:2]                 // Os 2 melhores
		worst2 = adGroups[len(adGroups)-2:] // Os 2 piores
	} else {
		accountResult := fmt.Sprintf("Na conta %s, a campanha 'modelos' tem poucos grupos de anúncios.", accountName)
		log.Println(accountResult)

		if err := WriteToGoogleSheetsLastModelos(ctx, token, []string{accountName}, []string{accountResult}, monthName); err != nil {
			log.Println("Erro ao escrever na planilha:", err)
			return "error"
		}

		return accountResult
	}

	// Garantir que sempre tenhamos 2 itens nos arrays
	for len(top2) < 2 {
		top2 = append(top2, models.AdGroup{Name: "Grupo insuficiente"})
	}
	for len(worst2) < 2 {
		worst2 = append(worst2, models.AdGroup{Name: "Grupo insuficiente"})
	}

	// Construir o resultado final com custo
	accountResult := fmt.Sprintf(
		"No mês %s, Tivemos um total de X leads gerados com um CPA Médio de R$XX.00. "+
			"A Campanha que mais gerou Leads foi a Campanha X. Na Campanha Modelos, "+
			"os modelos que mais tiveram buscas foram o %s e %s, e os menos buscados %s e %s. "+
			"Com isso, é válido estudar ações que maximizem ainda mais o resultados dos modelos mais Buscados, "+
			"fazendo assim ações que vão de acordo com o interesse do público. "+
			"Custo total: R$%.2f",
		monthName, top2[0].Name, top2[1].Name, worst2[0].Name, worst2[1].Name, totalCost)

	log.Println("Resultado final:", accountResult)

	// Salvar na planilha
	if err := WriteToGoogleSheetsLastModelos(ctx, token, []string{accountName}, []string{accountResult}, monthName); err != nil {
		log.Println("Erro ao escrever na planilha:", err)
		return "error"
	}

	return accountResult
}
