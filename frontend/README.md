# ATS Frontend - Next.js Application with TypeScript

## Setup Instructions

### Prerequisites

- Node.js 18+
- npm or yarn

### Local Development

1. **Install dependencies**

```bash
npm install
```

2. **Create .env.local file**

```bash
cp .env.local.example .env.local
# Edit .env.local with your API URL
NEXT_PUBLIC_API_URL=http://localhost:8080/api
```

3. **Run development server**

```bash
npm run dev
```

Open http://localhost:3000

### Build for Production

```bash
npm run build
npm start
```

### Deployment

Deploy to Vercel:

1. Push to GitHub
2. Import project in Vercel
3. Add environment variables
4. Deploy

### Project Structure

```
frontend/
├── app/                    # Next.js App Router pages
│   ├── admin/             # Admin pages
│   ├── jobs/              # Public job listings
│   └── apply/             # Application form
├── lib/                   # Utilities and API client
│   ├── api.ts            # API client with TypeScript types
│   └── store.ts          # Zustand state management
└── components/           # Reusable components (if any)
```
