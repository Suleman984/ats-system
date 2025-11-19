-- Add company_id to applications table for better tracking
-- This allows us to show applications even when jobs are deleted
-- Run this SQL in your Supabase SQL Editor

-- Add company_id column (nullable for now, will populate from existing jobs)
ALTER TABLE applications 
ADD COLUMN IF NOT EXISTS company_id UUID REFERENCES companies(id) ON DELETE CASCADE;

-- Populate company_id from existing jobs
UPDATE applications 
SET company_id = jobs.company_id 
FROM jobs 
WHERE applications.job_id = jobs.id 
AND applications.company_id IS NULL;

-- Make company_id NOT NULL after populating
-- Note: You may need to handle applications with deleted jobs manually
ALTER TABLE applications 
ALTER COLUMN company_id SET NOT NULL;

-- Create index for performance
CREATE INDEX IF NOT EXISTS idx_applications_company_id ON applications(company_id);

-- Now update the foreign key constraint for job_id to SET NULL
ALTER TABLE applications 
DROP CONSTRAINT IF EXISTS applications_job_id_fkey;

ALTER TABLE applications 
ADD CONSTRAINT applications_job_id_fkey 
FOREIGN KEY (job_id) 
REFERENCES jobs(id) 
ON DELETE SET NULL;

