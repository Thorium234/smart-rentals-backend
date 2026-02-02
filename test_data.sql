-- Insert Landlord
INSERT INTO users (id, email, password_hash, full_name, phone, role, created_at, updated_at) 
VALUES (1, 'test@landlord.com', 'hash', 'Test Landlord', '0700000000', 'landlord', NOW(), NOW()) 
ON CONFLICT (id) DO NOTHING;

-- Insert Config
INSERT INTO landlord_payment_configs (landlord_id, short_code, short_code_type, consumer_key, consumer_secret, created_at, updated_at) 
VALUES (1, '123456', 'paybill', 'key', 'secret', NOW(), NOW()) 
ON CONFLICT (landlord_id) DO UPDATE SET short_code = '123456';

-- Insert Property
INSERT INTO properties (id, landlord_id, title, description, location, property_type, created_at, updated_at) 
VALUES (1, 1, 'Test Property', 'Desc', 'Nairobi', 'Apartment', NOW(), NOW()) 
ON CONFLICT (id) DO NOTHING;

-- Insert Unit
INSERT INTO units (id, property_id, unit_name, unit_type, unit_price, created_at, updated_at) 
VALUES (1, 1, 'Unit 1', '1BHK', 1000, NOW(), NOW()) 
ON CONFLICT (id) DO NOTHING;

-- Insert Tenant
INSERT INTO tenants (id, tenant_name, payment_no1, rent, balance, unit_id, landlord_id, created_at, updated_at) 
VALUES (1, 'John Doe', '254712345678', 1000, 1000, 1, 1, NOW(), NOW()) 
ON CONFLICT (id) DO UPDATE SET balance = 1000;

-- Clean up
DELETE FROM payments WHERE receipt = 'TEST1234';
