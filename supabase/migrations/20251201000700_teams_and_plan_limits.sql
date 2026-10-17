-- Team accounts: every existing admin becomes their company's owner; invited teammates join as
-- owner or recruiter. Plans gain an admin-seat cap and a PKR price for Safepay checkout.
update admins set role = 'owner' where role = 'admin';
alter table admins alter column role set default 'owner';
alter table admins add constraint admins_role_check check (role in ('owner', 'recruiter'));

alter table subscription_plans add column max_admins integer not null default 1;
-- Placeholder PKR prices; set real ones before going live.
alter table subscription_plans add column price_pkr integer not null default 0 check (price_pkr >= 0);
update subscription_plans set max_admins = 1,  price_pkr = 8000  where name = 'Starter';
update subscription_plans set max_admins = 3,  price_pkr = 22000 where name = 'Professional';
update subscription_plans set max_admins = 10, price_pkr = 55000 where name = 'Enterprise';

create table team_invites (
    id          uuid primary key default gen_random_uuid(),
    company_id  uuid not null references companies(id) on delete cascade,
    email       varchar(255) not null,
    role        varchar(50) not null check (role in ('owner', 'recruiter')),
    token_hash  varchar(64) not null unique,
    invited_by  uuid references admins(id) on delete set null,
    expires_at  timestamptz not null,
    accepted_at timestamptz,
    created_at  timestamptz not null default now()
);

create index idx_team_invites_company on team_invites(company_id);
-- One open invite per email per company.
create unique index uq_team_invites_pending on team_invites(company_id, lower(email)) where accepted_at is null;

alter table team_invites enable row level security;
