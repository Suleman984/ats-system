-- Core ATS schema. The Go API connects as the postgres role, which bypasses RLS;
-- RLS is enabled with no policies so the public anon/authenticated roles can't reach any table.

create extension if not exists pgcrypto;

create table companies (
    id                  uuid primary key default gen_random_uuid(),
    company_name        varchar(255) not null,
    email               varchar(255) not null unique,
    company_website     varchar(255),
    embedded_mode       boolean not null default false,
    embed_domain        varchar(255),
    subscription_status varchar(50) not null default 'trial',
    subscription_tier   varchar(50) not null default 'starter',
    created_at          timestamptz not null default now(),
    updated_at          timestamptz not null default now()
);

create table admins (
    id            uuid primary key default gen_random_uuid(),
    company_id    uuid not null references companies(id) on delete cascade,
    name          varchar(255) not null,
    email         varchar(255) not null unique,
    password_hash varchar(255) not null,
    role          varchar(50) not null default 'admin',
    created_at    timestamptz not null default now()
);

create table super_admins (
    id            uuid primary key default gen_random_uuid(),
    name          varchar(255) not null,
    email         varchar(255) not null unique,
    password_hash varchar(255) not null,
    created_at    timestamptz not null default now()
);

create table jobs (
    id                 uuid primary key default gen_random_uuid(),
    company_id         uuid not null references companies(id) on delete cascade,
    title              varchar(255) not null,
    description        text not null,
    requirements       text,
    location           varchar(255),
    job_type           varchar(50),
    salary_range       varchar(100),
    deadline           date not null,
    status             varchar(50) not null default 'open' check (status in ('open', 'closed', 'archived')),
    auto_shortlist     boolean not null default true,
    shortlist_criteria jsonb,
    created_at         timestamptz not null default now(),
    updated_at         timestamptz not null default now()
);

-- job_id is set null when a job is deleted so candidates stay searchable in Find Candidates.
create table applications (
    id                     uuid primary key default gen_random_uuid(),
    job_id                 uuid references jobs(id) on delete set null,
    company_id             uuid not null references companies(id) on delete cascade,
    full_name              varchar(255) not null,
    email                  varchar(255) not null,
    phone                  varchar(50),
    resume_url             text not null,
    cover_letter           text,
    years_of_experience    integer not null default 0,
    current_position       varchar(255),
    linkedin_url           varchar(255),
    portfolio_url          varchar(255),
    status                 varchar(50) not null default 'pending'
                           check (status in ('pending', 'cv_viewed', 'shortlisted', 'rejected')),
    score                  integer not null default 0 check (score between 0 and 100),
    analysis_result        jsonb,
    parsed_cv_text         text,
    applied_at             timestamptz not null default now(),
    reviewed_at            timestamptz,
    reviewed_by            uuid references admins(id) on delete set null,
    cv_viewed_at           timestamptz,
    cv_viewed_by           uuid references admins(id) on delete set null,
    expected_response_date date,
    last_status_update     timestamptz,
    referral_source        varchar(255),
    referred_by_name       varchar(255),
    referred_by_email      varchar(255),
    referred_by_phone      varchar(50),
    in_talent_pool         boolean not null default false,
    talent_pool_added_at   timestamptz,
    talent_pool_added_by   uuid references admins(id) on delete set null
);

create table candidate_notes (
    id             uuid primary key default gen_random_uuid(),
    application_id uuid not null references applications(id) on delete cascade,
    admin_id       uuid not null references admins(id) on delete cascade,
    note           text not null,
    is_private     boolean not null default false,
    created_at     timestamptz not null default now(),
    updated_at     timestamptz not null default now()
);

create table messages (
    id             uuid primary key default gen_random_uuid(),
    application_id uuid not null references applications(id) on delete cascade,
    sender_type    varchar(20) not null check (sender_type in ('candidate', 'recruiter')),
    sender_id      uuid references admins(id) on delete set null,
    sender_email   varchar(255) not null,
    message        text not null,
    is_read        boolean not null default false,
    read_at        timestamptz,
    created_at     timestamptz not null default now()
);

-- company_id and admin_id are null for platform-level and system actions.
create table activity_logs (
    id          uuid primary key default gen_random_uuid(),
    company_id  uuid references companies(id) on delete cascade,
    admin_id    uuid references admins(id) on delete set null,
    action_type varchar(50) not null,
    entity_type varchar(50) not null,
    entity_id   uuid,
    description text,
    metadata    jsonb,
    created_at  timestamptz not null default now()
);

create index idx_admins_company_id on admins(company_id);
create index idx_jobs_company_status on jobs(company_id, status);
create index idx_applications_job_id on applications(job_id);
create index idx_applications_company_status on applications(company_id, status);
create index idx_applications_company_applied on applications(company_id, applied_at desc);
create index idx_applications_email on applications(email);
create index idx_applications_talent_pool on applications(company_id) where in_talent_pool;
create index idx_candidate_notes_application on candidate_notes(application_id, created_at desc);
create index idx_messages_application on messages(application_id, created_at);
create index idx_activity_logs_company_created on activity_logs(company_id, created_at desc);
create index idx_activity_logs_entity on activity_logs(entity_type, entity_id);

create or replace function set_updated_at()
returns trigger
language plpgsql
set search_path = ''
as $$
begin
    new.updated_at = now();
    return new;
end;
$$;

create trigger companies_updated_at before update on companies
    for each row execute function set_updated_at();
create trigger jobs_updated_at before update on jobs
    for each row execute function set_updated_at();
create trigger candidate_notes_updated_at before update on candidate_notes
    for each row execute function set_updated_at();

alter table companies       enable row level security;
alter table admins          enable row level security;
alter table super_admins    enable row level security;
alter table jobs            enable row level security;
alter table applications    enable row level security;
alter table candidate_notes enable row level security;
alter table messages        enable row level security;
alter table activity_logs   enable row level security;
