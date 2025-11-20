-- Additional indexes for optimized queries
-- Run this SQL in your Supabase SQL Editor after running ADD_COMPANY_ID_TO_APPLICATIONS.sql

-- Composite index for Find Candidates query (company_id + job_id NULL check)
-- This helps with queries that filter by company_id and check for NULL job_id
CREATE INDEX IF NOT EXISTS idx_applications_company_job_null 
ON applications(company_id) 
WHERE job_id IS NULL;

-- Composite index for active applications (company_id + job_id)
-- This helps with queries that join applications with jobs
CREATE INDEX IF NOT EXISTS idx_applications_company_job 
ON applications(company_id, job_id) 
WHERE job_id IS NOT NULL;

-- Index for status filtering (commonly used in both Applications and Find Candidates)
CREATE INDEX IF NOT EXISTS idx_applications_company_status 
ON applications(company_id, status);

-- Index for experience filtering (used in Find Candidates)
CREATE INDEX IF NOT EXISTS idx_applications_company_experience 
ON applications(company_id, years_of_experience);

-- Note: These indexes will significantly improve query performance for:
-- 1. Find Candidates search (especially with deleted jobs)
-- 2. Applications tab filtering
-- 3. Status-based queries
-- 4. Experience-based filtering

