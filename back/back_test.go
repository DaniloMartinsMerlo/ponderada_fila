package main

import (
    "bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
 
	"github.com/gin-gonic/gin"
)
 
// ── processarSensor ──────────────────────────────────────────────────────────
 
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

// Campos do output devem refletir exatamente o input
func TestProcessarSensorCamposRepassados(t *testing.T) {
    input := SensorInput{
        DeviceID:    "dev-99",
        Timestamp:   "2025-06-15T12:00:00Z",
        SensorType:  "umidade",
        ReadingType: "analogica",
        Value:       "55.0",
    }
    got, err := processarSensor(input)
    if err != nil {
        t.Fatalf("erro inesperado: %v", err)
    }
    if got.DeviceID != input.DeviceID {
        t.Errorf("DeviceID: esperava %q, obteve %q", input.DeviceID, got.DeviceID)
    }
    if got.Timestamp != input.Timestamp {
        t.Errorf("Timestamp: esperava %q, obteve %q", input.Timestamp, got.Timestamp)
    }
    if got.SensorType != input.SensorType {
        t.Errorf("SensorType: esperava %q, obteve %q", input.SensorType, got.SensorType)
    }
    if got.ReadingType != input.ReadingType {
        t.Errorf("ReadingType: esperava %q, obteve %q", input.ReadingType, got.ReadingType)
    }
}

// ReadingType com espaços/capitalização deve ser aceito
func TestProcessarSensorReadingTypeComEspacosEMaiusculas(t *testing.T) {
    casos := []string{"  Analogica  ", "ANALOGICA", " analogica"}
    for _, readtype := range casos {
        input := SensorInput{
            DeviceID:    "device-05",
            Timestamp:   "2024-01-01T00:00:00Z",
            SensorType:  "temperatura",
            ReadingType: readtype,
            Value:       "10.0",
        }
        _, err := processarSensor(input)
        if err != nil {
            t.Errorf("ReadingType %q deveria ser aceito, mas retornou erro: %v", rt, err)
        }
    }
}

// Valor numérico negativo é válido para analógica
func TestProcessarSensorAnalogicaValorNegativo(t *testing.T) {
    input := SensorInput{
        DeviceID:    "device-06",
        Timestamp:   "2024-01-01T00:00:00Z",
   -     SensorType:  "temperatura",
        ReadingType: "analogica",
        Value:       "-10.5",
    }
    got, err := processarSensor(input)
    if err != nil {
        t.Fatalf("erro inesperado: %v", err)
    }
    if got.NumericValue != -10.5 {
        t.Errorf("esperava NumericValue=-10.5, obteve %v", got.NumericValue)
    }
}

// Valor zero é válido para analógica
func TestProcessarSensorAnalogicaValorZero(t *testing.T) {
    input := SensorInput{
        DeviceID:    "device-07",
        Timestamp:   "2024-01-01T00:00:00Z",
        SensorType:  "temperatura",
        ReadingType: "analogica",
        Value:       "0",
    }
    got, err := processarSensor(input)
    if err != nil {
        t.Fatalf("erro inesperado: %v", err)
    }
    if got.NumericValue != 0 {
        t.Errorf("esperava NumericValue=0, obteve %v", got.NumericValue)
    }
}
 
// ── handler POST /dados (sem RabbitMQ) ───────────────────────────────────────
 
// configura um router Gin isolado que chama processarSensor mas não publica no RabbitMQ
func setupRouterSemRabbit() *gin.Engine {
    gin.SetMode(gin.TestMode)
    r := gin.New()
    r.POST("/dados", func(c *gin.Context) {
        var input SensorInput
        if err := c.ShouldBindJSON(&input); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }
        output, err := processarSensor(input)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, output)
    })
    return r
}
 
func TestPostDadosSucesso(t *testing.T) {
    r := setupRouterSemRabbit()
 
    payload := SensorInput{
        DeviceID:    "device-10",
        Timestamp:   "2024-01-01T00:00:00Z",
        SensorType:  "temperatura",
        ReadingType: "analogica",
        Value:       "36.6",
	}
    body, _ := json.Marshal(payload)

    req := httptest.NewRequest(http.MethodPost, "/dados", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)
 
    if w.Code != http.StatusOK {
        t.Fatalf("esperava status 200, obteve %d", w.Code)
    }

    var out SensorOutput
    if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
        t.Fatalf("resposta não é um SensorOutput válido: %v", err)
    }
    if out.NumericValue != 36.6 {
        t.Errorf("esperava NumericValue=36.6, obteve %v", out.NumericValue)
    }
}
 
func TestPostDadosBodyVazio(t *testing.T) {
    r := setupRouterSemRabbit()

    req := httptest.NewRequest(http.MethodPost, "/dados", bytes.NewBuffer([]byte(`{}`)))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)
 
    if w.Code != http.StatusBadRequest {
        t.Errorf("esperava status 400, obteve %d", w.Code)
    }
}

func TestPostDadosReadingTypeInvalido(t *testing.T) {
    r := setupRouterSemRabbit()
 
    payload := SensorInput{
        DeviceID:    "device-11",
        Timestamp:   "2024-01-01T00:00:00Z",
        SensorType:  "temperatura",
        ReadingType: "invalido",
        Value:       "10",
    }
    body, _ := json.Marshal(payload)

    req := httptest.NewRequest(http.MethodPost, "/dados", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)

    if w.Code != http.StatusBadRequest {
        t.Errorf("esperava status 400, obteve %d", w.Code)
    }
}

func TestPostDadosContentTypeErrado(t *testing.T) {
    r := setupRouterSemRabbit()

    body := []byte(`device_id=dev&timestamp=2024-01-01T00:00:00Z`)
    req := httptest.NewRequest(http.MethodPost, "/dados", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "text/plain")
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)

    if w.Code != http.StatusBadRequest {
        t.Errorf("esperava status 400, obteve %d", w.Code)
    }
}