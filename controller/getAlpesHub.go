package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
)

// Cliente HTTP com suporte a cookies
var httpClient *http.Client

func init() {
	// Inicializa o cliente HTTP com CookieJar
	jar, _ := cookiejar.New(nil)
	httpClient = &http.Client{Jar: jar}
}

// Função para buscar os tokens e fazer o login
func FetchTokensHandler(c *gin.Context) {
	resp, err := httpClient.Get("https://hub.alpes.one/admin/backend/auth/signin")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro ao fazer requisição GET: %v", err)})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Printf("Resposta: %s", string(body))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Status de resposta inválido: %d", resp.StatusCode)})
		return
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro ao processar HTML: %v", err)})
		return
	}

	sessionKey, exists := doc.Find("input[name='_session_key']").Attr("value")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "_session_key não encontrado no HTML"})
		return
	}

	token, exists := doc.Find("input[name='_token']").Attr("value")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "_token não encontrado no HTML"})
		return
	}

	Login(sessionKey, token, c)
}

func Login(sessionKey string, token string, c *gin.Context) {
	data := url.Values{}
	data.Set("_session_key", sessionKey)
	data.Set("_token", token)
	data.Set("postback", "1")
	data.Set("login", "gabriel.netto")
	data.Set("password", "mudar123")
	data.Set("useTerms", "1")

	resp, err := httpClient.PostForm("https://hub.alpes.one/admin/backend/auth/signin", data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro ao fazer requisição POST: %v", err)})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Status de resposta inválido: %d", resp.StatusCode)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login bem-sucedido!",
		"status":  resp.StatusCode,
	})

	//PostLeadReportHandler(c)
	ProcessAccounts(c)
}

type GeneralData struct {
	Conversions int `json:"conversions"`
}

type AdWordsData struct {
	GeneralData GeneralData `json:"general_data"`
}

type ResponseData struct {
	AdWordsData []AdWordsData `json:"adwords_data"`
}

func PostLeadReportHandler(c *gin.Context, AlpesHubId string) float64 {
	// Dados no formato codificado
	encodedData := "company_id=" + AlpesHubId + "&type_report%5B%5D=Adwords&month=01&year=2025&cache=1"

	// Criando a requisição POST
	req, err := http.NewRequest("POST", "https://hub.alpes.one/admin/alpesone/leads/reports", bytes.NewBufferString(encodedData))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro ao criar requisição POST: %v", err)})
		return 0
	}

	// Cabeçalhos necessários para a requisição
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br, zstd")
	req.Header.Set("Accept-Language", "en-GB,en;q=0.8")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Referer", "https://hub.alpes.one/admin/alpesone/leads/reports")
	req.Header.Set("Origin", "https://hub.alpes.one")
	//req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36")
	//req.Header.Set("X-CSRF-TOKEN", "<CSRF-TOKEN-AQUI>") // Coloque o valor real do CSRF token
	req.Header.Set("X-OCTOBER-REQUEST-HANDLER", "onLoadReports")
	req.Header.Set("X-OCTOBER-REQUEST-PARTIALS", "reports")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	// Cookies para manter a sessão
	req.Header.Set("Cookie", "_hjSessionUser_648805=eyJpZCI6ImU0YjE3MDk0LWZmNGEtNTdmNS04OTIyLTgyY2JlOWIyOTlkMyIsImNyZWF0ZWQiOjE3MzkwMzkzNjgzMTYsImV4aXN0aW5nIjp0cnVlfQ==; _hjSession_648805=eyJpZCI6ImUyZTI3YWRkLTJjN2YtNGNmOS1iOGIxLTE2MmRhMTBhZjJiYSIsImMiOjE3MzkyOTk1MDMyMTUsInMiOjAsInIiOjAsInNiIjowLCJzciI6MCwic2UiOjAsImZzIjowLCJzcCI6MH0=; admin_auth=eyJpdiI6IkZNTm42blpPb0tOTEU3TEttTndiTmc9PSIsInZhbHVlIjoiTDduOXptdUV2Z09idFBHVWhBNlRnZlpBd3VaXC9rMlg4dzZ0bzZqYk9kYjVRRzdaeWZcL01vbDc1aUtOSmRzcnlidUJaWlpORVhcL1VkcXRoVWNRd1o...")

	// Enviar a requisição
	resp, err := httpClient.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro ao enviar requisição POST: %v", err)})
		return 0
	}
	defer resp.Body.Close()

	// Lendo o corpo da resposta
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro ao ler o corpo da resposta: %v", err)})
		return 0
	}

	// Verifica se a resposta contém HTML
	if strings.HasPrefix(string(body), "<html>") {
		log.Printf("Resposta HTML detectada: %s", string(body))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "A resposta do servidor contém HTML, não JSON. Pode ser uma página de erro ou login.",
			"body":  string(body),
		})
		return 0
	}

	// Logando o corpo da resposta para depuração
	log.Printf("Corpo da resposta (antes de filtrar): %s", string(body))

	// Remover o campo "reports" do JSON, se necessário
	bodyStr := string(body)
	if strings.HasPrefix(bodyStr, "{\"reports\":") {
		index := strings.Index(bodyStr, "{")
		if index != -1 {
			bodyStr = bodyStr[index+1:] // Remove o prefixo até o próximo '{'
		}
	}

	// Reconstruindo o JSON válido
	cleanedBody := "{" + bodyStr
	log.Printf("Corpo da resposta (depois de filtrar): %s", cleanedBody)

	// Decodificando o JSON processado
	var filteredJSON map[string]interface{}
	err = json.Unmarshal([]byte(cleanedBody), &filteredJSON)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro ao processar o JSON limpo: %v", err)})
		return 0
	}

	// Acessando o valor de "conversions" dentro de "adwords_data"
	adwordsData, ok := filteredJSON["adwords_data"].([]interface{})
	if !ok || len(adwordsData) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Dados do adwords_data não encontrados"})
		return 0
	}

	// Acessando o "general_data" e "conversions"
	generalData, ok := adwordsData[0].(map[string]interface{})["general_data"].(map[string]interface{})
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Dados de 'general_data' não encontrados"})
		return 0
	}

	conversions, ok := generalData["conversions"].(float64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Valor de 'conversions' não encontrado"})
		return 0
	}

	// Retorna o valor de "conversions" como parte do JSON

	return conversions

}
