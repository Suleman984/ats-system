-- Migration for Candidate Communication Hub
-- Run this SQL in your Supabase SQL Editor

-- 1. Add new columns to applications table
ALTER TABLE applications 
ADD COLUMN IF NOT EXISTS cv_viewed_at TIMESTAMP,
ADD COLUMN IF NOT EXISTS cv_viewed_by UUID REFERENCES admins(id),
ADD COLUMN IF NOT EXISTS expected_response_date DATE,
ADD COLUMN IF NOT EXISTS last_status_update TIMESTAMP;

-- 2. Create messages table for two-way communication
CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    sender_type VARCHAR(20) NOT NULL CHECK (sender_type IN ('candidate', 'recruiter')),
    sender_id UUID REFERENCES admins(id), -- NULL if sender is candidate
    sender_email VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    is_read BOOLEAN DEFAULT false,
    read_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

-- 3. Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_messages_application_id ON messages(application_id);
CREATE INDEX IF NOT EXISTS idx_messages_sender_email ON messages(sender_email);
CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);
CREATE INDEX IF NOT EXISTS idx_applications_cv_viewed_at ON applications(cv_viewed_at);

-- 4. Update existing applications to set last_status_update
UPDATE applications 
SET last_status_update = COALESCE(reviewed_at, applied_at)
WHERE last_status_update IS NULL;

