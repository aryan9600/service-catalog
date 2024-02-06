CREATE TABLE versions (
    id SERIAL PRIMARY KEY,
    service_id INTEGER NOT NULL,
    version VARCHAR(50) NOT NULL,
    changelog TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (service_id) REFERENCES services(id),
    UNIQUE(service_id, version)
);
