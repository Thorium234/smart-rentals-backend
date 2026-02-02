ALTER TABLE landlord_payment_configs 
ADD COLUMN environment VARCHAR(20) DEFAULT 'sandbox' NOT NULL,
ADD COLUMN validation_enabled BOOLEAN DEFAULT TRUE NOT NULL;
