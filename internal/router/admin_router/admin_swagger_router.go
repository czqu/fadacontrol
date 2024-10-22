//go:build swag

package admin_router

import (
	_ "fadacontrol/docs/admin"
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
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

func init() {
	adminSwagHandler := ginSwagger.WrapHandler(swaggerFiles.Handler, func(config *ginSwagger.Config) {
		config.InstanceName = "admin"
	})
	SetSwagHandler(adminSwagHandler)
}

//swag init   -parseDependency=false  --instanceName=admin  --generalInfo=internal/router/admin_router/admin_swagger_router.go  --output docs/admin
