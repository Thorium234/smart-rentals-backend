CREATE TABLE IF NOT EXISTS landlord_payment_configs (
    id SERIAL PRIMARY KEY,
    landlord_id INTEGER UNIQUE NOT NULL,
    short_code VARCHAR(255) NOT NULL,
    short_code_type VARCHAR(50) NOT NULL, -- 'paybill' or 'till'
    consumer_key VARCHAR(255) NOT NULL,
    consumer_secret VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_landlord FOREIGN KEY (landlord_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_landlord_configs_short_code ON landlord_payment_configs(short_code);
