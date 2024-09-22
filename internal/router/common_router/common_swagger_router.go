package common_router

import (
	_ "fadacontrol/docs/webapi"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title						Remote Unlock Module API documentation
// @version					1.0
// @BasePath					/api/v1/
// @description				Remote unlock module API documentation
// @termsOfService				https://rfu.czqu.net
// @contact.name				API Support
// @contact.url				https://rfu.czqu.net
// @contact.email				me@czqu.net
// @host						localhost:2091
// @BasePath					/api/v1
// @externalDocs.description	OpenAPI
// @externalDocs.url			https://swagger.io/resources/open-api/
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
var commonSwagHandler gin.HandlerFunc

func init() {
	commonSwagHandler = ginSwagger.WrapHandler(swaggerFiles.Handler, func(config *ginSwagger.Config) {
		config.InstanceName = "webapi"
	})

}

// swag init   --parseDependency=false  --instanceName=webapi  --generalInfo=internal/router/common_router/common_swagger_router.go --exclude internal/controller/admin_controller/  --output docs/webapi
