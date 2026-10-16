-- Demo data for local development (supabase db reset). Every account's password is: demo12345
-- Never run this against production.

insert into companies (id, company_name, email, company_website, subscription_status)
values ('11111111-1111-4111-8111-111111111111', 'Acme Labs', 'admin@acme.test', 'https://acme.test', 'active');

insert into admins (id, company_id, name, email, password_hash)
values ('22222222-2222-4222-8222-222222222222', '11111111-1111-4111-8111-111111111111',
        'Ayesha Khan', 'admin@acme.test', '$2a$10$GLF4uCxB8wGvGPc/eseHAuoLf5.8Dlfzw0yjWlukJDd2FDZuKe4SC');

insert into super_admins (name, email, password_hash)
values ('Platform Owner', 'owner@ats.test', '$2a$10$GLF4uCxB8wGvGPc/eseHAuoLf5.8Dlfzw0yjWlukJDd2FDZuKe4SC');

insert into jobs (id, company_id, title, description, requirements, location, job_type, salary_range, deadline, shortlist_criteria)
values
    ('33333333-3333-4333-8333-333333333331', '11111111-1111-4111-8111-111111111111',
     'Senior Backend Engineer',
     'Build and scale our Go services and PostgreSQL data layer. You will own APIs used by thousands of recruiters.',
     '5+ years backend experience, Go, PostgreSQL, Docker',
     'Islamabad (Hybrid)', 'full-time', 'PKR 450k - 600k',
     current_date + interval '30 days',
     '{"required_skills": ["Go", "PostgreSQL", "Docker"], "min_experience": 5, "required_languages": ["English"], "match_job_description": true}'),
    ('33333333-3333-4333-8333-333333333332', '11111111-1111-4111-8111-111111111111',
     'Frontend Developer',
     'Craft fast, accessible interfaces with React and Next.js for our recruiter dashboard.',
     'React, TypeScript, Tailwind CSS',
     'Remote', 'full-time', 'PKR 250k - 350k',
     current_date + interval '21 days',
     null);

insert into applications (job_id, company_id, full_name, email, phone, resume_url, years_of_experience,
                          current_position, status, score, parsed_cv_text, applied_at)
values
    ('33333333-3333-4333-8333-333333333331', '11111111-1111-4111-8111-111111111111',
     'Bilal Ahmed', 'bilal@example.com', '+923001234567', 'https://example.com/resumes/bilal.pdf', 6,
     'Backend Engineer', 'shortlisted', 88,
     'Backend engineer with 6 years of experience in Go, PostgreSQL, Docker and Kubernetes. Fluent English.',
     now() - interval '3 days'),
    ('33333333-3333-4333-8333-333333333331', '11111111-1111-4111-8111-111111111111',
     'Sara Malik', 'sara@example.com', null, 'https://example.com/resumes/sara.pdf', 3,
     'Software Engineer', 'pending', 54,
     'Software engineer with 3 years of experience in Python, Django and PostgreSQL. English and Urdu.',
     now() - interval '1 day'),
    ('33333333-3333-4333-8333-333333333332', '11111111-1111-4111-8111-111111111111',
     'Hamza Iqbal', 'hamza@example.com', null, 'https://example.com/resumes/hamza.pdf', 4,
     'Frontend Developer', 'cv_viewed', 76,
     'Frontend developer, 4 years of experience with React, TypeScript, Next.js and Tailwind CSS.',
     now() - interval '2 days');
