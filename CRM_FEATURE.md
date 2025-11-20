# Candidate Relationship Management (CRM) Feature

## Overview

This feature transforms the ATS from a simple application tracker into a comprehensive candidate relationship management system, focusing on building long-term relationships with candidates rather than treating them as numbers.

## Features Implemented

### 1. Talent Pool Management ✅

- **Add to Talent Pool**: Mark promising candidates for future opportunities
- **Talent Pool Page**: Dedicated page to view all candidates in talent pool (`/admin/dashboard/talent-pool`)
- **Remove from Pool**: Easy removal when no longer needed
- **Tracking**: Records when and by whom candidates were added

### 2. Automated Nurture Campaigns ✅

- **Monthly Job Alerts**: Automated emails sent to talent pool candidates
- **Job Matching**: Send relevant job opportunities based on candidate profile
- **Check-in Emails**: Monthly check-in emails to keep candidates engaged
- **Campaign Tracking**: All emails logged in `nurture_campaigns` table
- **Service Function**: `ProcessMonthlyNurtureCampaigns()` - can be run as cron job

### 3. Candidate Notes ✅

- **Add Notes**: Recruiters can add personal notes about candidates
- **Private Notes**: Option to make notes private (only visible to creator)
- **View Notes**: See all notes for a candidate in a modal
- **Edit/Delete**: Update or delete your own notes
- **Notes History**: All notes stored with timestamps and author info

### 4. Relationship Timeline ✅

- **Complete History**: See all interactions with a candidate in chronological order
- **Event Types**:
  - Application submitted
  - CV viewed
  - Status changes (shortlisted, rejected, etc.)
  - Notes added
  - Messages sent/received
  - Activity logs
  - Talent pool additions
- **Visual Timeline**: Easy-to-read timeline with icons and timestamps

### 5. Referral Tracking ✅

- **Referral Source**: Track how candidates heard about the job
- **Referrer Information**: Store name, email, and phone of referrer
- **Update Anytime**: Edit referral information through modal
- **Analytics Ready**: Data available for referral program analytics

## Database Schema

### New Tables:

1. **candidate_notes** - Stores recruiter notes
2. **nurture_campaigns** - Tracks automated email campaigns
3. **nurture_preferences** - Stores candidate preferences for job alerts

### Updated Tables:

- **applications** - Added fields:
  - `referral_source`, `referred_by_name`, `referred_by_email`, `referred_by_phone`
  - `in_talent_pool`, `talent_pool_added_at`, `talent_pool_added_by`

## API Endpoints

### CRM Endpoints (Protected):

- `POST /api/crm/notes` - Add candidate note
- `GET /api/crm/applications/:id/notes` - Get all notes for application
- `PUT /api/crm/notes/:id` - Update note
- `DELETE /api/crm/notes/:id` - Delete note
- `POST /api/crm/talent-pool` - Add to talent pool
- `DELETE /api/crm/talent-pool/:id` - Remove from talent pool
- `GET /api/crm/talent-pool` - Get all talent pool candidates
- `PUT /api/crm/applications/:id/referral` - Update referral info
- `GET /api/crm/applications/:id/timeline` - Get relationship timeline

## Frontend Pages

### Talent Pool Page

- **Route**: `/admin/dashboard/talent-pool`
- **Features**:
  - View all candidates in talent pool
  - See when they were added
  - Quick actions (View CV, Remove from pool)
  - Empty state with link to Applications page

### Applications Page (Enhanced)

- **New Actions in Dropdown**:
  - 📝 Notes - View/add notes
  - 📅 Timeline - View relationship timeline
  - 👥 Referral Info - Edit referral information
  - ⭐ Add/Remove from Talent Pool

### Modals

- **Notes Modal**: View all notes, add new notes, mark as private
- **Timeline Modal**: Visual timeline of all interactions
- **Referral Modal**: Edit referral source and referrer information

## Usage

### Adding to Talent Pool:

1. Go to Applications page
2. Click ⋮ menu on any application
3. Click "⭐ Add to Talent Pool"
4. Candidate is now in talent pool

### Adding Notes:

1. Click ⋮ menu on application
2. Click "📝 Notes"
3. Enter note in modal
4. Optionally mark as private
5. Click "Add Note"

### Viewing Timeline:

1. Click ⋮ menu on application
2. Click "📅 Timeline"
3. See all interactions in chronological order

### Updating Referral Info:

1. Click ⋮ menu on application
2. Click "👥 Referral Info"
3. Fill in referral source and referrer details
4. Click "Save"

## Automated Nurture Campaigns

### Setup Monthly Campaigns:

The `ProcessMonthlyNurtureCampaigns()` function should be run as a scheduled job (cron) to:

- Send monthly check-in emails to talent pool candidates
- Send job alerts when new relevant jobs are posted

### Example Cron Job:

```bash
# Run monthly on the 1st of each month at 9 AM
0 9 1 * * /path/to/your/app ProcessMonthlyNurtureCampaigns
```

Or integrate with your Go application's scheduler.

## Migration Instructions

1. **Run Database Migration**:

   ```sql
   -- Run supabase/CRM_MIGRATION.sql in Supabase SQL Editor
   ```

2. **Restart Backend Server**:

   - The new models and controllers are ready to use

3. **Access Features**:
   - Talent Pool: Navigate to "⭐ Talent Pool" in the navbar
   - Notes/Timeline/Referral: Available in Applications page dropdown

## Future Enhancements

- [ ] Bulk add to talent pool
- [ ] Talent pool search and filters
- [ ] Custom nurture campaign templates
- [ ] Referral rewards tracking
- [ ] Candidate preferences UI (job types, locations, salary)
- [ ] Automated job matching for talent pool
- [ ] Email campaign analytics dashboard
- [ ] Export talent pool to CSV
- [ ] Tags/labels for candidates
- [ ] Candidate rating system

## Notes

- Private notes are only visible to the recruiter who created them
- Talent pool candidates can still be in other statuses (pending, shortlisted, etc.)
- Timeline includes all interactions automatically
- Referral tracking is optional - fields can be left empty
- Nurture campaigns are logged for analytics
