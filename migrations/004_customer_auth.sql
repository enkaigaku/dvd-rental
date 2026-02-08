-- Add password_hash to customer table for authentication
-- This allows customers to log in and manage their accounts

ALTER TABLE customer 
ADD COLUMN IF NOT EXISTS password_hash TEXT;

-- Add index on email for faster login lookups
CREATE INDEX IF NOT EXISTS idx_customer_email ON customer(email);

-- Set a default password for existing customers (password: 'changeme')
-- In production, you would require customers to set their password on first login
UPDATE customer 
SET password_hash = '$2b$10$yCKt4D99d1R8z4LeXU/B6.9weZIMzI5MChWAtCHXFnKyhzG4uBHqS'
WHERE password_hash IS NULL;
