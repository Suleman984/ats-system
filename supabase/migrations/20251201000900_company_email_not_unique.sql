-- Login uniqueness lives on admins.email. A company's email is only contact info, and keeping it
-- unique made sign-up fail silently when a company outlived its admin.
alter table companies drop constraint if exists companies_email_key;
