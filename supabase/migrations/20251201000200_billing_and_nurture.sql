-- Tables for planned features: subscription billing, talent-pool nurture emails and email delivery logs.
-- The Go models exist; handlers will start writing to these once the features ship.

create table subscription_plans (
    id               uuid primary key default gen_random_uuid(),
    name             varchar(100) not null unique,
    price            numeric(10, 2) not null check (price >= 0),
    currency         varchar(10) not null default 'USD',
    features         jsonb not null default '[]',
    max_jobs         integer not null default 5,
    max_applications integer not null default 100,
    ai_shortlisting  boolean not null default false,
    created_at       timestamptz not null default now(),
    updated_at       timestamptz not null default now()
);

create table subscriptions (
    id                     uuid primary key default gen_random_uuid(),
    company_id             uuid not null references companies(id) on delete cascade,
    plan_id                uuid not null references subscription_plans(id),
    status                 varchar(50) not null default 'active'
                           check (status in ('trial', 'active', 'cancelled', 'expired')),
    current_period_start   timestamptz not null,
    current_period_end     timestamptz not null,
    cancel_at_period_end   boolean not null default false,
    stripe_subscription_id varchar(255),
    paypal_subscription_id varchar(255),
    created_at             timestamptz not null default now(),
    updated_at             timestamptz not null default now()
);

create table payments (
    id                 uuid primary key default gen_random_uuid(),
    company_id         uuid not null references companies(id) on delete cascade,
    subscription_id    uuid references subscriptions(id) on delete set null,
    amount             numeric(10, 2) not null check (amount >= 0),
    currency           varchar(10) not null default 'USD',
    status             varchar(50) not null default 'pending'
                       check (status in ('pending', 'completed', 'failed', 'refunded')),
    payment_method     varchar(50) not null,
    payment_gateway_id varchar(255),
    transaction_id     varchar(255) not null unique,
    metadata           jsonb,
    created_at         timestamptz not null default now(),
    updated_at         timestamptz not null default now()
);

create table nurture_campaigns (
    id             uuid primary key default gen_random_uuid(),
    application_id uuid not null references applications(id) on delete cascade,
    job_id         uuid references jobs(id) on delete set null,
    email_sent_at  timestamptz not null default now(),
    email_type     varchar(50) not null check (email_type in ('job_alert', 'check_in', 'opportunity')),
    subject        varchar(255),
    status         varchar(50) not null default 'sent' check (status in ('sent', 'opened', 'clicked', 'bounced')),
    created_at     timestamptz not null default now()
);

create table nurture_preferences (
    id                uuid primary key default gen_random_uuid(),
    application_id    uuid not null references applications(id) on delete cascade,
    email             varchar(255) not null,
    preferences       jsonb,
    is_active         boolean not null default true,
    last_contacted_at timestamptz,
    created_at        timestamptz not null default now(),
    updated_at        timestamptz not null default now(),
    unique (application_id, email)
);

create table email_logs (
    id             uuid primary key default gen_random_uuid(),
    application_id uuid references applications(id) on delete cascade,
    email_type     varchar(50),
    sent_to        varchar(255),
    sent_at        timestamptz not null default now(),
    status         varchar(50)
);

create index idx_subscriptions_company on subscriptions(company_id, status);
create index idx_payments_company on payments(company_id, created_at desc);
create index idx_nurture_campaigns_application on nurture_campaigns(application_id, email_type, email_sent_at desc);
create index idx_nurture_preferences_active on nurture_preferences(email) where is_active;
create index idx_email_logs_application on email_logs(application_id);

create trigger subscription_plans_updated_at before update on subscription_plans
    for each row execute function set_updated_at();
create trigger subscriptions_updated_at before update on subscriptions
    for each row execute function set_updated_at();
create trigger payments_updated_at before update on payments
    for each row execute function set_updated_at();
create trigger nurture_preferences_updated_at before update on nurture_preferences
    for each row execute function set_updated_at();

alter table subscription_plans  enable row level security;
alter table subscriptions       enable row level security;
alter table payments            enable row level security;
alter table nurture_campaigns   enable row level security;
alter table nurture_preferences enable row level security;
alter table email_logs          enable row level security;

-- Plans advertised on the landing page.
insert into subscription_plans (name, price, features, max_jobs, max_applications, ai_shortlisting)
values
    ('Starter', 29,
     '["5 active job postings", "200 applications/month", "Basic auto-shortlisting", "Email automation", "1 admin account"]',
     5, 200, false),
    ('Professional', 79,
     '["20 active job postings", "1000 applications/month", "Advanced AI shortlisting", "3 admin accounts", "Priority support", "Custom branding"]',
     20, 1000, true),
    ('Enterprise', 199,
     '["Unlimited job postings", "Unlimited applications", "10 admin accounts", "API access", "Dedicated support", "White-label option"]',
     1000000, 1000000, true);
