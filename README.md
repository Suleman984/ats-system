# ATS System

A multi-tenant applicant tracking system. Companies post jobs, candidates apply through a public careers page, and every CV is parsed and scored against the job's criteria so recruiters can shortlist faster.

**Backend:** Go · Gin · GORM · PostgreSQL (Supabase) &nbsp;|&nbsp; **Frontend:** Next.js 16 · React 19 · TypeScript · Tailwind CSS 4 · Zustand

## Features

- **Company workspaces**: each company registers its own account; data is scoped by `company_id` on every query.
- **Jobs**: create, edit, close and archive postings. Expired jobs close automatically at their deadline.
- **Public careers page and apply form**: candidates upload a CV (PDF/DOC/DOCX) or paste a link.
- **CV parsing and scoring**: text is extracted from PDF and DOCX files and scored on skills (with synonyms), experience, languages and job-description overlap. Scoring is local keyword matching; no external AI service is called.
- **Candidate search**: search every CV your company has received, including candidates from deleted jobs.
- **CRM**: private or shared recruiter notes, a talent pool, referral tracking and a per-candidate timeline.
- **Candidate portal**: applicants check their status with their email and application ID.
- **Notifications**: status emails via SendGrid or Resend, with optional Twilio SMS.
- **Embeddable dashboard**: an iframe-ready view of the recruiter dashboard for company websites.
- **Super-admin console**: platform-wide stats, companies and an audit log of every action.

## Architecture

```mermaid
flowchart LR
    subgraph Browser
        A[Recruiter dashboard]
        B[Careers page / apply form]
        C[Candidate portal]
    end
    A & B & C -->|REST + JWT| API[Go API · Gin]
    API -->|GORM| DB[(Supabase Postgres)]
    API -->|service-role key| S[(Supabase Storage)]
    API -.async.-> CV[CV parser + scorer]
    CV -->|SSRF-safe download| S
    API -.async.-> N[SendGrid / Resend / Twilio]
```

The API is the only database client. It connects as the `postgres` role, and Row Level Security is enabled on every table with no policies, so the public Supabase keys cannot read or write anything.

## Security

- Company-admin and super-admin JWTs carry different audiences, so one can never be used as the other. `JWT_SECRET` is required and must be at least 32 characters.
- Every recruiter endpoint resolves records through `company_id`, so one company can't read or modify another company's data.
- Public endpoints accept only the fields an applicant may set; scores and statuses can't be injected.
- CV links are downloaded through a client that refuses private, loopback and link-local addresses (checked at connect time), with a 10 MB cap and timeouts.
- Uploads are checked by extension and file signature, then stored under random UUID names.
- Login, registration and public endpoints are rate-limited per IP. CORS is restricted to `ALLOWED_ORIGINS`.
- Every value placed in an email template is HTML-escaped.

## Getting started

### Prerequisites

Go 1.26+, Node.js 20.9+, the [Supabase CLI](https://supabase.com/docs/guides/cli), and a Supabase project (or Docker for a local one).

### 1. Database

The schema lives in [`supabase/migrations`](supabase/migrations), and [`supabase/seed.sql`](supabase/seed.sql) holds demo data.

**Hosted Supabase**

```bash
supabase link --project-ref <your-project-ref>
supabase db push            # applies the migrations, including the storage buckets
```

**Local Supabase (Docker)**

```bash
supabase start
supabase db reset           # applies the migrations, then seed.sql
```

The seed creates a demo company (`admin@acme.test`) and a super admin (`owner@ats.test`), both with the password `demo12345`.

**Existing databases** (created by the old GORM auto-migration): mark the base migrations as applied, then push. `20251201000300_align_legacy_schema.sql` fixes constraints, types and RLS in place without touching data.

```bash
supabase migration repair --status applied 20251201000000 20251201000100 20251201000200
supabase db push
```

Uploads go through the API with `SUPABASE_SERVICE_ROLE_KEY`, so once that backend is deployed, the old anonymous storage policies can be dropped:

```sql
drop policy if exists "Allow public uploads to resumes" on storage.objects;
drop policy if exists "Allow public uploads to portfolios" on storage.objects;
drop policy if exists "Allow public reads from resumes" on storage.objects;
drop policy if exists "Allow public reads from portfolios" on storage.objects;
```

### 2. Backend

```bash
cd backend
cp .env.example .env        # fill in DATABASE_URL, JWT_SECRET, SUPABASE_URL, SUPABASE_SERVICE_ROLE_KEY
go run .
```

The API listens on `http://localhost:8080`; `GET /health` returns `{"status":"ok"}`.

### 3. Frontend

```bash
cd frontend
cp .env.local.example .env.local
npm install
npm run dev
```

Open `http://localhost:3000`.

## Environment variables

| Variable | Required | Description |
|---|---|---|
| `DATABASE_URL` | yes | Supabase Postgres connection string (session pooler) |
| `JWT_SECRET` | yes | At least 32 random characters |
| `SUPABASE_URL` | for uploads | `https://<ref>.supabase.co` |
| `SUPABASE_SERVICE_ROLE_KEY` | for uploads | Used server-side only, to write to Storage |
| `FRONTEND_URL` | no | Used in email links. Default `http://localhost:3000` |
| `ALLOWED_ORIGINS` | no | Comma-separated CORS origins. Defaults to `FRONTEND_URL` |
| `TRUSTED_PROXIES` | no | Proxy IPs/CIDRs whose `X-Forwarded-For` is trusted |
| `APP_MODE` | no | `production` enables release mode and quieter logs |
| `EMAIL_PROVIDER` | no | `sendgrid` (default) or `resend` |
| `SENDGRID_API_KEY`, `SENDGRID_FROM_EMAIL` | for SendGrid | |
| `RESEND_API_KEY`, `RESEND_FROM_EMAIL` | for Resend | |
| `SMS_PROVIDER`, `TWILIO_*` | no | Set `SMS_PROVIDER=twilio` to enable SMS |

The frontend reads `NEXT_PUBLIC_API_URL`, `NEXT_PUBLIC_FRONTEND_URL` and `NEXT_PUBLIC_APP_MODE`.

## API overview

| Area | Endpoints |
|---|---|
| Auth | `POST /api/auth/register`, `POST /api/auth/login`, `POST /api/super-admin/login` |
| Public | `GET /api/jobs/public/:companyId`, `GET /api/public/jobs/:id`, `POST /api/applications`, `POST /api/upload/cv`, `POST /api/upload/portfolio` |
| Candidate portal | `POST /api/candidate/status`, `GET /api/candidate/messages`, `POST /api/candidate/messages/send` |
| Jobs | `GET/POST /api/jobs`, `GET/PUT/DELETE /api/jobs/:id` |
| Applications | `GET /api/applications`, `PUT /api/applications/:id/{shortlist,reject}`, `POST /api/applications/:id/track-cv-view`, `DELETE /api/applications/:id`, `POST /api/applications/bulk-delete`, `GET/POST /api/applications/:id/messages` |
| Scoring | `POST /api/applications/ai-shortlist`, `POST /api/applications/ai-shortlist-batch` |
| Candidates | `POST /api/candidates/search`, `GET /api/candidates/:id`, `POST /api/candidates/manual`, `POST /api/candidates/reparse-all`, `POST /api/candidates/:id/reparse` |
| CRM | `/api/crm/notes`, `/api/crm/talent-pool`, `/api/crm/applications/:id/{notes,referral,timeline}` |
| Super admin | `GET /api/super-admin/{stats,companies,activity-logs}` |
| Activity | `GET /api/activity-logs` |

## Tests

```bash
cd backend && go test ./...
cd frontend && npm run lint && npm run typecheck && npm run build
```

The backend tests cover token audience separation, SSRF blocking, CV parsing and scoring, and rate limiting. CI runs all of the above on every push.

## Project structure

```
backend/
  config/       env and database setup
  controllers/  HTTP handlers
  middleware/   auth and rate limiting
  models/       GORM models
  routes/       route table
  services/     CV parsing, storage, email, SMS, audit log
  utils/        JWT
frontend/
  app/          Next.js App Router pages (admin, embed, super-admin, public)
  components/   shared UI (job form, application modals, toasts)
  lib/          typed API client and auth store
supabase/
  migrations/   schema, indexes, RLS and storage buckets
  seed.sql      demo data
```

## License

[MIT](LICENSE)
