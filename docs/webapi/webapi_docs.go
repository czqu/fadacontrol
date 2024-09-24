// Package webapi Code generated by swaggo/swag. DO NOT EDIT
package webapi

import "github.com/swaggo/swag"

const docTemplatewebapi = `{
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

// SwaggerInfowebapi holds exported Swagger Info so clients can modify it
var SwaggerInfowebapi = &swag.Spec{
	Version:          "1.0",
	Host:             "localhost:2091",
	BasePath:         "/api/v1",
	Schemes:          []string{},
	Title:            "Remote Unlock Module API documentation",
	Description:      "Remote unlock module API documentation",
	InfoInstanceName: "webapi",
	SwaggerTemplate:  docTemplatewebapi,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfowebapi.InstanceName(), SwaggerInfowebapi)
}
