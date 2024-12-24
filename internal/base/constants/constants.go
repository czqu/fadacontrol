package constants

type CountryCode string

const (
	CountryGlobal CountryCode = "global"
	CountryCodeCN CountryCode = "cn"
	CountryCodeUS CountryCode = "us"
)

type LanguageCode string

const (
	LanguageCodeEN LanguageCode = "en"
	LanguageCodeCN LanguageCode = "cn"
)
const ServiceName = "FadaControlService"
const ConfKey = "conf"
const ClientIdKey = "client_id"
