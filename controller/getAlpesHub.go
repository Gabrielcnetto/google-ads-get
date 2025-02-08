package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
)

func FetchTokensHandler(c *gin.Context) {
	// Fazendo a requisição GET para o endpoint
	resp, err := http.Get("https://hub.alpes.one/admin/backend/auth/signin")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro ao fazer requisição GET: %v", err)})
		return
	}
	defer resp.Body.Close()

	// Verifica se a resposta foi bem-sucedida
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)     // Para fins de depuração
		log.Printf("Resposta: %s", string(body)) // Loga o HTML retornado
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Status de resposta inválido: %d", resp.StatusCode)})
		return
	}

	// Carrega o HTML com goquery
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro ao processar HTML: %v", err)})
		return
	}

	// Procura o valor de _session_key
	sessionKey, exists := doc.Find("input[name='_session_key']").Attr("value")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "_session_key não encontrado no HTML"})
		return
	}

	// Procura o valor de _token
	token, exists := doc.Find("input[name='_token']").Attr("value")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "_token não encontrado no HTML"})
		return
	}

	// Chama a função Login passando os tokens encontrados
	Login(sessionKey, token, c)
}

// Função para fazer o login usando os tokens (_session_key e _token)
func Login(sessionKey string, token string, c *gin.Context) {
	// Definindo os dados do corpo para o POST
	data := url.Values{}
	data.Set("_session_key", sessionKey)
	data.Set("_token", token)
	data.Set("postback", "1") // Valor fixo
	data.Set("login", "gabriel.netto")
	data.Set("password", "mudar123")
	data.Set("useTerms", "1") // Representando true como "1" (float com valor verdadeiro)

	// Criando a requisição POST para o login
	resp, err := http.PostForm("https://hub.alpes.one/admin/backend/auth/signin", data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro ao fazer requisição POST: %v", err)})
		return
	}
	defer resp.Body.Close()

	// Verificando a resposta do login
	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Status de resposta inválido: %d", resp.StatusCode)})
		return
	}

	// Loga a resposta para depuração
	c.JSON(http.StatusOK, gin.H{
		"message": "Login bem-sucedido!",
		"status":  resp.StatusCode,
	})

	// Após o login bem-sucedido, chamar a função que envia o segundo POST para obter o relatório
	PostLeadReportHandler(c)
}

// Estruturas para mapear a resposta JSON
type GeneralData struct {
	Conversions int `json:"conversions"`
}

type AdWordsData struct {
	GeneralData GeneralData `json:"general_data"`
}

type ResponseData struct {
	AdWordsData []AdWordsData `json:"adwords_data"`
}

// Função para enviar o POST para o relatório e processar o JSON de resposta
func PostLeadReportHandler(c *gin.Context) {
	// Definindo os dados do corpo para o POST
	data := url.Values{}
	data.Set("company_id", "3")          // company_id como int 3
	data.Set("type_report[]", "adwords") // type_report[] como string "adwords"
	data.Set("month", "01")              // mês como int 01
	data.Set("year", "2025")             // ano como int 2025
	data.Set("cache", "01")              // cache como int 01

	// Criando a requisição POST com headers customizados
	req, err := http.NewRequest("POST", "https://hub.alpes.one/admin/alpesone/leads/reports", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro ao criar requisição POST: %v", err)})
		return
	}

	// Adicionando o cabeçalho Content-Type
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("cookie", "_hjSession_648805=eyJpZCI6IjgwNWViMzY2LThjMGEtNGUyYi04OWY0LWMzOWIzMWU0OWU3NCIsImMiOjE3MzkwMzkzNjgzMjQsInMiOjAsInIiOjAsInNiIjowLCJzciI6MCwic2UiOjAsImZzIjoxLCJzcCI6MH0=")                                                                                                                                                                                                             // Exemplo de cookie
	req.Header.Set("admin_auth", "eyJpdiI6InFzTGVEdnBWM0MrekFuSXNqZkhmY2c9PSIsInZhbHVlIjoiYUVBWjI4UWdpVFVTbWNpemdSQ1JzT1Jhdm56cDU0UGZxOCtmUmlXNENDT3ZqSkVFWnVvb1wvQVpNdExPeHFBV2dQbDRwM3lOdmRBa09hOTQzSUhqWkZ2OUNIWDRMU25XakVlZWgzXC9jdUI2M3E0SUZ2MjNTXC8xcjk4QkF5djZzRDd3TmZkWHl4TmVFdHY5U1BrSCs2OUR3PT0iLCJtYWMiOiJiNDBiNDIyODFjNDY5ZjU5ZmVmM2ZlYzhkZmM0NzZlNmM0YzA5ZmM2MjAyMTMyM2IwZmE5N2ViNTJmMGE4MWI5In0%3D") // Exemplo de chave admin_auth

	// Definindo o corpo da requisição
	req.Body = ioutil.NopCloser(bytes.NewBufferString(data.Encode()))

	// Enviar a requisição
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro ao enviar requisição POST: %v", err)})
		return
	}
	defer resp.Body.Close()

	// Lendo o corpo da resposta
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro ao ler o corpo da resposta: %v", err)})
		return
	}

	// Logando o corpo da resposta para depuração
	log.Printf("Corpo da resposta: %s", string(body))

	// Variável para mapear o JSON
	var response ResponseData

	// Decodificando o JSON da resposta
	err = json.Unmarshal(body, &response)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro ao processar o JSON: %v", err)})
		return
	}

	// Extraindo o valor de 'conversions' de 'adwords_data'
	if len(response.AdWordsData) > 0 {
		conversions := response.AdWordsData[0].GeneralData.Conversions
		// Retornando o valor de 'conversions'
		c.JSON(http.StatusOK, gin.H{
			"conversions": conversions,
		})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "adwords_data não encontrado"})
	}
}
