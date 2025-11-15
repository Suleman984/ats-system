# Embedded Dashboard Mode Guide

## Overview

The ATS system now supports **Embedded Dashboard Mode**, allowing companies to embed the entire ATS dashboard into their existing website. This provides a white-label solution where the dashboard appears seamlessly within the company's website.

## Features

### Two Embed Options

1. **Full Dashboard Embed** (`/embed/dashboard`)

   - Complete ATS dashboard with login
   - All features: Dashboard, Jobs Management, Applications
   - Perfect for companies wanting a fully integrated experience

2. **Job Portal Only** (`/jobs/:companyId`)
   - Public job listings only
   - For displaying open positions to candidates
   - No login required (public view)

## Setup

### 1. Company Registration

During registration, companies can choose to enable embedded mode:

- **Checkbox**: "Use Embedded Dashboard Mode"
- **Optional Domain Field**: Enter your website domain (e.g., `example.com`) for security
  - If provided, the dashboard can only be embedded on this domain

### 2. Database Schema

The following fields were added to the `companies` table:

```sql
embedded_mode BOOLEAN DEFAULT false
embed_domain VARCHAR(255) -- Optional, for security
```

### 3. Getting Embed Code

After logging in as an admin:

1. Navigate to **Dashboard → Embed Code**
2. You'll see two sections:
   - **📊 Embed Full Dashboard** - Complete dashboard embed code
   - **💼 Embed Job Portal Only** - Public job listings embed code
3. Click "Copy" to copy the embed code
4. Paste it into your website

## Embed Code Examples

### Full Dashboard Embed

```html
<iframe
  src="https://your-ats-domain.com/embed/dashboard"
  width="100%"
  height="900px"
  frameborder="0"
  allow="clipboard-read; clipboard-write"
>
</iframe>
```

### Job Portal Only

```html
<iframe
  src="https://your-ats-domain.com/jobs/{company_id}"
  width="100%"
  height="800px"
  frameborder="0"
>
</iframe>
```

## Implementation Details

### Frontend Routes

- `/embed/login` - Login page for embedded mode
- `/embed/dashboard` - Main dashboard (redirects to login if not authenticated)
- `/embed/dashboard/jobs` - Jobs management
- `/embed/dashboard/applications` - Applications management

### Backend Changes

- `RegisterRequest` now includes:

  - `embedded_mode` (boolean)
  - `embed_domain` (string, optional)

- `Company` model includes:
  - `EmbeddedMode` field
  - `EmbedDomain` field (nullable)

### Authentication

- Embedded mode uses the same authentication system
- JWT tokens work in both normal and embedded modes
- Login happens within the iframe

## Usage Instructions

### For WordPress

1. Go to WordPress admin panel
2. Edit the page where you want to embed the dashboard
3. Add a "Custom HTML" block
4. Paste the embed code
5. Publish the page

### For HTML Website

1. Open your HTML file in a text editor
2. Find where you want to display the dashboard
3. Paste the embed code
4. Save and upload the file

### For Shopify

1. Go to Online Store → Pages
2. Create or edit a page (e.g., "Careers" or "ATS Dashboard")
3. Click "Show HTML"
4. Paste the embed code
5. Save

## Security Considerations

### Domain Restriction (Optional)

If you provide an `embed_domain` during registration:

- The system can validate that requests are coming from the allowed domain
- This prevents unauthorized embedding on other websites
- **Note**: Full domain validation requires additional backend middleware (can be implemented if needed)

### Iframe Security

- The embedded dashboard uses standard iframe security
- Cookies and localStorage work within the iframe
- Authentication tokens are stored in localStorage

## Differences: Normal vs Embedded Mode

| Feature    | Normal Mode                   | Embedded Mode             |
| ---------- | ----------------------------- | ------------------------- |
| Access URL | `/admin/dashboard`            | `/embed/dashboard`        |
| Layout     | Full page with mode indicator | Minimal layout for iframe |
| Navigation | Full navigation bar           | Compact navigation        |
| Styling    | Standard spacing              | Compact spacing           |
| Login      | Separate login page           | Login within iframe       |

## Super Admin

- **Super Admin always uses normal mode**
- No embedded option for super admin accounts
- Super admin login remains at `/super-admin/login`

## Troubleshooting

### Iframe Not Loading

- Check that the iframe `src` URL is correct
- Ensure the iframe height is sufficient (recommended: 900px+)
- Check browser console for CORS or security errors

### Login Issues

- Clear browser cache and cookies
- Ensure localStorage is enabled
- Check that the iframe allows cookies

### Styling Issues

- The embedded layout uses compact styling
- Adjust iframe height if content is cut off
- Check parent page CSS for conflicts

## Future Enhancements

Potential improvements:

- Domain validation middleware for enhanced security
- Custom branding options for embedded mode
- API endpoints for programmatic embed code generation
- Analytics tracking for embedded usage
