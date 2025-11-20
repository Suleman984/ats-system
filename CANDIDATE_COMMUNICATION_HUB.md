# Candidate Communication Hub Implementation

## Overview

This feature implements a comprehensive candidate communication system that addresses the lack of communication during the hiring process. It includes real-time status tracking, automated updates, two-way messaging, and SMS notifications.

## Features Implemented

### 1. Real-time Status Portal ✅

- Candidates can log in using their email and application ID
- See exactly where they are in the hiring process
- View status history and timeline
- Access at `/application-status`

### 2. Automated Status Updates ✅

- **CV Viewed**: When recruiter views CV, status automatically changes from "pending" to "cv_viewed"
- **Shortlisted**: Status changes to "shortlisted" with email and SMS notifications
- **Rejected**: Status changes to "rejected" with email and SMS notifications
- **Under Review**: Can be set manually or automatically
- **Interview Scheduled**: Can be set for future use
- **Decision Pending**: Can be set for future use

### 3. Expected Timeline ✅

- Shows "You'll hear from us within X days" based on expected response date
- Automatically calculates expected response date (5 days from status update)
- Displays countdown to expected response

### 4. Two-way Messaging ✅

- Candidates can send messages to recruiters
- Recruiters can send messages to candidates
- Messages are stored in database
- Email notifications sent when recruiter messages candidate
- SMS notifications sent when recruiter messages candidate (if phone number available)
- Read receipts for messages

### 5. SMS Notifications ✅

- Integrated with Twilio for SMS notifications
- Sends SMS for:
  - CV viewed
  - Shortlisted
  - Rejected
  - New messages from recruiters
- Gracefully handles missing phone numbers or unconfigured SMS

## Database Changes

### New Columns in `applications` table:

- `cv_viewed_at` - Timestamp when CV was first viewed
- `cv_viewed_by` - Admin ID who viewed the CV
- `expected_response_date` - Expected date for response
- `last_status_update` - Last time status was updated

### New `messages` table:

- Stores all messages between candidates and recruiters
- Tracks sender type (candidate/recruiter)
- Tracks read status
- Links to application

## API Endpoints

### Public Endpoints (No Auth):

- `POST /api/candidate/status` - Check application status
- `GET /api/candidate/applications` - Get all applications by email
- `POST /api/candidate/messages/send` - Send message (candidate)
- `GET /api/candidate/messages` - Get messages (candidate)

### Protected Endpoints (Admin Auth):

- `POST /api/applications/:id/track-cv-view` - Track CV view
- `POST /api/applications/:id/messages` - Send message (recruiter)
- `GET /api/applications/:id/messages` - Get messages (recruiter)

## Environment Variables

### SMS Configuration (Optional):

```bash
SMS_PROVIDER=twilio  # Currently only Twilio supported
TWILIO_ACCOUNT_SID=your_account_sid
TWILIO_AUTH_TOKEN=your_auth_token
TWILIO_FROM_NUMBER=+1234567890  # Your Twilio phone number
```

### Email Configuration (Already configured):

- Uses existing email service (SendGrid or Resend)

## Frontend Pages

### Candidate Portal

- **Route**: `/application-status`
- **Features**:
  - Status check form (email + application ID)
  - Real-time status display with timeline
  - Expected response date countdown
  - Two-way messaging interface
  - Message history
  - Unread message indicators

### Admin Applications Page

- Automatically tracks CV views when recruiters click "View CV"
- Status updates automatically based on actions
- Email and SMS sent automatically on status changes

## Status Flow

1. **Pending** → Application submitted
2. **CV Viewed** → Recruiter views CV (automatic)
3. **Under Review** → Can be set manually
4. **Shortlisted** → Recruiter shortlists (automatic, sends email + SMS)
5. **Rejected** → Recruiter rejects (automatic, sends email + SMS)
6. **Interview Scheduled** → Can be set manually
7. **Decision Pending** → Can be set manually

## Migration Instructions

1. Run the database migration:

   ```sql
   -- Run this in Supabase SQL Editor
   -- File: supabase/CANDIDATE_COMMUNICATION_MIGRATION.sql
   ```

2. Update environment variables (optional for SMS):

   ```bash
   SMS_PROVIDER=twilio
   TWILIO_ACCOUNT_SID=your_account_sid
   TWILIO_AUTH_TOKEN=your_auth_token
   TWILIO_FROM_NUMBER=+1234567890
   ```

3. Restart backend server

## Usage Examples

### For Candidates:

1. Visit `/application-status`
2. Enter email and application ID
3. View status, timeline, and messages
4. Send messages to recruiters

### For Recruiters:

1. View CV → Status automatically changes to "cv_viewed"
2. Shortlist → Status changes to "shortlisted" + email + SMS sent
3. Reject → Status changes to "rejected" + email + SMS sent
4. Send messages via API or future UI integration

## Future Enhancements

- [ ] Add messaging UI to admin dashboard
- [ ] Add status update UI for recruiters (interview scheduled, etc.)
- [ ] Add email templates customization
- [ ] Add SMS template customization
- [ ] Add push notifications
- [ ] Add status change history log
- [ ] Add bulk messaging capabilities

## Notes

- SMS notifications are optional and won't fail if not configured
- Email notifications use existing email service
- All status changes are logged in activity logs
- Messages are stored permanently for audit trail
- CV view tracking only happens once per application
