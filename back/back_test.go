package main

import "testing"

func TestProcessarSensorAnalogicaSucesso(t *testing.T) {
    input := SensorInput{
        DeviceID:    "device-01",
        Timestamp:   "2024-01-01T00:00:00Z",
        SensorType:  "temperatura",
        ReadingType: "analogica",
        Value:       "23.5",
    }

    got, err := processarSensor(input)

    if err != nil {
        t.Fatalf("erro inesperado: %v", err)
    }
    if got.NumericValue != 23.5 {
        t.Errorf("esperava NumericValue=23.5, obteve %v", got.NumericValue)
    }
    if got.DiscreteValue != "0" {
        t.Errorf("esperava DiscreteValue='0', obteve %q", got.DiscreteValue)
    }
}

func TestProcessarSensorDiscretoSucesso(t *testing.T) {
    input := SensorInput{
        DeviceID:    "device-02",
        Timestamp:   "2024-01-01T00:00:00Z",
        SensorType:  "presenca",
        ReadingType: "discreto",
        Value:       "ativo",
    }

    got, err := processarSensor(input)

    if err != nil {
        t.Fatalf("erro inesperado: %v", err)
    }
    if got.DiscreteValue != "ativo" {
        t.Errorf("esperava DiscreteValue='ativo', obteve %q", got.DiscreteValue)
    }
    if got.NumericValue != 0 {
        t.Errorf("esperava NumericValue=0, obteve %v", got.NumericValue)
    }
}

func TestProcessarSensorAnalogicaValorInvalido(t *testing.T) {
    input := SensorInput{
        DeviceID:    "device-03",
        Timestamp:   "2024-01-01T00:00:00Z",
        SensorType:  "temperatura",
        ReadingType: "analogica",
        Value:       "dez",
    }

    _, err := processarSensor(input)

    if err == nil {
        t.Error("esperava erro para valor não numérico, mas não obteve")
    }
}

func TestProcessarSensorReadingTypeInvalido(t *testing.T) {
    input := SensorInput{
        DeviceID:    "device-04",
        Timestamp:   "2024-01-01T00:00:00Z",
        SensorType:  "temperatura",
        ReadingType: "Nenhum",
        Value:       "10",
    }

    _, err := processarSensor(input)

    if err == nil {
        t.Error("esperava erro para reading_type inválido, mas não obteve")
    }
}