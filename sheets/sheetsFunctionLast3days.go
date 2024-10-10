package sheets

import (
	"context"
	"fmt"
	"google-ads-get/models"

	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func WriteToGoogleSheetsLast3Days(ctx context.Context, accessToken string, campaigns []models.AdsAccount) error {
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
	sheetData = append(sheetData, []interface{}{"Nome da Conta", "Custo", "Impressões", "Cliques", "AccountId"})

	// Adicionar os dados das campanhas
	for _, campaign := range campaigns {
		sheetData = append(sheetData, []interface{}{
			campaign.Name,
			campaign.Cost,
			campaign.Impressions,
			campaign.Click,
			campaign.AccountId,
		})
	}

	// Definir o intervalo da planilha onde os dados serão inseridos
	spreadsheetId := "1rFYqz1HtqMA6ttFHhp5e_2uLjd3eL9yieLM4IB1zidg"
	writeRange := "apigetreedays!A1:E" // Adaptado para 5 colunas

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
