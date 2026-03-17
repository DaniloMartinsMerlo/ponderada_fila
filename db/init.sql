CREATE TABLE sensor_data (
    id SERIAL PRIMARY KEY,
    device_id VARCHAR(255),
    timestamp TIMESTAMP,
    sensor_type VARCHAR(100),
    reading_type VARCHAR(100),
    discrete_value VARCHAR(255),
    numeric_value NUMERIC
);