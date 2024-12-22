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
	// 1000x

	ErrUserLogonFailure = &Exception{
		Code: 10001,
		Msg:  "Logon failure",
	}

	ErrUserAccountRestriction = &Exception{
		Code: 10002,
		Msg:  "Account restriction",
	}
	ErrUserWrongPassword = &Exception{
		Code: 10003,
		Msg:  "Wrong password",
	}
	ErrUserAccountDisabled = &Exception{
		Code: 10004,
		Msg:  "Account disabled",
	}
	ErrUserParameterError = &Exception{
		Code: 10005,
		Msg:  "Parameter errors",
	}
	ErrUserParameterLengthExceeds = &Exception{
		Code: 10006,
		Msg:  "Parameter length exceeds limit",
	}
	ErrUserIllegalCharacter = &Exception{
		Code: 10007,
		Msg:  "Illegal character in parameter",
	}
	ErrUserResourceNotFound = &Exception{
		Code: 10008,
		Msg:  "Resource not found",
	}
	ErrUsernamePasswordEmpty = &Exception{
		Code: 10009,
		Msg:  "Username or password cannot be empty",
	}
	ErrUserUnauthorizedAccess = &Exception{
		Code: 10010,
		Msg:  "Unauthorized access",
	}
	ErrUserSecretKeyError = &Exception{
		Code: 10011,
		Msg:  "Secret key error, please check",
	}
	ErrUserControlPacketStructureError = &Exception{
		Code: 10012,
		Msg:  "Control packet structure error, please check",
	}
	ErrUserMessageDecryptionFailed = &Exception{
		Code: 10013,
		Msg:  "Message decryption failed, please check the password",
	}
	ErrUserMessageDeserializationFailed = &Exception{
		Code: 10014,
		Msg:  "Message deserialization failed, please check the password",
	}
	ErrUserInvalidHostAddress = &Exception{
		Code: 10015,
		Msg:  "Invalid host address",
	}
	ErrUserUnsupportedEncryptionType = &Exception{
		Code: 10016,
		Msg:  "Unsupported encryption type",
	}
	ErrUserOperationNotSupportedOnOS = &Exception{
		Code: 10017,
		Msg:  "This operation is not supported on the current operating system",
	}
	ErrUserCertificateFormatError = &Exception{
		Code: 10018,
		Msg:  "Certificate format error",
	}
	ErrUserMethodNotAllowed = &Exception{
		Code: 10019,
		Msg:  "Method not allowed",
	}
	ErrUserUnlockNotInLockScreenState = &Exception{
		Code: 10020,
		Msg:  "The verification of the username and password was successful, but there's no need to unlock it on the non-lock screen interface!",
	}

	ErrUserTooManyRequests = &Exception{
		Code: 11205,
		Msg:  "Too many requests",
	}
	//2xx

	ErrSystemInsufficientMemory = &Exception{
		Code: 20001,
		Msg:  "Insufficient memory",
	}
	ErrSystemCredentialProviderError = &Exception{
		Code: 20002,
		Msg:  "Credential provider errors",
	}
	ErrSystemPluginManagerException = &Exception{
		Code: 20003,
		Msg:  "Plugin manager exception",
	}
	ErrSystemUnknownException = &Exception{
		Code: 20004,
		Msg:  "internal error",
	}
	ErrSystemInternalParameterError = &Exception{
		Code: 20005,
		Msg:  "System internal parameter error",
	}
	ErrSystemBluetoothServiceInitFailure = &Exception{
		Code: 20007,
		Msg:  "Bluetooth service init failure",
	}
	ErrSystemBluetoothServiceStopFailure = &Exception{
		Code: 20008,
		Msg:  "Bluetooth service stop failure",
	}
	ErrSystemServiceAlreadyRunning = &Exception{
		Code: 20009,
		Msg:  "Service already running",
	}

	ErrSystemGenTokenErr = &Exception{
		Code: 20010,
		Msg:  "Generating token failed",
	}
	ErrSystemSetPowerSaveModeError = &Exception{
		Code: 20011,
		Msg:  "Set power save mode error",
	}
	ErrSystemMessageSerializationFailed = &Exception{
		Code: 20015,
		Msg:  "Message serialization failed",
	}
	ErrSystemSevereConfigurationError = &Exception{
		Code: 20016,
		Msg:  "Severe error in current configuration, please reinitialize the configuration",
	}
	ErrSystemInvalidAlgoKeyLen = &Exception{
		Code: 20017,
		Msg:  "Invalid algo key length,Please check System configuration",
	}
	ErrSystemServiceNotFullyStarted = &Exception{
		Code: 20018,
		Msg:  "Service not fully started",
	}
	ErrSystemRequestTimeout = &Exception{
		Code: 20019,
		Msg:  "Request timeout!",
	}
	//9xx
	ErrUnknownLoginFailure = &Exception{
		Code: 90001,
		Msg:  "Unknown login failure",
	}
)
var errorMap = map[int]*Exception{
	-1:    ErrUnknownException,
	0:     ErrSuccess,
	10001: ErrUserLogonFailure,
	10002: ErrUserAccountRestriction,
	10003: ErrUserWrongPassword,
	10004: ErrUserAccountDisabled,
	10005: ErrUserParameterError,
	10006: ErrUserParameterLengthExceeds,
	10007: ErrUserIllegalCharacter,
	10008: ErrUserResourceNotFound,
	10009: ErrUsernamePasswordEmpty,
	10010: ErrUserUnauthorizedAccess,
	10011: ErrUserSecretKeyError,
	10012: ErrUserControlPacketStructureError,
	10013: ErrUserMessageDecryptionFailed,
	10014: ErrUserMessageDeserializationFailed,
	10015: ErrUserInvalidHostAddress,
	10016: ErrUserUnsupportedEncryptionType,
	10017: ErrUserOperationNotSupportedOnOS,
	10018: ErrUserCertificateFormatError,
	10019: ErrUserMethodNotAllowed,
	10020: ErrUserUnlockNotInLockScreenState,

	20001: ErrSystemInsufficientMemory,
	20002: ErrSystemCredentialProviderError,
	20003: ErrSystemPluginManagerException,
	20004: ErrSystemUnknownException,
	20005: ErrSystemInternalParameterError,
	20007: ErrSystemBluetoothServiceInitFailure,
	20008: ErrSystemBluetoothServiceStopFailure,
	20009: ErrSystemServiceAlreadyRunning,

	20011: ErrSystemSetPowerSaveModeError,

	20010: ErrSystemGenTokenErr,
	20015: ErrSystemMessageSerializationFailed,
	20016: ErrSystemSevereConfigurationError,
	20017: ErrSystemInvalidAlgoKeyLen,
	20018: ErrSystemServiceNotFullyStarted,
	20019: ErrSystemRequestTimeout,
	//9xx
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
