import http from 'k6/http'
import { check, sleep } from 'k6'

export const options = {
  vus: 1000,
  duration: '30s'
}

export default function () {
  const data = {
    device_id:    "device-01",
    timestamp:    "2024-01-01T00:00:00Z",
    sensor_type:  "temperatura",
    reading_type: "analogica",
    value:        "23.5",
  }

  const params = {
    headers: { 'Content-Type': 'application/json' },
  }

  let res = http.post('http://localhost:8088/dados', JSON.stringify(data), params)
  check(res, { 'status 200': (r) => r.status === 200 })
  sleep(0.3)
}