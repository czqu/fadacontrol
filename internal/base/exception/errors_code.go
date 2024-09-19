package exception

import (
	"math/rand"
	"time"
)

const (
	Success = 0

	UnknownError = -1

	// The range of user source error codes
	UserErrorStart = 10001
	UserErrorEnd   = 19999

	// The range of error codes from the source of the system
	SystemErrorStart = 20001
	SystemErrorEnd   = 29999

	// The range of error codes from third-party service sources
	ThirdPartyErrorStart = 30001
	ThirdPartyErrorEnd   = 39999

	// The range of error codes that originated could not be determined
	UncertainErrorStart = 90001
	UncertainErrorEnd   = 99999
)

var (
	ErrUnknownException = &Exception{
		Code: -1,
		Msg:  "Unknown exception",
	}
	ErrSuccess = &Exception{
		Code: 0,
		Msg:  "Success",
	}
	ErrLogonFailure = &Exception{
		Code: 10001,
		Msg:  "Logon failure",
	}

	ErrControlPacketParseError = &Exception{
		Code: 20012,
		Msg:  "Control packet parse error",
	}
	ErrDecryptDataError = &Exception{
		Code: 20013,
		Msg:  "Decrypt data error",
	}
	ErrDeserializationError = &Exception{
		Code: 20014,
		Msg:  "Deserialization error",
	}
	ErrSerializationError = &Exception{
		Code: 20015,
		Msg:  "Serialization error",
	}
	ErrUnsupportedCryptographicAlgorithm = &Exception{
		Code: 20016,
		Msg:  "Encryption & Decryption",
	}
	ErrAccountRestriction = &Exception{
		Code: 10002,
		Msg:  "Account restriction",
	}
	ErrWrongPassword = &Exception{
		Code: 10003,
		Msg:  "Wrong password",
	}
	ErrAccountDisabled = &Exception{
		Code: 10004,
		Msg:  "Account disabled",
	}
	ErrParameterError = &Exception{
		Code: 10005,
		Msg:  "Parameter errors",
	}
	ErrParameterLengthExceeds = &Exception{
		Code: 10006,
		Msg:  "Parameter length exceeds limit",
	}
	ErrIllegalCharacter = &Exception{
		Code: 10007,
		Msg:  "Illegal character in parameter",
	}
	ErrResourceNotFound = &Exception{
		Code: 10008,
		Msg:  "Resource not found",
	}
	ErrUsernamePasswordEmpty = &Exception{
		Code: 10009,
		Msg:  "Username or password cannot be empty",
	}
	ErrUnauthorizedAccess = &Exception{
		Code: 10010,
		Msg:  "Unauthorized access",
	}

	ErrInsufficientMemory = &Exception{
		Code: 20001,
		Msg:  "Insufficient memory",
	}
	ErrCredentialProviderError = &Exception{
		Code: 20002,
		Msg:  "Credential provider errors",
	}
	ErrPluginManagerException = &Exception{
		Code: 20003,
		Msg:  "Plugin manager exception",
	}
	ErrSystemUnknownException = &Exception{
		Code: 20004,
		Msg:  "internal error",
	}
	ErrUnknownLoginFailure = &Exception{
		Code: 90001,
		Msg:  "Unknown login failure",
	}
)
var errorMap = map[int]*Exception{
	-1:    ErrUnknownException,
	0:     ErrSuccess,
	10001: ErrLogonFailure,
	10002: ErrAccountRestriction,
	10003: ErrWrongPassword,
	10004: ErrAccountDisabled,
	10005: ErrParameterError,
	10006: ErrParameterLengthExceeds,
	10007: ErrIllegalCharacter,
	10008: ErrResourceNotFound,
	10009: ErrUsernamePasswordEmpty,
	10010: ErrUnauthorizedAccess,
	20001: ErrInsufficientMemory,
	20002: ErrCredentialProviderError,
	20003: ErrPluginManagerException,
	20004: ErrSystemUnknownException,
	20012: ErrControlPacketParseError,
	20013: ErrDecryptDataError,
	20014: ErrDeserializationError,
	20015: ErrSerializationError,
	20016: ErrUnsupportedCryptographicAlgorithm,
	90001: ErrUnknownLoginFailure,
}

func GetErrorByCode(code int) *Exception {
	e, ok := errorMap[code]
	if !ok {
		return ErrUnknownException
	}
	return e

}
func GetErrRandom() *Exception {

	seed := time.Now().UnixNano()
	r := rand.New(rand.NewSource(seed))

	keys := make([]int, 0, len(errorMap))
	for key := range errorMap {
		keys = append(keys, key)
	}
	randomIndex := r.Intn(len(keys))
	randomKey := keys[randomIndex]

	randomError := errorMap[randomKey]
	return randomError
}
