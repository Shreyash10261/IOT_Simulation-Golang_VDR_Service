package worker

import (
	"math/rand"
	"time"

	"github.com/team/vdr/internal/models"
)

// GenerateTelemetryValue generates the next simulated value for a telemetry field.
// It uses random walks for numerical types to simulate physical environments.
func GenerateTelemetryValue(field models.TelemetryField, currentVal interface{}) interface{} {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	switch field.DataType {
	case "int", "integer":
		if currentVal == nil {
			return 100
		}
		if val, ok := currentVal.(int); ok {
			if r.Float64() < 0.2 {
				return val + 1
			}
			return val
		}
		if val, ok := currentVal.(float64); ok {
			if r.Float64() < 0.2 {
				return int(val) + 1
			}
			return int(val)
		}
		return 100

	case "float", "double", "number":
		if currentVal == nil {
			return 40.0
		}
		var currentFloat float64
		switch v := currentVal.(type) {
		case float64:
			currentFloat = v
		case float32:
			currentFloat = float64(v)
		case int:
			currentFloat = float64(v)
		default:
			return 40.0
		}

		// Random walk
		change := (r.Float64() - 0.5) * 1.0
		newVal := currentFloat + change
		if newVal < 30.0 {
			newVal = 30.0
		}
		if newVal > 80.0 {
			newVal = 80.0
		}
		return newVal

	case "bool", "boolean":
		if currentVal == nil {
			return true
		}
		if val, ok := currentVal.(bool); ok {
			if r.Float64() < 0.05 {
				return !val
			}
			return val
		}
		return true

	case "string":
		return "NORMAL"

	default:
		return "NORMAL"
	}
}
