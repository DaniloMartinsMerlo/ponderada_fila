package main

import (
    "fmt"
    "strconv"
    "strings"
)

func processarSensor(input SensorInput) (SensorOutput, error) {
    output := SensorOutput{
        DeviceID:    input.DeviceID,
        Timestamp:   input.Timestamp,
        SensorType:  input.SensorType,
        ReadingType: input.ReadingType,
    }

    readingType := strings.ToLower(strings.TrimSpace(input.ReadingType))
    switch readingType {
    case "discreto":
        output.DiscreteValue = input.Value
        output.NumericValue = 0

    case "analogica":
        output.DiscreteValue = "0"
        numValue, err := strconv.ParseFloat(input.Value, 64)
        if err != nil {
            return SensorOutput{}, fmt.Errorf("valor inválido para reading_type 'analogica'")
        }
        output.NumericValue = numValue

    default:
        return SensorOutput{}, fmt.Errorf("reading_type deve ser 'discreto' ou 'analogica'")
    }

    return output, nil
}