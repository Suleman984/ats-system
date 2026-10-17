-- One application per candidate per job. Fails (without deleting anything) if duplicates already exist;
-- find them with: select job_id, lower(email), count(*) from applications group by 1, 2 having count(*) > 1;
create unique index uq_applications_job_email on applications (job_id, lower(email));
