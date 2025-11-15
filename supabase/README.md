# Supabase Database Setup

## Instructions

1. Go to your Supabase project dashboard
2. Navigate to SQL Editor
3. Copy and paste the contents of `schema.sql`
4. Click "Run" to execute the SQL
5. Verify tables are created in the Table Editor

## Database Connection

After creating the tables, get your database connection string from:

- Supabase Dashboard → Settings → Database → Connection String
- Use the "URI" format for your `DATABASE_URL` environment variable

Format: `postgresql://postgres:[YOUR-PASSWORD]@db.[PROJECT-REF].supabase.co:5432/postgres`

## Tables Created

- `companies` - Company information
- `admins` - Admin users for each company
- `jobs` - Job postings
- `applications` - Job applications
- `super_admin` - Super admin accounts
- `email_logs` - Email sending logs
