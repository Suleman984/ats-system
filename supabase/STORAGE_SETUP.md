# Supabase Storage Setup Guide

This guide will help you set up Supabase Storage for CV and portfolio file uploads.

## Prerequisites

- Supabase account (free tier includes 1GB storage, 2GB bandwidth/month)
- Your Supabase project URL and API keys

## Step 1: Get Your Supabase Credentials

1. Go to your Supabase project dashboard: https://app.supabase.com
2. Navigate to **Settings** → **API** (or click the gear icon in the left sidebar)
3. In the **Project API keys** section, you'll see:
   - **Project URL** (e.g., `https://xxxxx.supabase.co`) - at the top
   - **anon public** key - this is the one you need! (starts with `eyJ...`)
   - **service_role** key - **OPTIONAL** (only if you want automatic bucket creation)
     - To find it: Scroll down in the API settings page
     - It's in a separate section, sometimes you need to click "Reveal" to see it
     - **Note:** If you can't find it, don't worry! Just create buckets manually (see Step 3)

## Step 2: Add Environment Variables

Add these to your `backend/.env` file:

```env
SUPABASE_URL=https://xxxxx.supabase.co
SUPABASE_ANON_KEY=your_anon_key_here
SUPABASE_SERVICE_ROLE_KEY=your_service_role_key_here
```

**Important:**

- `SUPABASE_ANON_KEY` is used for public file uploads
- `SUPABASE_SERVICE_ROLE_KEY` is used for creating buckets and admin operations
- Never commit the service role key to version control!

## Step 3: Create Storage Buckets (Automatic)

The backend will automatically create the required buckets when you start the server:

- `resumes` - for CV/resume files (public)
- `portfolios` - for portfolio files (public)

If buckets are not created automatically, you can create them manually:

### Method 1: Via Supabase Dashboard (Easiest - No Service Role Key Needed!)

1. Go to your Supabase project dashboard
2. Click **Storage** in the left sidebar (or go to https://app.supabase.com/project/YOUR_PROJECT/storage/buckets)
3. Click the **"New bucket"** button (usually top right)
4. Create bucket named `resumes`:
   - **Name:** `resumes` (must be exactly this)
   - **Public bucket:** ✅ **Toggle this ON** (very important!)
   - Click **"Create bucket"**
5. Create bucket named `portfolios`:
   - Click **"New bucket"** again
   - **Name:** `portfolios` (must be exactly this)
   - **Public bucket:** ✅ **Toggle this ON** (very important!)
   - Click **"Create bucket"**

That's it! You're done with bucket creation. Now proceed to Step 4 for policies.

### Method 2: Via SQL (Alternative - If you prefer SQL)

Run this SQL in the Supabase SQL Editor:

```sql
-- Create resumes bucket
INSERT INTO storage.buckets (id, name, public)
VALUES ('resumes', 'resumes', true)
ON CONFLICT (id) DO NOTHING;

-- Create portfolios bucket
INSERT INTO storage.buckets (id, name, public)
VALUES ('portfolios', 'portfolios', true)
ON CONFLICT (id) DO NOTHING;
```

## Step 4: Set Up Storage Policies (Important!)

**⚠️ CRITICAL:** You must set up policies or file uploads won't work!

For public buckets, you need to set up policies to allow uploads and reads:

### Via Supabase Dashboard:

1. Go to **Storage** → **Policies**
2. For `resumes` bucket:

   - Click **New Policy**
   - Policy name: `Allow public uploads`
   - Allowed operation: **INSERT**
   - Policy definition:
     ```sql
     true
     ```
   - Click **Save**

   - Create another policy:
   - Policy name: `Allow public reads`
   - Allowed operation: **SELECT**
   - Policy definition:
     ```sql
     true
     ```
   - Click **Save**

3. Repeat the same policies for `portfolios` bucket

Run this SQL in the Supabase SQL Editor (go to SQL Editor in left sidebar):

```sql
-- Allow public uploads to resumes bucket
CREATE POLICY "Allow public uploads to resumes"
ON storage.objects FOR INSERT
TO public
WITH CHECK (bucket_id = 'resumes');

-- Allow public reads from resumes bucket
CREATE POLICY "Allow public reads from resumes"
ON storage.objects FOR SELECT
TO public
USING (bucket_id = 'resumes');

-- Allow public uploads to portfolios bucket
CREATE POLICY "Allow public uploads to portfolios"
ON storage.objects FOR INSERT
TO public
WITH CHECK (bucket_id = 'portfolios');

-- Allow public reads from portfolios bucket
CREATE POLICY "Allow public reads from portfolios"
ON storage.objects FOR SELECT
TO public
USING (bucket_id = 'portfolios');
```

## Step 5: Test the Setup

1. Start your backend server:

   ```bash
   cd backend
   go run main.go
   ```

2. You should see logs indicating buckets were created (or already exist)

3. Test file upload via the application form

## Troubleshooting

### Error: "Bucket not found"

- Make sure buckets are created in Supabase dashboard
- Check bucket names match exactly: `resumes` and `portfolios`
- Verify buckets are set to **Public**

### Error: "Permission denied" or "403 Forbidden"

- Check storage policies are set up correctly
- Verify `SUPABASE_ANON_KEY` is correct
- Make sure buckets are public

### Error: "Failed to upload file"

- Check file size (max 10MB)
- Verify file type is allowed (PDF, DOC, DOCX for CVs)
- Check Supabase Storage quota hasn't been exceeded

### Files not accessible

- Ensure buckets are set to **Public**
- Check storage policies allow SELECT operations
- Verify the public URL format is correct

## Storage Limits (Free Tier)

- **Storage:** 1GB
- **Bandwidth:** 2GB/month
- **File size limit:** 50MB per file (we limit to 10MB in code)

## Security Notes

- Public buckets allow anyone to upload files
- Consider implementing rate limiting for production
- File validation is done on both frontend and backend
- Consider adding virus scanning for production use

## Next Steps

Once storage is set up:

1. Test file uploads from the application form
2. Verify files are accessible via public URLs
3. Check that admins can view CVs in the dashboard

For production, consider:

- Adding file virus scanning
- Implementing rate limiting
- Adding file expiration policies
- Setting up CDN for faster file delivery
