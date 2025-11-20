-- Set NULL to preserve applications when jobs are deleted
-- This allows applications to remain in database for "Find Candidates" search
-- but they won't show in the Applications tab (filtered out by job_id IS NOT NULL)
-- Run this SQL in your Supabase SQL Editor

-- Step 1: Fix email_logs constraint first (to allow cascade delete of email_logs when application is deleted)
ALTER TABLE email_logs 
DROP CONSTRAINT IF EXISTS email_logs_application_id_fkey;

ALTER TABLE email_logs 
ADD CONSTRAINT email_logs_application_id_fkey 
FOREIGN KEY (application_id) 
REFERENCES applications(id) 
ON DELETE CASCADE;

-- Step 2: Drop the existing foreign key constraint on applications
-- The constraint name might be "fk_jobs_applications" or "applications_job_id_fkey"
ALTER TABLE applications 
DROP CONSTRAINT IF EXISTS fk_jobs_applications;

ALTER TABLE applications 
DROP CONSTRAINT IF EXISTS applications_job_id_fkey;

-- Step 3: Recreate the foreign key with SET NULL
-- This way, when a job is deleted, applications remain in database but job_id becomes NULL
-- Applications with NULL job_id will:
-- - NOT show in Applications tab (filtered out)
-- - STILL be searchable in Find Candidates tab
ALTER TABLE applications 
ADD CONSTRAINT applications_job_id_fkey 
FOREIGN KEY (job_id) 
REFERENCES jobs(id) 
ON DELETE SET NULL;

-- Note: After running this:
-- 1. When you delete a job, applications remain in database (job_id becomes NULL)
-- 2. Applications tab will NOT show applications with deleted jobs (filtered by job_id IS NOT NULL)
-- 3. Find Candidates will STILL show all applications including those with deleted jobs
-- 4. This preserves candidate data for future searches while keeping Applications tab clean

