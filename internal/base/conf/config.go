package conf

const DefaultLogLevel = "info"
const DefaultMasterLogName = "service.log"
const DefaultSlaveLogName = "slave.log"

var RootPassword = "1234"
var ResetPassword = false
var Http3Enabled = false
var Http3Port = 2091
var IgnoredPaths = []string{
	"/api/v1/ping",
	"/api/v1/unlock",
	"/api/v1/login",
	"/admin/api/v1/ping",
	"/admin/api/v1/unlock",
	"/admin/api/v1/login",
	"/swagger/*",
}

type ProductLanguage string

const (
	LanguageChineseSimple      ProductLanguage = "zh-cn" // Simplified Chinese
	LanguageEnglish            ProductLanguage = "en"    // English
	LanguageFrench             ProductLanguage = "fr"    // French
	LanguageGerman             ProductLanguage = "de"    // German
	LanguageItalian            ProductLanguage = "it"    // Italian
	LanguageSpanish            ProductLanguage = "es"    // Spanish
	LanguageRussian            ProductLanguage = "ru"    // Russian
	LanguageJapanese           ProductLanguage = "ja"    // Japanese
	LanguageKorean             ProductLanguage = "ko"    // Korean
	LanguagePortuguese         ProductLanguage = "pt"    // Portuguese
	LanguageChineseTraditional ProductLanguage = "zh-tw" // Traditional Chinese
)

func (l ProductLanguage) String() string {
	return string(l)
}
func ProductLanguageFromString(s string) ProductLanguage {
	switch s {
	case "ru":
		return LanguageRussian
	case "es":
		return LanguageSpanish
	case "zh-cn":
		return LanguageChineseSimple
	case "en":
		return LanguageEnglish
	case "fr":
		return LanguageFrench
	case "ge":
		return LanguageGerman
	case "it":
		return LanguageItalian
	case "ja":
		return LanguageJapanese
	case "ko":
		return LanguageKorean
	case "pt":
		return LanguagePortuguese
	case "zh-tw":
		return LanguageChineseTraditional
	default:
		return LanguageEnglish

	}
}
