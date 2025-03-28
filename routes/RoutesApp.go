package routes

import (
	"google-ads-get/controller"

	"github.com/gin-gonic/gin"
)

func MainRoutes() {
	r := gin.Default()
	controller.Init()

	//apenas uma Conta
	r.GET("/Auth/:customerID", controller.AuthGetAcessToken)
	r.GET("/oauth2callback", controller.OAuth2Callback)
	r.GET("/Mcc/:customerID", controller.GetGoogleAdsDataAutorizations)

	//=> separador

	//<= Separador
	//multiplasContas
	r.GET("/Auth/multipleAccounts", controller.AuthGetAcessTokenMultipleAccounts)
	r.GET("/oauth2callbackMultipleAccounts", controller.OAuth2CallbackMultipleAccounts)
	r.GET("/MultipleAccounts", controller.GetGoogleAdsDataForMultipleAccounts)
	//=> separador

	//<= Separador
	//multiplasContas 3 dias
	r.GET("/Auth/multipleAccountstresdias", controller.AuthGetAcessTokenMultipleAccountsLast3days)
	r.GET("/oauth2callbackMultipleAccountstresdias", controller.OAuth2CallbackMultipleAccountsLast3days)
	r.GET("/MultipleAccountstresdias", controller.GetGoogleAdsDataForMultipleAccountsLast3days)

	//pegar dados da alpes e google ads
	//r.GET("/loginAlpes", controller.FetchTokensHandler)
	r.GET("/oauth2callbackMultipleGetModelos", controller.OAuth2CallbackMultipleAccountsGetModelosGETFOR)
	r.GET("/forGoogleAds", controller.StartOAuthFlow)
	r.Run(":7070")

}
