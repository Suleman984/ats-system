-- Uploads go through the API with the service-role key, so no storage.objects policies are granted.
-- Objects get random UUID names, which makes the public URLs unguessable.

insert into storage.buckets (id, name, public, file_size_limit, allowed_mime_types)
values
    ('resumes', 'resumes', true, 10485760, array[
        'application/pdf',
        'application/msword',
        'application/vnd.openxmlformats-officedocument.wordprocessingml.document'
    ]),
    ('portfolios', 'portfolios', true, 10485760, array[
        'application/pdf',
        'application/zip',
        'application/x-rar-compressed',
        'application/x-7z-compressed'
    ])
on conflict (id) do update
set public = excluded.public,
    file_size_limit = excluded.file_size_limit,
    allowed_mime_types = excluded.allowed_mime_types;
