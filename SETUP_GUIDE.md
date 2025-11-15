# Complete Setup Guide - ATS System

This guide will walk you through setting up the entire ATS system from scratch.

## Prerequisites

- Go 1.21+ installed
- Node.js 18+ installed
- Supabase account (free tier works)
- Resend.dev account (for emails)
- Git installed

## Step 1: Database Setup (Supabase)

1. **Create Supabase Project**

   - Go to https://supabase.com
   - Sign up/login
   - Click "New Project"
   - Fill in project details
   - Wait for project to be created

2. **Run Database Schema**

   - Go to SQL Editor in Supabase dashboard
   - Open `supabase/schema.sql` from this project
   - Copy all contents
   - Paste into SQL Editor
   - Click "Run" or press Cmd/Ctrl + Enter
   - Verify tables are created in Table Editor

3. **Get Database Connection String**
   - Go to Settings → Database
   - Find "Connection string" section
   - Copy the "URI" format
   - It looks like: `postgresql://postgres:[YOUR-PASSWORD]@db.[PROJECT-REF].supabase.co:5432/postgres`
   - Save this for backend `.env` file

## Step 2: Email Service Setup (Resend)

1. **Create Resend Account**

   - Go to https://resend.dev
   - Sign up for free account
   - Verify your email

2. **Get API Key**

   - Go to API Keys section
   - Create a new API key
   - Copy the key (starts with `re_`)
   - Save this for backend `.env` file

3. **Domain Setup (Optional for MVP)**
   - For MVP, you can use Resend's test domain
   - For production, add your own domain

## Step 3: Backend Setup

1. **Navigate to Backend Directory**

   ```bash
   cd backend
   ```

2. **Install Dependencies**

   ```bash
   go mod download
   ```

3. **Create Environment File**

   ```bash
   cp .env.example .env
   ```

4. **Edit .env File**
   Open `.env` and fill in:

   ```
   DATABASE_URL=postgresql://postgres:[YOUR-PASSWORD]@db.[PROJECT-REF].supabase.co:5432/postgres
   JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
   PORT=8080
   GIN_MODE=release
   RESEND_API_KEY=re_xxxxxxxxxxxxx
   RESEND_FROM_EMAIL=noreply@yourdomain.com
   ```

5. **Run Backend**

   ```bash
   go run main.go
   ```

   You should see:

   ```
   ✅ Database connected successfully!
   ✅ Database tables created!
   Server starting on port 8080
   ```

6. **Test Backend**
   Open another terminal and test:
   ```bash
   curl http://localhost:8080/api/jobs/test/public
   ```
   Should return `{"jobs":[]}`

## Step 4: Frontend Setup

1. **Navigate to Frontend Directory**

   ```bash
   cd frontend
   ```

2. **Install Dependencies**

   ```bash
   npm install
   ```

3. **Create Environment File**

   ```bash
   cp .env.local.example .env.local
   ```

4. **Edit .env.local File**

   ```
   NEXT_PUBLIC_API_URL=http://localhost:8080/api
   ```

5. **Run Frontend**

   ```bash
   npm run dev
   ```

   Frontend will be available at http://localhost:3000

## Step 5: Test Complete Flow

### 1. Register a Company

- Go to http://localhost:3000/admin/register
- Fill in:
  - Company Name: "Test Company"
  - Your Name: "Admin User"
  - Email: "admin@test.com"
  - Password: "password123"
- Click Register
- You should be redirected to dashboard

### 2. Create a Job

- In dashboard, click "Post New Job"
- Fill in job details:
  - Title: "Frontend Developer"
  - Description: "Looking for React developer"
  - Deadline: (future date)
  - Other fields as needed
- Click "Post Job"
- Job should appear in Jobs list

### 3. View Public Jobs

- Note your company ID from the URL or database
- Go to: http://localhost:3000/jobs/[YOUR-COMPANY-ID]
- You should see the job you just created

### 4. Apply for Job

- Click "Apply Now" on the job
- Fill in application form
- Use a Google Drive link for resume URL (for MVP)
- Submit application
- Check email for confirmation

### 5. Review Application

- Go back to admin dashboard
- Navigate to Applications
- You should see the application
- Test Shortlist and Reject buttons
- Check emails are sent

## Step 6: Deployment

### Backend Deployment (Render.com)

1. **Push to GitHub**

   ```bash
   git init
   git add .
   git commit -m "Initial commit"
   # Create repo on GitHub
   git remote add origin https://github.com/yourusername/ats-backend.git
   git push -u origin main
   ```

2. **Deploy on Render**
   - Go to https://render.com
   - Sign up/login
   - Click "New +" → "Web Service"
   - Connect GitHub repository
   - Configure:
     - Name: `ats-backend`
     - Environment: `Go`
     - Build Command: `go build -o main`
     - Start Command: `./main`
     - Instance Type: Free
   - Add Environment Variables (from your `.env`)
   - Click "Create Web Service"
   - Wait for deployment
   - Copy the URL (e.g., `https://ats-backend-xxxx.onrender.com`)

### Frontend Deployment (Vercel)

1. **Push to GitHub**

   ```bash
   cd frontend
   git init
   git add .
   git commit -m "Initial commit"
   # Create repo on GitHub
   git remote add origin https://github.com/yourusername/ats-frontend.git
   git push -u origin main
   ```

2. **Deploy on Vercel**

   - Go to https://vercel.com
   - Sign up/login with GitHub
   - Click "Add New Project"
   - Import your frontend repository
   - Configure:
     - Framework Preset: Next.js (auto-detected)
   - Add Environment Variable:
     - `NEXT_PUBLIC_API_URL` = Your Render backend URL + `/api`
   - Click "Deploy"
   - Wait for deployment
   - Copy the URL (e.g., `https://ats-frontend.vercel.app`)

3. **Update Frontend Environment**
   - Update `NEXT_PUBLIC_API_URL` in Vercel to point to deployed backend
   - Redeploy if needed

## Troubleshooting

### Database Connection Issues

- Verify `DATABASE_URL` is correct
- Check Supabase project is active
- Ensure password is URL-encoded if it contains special characters

### CORS Errors

- Backend already has CORS middleware
- If issues persist, check frontend API URL is correct

### Email Not Sending

- Verify `RESEND_API_KEY` is correct
- Check Resend dashboard for any errors
- Ensure `RESEND_FROM_EMAIL` is set

### Token Issues

- Check JWT_SECRET is set
- Verify token is stored in localStorage
- Check token expiration (24 hours default)

### Build Errors

- Backend: Ensure Go 1.21+ is installed
- Frontend: Ensure Node.js 18+ is installed
- Run `go mod tidy` for backend
- Run `npm install` for frontend

## Next Steps

1. **Customize Branding**

   - Update colors in Tailwind config
   - Change logo and company name
   - Customize email templates

2. **Add File Upload**

   - Implement Cloudflare R2 integration
   - Add file upload to application form
   - Update resume handling

3. **Add Auto-Shortlisting**

   - Implement scoring algorithm
   - Add criteria matching
   - Automate shortlist process

4. **Add Payment Integration**
   - Integrate Stripe
   - Add subscription management
   - Implement tier restrictions

## Support

For issues or questions:

- Check the main README.md
- Review code comments
- Check Supabase/Resend documentation
- Review Next.js and Go documentation

## Success!

You now have a fully functional ATS system! 🎉
