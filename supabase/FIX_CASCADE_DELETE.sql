-- Fix cascade delete for applications
-- Run this SQL in your Supabase SQL Editor to prevent applications from being deleted when jobs are deleted

-- IMPORTANT: This fixes the issue where deleting a job causes a 500 error
-- The problem is that CASCADE tries to delete applications, but email_logs references applications
-- without ON DELETE CASCADE, causing a constraint violation.

-- NOTE: If you want to DELETE applications when jobs are deleted, use ENABLE_CASCADE_DELETE.sql instead
-- This file preserves applications when jobs are deleted (SET NULL)

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

-- Step 3: Recreate the foreign key with SET NULL instead of CASCADE
-- This way, when a job is deleted, applications remain but job_id becomes NULL
ALTER TABLE applications 
ADD CONSTRAINT applications_job_id_fkey 
FOREIGN KEY (job_id) 
REFERENCES jobs(id) 
ON DELETE SET NULL;

-- Note: After running this, existing applications will remain even if their jobs are deleted
-- The job_id will be NULL for applications whose jobs were deleted before this change

