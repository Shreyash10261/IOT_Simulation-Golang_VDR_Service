package worker

import (
	"math"
	"math/rand"
	"time"

	"github.com/team/vdr/internal/models"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// fieldProfile defines the simulation parameters for a specific telemetry field.
type fieldProfile struct {
	InitialValue float64
	MinValue     float64
	MaxValue     float64
	StepSize     float64 // max absolute change per tick
	Bias         float64 // positive = trend upward, negative = trend downward, 0 = neutral
}

// knownFieldProfiles maps field names to their simulation profiles.
// This ensures each metric type produces realistic, visually distinct telemetry.
var knownFieldProfiles = map[string]fieldProfile{
	"temperature": {InitialValue: 22.0, MinValue: 15.0, MaxValue: 35.0, StepSize: 1.5, Bias: 0.05},
	"brightness":  {InitialValue: 70.0, MinValue: 20.0, MaxValue: 100.0, StepSize: 3.0, Bias: 0.0},
	"fps":         {InitialValue: 30.0, MinValue: 24.0, MaxValue: 60.0, StepSize: 2.0, Bias: 0.0},
	"volume":      {InitialValue: 50.0, MinValue: 0.0, MaxValue: 100.0, StepSize: 4.0, Bias: 0.0},
	"battery":     {InitialValue: 85.0, MinValue: 10.0, MaxValue: 100.0, StepSize: 1.0, Bias: -0.3},
	"humidity":    {InitialValue: 45.0, MinValue: 20.0, MaxValue: 90.0, StepSize: 2.0, Bias: 0.0},
}

// defaultFloatProfile is used for any float field not in the known map.
var defaultFloatProfile = fieldProfile{
	InitialValue: 50.0, MinValue: 10.0, MaxValue: 90.0, StepSize: 2.0, Bias: 0.0,
}

// getProfile returns the simulation profile for a field, falling back to defaults.
func getProfile(fieldName string) fieldProfile {
	if p, ok := knownFieldProfiles[fieldName]; ok {
		return p
	}
	return defaultFloatProfile
}

// GenerateTelemetryValue generates the next simulated value for a telemetry field.
// It uses field-name-aware random walks with realistic ranges and step sizes.
func GenerateTelemetryValue(field models.TelemetryField, currentVal interface{}) interface{} {

	switch field.DataType {
	case "int", "integer":
		return generateInt(field.FieldName, currentVal)

	case "float", "double", "number":
		return generateFloat(field.FieldName, currentVal)

	case "bool", "boolean":
		if currentVal == nil {
			return true
		}
		if val, ok := currentVal.(bool); ok {
			if rand.Float64() < 0.05 {
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

// generateFloat produces a random-walk float value using the field's profile.
func generateFloat(fieldName string, currentVal interface{}) interface{} {
	p := getProfile(fieldName)

	if currentVal == nil {
		// Start at the initial value with a small random offset for variety between devices
		return p.InitialValue + (rand.Float64()-0.5)*p.StepSize
	}

	var current float64
	switch v := currentVal.(type) {
	case float64:
		current = v
	case float32:
		current = float64(v)
	case int:
		current = float64(v)
	default:
		return p.InitialValue
	}

	// Random walk: step = random(-StepSize, +StepSize) + Bias
	step := (rand.Float64()*2.0 - 1.0) * p.StepSize
	step += p.Bias

	// Occasionally add a larger "spike" for visual interest (10% chance)
	if rand.Float64() < 0.10 {
		step *= 2.5
	}

	newVal := current + step

	// Clamp to valid range
	newVal = math.Max(p.MinValue, math.Min(p.MaxValue, newVal))

	// Round to 2 decimal places for cleaner display
	return math.Round(newVal*100) / 100
}

// generateInt produces a random-walk integer value.
func generateInt(fieldName string, currentVal interface{}) interface{} {
	if currentVal == nil {
		if fieldName == "lamp_hours" {
			// Lamp hours start at a realistic value
			return 100 + rand.Intn(500)
		}
		return 50 + rand.Intn(50)
	}

	var current int
	switch v := currentVal.(type) {
	case int:
		current = v
	case float64:
		current = int(v)
	default:
		return 50
	}

	// Different behavior per field
	switch fieldName {
	case "lamp_hours":
		// Lamp hours only increase (realistic: projector accumulates hours)
		if rand.Float64() < 0.3 {
			return current + 1
		}
		return current
	default:
		// General int: random walk up or down
		change := 0
		roll := rand.Float64()
		if roll < 0.35 {
			change = 1
		} else if roll < 0.70 {
			change = -1
		}
		// else no change (30% of the time)

		result := current + change
		if result < 0 {
			result = 0
		}
		if result > 1000 {
			result = 1000
		}
		return result
	}
}

