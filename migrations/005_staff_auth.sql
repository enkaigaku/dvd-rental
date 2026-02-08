-- Add password_hash to staff table for secure authentication
-- Pagila uses plaintext password in 'password' column, we add bcrypt hash

ALTER TABLE staff 
ADD COLUMN IF NOT EXISTS password_hash TEXT;

-- Create index on username for faster login
CREATE INDEX IF NOT EXISTS idx_staff_username ON staff(username);

-- Set a default password for existing staff (password: 'admin123')
-- bcrypt hash of 'admin123'
UPDATE staff 
SET password_hash = '$2b$10$yCKt4D99d1R8z4LeXU/B6.9weZIMzI5MChWAtCHXFnKyhzG4uBHqS'
WHERE password_hash IS NULL;
