package remote_schema

type RegisterDeviceRequest struct {
	DeviceName string `json:"device_name"`
}
type RegisterDeviceResponse struct {
	DeviceId string `json:"device_id"`
}
type DeviceRefreshTokenRequest struct {
	DeviceId string `json:"device_id"`
}
type DeviceRefreshTokenResponse struct {
	NewToken  string `json:"new_token"`
	DeviceId  string `json:"device_id"`
	ExpiresAt int64  `json:"expires_at"`
}
type RmttMsgUrls struct {
	Url    string `json:"url"`
	Weight int    `json:"weight"`
}
