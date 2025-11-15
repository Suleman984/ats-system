# Email Service Setup Guide

This guide will help you set up email sending for the ATS system. You can choose between **SendGrid** (recommended for testing) or **Resend**.

## Quick Comparison

| Feature                  | SendGrid (Free)       | Resend (Free)                      |
| ------------------------ | --------------------- | ---------------------------------- |
| **Free Emails/Day**      | 100                   | 3,000/month                        |
| **Send to Any Email**    | ✅ Yes                | ❌ Only verified email (test mode) |
| **Credit Card Required** | ❌ No                 | ❌ No                              |
| **Best For**             | Testing & Development | Production (with domain)           |

## Option 1: SendGrid (Recommended for Testing) ⭐

SendGrid allows you to send emails to **any email address** without restrictions, making it perfect for testing.

### Step 1: Create SendGrid Account

1. Go to https://signup.sendgrid.com/
2. Sign up for a free account (no credit card required)
3. Verify your email address

### Step 2: Create API Key

1. After logging in, go to **Settings** → **API Keys**
2. Click **"Create API Key"**
3. Name it: `ATS System`
4. Select **"Full Access"** (or "Mail Send" permissions)
5. Click **"Create & View"**
6. **Copy the API key immediately** (you won't see it again!)

### Step 3: Verify Sender Email (Required - But Easy!)

**Important:** SendGrid requires sender verification, but you can verify a **single email address** (not a domain) - perfect for localhost development!

1. Go to **Settings** → **Sender Authentication**
2. Click **"Verify a Single Sender"** (NOT "Authenticate Your Domain")
3. Fill in the form:
   - **From Email Address:** Use your personal email (e.g., `yourname@gmail.com`)
   - **From Name:** Your name or company name
   - **Reply To:** Same email (or leave blank)
   - **Address:** Your address (can be any address for testing)
   - **City, State, Zip:** Your location
   - **Country:** Your country
4. Click **"Create"**
5. **Check your email inbox** - SendGrid will send a verification email
6. Click the verification link in the email
7. Once verified, use this email as `SENDGRID_FROM_EMAIL`

**Note:**

- You can use your personal Gmail/Outlook/etc. email - no domain needed!
- This works perfectly for localhost development
- The verification usually takes just a few minutes

### Step 4: Add to Backend .env

Add these to your `backend/.env` file:

```env
# Email Provider (sendgrid or resend)
EMAIL_PROVIDER=sendgrid

# SendGrid Configuration
SENDGRID_API_KEY=SG.your-api-key-here
SENDGRID_FROM_EMAIL=your-verified-email@gmail.com
```

**Example (using your personal Gmail):**

```env
EMAIL_PROVIDER=sendgrid
SENDGRID_API_KEY=SG.abc123xyz789...
SENDGRID_FROM_EMAIL=yourname@gmail.com
```

**Important:**

- Use the **exact same email** you verified in Step 3
- This can be your personal Gmail/Outlook/etc. - no domain needed!
- Perfect for localhost development

### Step 5: Test It!

1. Restart your backend server
2. Submit a test application
3. Check the recipient's email inbox

## Option 2: Resend (For Production)

Resend is great for production but has limitations in test mode.

### Step 1: Create Resend Account

1. Go to https://resend.com
2. Sign up for a free account
3. Verify your email

### Step 2: Get API Key

1. Go to **API Keys** section
2. Click **"Create API Key"**
3. Name it: `ATS System`
4. Copy the API key (starts with `re_`)

### Step 3: Domain Verification (Required for Production)

**For Testing:**

- You can use `onboarding@resend.dev` (but can only send to your verified email)

**For Production:**

1. Go to **Domains** section
2. Click **"Add Domain"**
3. Add your domain (e.g., `yourcompany.com`)
4. Add the DNS records to your domain provider
5. Wait for verification (usually a few minutes)

### Step 4: Add to Backend .env

```env
# Email Provider
EMAIL_PROVIDER=resend

# Resend Configuration
RESEND_API_KEY=re_your-api-key-here
RESEND_FROM_EMAIL=onboarding@resend.dev  # For testing
# OR
RESEND_FROM_EMAIL=noreply@yourdomain.com  # For production (verified domain)
```

## Switching Between Providers

Simply change the `EMAIL_PROVIDER` in your `.env` file:

```env
# Use SendGrid
EMAIL_PROVIDER=sendgrid
SENDGRID_API_KEY=...
SENDGRID_FROM_EMAIL=...

# OR use Resend
EMAIL_PROVIDER=resend
RESEND_API_KEY=...
RESEND_FROM_EMAIL=...
```

## Troubleshooting

### SendGrid Issues

**"API key invalid"**

- Make sure you copied the full API key
- Check that the API key has "Mail Send" permissions

**"Sender email not verified"**

- Verify your sender email in SendGrid dashboard
- Or use a verified email address

**"Email not received"**

- Check spam folder
- Verify sender email is correct
- Check SendGrid activity feed for delivery status

### Resend Issues

**"Can only send to verified email"**

- This is Resend's test mode limitation
- Switch to SendGrid for testing, or verify a domain in Resend

**"Domain not verified"**

- Complete domain verification in Resend dashboard
- Or use `onboarding@resend.dev` for testing (limited)

## Recommended Setup

**For Development/Testing:**

- Use **SendGrid** (100 emails/day, no restrictions)

**For Production:**

- Use **Resend** with verified domain (3,000 emails/month)
- OR use **SendGrid** (upgrade if needed)

## Environment Variables Summary

### SendGrid

```env
EMAIL_PROVIDER=sendgrid
SENDGRID_API_KEY=SG.xxx
SENDGRID_FROM_EMAIL=your-email@example.com
```

### Resend

```env
EMAIL_PROVIDER=resend
RESEND_API_KEY=re_xxx
RESEND_FROM_EMAIL=onboarding@resend.dev
```

## Next Steps

1. Choose your provider (SendGrid recommended for testing)
2. Set up your account and get API key
3. Add environment variables to `backend/.env`
4. Restart your backend server
5. Test by submitting an application

That's it! Your email system is ready to go! 🎉
