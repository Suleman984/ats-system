-- The dashboard bell and row badges count unread candidate messages every minute; this partial
-- index covers only those rows, so it stays tiny as read history grows.
create index if not exists idx_messages_unread_from_candidate
    on messages(application_id) where sender_type = 'candidate' and not is_read;
