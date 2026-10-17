-- CVs and portfolios move to private buckets. Files are referenced by object path and served only
-- through short-lived signed URLs the API issues after checking the recruiter's company.
update storage.buckets set public = false where id in ('resumes', 'portfolios');

alter table applications add column resume_path text, add column portfolio_path text;
alter table applications alter column resume_url set default '';

-- Earlier uploads were stored as public URLs; keep the object path and drop the now-dead URL.
update applications
set resume_path = substring(resume_url from '/storage/v1/object/public/resumes/(.+)$'), resume_url = ''
where resume_url ~ '/storage/v1/object/public/resumes/.+$';
update applications
set portfolio_path = substring(portfolio_url from '/storage/v1/object/public/portfolios/(.+)$'), portfolio_url = null
where portfolio_url ~ '/storage/v1/object/public/portfolios/.+$';

alter table applications add constraint applications_resume_present
    check (resume_url <> '' or resume_path is not null);

-- Each upload is claimed by at most one application; unclaimed ones are purged after 24 hours.
create table uploads (
    id            uuid primary key default gen_random_uuid(),
    kind          varchar(20) not null check (kind in ('resume', 'portfolio')),
    object_path   text not null unique,
    content_type  varchar(100) not null,
    size_bytes    integer not null check (size_bytes > 0),
    original_name varchar(255) not null default '',
    parsed_text   text,
    claimed_at    timestamptz,
    created_at    timestamptz not null default now()
);
create index idx_uploads_unclaimed on uploads(created_at) where claimed_at is null;
alter table uploads enable row level security;

-- Email verification: existing admins are grandfathered in as verified.
alter table admins add column email_verified_at timestamptz;
update admins set email_verified_at = now();

create table email_verification_tokens (
    id         uuid primary key default gen_random_uuid(),
    admin_id   uuid not null references admins(id) on delete cascade,
    token_hash varchar(64) not null unique,
    expires_at timestamptz not null,
    used_at    timestamptz,
    created_at timestamptz not null default now()
);
create index idx_email_verification_tokens_admin on email_verification_tokens(admin_id);
alter table email_verification_tokens enable row level security;
