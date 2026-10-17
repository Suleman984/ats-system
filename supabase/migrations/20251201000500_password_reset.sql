-- Password reset: single-use tokens (only the SHA-256 hash is stored) and a timestamp that
-- invalidates JWTs issued before the last password change.
alter table admins add column password_changed_at timestamptz;

create table password_reset_tokens (
    id         uuid primary key default gen_random_uuid(),
    admin_id   uuid not null references admins(id) on delete cascade,
    token_hash varchar(64) not null unique,
    expires_at timestamptz not null,
    used_at    timestamptz,
    created_at timestamptz not null default now()
);

create index idx_password_reset_tokens_admin on password_reset_tokens(admin_id);

-- Only the backend (which bypasses RLS) may touch reset tokens.
alter table password_reset_tokens enable row level security;
