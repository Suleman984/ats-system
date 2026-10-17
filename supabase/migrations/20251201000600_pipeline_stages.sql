-- Hiring stages after shortlisting: interview -> offer -> hired.
alter table applications drop constraint applications_status_check;
alter table applications add constraint applications_status_check
    check (status in ('pending', 'cv_viewed', 'shortlisted', 'interview', 'offer', 'hired', 'rejected'));
