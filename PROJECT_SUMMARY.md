# ATS System - Complete Project Summary

## 📋 Table of Contents

1. [Project Overview](#project-overview)
2. [Technology Stack](#technology-stack)
3. [System Architecture](#system-architecture)
4. [Features Implemented](#features-implemented)
5. [Database Schema](#database-schema)
6. [API Endpoints](#api-endpoints)
7. [Frontend Structure](#frontend-structure)
8. [Security Features](#security-features)
9. [Deployment & Setup](#deployment--setup)
10. [File Structure](#file-structure)
11. [Key Components](#key-components)

---

## 🎯 Project Overview

**Applicant Tracking System (ATS)** - A comprehensive SaaS platform for companies to manage job postings, receive applications, and track candidates through the hiring process.

### Core Purpose

- Enable companies to post job openings
- Allow candidates to apply for jobs
- Help admins manage applications and shortlist candidates
- Provide AI-powered CV analysis and matching
- Support embedded dashboard integration
- Track all activities with comprehensive logging

### User Roles

1. **Super Admin** - Platform owner, manages all companies
2. **Company Admin** - Manages jobs and applications for their company
3. **Candidates** - Apply for jobs (public access)

---

## 🛠 Technology Stack

### Backend

- **Language**: Go (Golang)
- **Framework**: Gin Web Framework
- **ORM**: GORM
- **Database**: PostgreSQL (via Supabase)
- **Authentication**: JWT (JSON Web Tokens)
- **File Storage**: Supabase Storage
- **Email Service**: SendGrid (with Resend fallback)
- **CV Parsing**: Local text extraction (PDF/DOCX/DOC/TXT)
- **CV Matching**: Custom keyword-based algorithm

### Frontend

- **Framework**: Next.js 14 (React)
- **Language**: TypeScript
- **Styling**: Tailwind CSS
- **State Management**: Zustand
- **HTTP Client**: Axios
- **Build Tool**: Next.js built-in

### Infrastructure

- **Database**: Supabase (PostgreSQL)
- **File Storage**: Supabase Storage
- **Hosting**: Can be deployed on Vercel (frontend) + any Go hosting (backend)

---

## 🏗 System Architecture

```
┌─────────────────┐
│   Frontend      │  Next.js (TypeScript)
│   (Next.js)     │  └── Admin Dashboard
│                 │  └── Super Admin Dashboard
│                 │  └── Public Job Portal
│                 │  └── Embedded Dashboard
└────────┬────────┘
         │
         │ HTTP/REST API
         │
┌────────▼────────┐
│   Backend       │  Go (Gin Framework)
│   (Golang)      │  └── REST API Endpoints
│                 │  └── Authentication
│                 │  └── Business Logic
│                 │  └── CV Parsing & Matching
└────────┬────────┘
         │
         │ GORM
         │
┌────────▼────────┐
│   Database      │  PostgreSQL (Supabase)
│   (Supabase)    │  └── Companies
│                 │  └── Admins
│                 │  └── Jobs
│                 │  └── Applications
│                 │  └── Activity Logs
└─────────────────┘
```

---

## ✨ Features Implemented

### 1. **Authentication & Authorization**

- ✅ Company registration with admin account creation
- ✅ Admin login with JWT tokens
- ✅ Super admin login (separate system)
- ✅ Token-based authentication for all protected routes
- ✅ Company ID validation in all requests
- ✅ Secure password hashing (bcrypt)

### 2. **Company Management**

- ✅ Company registration
- ✅ Company profile management
- ✅ Embedded mode configuration
- ✅ Subscription status tracking
- ✅ Super admin can view all companies

### 3. **Job Management**

- ✅ Create job postings (title, description, requirements, location, salary, deadline)
- ✅ View all jobs for company
- ✅ Edit existing jobs
- ✅ Delete jobs
- ✅ Job status management (open, closed, archived)
- ✅ Job posting date tracking
- ✅ Public job listings (company-specific)

### 4. **Application Management**

- ✅ Candidates can submit applications
- ✅ File upload support (CV, Portfolio) - URL or file upload
- ✅ View all applications for company's jobs
- ✅ Filter applications by:
  - Job title
  - Status (pending, shortlisted, rejected)
  - Date range
- ✅ Shortlist applications
- ✅ Reject applications
- ✅ View application details
- ✅ CV viewing/downloading
- ✅ Application score tracking (AI matching)

### 5. **AI-Powered CV Analysis**

- ✅ Automatic CV parsing (PDF, DOCX, DOC, TXT)
- ✅ Text extraction from CVs
- ✅ Keyword-based matching algorithm
- ✅ Match score calculation (0-100%)
- ✅ Criteria-based matching:
  - Required skills
  - Minimum experience
  - Required languages
  - Job description matching
- ✅ Score display on applications page
- ✅ Automatic analysis on application submission

### 6. **Email Notifications**

- ✅ Application confirmation emails
- ✅ Shortlist notification emails
- ✅ Rejection notification emails
- ✅ SendGrid integration (with Resend fallback)
- ✅ Email logging

### 7. **File Storage**

- ✅ Supabase Storage integration
- ✅ CV file uploads
- ✅ Portfolio file uploads
- ✅ URL-based file references
- ✅ Secure file access

### 8. **Embedded Dashboard**

- ✅ Full dashboard embedding via iframe
- ✅ Company-specific embed codes
- ✅ Security validation (company_id matching)
- ✅ Embedded login flow
- ✅ Embedded dashboard pages (jobs, applications)
- ✅ Public job portal embedding

### 9. **Activity Logging**

- ✅ Comprehensive activity tracking
- ✅ Logs for:
  - Company registration
  - Job creation/update/deletion
  - Job status changes
  - Application shortlisting/rejection
  - Application status changes
- ✅ Admin view (company-specific logs)
- ✅ Super admin view (all companies)
- ✅ Filtering by action type, entity type, date range
- ✅ Company filtering (super admin)

### 10. **Super Admin Dashboard**

- ✅ Platform statistics
- ✅ View all companies
- ✅ Company statistics (jobs, applications)
- ✅ Activity logs across all companies
- ✅ Separate authentication system

### 11. **Development/Production Modes**

- ✅ Environment-based mode switching
- ✅ Mode indicator on frontend
- ✅ Different configurations per mode

### 12. **Subscription & Payment (Schema Ready)**

- ✅ Database schema for subscriptions
- ✅ Payment tracking
- ✅ Subscription plans structure
- ⚠️ Payment gateway integration (structure ready, not implemented)

---

## 🗄 Database Schema

### Tables

1. **companies**

   - `id` (UUID, Primary Key)
   - `company_name` (VARCHAR)
   - `email` (VARCHAR, Unique)
   - `company_website` (VARCHAR)
   - `embedded_mode` (BOOLEAN)
   - `embed_domain` (VARCHAR)
   - `subscription_status` (VARCHAR)
   - `subscription_tier` (VARCHAR)
   - `created_at`, `updated_at` (TIMESTAMP)

2. **admins**

   - `id` (UUID, Primary Key)
   - `company_id` (UUID, Foreign Key → companies)
   - `name` (VARCHAR)
   - `email` (VARCHAR, Unique)
   - `password_hash` (VARCHAR)
   - `role` (VARCHAR)
   - `created_at` (TIMESTAMP)

3. **jobs**

   - `id` (UUID, Primary Key)
   - `company_id` (UUID, Foreign Key → companies)
   - `title` (VARCHAR)
   - `description` (TEXT)
   - `requirements` (TEXT)
   - `location` (VARCHAR)
   - `job_type` (VARCHAR)
   - `salary_range` (VARCHAR)
   - `deadline` (DATE)
   - `status` (VARCHAR)
   - `auto_shortlist` (BOOLEAN)
   - `shortlist_criteria` (JSONB)
   - `created_at`, `updated_at` (TIMESTAMP)

4. **applications**

   - `id` (UUID, Primary Key)
   - `job_id` (UUID, Foreign Key → jobs)
   - `full_name` (VARCHAR)
   - `email` (VARCHAR)
   - `phone` (VARCHAR)
   - `resume_url` (TEXT)
   - `cover_letter` (TEXT)
   - `years_of_experience` (INT)
   - `current_position` (VARCHAR)
   - `linkedin_url` (VARCHAR)
   - `portfolio_url` (VARCHAR)
   - `status` (VARCHAR)
   - `score` (INT) - AI match score
   - `analysis_result` (JSONB) - AI analysis details
   - `applied_at` (TIMESTAMP)
   - `reviewed_at` (TIMESTAMP)
   - `reviewed_by` (UUID, Foreign Key → admins)

5. **super_admin**

   - `id` (UUID, Primary Key)
   - `name` (VARCHAR)
   - `email` (VARCHAR, Unique)
   - `password_hash` (VARCHAR)
   - `created_at` (TIMESTAMP)

6. **email_logs**

   - `id` (UUID, Primary Key)
   - `application_id` (UUID, Foreign Key → applications)
   - `email_type` (VARCHAR)
   - `sent_to` (VARCHAR)
   - `sent_at` (TIMESTAMP)
   - `status` (VARCHAR)

7. **activity_logs**

   - `id` (UUID, Primary Key)
   - `company_id` (UUID, Foreign Key → companies)
   - `admin_id` (UUID, Foreign Key → admins)
   - `action_type` (VARCHAR)
   - `entity_type` (VARCHAR)
   - `entity_id` (UUID)
   - `description` (TEXT)
   - `metadata` (JSONB)
   - `created_at` (TIMESTAMP)

8. **subscription_plans** (Schema ready)
9. **subscriptions** (Schema ready)
10. **payments** (Schema ready)

---

## 🔌 API Endpoints

### Public Endpoints

- `POST /api/auth/register` - Company registration
- `POST /api/auth/login` - Admin login
- `GET /api/jobs/public/:companyId` - Public job listings
- `POST /api/applications` - Submit application
- `POST /api/upload/cv` - Upload CV file
- `POST /api/upload/portfolio` - Upload portfolio file
- `POST /api/super-admin/login` - Super admin login

### Protected Endpoints (Admin Auth Required)

- `POST /api/jobs` - Create job
- `GET /api/jobs` - Get all jobs (company-specific)
- `GET /api/jobs/:id` - Get single job
- `PUT /api/jobs/:id` - Update job
- `DELETE /api/jobs/:id` - Delete job
- `GET /api/applications` - Get all applications
- `PUT /api/applications/:id/shortlist` - Shortlist application
- `PUT /api/applications/:id/reject` - Reject application
- `POST /api/applications/ai-shortlist` - AI analyze single application
- `POST /api/applications/ai-shortlist-batch` - AI analyze multiple applications
- `GET /api/activity-logs` - Get activity logs (company-specific)

### Super Admin Endpoints

- `GET /api/super-admin/stats` - Platform statistics
- `GET /api/super-admin/companies` - Get all companies
- `GET /api/super-admin/activity-logs` - Get all activity logs

---

## 🎨 Frontend Structure

### Admin Dashboard (`/admin/dashboard`)

- **Dashboard** (`/admin/dashboard`) - Overview with stats
- **Jobs** (`/admin/dashboard/jobs`) - Job management
  - List all jobs
  - Create job (`/admin/dashboard/jobs/create`)
  - Edit job (`/admin/dashboard/jobs/[id]/edit`)
- **Applications** (`/admin/dashboard/applications`) - Application management
- **Embed Code** (`/admin/dashboard/embed`) - Get embed codes
- **Activity Logs** (`/admin/dashboard/activity-logs`) - View activity logs

### Super Admin Dashboard (`/super-admin/dashboard`)

- **Dashboard** (`/super-admin/dashboard`) - Platform overview
- **Companies** (`/super-admin/dashboard/companies`) - All companies
- **Activity Logs** (`/super-admin/dashboard/activity-logs`) - All activity logs

### Public Pages

- **Job Portal** (`/jobs/[companyId]`) - Public job listings
- **Application Form** - Submit application

### Embedded Pages (`/embed/*`)

- **Login** (`/embed/login`) - Embedded login
- **Dashboard** (`/embed/dashboard`) - Embedded dashboard
- **Jobs** (`/embed/dashboard/jobs`) - Embedded jobs management
- **Applications** (`/embed/dashboard/applications`) - Embedded applications

---

## 🔒 Security Features

### Authentication

- ✅ JWT token-based authentication
- ✅ Secure password hashing (bcrypt)
- ✅ Token expiration (24 hours)
- ✅ Separate authentication for super admin

### Authorization

- ✅ Company ID validation on all requests
- ✅ Admin can only access their company's data
- ✅ Super admin can access all data
- ✅ Embed code validation (company_id matching)

### Data Protection

- ✅ SQL injection prevention (GORM parameterized queries)
- ✅ XSS protection (React automatic escaping)
- ✅ CORS configuration
- ✅ Input validation on all endpoints
- ✅ File upload validation

### Embed Security

- ✅ Company-specific embed codes
- ✅ URL parameter validation (company_id)
- ✅ Login validation against company_id
- ✅ Cross-company access prevention

---

## 🚀 Deployment & Setup

### Environment Variables

#### Backend (`.env`)

```env
DATABASE_URL=postgresql://user:password@host:port/database
JWT_SECRET=your-secret-key-change-in-production
RESEND_API_KEY=your-resend-api-key
SENDGRID_API_KEY=your-sendgrid-api-key
SENDGRID_FROM_EMAIL=your-email@domain.com
SUPABASE_URL=your-supabase-url
SUPABASE_SERVICE_ROLE_KEY=your-service-role-key
APP_MODE=development # or production
```

#### Frontend (`.env.local`)

```env
NEXT_PUBLIC_API_URL=http://localhost:8080/api
NEXT_PUBLIC_FRONTEND_URL=http://localhost:3000
NEXT_PUBLIC_APP_MODE=development
```

### Setup Steps

1. **Database Setup**

   - Run SQL schema from `supabase/schema.sql` in Supabase
   - Configure connection pooling (port 6543)

2. **Backend Setup**

   ```bash
   cd backend
   go mod download
   go run main.go
   ```

3. **Frontend Setup**

   ```bash
   cd frontend
   npm install
   npm run dev
   ```

4. **Supabase Storage**
   - Create `cvs` and `portfolios` buckets
   - Configure public access if needed

---

## 📁 File Structure

```
ats-system/
├── backend/
│   ├── config/
│   │   ├── database.go          # Database connection & migration
│   │   └── config.go            # Configuration helpers
│   ├── controllers/
│   │   ├── auth_controller.go  # Authentication endpoints
│   │   ├── job_controller.go    # Job management
│   │   ├── application_controller.go  # Application management
│   │   ├── ai_controller.go     # AI CV analysis
│   │   ├── activity_log_controller.go  # Activity logs
│   │   └── super_admin_controller.go   # Super admin endpoints
│   ├── middleware/
│   │   ├── auth.go              # JWT authentication middleware
│   │   └── super_admin_auth.go  # Super admin auth middleware
│   ├── models/
│   │   ├── company.go
│   │   ├── admin.go
│   │   ├── job.go
│   │   ├── application.go
│   │   ├── super_admin.go
│   │   ├── activity_log.go
│   │   └── subscription.go
│   ├── services/
│   │   ├── email_service.go     # Email sending (SendGrid/Resend)
│   │   ├── storage_service.go   # File uploads (Supabase)
│   │   ├── cv_matcher.go        # CV parsing & matching
│   │   ├── activity_logger.go   # Activity logging
│   │   └── payment_service.go   # Payment processing (structure)
│   ├── routes/
│   │   └── routes.go            # API route definitions
│   ├── utils/
│   │   └── jwt.go               # JWT token generation/verification
│   ├── main.go                  # Application entry point
│   └── go.mod                   # Go dependencies
│
├── frontend/
│   ├── app/
│   │   ├── admin/
│   │   │   ├── login/           # Admin login page
│   │   │   ├── register/        # Company registration
│   │   │   └── dashboard/       # Admin dashboard
│   │   ├── super-admin/
│   │   │   ├── login/           # Super admin login
│   │   │   └── dashboard/       # Super admin dashboard
│   │   ├── jobs/
│   │   │   └── [companyId]/     # Public job portal
│   │   ├── embed/               # Embedded dashboard pages
│   │   └── page.tsx             # Landing page
│   ├── lib/
│   │   ├── api.ts               # API client functions
│   │   └── store.ts             # Zustand state management
│   ├── public/                  # Static assets
│   └── package.json
│
├── supabase/
│   ├── schema.sql               # Database schema
│   └── README.md
│
└── Documentation/
    ├── EMBED_SECURITY_EXPLANATION.md
    └── PROJECT_SUMMARY.md (this file)
```

---

## 🔑 Key Components

### Backend Components

1. **Authentication System**

   - JWT token generation/verification
   - Password hashing (bcrypt)
   - Middleware for route protection
   - Company ID validation

2. **CV Analysis System**

   - Text extraction from PDF/DOCX/DOC/TXT
   - Keyword-based matching
   - Score calculation (0-100%)
   - Criteria-based filtering

3. **Activity Logging**

   - Automatic logging on all actions
   - Metadata storage (JSONB)
   - Filtering and querying
   - Admin and super admin views

4. **File Storage**
   - Supabase Storage integration
   - CV and portfolio uploads
   - URL and file-based uploads
   - Secure file access

### Frontend Components

1. **State Management (Zustand)**

   - Auth store (admin authentication)
   - Super admin store
   - Token management
   - User data storage

2. **API Client**

   - Axios-based HTTP client
   - Automatic token injection
   - Error handling
   - Type-safe API calls

3. **Dashboard Components**
   - Stats cards
   - Data tables
   - Filters
   - Forms
   - Modals

---

## 📊 Statistics & Metrics

### Current Implementation Status

- ✅ **100%** - Core features implemented
- ✅ **100%** - Authentication & authorization
- ✅ **100%** - Job management
- ✅ **100%** - Application management
- ✅ **100%** - AI CV analysis
- ✅ **100%** - Activity logging
- ✅ **100%** - Embedded dashboard
- ✅ **100%** - Email notifications
- ⚠️ **50%** - Payment integration (schema ready, gateway not implemented)

### Database Tables

- **10 tables** created
- **Indexes** optimized for performance
- **Foreign keys** properly configured
- **JSONB** fields for flexible data storage

### API Endpoints

- **20+** REST API endpoints
- **Public** and **protected** routes
- **Super admin** specific endpoints
- **File upload** endpoints

---

## 🎯 Future Enhancements (Not Implemented)

1. **Payment Integration**

   - Stripe integration
   - PayPal integration
   - Pakistani payment methods (EasyPaisa, JazzCash)
   - Subscription management UI

2. **Advanced Features**

   - Interview scheduling
   - Email templates customization
   - Advanced analytics
   - Export functionality (PDF, Excel)
   - Multi-language support
   - Mobile app

3. **Performance**
   - Caching layer
   - CDN for static assets
   - Database query optimization
   - Image optimization

---

## 📝 Notes

- **Connection Pooling**: Uses Supabase connection pooling port (6543)
- **Prepared Statements**: Disabled for Supabase compatibility
- **Date Handling**: Custom DateOnly type for date-only fields
- **Error Handling**: Comprehensive error messages throughout
- **Logging**: Detailed logging for debugging
- **Security**: Company ID validation on all operations

---

## 🎉 Project Status

**Status**: ✅ **Production Ready** (Core Features)

The ATS system is fully functional with all core features implemented. The system is ready for deployment and use, with comprehensive security, logging, and user management features.

---

**Last Updated**: 2024
**Version**: 1.0.0
