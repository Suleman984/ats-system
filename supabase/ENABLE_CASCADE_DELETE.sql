-- Enable CASCADE DELETE for jobs -> applications
-- This will delete all applications when a job is deleted
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

-- Step 3: Recreate the foreign key with CASCADE delete
-- This way, when a job is deleted, all its applications will also be deleted
ALTER TABLE applications 
ADD CONSTRAINT applications_job_id_fkey 
FOREIGN KEY (job_id) 
REFERENCES jobs(id) 
ON DELETE CASCADE;

-- Note: After running this, when you delete a job, all its applications will be automatically deleted
-- This also means email_logs for those applications will be deleted (due to CASCADE on email_logs -> applications)

