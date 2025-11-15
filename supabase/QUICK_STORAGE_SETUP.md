# Quick Storage Setup (No Service Role Key Needed!)

If you can't find the service role key, don't worry! Here's the **easiest way** to set up file storage:

## Step 1: Get Your Anon Key (Required)

1. Go to https://app.supabase.com
2. Select your project
3. Click **Settings** (gear icon) → **API**
4. Copy:
   - **Project URL** (e.g., `https://xxxxx.supabase.co`)
   - **anon public** key (the long key starting with `eyJ...`)

Add to `backend/.env`:

```env
SUPABASE_URL=https://xxxxx.supabase.co
SUPABASE_ANON_KEY=your-anon-key-here
```

## Step 2: Create Buckets (2 minutes)

1. In Supabase dashboard, click **Storage** (left sidebar)
2. Click **"New bucket"** button
3. Create `resumes` bucket:
   - Name: `resumes`
   - **Toggle "Public bucket" ON** ✅
   - Click **Create bucket**
4. Click **"New bucket"** again
5. Create `portfolios` bucket:
   - Name: `portfolios`
   - **Toggle "Public bucket" ON** ✅
   - Click **Create bucket**

## Step 3: Set Up Policies (Important!)

For each bucket (`resumes` and `portfolios`):

1. Click on the bucket name
2. Click **"Policies"** tab
3. Click **"New Policy"**
4. Create **Policy 1:**
   - Name: `Allow public uploads`
   - Operation: **INSERT**
   - Definition: Type `true`
   - Save
5. Click **"New Policy"** again
6. Create **Policy 2:**
   - Name: `Allow public reads`
   - Operation: **SELECT**
   - Definition: Type `true`
   - Save

Repeat for both buckets.

## Step 4: Test It!

1. Start your backend: `cd backend && go run main.go`
2. Try uploading a file from the application form
3. Check if the file appears in Supabase Storage

## That's It! 🎉

You don't need the service role key at all if you create buckets manually!

## Troubleshooting

**"Bucket not found" error:**

- Make sure bucket names are exactly `resumes` and `portfolios`
- Check that buckets are set to **Public**

**"Permission denied" error:**

- Make sure you created the policies (Step 3)
- Policies must allow both INSERT and SELECT

**Can't find Storage in dashboard:**

- Look in the left sidebar
- Or go to: `https://app.supabase.com/project/YOUR_PROJECT/storage/buckets`
