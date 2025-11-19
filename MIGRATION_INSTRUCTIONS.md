# Database Migration Instructions

## Fix Cascade Delete Issue

To prevent applications from being deleted when jobs are deleted, you need to run two SQL migrations:

### Step 1: Fix Foreign Key Constraint

Run `supabase/FIX_CASCADE_DELETE.sql` in your Supabase SQL Editor:

```sql
-- This changes the foreign key from CASCADE to SET NULL
-- So when a job is deleted, applications remain but job_id becomes NULL
```

### Step 2: Add Company ID to Applications (Recommended)

Run `supabase/ADD_COMPANY_ID_TO_APPLICATIONS.sql` in your Supabase SQL Editor:

```sql
-- This adds company_id column to applications table
-- This allows us to track which company an application belongs to
-- even if the job is deleted
```

**Why this is important:**

- Without company_id, we can't filter applications by company if their job was deleted
- With company_id, we can show all applications for a company regardless of job status

### Step 3: Restart Backend

After running the migrations, restart your backend server so GORM picks up the schema changes.

## What This Fixes

1. ✅ Applications no longer deleted when jobs are deleted
2. ✅ Applications remain visible in admin dashboard even if job is deleted
3. ✅ Applications show "Job Deleted" instead of disappearing
4. ✅ Admin can still view, shortlist, reject, and delete applications
5. ✅ Bulk delete by status works correctly

## Verification

After migration:

1. Delete a job that has applications
2. Check Applications page - applications should still be visible
3. Applications should show "Job Deleted" in the Job column
4. You should be able to delete applications individually or in bulk
