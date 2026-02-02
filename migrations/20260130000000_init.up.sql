-- Clean up existing tables
DROP TABLE IF EXISTS paybill_transactions CASCADE;
DROP TABLE IF EXISTS till_transactions CASCADE;
DROP TABLE IF EXISTS paybills CASCADE;
DROP TABLE IF EXISTS tills CASCADE;
DROP TABLE IF EXISTS landlord_payment_configs CASCADE;
DROP TABLE IF EXISTS payments CASCADE;
DROP TABLE IF EXISTS tenants CASCADE;
DROP TABLE IF EXISTS units CASCADE;
DROP TABLE IF EXISTS properties CASCADE;
DROP TABLE IF EXISTS users CASCADE;

-- Create users table
CREATE TABLE users (
    id              SERIAL PRIMARY KEY,
    email           VARCHAR(255) UNIQUE NOT NULL,
    password_hash   TEXT NOT NULL,
    full_name       VARCHAR(255) NOT NULL,
    phone           VARCHAR(50),
    role            VARCHAR(50) NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create properties table
CREATE TABLE properties (
    id              SERIAL PRIMARY KEY,
    landlord_id     INTEGER NOT NULL,
    title           VARCHAR(255) NOT NULL,
    description     TEXT,
    location        VARCHAR(255) NOT NULL,
    property_type   VARCHAR(100),
    vacancy         BOOLEAN NOT NULL DEFAULT TRUE,
    total_rent      INTEGER NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_properties_landlord
        FOREIGN KEY (landlord_id)
        REFERENCES users (id)
        ON DELETE CASCADE
);

-- Create units table
CREATE TABLE units (
    id              SERIAL PRIMARY KEY,
    property_id     INTEGER NOT NULL,
    unit_name       VARCHAR(100) NOT NULL,
    unit_type       VARCHAR(100),
    unit_price      INTEGER NOT NULL,
    vacancy         BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_units_property
        FOREIGN KEY (property_id)
        REFERENCES properties (id)
        ON DELETE CASCADE
);

-- Create tenants table
CREATE TABLE tenants (
    id              SERIAL PRIMARY KEY,
    tenant_name     VARCHAR(255) NOT NULL,
    payment_no1     VARCHAR(100),
    payment_no2     VARCHAR(100),
    unit_id         INTEGER NOT NULL,
    landlord_id     INTEGER NOT NULL,
    rent            NUMERIC(12,2),
    balance         NUMERIC(12, 2),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_tenants_unit
        FOREIGN KEY (unit_id)
        REFERENCES units (id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_tenant_landlord
        FOREIGN KEY (landlord_id)
        REFERENCES users (id)
        ON DELETE RESTRICT
);

-- Helper function and trigger for default balance
CREATE OR REPLACE FUNCTION set_default_balance()
RETURNS trigger AS $$
BEGIN
    IF NEW.balance IS NULL THEN
        NEW.balance := NEW.rent;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_set_default_balance
BEFORE INSERT ON tenants
FOR EACH ROW
EXECUTE FUNCTION set_default_balance();

-- Create tills and paybills
CREATE TABLE tills (
    id              BIGSERIAL PRIMARY KEY,
    till_number     VARCHAR(100) NOT NULL,
    landlord_id     BIGINT NOT NULL REFERENCES users(id),
    active          BOOLEAN DEFAULT true
);

CREATE TABLE paybills (
    id              BIGSERIAL PRIMARY KEY,
    paybill         VARCHAR(100) NOT NULL,
    account_number  VARCHAR(100) NOT NULL,
    landlord_id     BIGINT NOT NULL REFERENCES users(id),
    active          BOOLEAN DEFAULT true
);

-- Create transaction tables
CREATE TABLE till_transactions (
    id                  BIGSERIAL PRIMARY KEY,
    receipt_number      VARCHAR(100) NOT NULL,
    till_number         VARCHAR(100) NOT NULL,
    phone_number        VARCHAR(100) NOT NULL,
    amount              NUMERIC(12,2),
    trans_time          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    tenant_id           INTEGER NOT NULL REFERENCES tenants(id),
    landlord_id         INTEGER NOT NULL REFERENCES users(id)
);

CREATE TABLE paybill_transactions (
    id                  BIGSERIAL PRIMARY KEY,
    receipt_number      VARCHAR(100) NOT NULL,
    paybill             VARCHAR(100) NOT NULL,
    account_number      VARCHAR(100) NOT NULL,
    phone_number        VARCHAR(100) NOT NULL,
    amount              NUMERIC(12,2),
    trans_time          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    tenant_id           INTEGER NOT NULL REFERENCES tenants(id),
    landlord_id         INTEGER NOT NULL REFERENCES users(id)
);
