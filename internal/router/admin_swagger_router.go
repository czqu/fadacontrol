package router

import (
	_ "fadacontrol/docs/admin"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title						Remote Unlock Module Admin API documentation
// @version					1.0
// @BasePath					/admin/api/v1/
// @description				Remote unlock module API documentation
// @termsOfService				https://rfu.czqu.net
// @contact.name				API Support
// @contact.url				https://rfu.czqu.net
// @contact.email				me@czqu.net
// @host						localhost:2093
// @BasePath					/admin/api/v1/
// @externalDocs.description	OpenAPI
// @externalDocs.url			https://swagger.io/resources/open-api/
var adminSwagHandler gin.HandlerFunc

func init() {
	adminSwagHandler = ginSwagger.WrapHandler(swaggerFiles.Handler, func(config *ginSwagger.Config) {
		config.InstanceName = "admin"
	})

}

//swag init   --parseDependency  --instanceName=admin  --generalInfo=internal/router/admin_swagger_router.go  --output docs/admin
