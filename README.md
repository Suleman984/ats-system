# ATS System - Complete Project

A complete Applicant Tracking System (ATS) built with Golang backend and Next.js TypeScript frontend.

## Project Structure

```
ats-system/
├── backend/              # Golang API server
│   ├── config/          # Configuration files
│   ├── controllers/     # Request handlers
│   ├── models/          # Database models
│   ├── middleware/      # Middleware (auth, CORS)
│   ├── routes/          # API routes
│   ├── services/        # Business logic (email, etc.)
│   └── utils/           # Utility functions
├── frontend/            # Next.js TypeScript frontend
│   ├── app/            # Next.js App Router pages
│   └── lib/            # API client and state management
└── supabase/           # Database schema files
    └── schema.sql      # SQL schema for Supabase
```

## Quick Start

### 1. Database Setup

1. Go to [Supabase](https://supabase.com) and create a project
2. Navigate to SQL Editor
3. Copy and paste the contents of `supabase/schema.sql`
4. Run the SQL to create all tables
5. Get your database connection string from Settings → Database

### 2. Backend Setup

```bash
cd backend
go mod download
cp .env.example .env
# Edit .env with your credentials
go run main.go
```

Backend will run on http://localhost:8080

### 3. Frontend Setup

```bash
cd frontend
npm install
cp .env.local.example .env.local
# Edit .env.local with your API URL
npm run dev
```

Frontend will run on http://localhost:3000

## Environment Variables

### Backend (.env)

- `DATABASE_URL` - Supabase PostgreSQL connection string
- `JWT_SECRET` - Secret key for JWT tokens
- `PORT` - Server port (default: 8080)
- `RESEND_API_KEY` - Resend.dev API key for emails
- `RESEND_FROM_EMAIL` - Email address to send from

### Frontend (.env.local)

- `NEXT_PUBLIC_API_URL` - Backend API URL (e.g., http://localhost:8080/api)

## Features

✅ Company registration and authentication
✅ Job posting and management
✅ Public job listings
✅ Application submission
✅ Application management (shortlist/reject)
✅ Email notifications
✅ Admin dashboard
✅ Embed code generation

## API Endpoints

### Public

- `POST /api/auth/register` - Register company
- `POST /api/auth/login` - Admin login
- `GET /api/jobs/:companyId/public` - Get public jobs
- `POST /api/applications` - Submit application

### Protected (Requires JWT)

- `POST /api/jobs` - Create job
- `GET /api/jobs` - Get company jobs
- `PUT /api/jobs/:id` - Update job
- `DELETE /api/jobs/:id` - Delete job
- `GET /api/applications` - Get applications
- `PUT /api/applications/:id/shortlist` - Shortlist candidate
- `PUT /api/applications/:id/reject` - Reject candidate

## Deployment

### Backend (Render.com)

1. Push backend to GitHub
2. Connect repository to Render
3. Set environment variables
4. Deploy

### Frontend (Vercel)

1. Push frontend to GitHub
2. Import to Vercel
3. Set environment variables
4. Deploy

## Technology Stack

- **Backend**: Golang, Gin, GORM, PostgreSQL
- **Frontend**: Next.js 14, TypeScript, Tailwind CSS, Zustand
- **Database**: Supabase (PostgreSQL)
- **Email**: Resend.dev
- **Storage**: Cloudflare R2 (optional for MVP)

## License

MIT
