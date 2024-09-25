// Package admin Code generated by swaggo/swag. DO NOT EDIT
package admin

import "github.com/swaggo/swag"

const docTemplateadmin = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "termsOfService": "https://rfu.czqu.net",
        "contact": {
            "name": "API Support",
            "url": "https://rfu.czqu.net",
            "email": "me@czqu.net"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/control-pc/interface": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Based on the specified IP version type, a list of valid network interfaces is returned",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Network interfaces"
                ],
                "summary": "Returns a valid network interface",
                "parameters": [
                    {
                        "enum": [
                            "4",
                            "6"
                        ],
                        "type": "string",
                        "default": "4",
                        "description": "IP Version Type (4 or 6)",
                        "name": "type",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "A list of valid network interfaces is successfully returned",
                        "schema": {
                            "$ref": "#/definitions/schema.ResponseData"
                        }
                    },
                    "400": {
                        "description": "The request parameter is incorrect",
                        "schema": {
                            "$ref": "#/definitions/schema.ResponseData"
                        }
                    },
                    "500": {
                        "description": "Server internal error",
                        "schema": {
                            "$ref": "#/definitions/schema.ResponseData"
                        }
                    }
                }
            }
        },
        "/control-pc/{action}/": {
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Control the operation of the computer according to the transmitted parameters",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "Control  computer",
                "parameters": [
                    {
                        "enum": [
                            "shutdown",
                            "standby",
                            "lock"
                        ],
                        "type": "string",
                        "description": "The type of operation（shutdown、standby、lock）",
                        "name": "action",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Delay time in seconds",
                        "name": "delay",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "The type of shutdown",
                        "name": "shutdown_type",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "success",
                        "schema": {
                            "$ref": "#/definitions/schema.ResponseData"
                        }
                    },
                    "400": {
                        "description": "Invalid action type",
                        "schema": {
                            "$ref": "#/definitions/schema.ResponseData"
                        }
                    },
                    "500": {
                        "description": "The operation failed",
                        "schema": {
                            "$ref": "#/definitions/schema.ResponseData"
                        }
                    }
                }
            }
        },
        "/discovery/config": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Get the Discovery Service configuration",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Discover"
                ],
                "summary": "Get Discover Service Config",
                "responses": {
                    "200": {
                        "description": "success",
                        "schema": {
                            "$ref": "#/definitions/schema.ResponseData"
                        }
                    },
                    "500": {
                        "description": "Server internal error",
                        "schema": {
                            "$ref": "#/definitions/schema.ResponseData"
                        }
                    }
                }
            },
            "patch": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Update the configuration of the Discover service with the provided settings.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Discover"
                ],
                "summary": "Update Discover Service Configuration",
                "parameters": [
                    {
                        "description": "New configuration settings",
                        "name": "config",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/schema.DiscoverSchema"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Successfully updated configuration.",
                        "schema": {
                            "$ref": "#/definitions/schema.ResponseData"
                        }
                    },
                    "400": {
                        "description": "Invalid request parameters.",
                        "schema": {
                            "$ref": "#/definitions/schema.ResponseData"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/schema.ResponseData"
                        }
                    }
                }
            }
        },
        "/discovery/restart": {
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Restart the discover service.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Discover"
                ],
                "summary": "Restart Discover Service",
                "responses": {
                    "200": {
                        "description": "Service restarted successfully.",
                        "schema": {
                            "$ref": "#/definitions/schema.ResponseData"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/schema.ResponseData"
                        }
                    }
                }
            }
        },
        "/http/config": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Retrieve the current HTTP configuration based on the provided type.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "HTTP"
                ],
                "summary": "Get HTTP Configuration",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Configuration type (HTTP_SERVICE_API or HTTPS_SERVICE_API)",
                        "name": "type",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Successfully retrieved configuration.",
                        "schema": {
                            "$ref": "#/definitions/schema.ResponseData"
                        }
                    },
                    "400": {
                        "description": "Invalid request parameters.",
                        "schema": {
                            "$ref": "#/definitions/schema.ResponseData"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/schema.ResponseData"
                        }
                    }
                }
            },
            "put": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Update the HTTP configuration based on the provided type and settings.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "HTTP"
                ],
                "summary": "Update HTTP Configuration",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Configuration type (HTTP_SERVICE_API or HTTPS_SERVICE_API)",
                        "name": "type",
                        "in": "query",
                        "required": true
                    },
                    {
                        "description": "Configuration settings",
                        "name": "config",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/http_schema.HttpConfigRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Successfully updated configuration.",
                        "schema": {
                            "$ref": "#/definitions/schema.ResponseData"
                        }
                    },
                    "400": {
                        "description": "Invalid request parameters.",
                        "schema": {
                            "$ref": "#/definitions/schema.ResponseData"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/schema.ResponseData"
                        }
                    }
                }
            },
            "patch": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Partially update the HTTP configuration based on the provided type and settings.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "HTTP"
                ],
                "summary": "Patch HTTP Configuration",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Configuration type (HTTP_SERVICE_API or HTTPS_SERVICE_API)",
                        "name": "type",
                        "in": "query",
                        "required": true
                    },
                    {
                        "description": "Partial configuration settings",
                        "name": "config",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/http_schema.HttpConfigRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Successfully updated configuration.",
                        "schema": {
                            "$ref": "#/definitions/schema.ResponseData"
                        }
                    },
                    "400": {
                        "description": "Invalid request parameters.",
                        "schema": {
                            "$ref": "#/definitions/schema.ResponseData"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/schema.ResponseData"
                        }
                    }
                }
            }
        },
        "/http/restart": {
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Restart the server based on the provided type.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "HTTP"
                ],
                "summary": "Restart Service",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Service type (HTTP_SERVICE_API or HTTPS_SERVICE_API)",
                        "name": "type",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Service restarted successfully.",
                        "schema": {
                            "$ref": "#/definitions/schema.ResponseData"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/schema.ResponseData"
                        }
                    }
                }
            }
        },
        "/info": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Get the software version information",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "System"
                ],
                "summary": "Get Software Info",
                "responses": {
                    "200": {
                        "description": "success",
                        "schema": {
                            "$ref": "#/definitions/schema.ResponseData"
                        }
                    },
                    "500": {
                        "description": "Server internal error",
                        "schema": {
                            "$ref": "#/definitions/schema.ResponseData"
                        }
                    }
                }
            }
        },
        "/interface/{ip}": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "Obtain the MAC address based on the IP address",
                "responses": {
                    "200": {
                        "description": "success",
                        "schema": {
                            "$ref": "#/definitions/schema.ResponseData"
                        }
                    },
                    "400": {
                        "description": "The request is incorrect",
                        "schema": {
                            "$ref": "#/definitions/schema.ResponseData"
                        }
                    },
                    "500": {
                        "description": "Internal errors",
                        "schema": {
                            "$ref": "#/definitions/schema.ResponseData"
                        }
                    }
                }
            }
        },
        "/interface/{ip}/all": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "Obtain the interface information based on the IP address",
                "responses": {
                    "200": {
                        "description": "success"
                    },
                    "400": {
                        "description": "The request is incorrect"
                    },
                    "500": {
                        "description": "Internal errors"
                    }
                }
            }
        },
        "/login": {
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Authenticate user with username and password to obtain a JWT token.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "User"
                ],
                "summary": "User Login",
                "parameters": [
                    {
                        "description": "User login credentials",
                        "name": "loginData",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/schema.LoginRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Successfully authenticated.",
                        "schema": {
                            "$ref": "#/definitions/schema.ResponseData"
                        }
                    },
                    "400": {
                        "description": "Invalid request parameters.",
                        "schema": {
                            "$ref": "#/definitions/schema.ResponseData"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/schema.ResponseData"
                        }
                    }
                }
            }
        },
        "/logs/{module}": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Stream the system logs in real-time. This endpoint opens a connection to the log buffer and sends log entries as they are generated. The connection remains open until explicitly closed or an error occurs. If the buffer is not available, it returns an error response.",
                "produces": [
                    "text/event-stream"
                ],
                "tags": [
                    "System"
                ],
                "summary": "Get System Logs",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Specify the module to retrieve logs from (must be 'service')",
                        "name": "module",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Stream of system logs.",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Invalid module specified.",
                        "schema": {
                            "$ref": "#/definitions/schema.ResponseData"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/schema.ResponseData"
                        }
                    }
                }
            }
        },
        "/ping": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "summary": "Ping",
                "responses": {
                    "200": {
                        "description": "success"
                    }
                }
            }
        },
        "/power-saving": {
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Enable or disable power saving mode. This function first records the setting in the database. If the database write fails, it returns immediately. If the database write succeeds, it sets the power saving mode. A failure in setting the mode does not affect the database value.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "System"
                ],
                "summary": "Set Power Saving Mode",
                "parameters": [
                    {
                        "type": "string",
                        "default": "true",
                        "description": "Enable power saving mode (true or false)",
                        "name": "enable",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Power saving mode set successfully.",
                        "schema": {
                            "$ref": "#/definitions/schema.ResponseData"
                        }
                    },
                    "400": {
                        "description": "Invalid request parameters.",
                        "schema": {
                            "$ref": "#/definitions/schema.ResponseData"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/schema.ResponseData"
                        }
                    }
                }
            }
        },
        "/remote/config": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Retrieve the current configuration for remote connections.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Remote"
                ],
                "summary": "Get Remote Connect Configuration",
                "responses": {
                    "200": {
                        "description": "Successfully retrieved configuration.",
                        "schema": {
                            "$ref": "#/definitions/schema.ResponseData"
                        }
                    },
                    "400": {
                        "description": "Invalid request parameters.",
                        "schema": {
                            "$ref": "#/definitions/schema.ResponseData"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/schema.ResponseData"
                        }
                    }
                }
            },
            "put": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Update the configuration for remote connections with the provided settings.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Remote"
                ],
                "summary": "Update Remote Connect Configuration",
                "parameters": [
                    {
                        "description": "New configuration settings",
                        "name": "config",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/remote_schema.RemoteConnectConfigRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Successfully updated configuration.",
                        "schema": {
                            "$ref": "#/definitions/schema.ResponseData"
                        }
                    },
                    "400": {
                        "description": "Invalid request parameters.",
                        "schema": {
                            "$ref": "#/definitions/schema.ResponseData"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/schema.ResponseData"
                        }
                    }
                }
            },
            "patch": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Partially update the configuration for remote connections with the provided settings.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Remote"
                ],
                "summary": "Patch Remote Connect Configuration",
                "parameters": [
                    {
                        "description": "Partial configuration settings",
                        "name": "config",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/remote_schema.RemoteConnectConfigRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Successfully updated configuration.",
                        "schema": {
                            "$ref": "#/definitions/schema.ResponseData"
                        }
                    },
                    "400": {
                        "description": "Invalid request parameters.",
                        "schema": {
                            "$ref": "#/definitions/schema.ResponseData"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/schema.ResponseData"
                        }
                    }
                }
            }
        },
        "/remote/restart": {
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Restart the remote service.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Remote"
                ],
                "summary": "Restart Remote Service",
                "responses": {
                    "200": {
                        "description": "Service restarted successfully.",
                        "schema": {
                            "$ref": "#/definitions/schema.ResponseData"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/schema.ResponseData"
                        }
                    }
                }
            }
        },
        "/unlock": {
            "post": {
                "produces": [
                    "application/json"
                ],
                "summary": "Unlock your computer",
                "parameters": [
                    {
                        "description": "User Information",
                        "name": "UserInfo",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/schema.PcUserInfo"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "success",
                        "schema": {
                            "$ref": "#/definitions/schema.ResponseData"
                        }
                    },
                    "400": {
                        "description": "The request is incorrect",
                        "schema": {
                            "$ref": "#/definitions/schema.ResponseData"
                        }
                    },
                    "500": {
                        "description": "Internal errors",
                        "schema": {
                            "$ref": "#/definitions/schema.ResponseData"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "http_schema.HttpConfigRequest": {
            "type": "object",
            "properties": {
                "enable": {
                    "type": "boolean"
                },
                "host": {
                    "type": "string"
                },
                "port": {
                    "type": "integer"
                }
            }
        },
        "remote_schema.RemoteConnectConfigRequest": {
            "type": "object",
            "properties": {
                "api_server_url": {
                    "type": "string"
                },
                "client_id": {
                    "type": "string"
                },
                "enable": {
                    "type": "boolean"
                },
                "msg_server_urls": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "time_stamp_check": {
                    "type": "boolean"
                }
            }
        },
        "schema.DiscoverSchema": {
            "type": "object",
            "properties": {
                "enabled": {
                    "type": "boolean"
                }
            }
        },
        "schema.LoginRequest": {
            "type": "object",
            "required": [
                "password",
                "username"
            ],
            "properties": {
                "password": {
                    "type": "string"
                },
                "username": {
                    "type": "string"
                }
            }
        },
        "schema.PcUserInfo": {
            "type": "object",
            "properties": {
                "password": {
                    "type": "string"
                },
                "username": {
                    "type": "string"
                }
            }
        },
        "schema.ResponseData": {
            "type": "object",
            "properties": {
                "code": {
                    "type": "integer"
                },
                "data": {},
                "msg": {
                    "type": "string"
                },
                "request_id": {
                    "type": "string"
                }
            }
        }
    },
    "securityDefinitions": {
        "ApiKeyAuth": {
            "type": "apiKey",
            "name": "Authorization",
            "in": "header"
        }
    },
    "externalDocs": {
        "description": "OpenAPI",
        "url": "https://swagger.io/resources/open-api/"
    }
}`

// SwaggerInfoadmin holds exported Swagger Info so clients can modify it
var SwaggerInfoadmin = &swag.Spec{
	Version:          "1.0",
	Host:             "localhost:2093",
	BasePath:         "/admin/api/v1/",
	Schemes:          []string{},
	Title:            "Remote Unlock Module Admin API documentation",
	Description:      "Remote unlock module API documentation",
	InfoInstanceName: "admin",
	SwaggerTemplate:  docTemplateadmin,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfoadmin.InstanceName(), SwaggerInfoadmin)
}