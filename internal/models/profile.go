package models

type ProtocolEndpoint struct {
	Protocol         string                 `json:"protocol"`
	Command          string                 `json:"command"`
	Description      string                 `json:"description"`
	ExpectedResponse string                 `json:"expected_response"`
	Variables        map[string]interface{} `json:"variables,omitempty"`
}

type TelemetryField struct {
	FieldName string `json:"field_name"`
	DataType  string `json:"data_type"`
	Unit      string `json:"unit,omitempty"`
}

type DeviceProfile struct {
	DeviceType   string             `json:"device_type"`
	Manufacturer string             `json:"manufacturer"`
	Model        string             `json:"model"`
	Endpoints    []ProtocolEndpoint `json:"endpoints"`
	Telemetry    []TelemetryField   `json:"telemetry"`
}
