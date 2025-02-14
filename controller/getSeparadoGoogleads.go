package controller

import (
	"context"
	"fmt"
	"google-ads-get/models"
	"log"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/shenzhencenter/google-ads-pb/services"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func GetTopAndWorstAdGroupsForModelosFOR(c *gin.Context, accountID string, token string) string {
	print("Token passado por construtor:", token)
	if token == "" {
		log.Println("Token não encontrado!")
		return ""
	}
	print("Entrei nesta funcao")

	// Definir as credenciais do Google Ads API
	developerToken := "MvSisVf6otSXPLwUvGbUaw" // Coloque seu developer token correto aqui
	loginCustomerID := "9155619419"            // ID da MCC

	// Criar um novo contexto e adicionar os headers necessários
	ctx := context.Background()
	headers := metadata.Pairs(
		"authorization", "Bearer "+TokenFinal, // Usar o token passado como argumento
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
		if err := WriteToGoogleSheetsLastModelos(ctx, token, []string{accountName}, []string{accountResult}, monthName); err != nil {
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
			if err := WriteToGoogleSheetsLastModelos(ctx, token, []string{accountName}, []string{accountResult}, monthName); err != nil {
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
		if err := WriteToGoogleSheetsLastModelos(ctx, token, []string{accountName}, []string{accountResult}, monthName); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao escrever os dados na planilha"})
			return "error"
		}

		return accountResult
	}
}
