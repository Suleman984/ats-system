# ATS Backend - Golang API

## Setup Instructions

### Prerequisites

- Go 1.21+
- PostgreSQL database (Supabase account)

### Local Development

1. **Clone the repository**

```bash
cd backend
```

2. **Install dependencies**

```bash
go mod download
```

3. **Create .env file**

```bash
cp .env.example .env
# Edit .env with your credentials
```

4. **Run the server**

```bash
go run main.go
```

Server will start on http://localhost:8080

### API Endpoints

#### Authentication

- `POST /api/auth/register` - Register new company
- `POST /api/auth/login` - Admin login

#### Jobs (Protected)

- `POST /api/jobs` - Create job
- `GET /api/jobs` - Get all jobs for company
- `GET /api/jobs/:companyId/public` - Get public jobs
- `PUT /api/jobs/:id` - Update job
- `DELETE /api/jobs/:id` - Delete job

#### Applications

- `POST /api/applications` - Submit application (public)
- `GET /api/applications` - Get all applications (protected)
- `PUT /api/applications/:id/shortlist` - Shortlist candidate
- `PUT /api/applications/:id/reject` - Reject candidate

### Deployment

Deploy to Render.com:

1. Connect GitHub repository
2. Add environment variables
3. Deploy
