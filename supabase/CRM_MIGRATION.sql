-- Migration for Candidate Relationship Management (CRM)
-- Run this SQL in your Supabase SQL Editor

-- 1. Add referral tracking fields to applications
ALTER TABLE applications 
ADD COLUMN IF NOT EXISTS referral_source VARCHAR(255),
ADD COLUMN IF NOT EXISTS referred_by_name VARCHAR(255),
ADD COLUMN IF NOT EXISTS referred_by_email VARCHAR(255),
ADD COLUMN IF NOT EXISTS referred_by_phone VARCHAR(50);

-- 2. Add talent pool flag to applications
ALTER TABLE applications 
ADD COLUMN IF NOT EXISTS in_talent_pool BOOLEAN DEFAULT false,
ADD COLUMN IF NOT EXISTS talent_pool_added_at TIMESTAMP,
ADD COLUMN IF NOT EXISTS talent_pool_added_by UUID REFERENCES admins(id);

-- 3. Create candidate_notes table for recruiter notes
CREATE TABLE IF NOT EXISTS candidate_notes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    admin_id UUID NOT NULL REFERENCES admins(id) ON DELETE CASCADE,
    note TEXT NOT NULL,
    is_private BOOLEAN DEFAULT false, -- Private notes only visible to creator
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 4. Create nurture_campaigns table for automated job alerts
CREATE TABLE IF NOT EXISTS nurture_campaigns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    job_id UUID REFERENCES jobs(id) ON DELETE SET NULL,
    email_sent_at TIMESTAMP NOT NULL,
    email_type VARCHAR(50) NOT NULL, -- 'job_alert', 'check_in', 'opportunity'
    subject VARCHAR(255),
    status VARCHAR(50) DEFAULT 'sent', -- 'sent', 'opened', 'clicked', 'bounced'
    created_at TIMESTAMP DEFAULT NOW()
);

-- 5. Create nurture_preferences table for candidate preferences
CREATE TABLE IF NOT EXISTS nurture_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    preferences JSONB, -- Store job preferences, location, salary range, etc.
    is_active BOOLEAN DEFAULT true,
    last_contacted_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(application_id, email)
);

-- 6. Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_applications_talent_pool ON applications(in_talent_pool) WHERE in_talent_pool = true;
CREATE INDEX IF NOT EXISTS idx_applications_referral_source ON applications(referral_source);
CREATE INDEX IF NOT EXISTS idx_candidate_notes_application_id ON candidate_notes(application_id);
CREATE INDEX IF NOT EXISTS idx_candidate_notes_admin_id ON candidate_notes(admin_id);
CREATE INDEX IF NOT EXISTS idx_candidate_notes_created_at ON candidate_notes(created_at);
CREATE INDEX IF NOT EXISTS idx_nurture_campaigns_application_id ON nurture_campaigns(application_id);
CREATE INDEX IF NOT EXISTS idx_nurture_campaigns_email_sent_at ON nurture_campaigns(email_sent_at);
CREATE INDEX IF NOT EXISTS idx_nurture_preferences_email ON nurture_preferences(email);
CREATE INDEX IF NOT EXISTS idx_nurture_preferences_active ON nurture_preferences(is_active) WHERE is_active = true;

-- 7. Add updated_at trigger for candidate_notes
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_candidate_notes_updated_at BEFORE UPDATE ON candidate_notes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_nurture_preferences_updated_at BEFORE UPDATE ON nurture_preferences
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

