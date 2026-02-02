-- Create payments table to support payment tracking and tenant balance management
CREATE TABLE payments (
    id              BIGSERIAL PRIMARY KEY,
    landlord_id     INTEGER NOT NULL,
    tenant_id       INTEGER,
    amount          NUMERIC(12,2) NOT NULL,
    status          VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    method          VARCHAR(50) NOT NULL,
    receipt         VARCHAR(255),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_payments_landlord
        FOREIGN KEY (landlord_id)
        REFERENCES users (id)
        ON DELETE CASCADE,

    CONSTRAINT fk_payments_tenant
        FOREIGN KEY (tenant_id)
        REFERENCES tenants (id)
        ON DELETE SET NULL
);

-- Create indexes for performance
CREATE INDEX idx_payments_landlord ON payments(landlord_id);
CREATE INDEX idx_payments_tenant ON payments(tenant_id);
CREATE INDEX idx_payments_status ON payments(status);
CREATE INDEX idx_payments_created_at ON payments(created_at DESC);

-- Add comment for documentation
COMMENT ON TABLE payments IS 'Stores all payment transactions including cash and M-Pesa payments';
COMMENT ON COLUMN payments.tenant_id IS 'Nullable to support unassigned M-Pesa payments that need manual assignment';
COMMENT ON COLUMN payments.status IS 'Payment status: PENDING, COMPLETED, FAILED, etc.';
COMMENT ON COLUMN payments.method IS 'Payment method: CASH, MPESA_TILL, MPESA_PAYBILL, etc.';
