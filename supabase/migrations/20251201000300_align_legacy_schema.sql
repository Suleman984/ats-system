-- Brings databases created by the old GORM AutoMigrate and hand-run SQL in line with the migrations above.
-- Every statement is idempotent, so on a database built from these migrations it changes nothing.

-- GORM added duplicate foreign keys with no ON DELETE action, which block job and company deletes.
alter table applications  drop constraint if exists fk_jobs_applications;
alter table admins        drop constraint if exists fk_admins_company;
alter table activity_logs drop constraint if exists fk_activity_logs_admin;
alter table activity_logs drop constraint if exists fk_activity_logs_company;
alter table payments      drop constraint if exists fk_payments_company;
alter table subscriptions drop constraint if exists fk_subscriptions_company;
alter table subscriptions drop constraint if exists fk_subscriptions_plan;

do $$
declare
    fk record;
begin
    for fk in select * from (values
        ('applications',  'applications_reviewed_by_fkey',          'reviewed_by',          'admins(id) on delete set null'),
        ('applications',  'applications_cv_viewed_by_fkey',         'cv_viewed_by',         'admins(id) on delete set null'),
        ('applications',  'applications_talent_pool_added_by_fkey', 'talent_pool_added_by', 'admins(id) on delete set null'),
        ('messages',      'messages_sender_id_fkey',                'sender_id',            'admins(id) on delete set null'),
        ('payments',      'payments_company_id_fkey',               'company_id',           'companies(id) on delete cascade'),
        ('payments',      'payments_subscription_id_fkey',          'subscription_id',      'subscriptions(id) on delete set null'),
        ('subscriptions', 'subscriptions_company_id_fkey',          'company_id',           'companies(id) on delete cascade'),
        ('subscriptions', 'subscriptions_plan_id_fkey',             'plan_id',              'subscription_plans(id)')
    ) as t(tbl, name, col, target)
    loop
        execute format('alter table %I drop constraint if exists %I', fk.tbl, fk.name);
        execute format('alter table %I add constraint %I foreign key (%I) references %s', fk.tbl, fk.name, fk.col, fk.target);
    end loop;
end $$;

-- GORM named this column pay_pal_subscription_id; some databases also have the hand-written one.
do $$
begin
    if exists (select 1 from information_schema.columns
               where table_schema = 'public' and table_name = 'subscriptions' and column_name = 'pay_pal_subscription_id') then
        if exists (select 1 from information_schema.columns
                   where table_schema = 'public' and table_name = 'subscriptions' and column_name = 'paypal_subscription_id') then
            update subscriptions set paypal_subscription_id = coalesce(paypal_subscription_id, pay_pal_subscription_id);
            alter table subscriptions drop column pay_pal_subscription_id;
        else
            alter table subscriptions rename column pay_pal_subscription_id to paypal_subscription_id;
        end if;
    end if;
end $$;

-- Store every timestamp with its time zone; existing values were written in UTC.
do $$
declare
    c record;
begin
    for c in select table_name, column_name from information_schema.columns
             where table_schema = 'public' and data_type = 'timestamp without time zone'
    loop
        execute format('alter table %I alter column %I type timestamptz using %I at time zone ''UTC''',
                       c.table_name, c.column_name, c.column_name);
    end loop;
end $$;

alter table applications       alter column years_of_experience type integer;
alter table subscription_plans alter column max_jobs type integer;
alter table subscription_plans alter column max_applications type integer;

-- Backfill NULLs, then enforce the defaults and NOT NULL constraints the code relies on.
do $$
declare
    c record;
begin
    for c in select * from (values
        ('activity_logs', 'created_at', 'now()'),
        ('admins', 'company_id', null), ('admins', 'created_at', 'now()'), ('admins', 'role', '''admin'''),
        ('applications', 'applied_at', 'now()'), ('applications', 'in_talent_pool', 'false'),
        ('applications', 'score', '0'), ('applications', 'status', '''pending'''),
        ('applications', 'years_of_experience', '0'),
        ('candidate_notes', 'created_at', 'now()'), ('candidate_notes', 'updated_at', 'now()'),
        ('candidate_notes', 'is_private', 'false'),
        ('companies', 'created_at', 'now()'), ('companies', 'updated_at', 'now()'),
        ('companies', 'embedded_mode', 'false'), ('companies', 'subscription_status', '''trial'''),
        ('companies', 'subscription_tier', '''starter'''),
        ('email_logs', 'sent_at', 'now()'),
        ('jobs', 'company_id', null), ('jobs', 'auto_shortlist', 'true'), ('jobs', 'status', '''open'''),
        ('jobs', 'created_at', 'now()'), ('jobs', 'updated_at', 'now()'),
        ('messages', 'created_at', 'now()'), ('messages', 'is_read', 'false'),
        ('nurture_campaigns', 'created_at', 'now()'), ('nurture_campaigns', 'email_sent_at', 'now()'),
        ('nurture_campaigns', 'status', '''sent'''),
        ('nurture_preferences', 'created_at', 'now()'), ('nurture_preferences', 'updated_at', 'now()'),
        ('nurture_preferences', 'is_active', 'true'),
        ('payments', 'created_at', 'now()'), ('payments', 'updated_at', 'now()'),
        ('payments', 'currency', '''USD'''), ('payments', 'status', '''pending'''),
        ('payments', 'payment_method', null), ('payments', 'transaction_id', null),
        ('subscription_plans', 'created_at', 'now()'), ('subscription_plans', 'updated_at', 'now()'),
        ('subscription_plans', 'currency', '''USD'''), ('subscription_plans', 'features', '''[]'''),
        ('subscription_plans', 'ai_shortlisting', 'false'), ('subscription_plans', 'max_jobs', '5'),
        ('subscription_plans', 'max_applications', '100'),
        ('subscriptions', 'created_at', 'now()'), ('subscriptions', 'updated_at', 'now()'),
        ('subscriptions', 'status', '''active'''), ('subscriptions', 'cancel_at_period_end', 'false'),
        ('subscriptions', 'current_period_start', null), ('subscriptions', 'current_period_end', null),
        ('super_admins', 'created_at', 'now()')
    ) as t(tbl, col, dflt)
    loop
        if c.dflt is not null then
            execute format('alter table %I alter column %I set default %s', c.tbl, c.col, c.dflt);
            execute format('update %I set %I = %s where %I is null', c.tbl, c.col, c.dflt, c.col);
        end if;
        execute format('alter table %I alter column %I set not null', c.tbl, c.col);
    end loop;
end $$;

-- Check and unique constraints, named the way Postgres names them in the init migrations.
do $$
declare
    k record;
begin
    for k in select * from (values
        ('jobs', 'jobs_status_check', 'check (status in (''open'', ''closed'', ''archived''))'),
        ('applications', 'applications_status_check',
         'check (status in (''pending'', ''cv_viewed'', ''shortlisted'', ''rejected''))'),
        ('applications', 'applications_score_check', 'check (score between 0 and 100)'),
        ('subscription_plans', 'subscription_plans_name_key', 'unique (name)'),
        ('subscription_plans', 'subscription_plans_price_check', 'check (price >= 0)'),
        ('subscriptions', 'subscriptions_status_check',
         'check (status in (''trial'', ''active'', ''cancelled'', ''expired''))'),
        ('payments', 'payments_amount_check', 'check (amount >= 0)'),
        ('payments', 'payments_status_check', 'check (status in (''pending'', ''completed'', ''failed'', ''refunded''))'),
        ('payments', 'payments_transaction_id_key', 'unique (transaction_id)'),
        ('nurture_campaigns', 'nurture_campaigns_email_type_check',
         'check (email_type in (''job_alert'', ''check_in'', ''opportunity''))'),
        ('nurture_campaigns', 'nurture_campaigns_status_check',
         'check (status in (''sent'', ''opened'', ''clicked'', ''bounced''))')
    ) as t(tbl, name, def)
    loop
        if not exists (select 1 from pg_constraint where conname = k.name and conrelid = k.tbl::regclass) then
            execute format('alter table %I add constraint %I %s', k.tbl, k.name, k.def);
        end if;
    end loop;
end $$;

create index if not exists idx_admins_company_id on admins(company_id);
create index if not exists idx_jobs_company_status on jobs(company_id, status);
create index if not exists idx_applications_job_id on applications(job_id);
create index if not exists idx_applications_company_status on applications(company_id, status);
create index if not exists idx_applications_company_applied on applications(company_id, applied_at desc);
create index if not exists idx_applications_email on applications(email);
create index if not exists idx_applications_talent_pool on applications(company_id) where in_talent_pool;
create index if not exists idx_candidate_notes_application on candidate_notes(application_id, created_at desc);
create index if not exists idx_messages_application on messages(application_id, created_at);
create index if not exists idx_activity_logs_company_created on activity_logs(company_id, created_at desc);
create index if not exists idx_activity_logs_entity on activity_logs(entity_type, entity_id);
create index if not exists idx_subscriptions_company on subscriptions(company_id, status);
create index if not exists idx_payments_company on payments(company_id, created_at desc);
create index if not exists idx_nurture_campaigns_application on nurture_campaigns(application_id, email_type, email_sent_at desc);
create index if not exists idx_nurture_preferences_active on nurture_preferences(email) where is_active;
create index if not exists idx_email_logs_application on email_logs(application_id);

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

drop trigger if exists update_candidate_notes_updated_at on candidate_notes;
drop trigger if exists update_nurture_preferences_updated_at on nurture_preferences;

do $$
declare
    t text;
begin
    foreach t in array array['companies', 'jobs', 'candidate_notes', 'subscription_plans', 'subscriptions',
                             'payments', 'nurture_preferences']
    loop
        execute format('drop trigger if exists %I on %I', t || '_updated_at', t);
        execute format('create trigger %I before update on %I for each row execute function set_updated_at()',
                       t || '_updated_at', t);
    end loop;
end $$;

-- RLS on every public table, including legacy ones the code no longer uses.
do $$
declare
    t record;
begin
    for t in select tablename from pg_tables where schemaname = 'public' loop
        execute format('alter table %I enable row level security', t.tablename);
    end loop;
end $$;

-- Wildcard policies opened every bucket to anonymous listing and uploads.
drop policy if exists "Allow public reads i5g8va_0" on storage.objects;
drop policy if exists "Allow public uploads i5g8va_0" on storage.objects;

update storage.buckets
set file_size_limit = 10485760,
    allowed_mime_types = case id
        when 'resumes' then array['application/pdf', 'application/msword',
                                  'application/vnd.openxmlformats-officedocument.wordprocessingml.document']
        else array['application/pdf', 'application/zip', 'application/x-rar-compressed', 'application/x-7z-compressed']
    end
where id in ('resumes', 'portfolios');

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
     1000000, 1000000, true)
on conflict (name) do nothing;
