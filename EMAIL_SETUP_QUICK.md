# Quick Email Setup for Localhost (SendGrid)

**Perfect for localhost development - no domain needed!**

## 5-Minute Setup

### 1. Sign Up for SendGrid (Free)

- Go to https://signup.sendgrid.com/
- Sign up with your email (no credit card needed)
- Verify your email address

### 2. Create API Key

- Go to **Settings** → **API Keys**
- Click **"Create API Key"**
- Name: `ATS System`
- Permissions: **Full Access** (or "Mail Send")
- **Copy the key** (starts with `SG.`)

### 3. Verify Your Email as Sender (No Domain Needed!)

- Go to **Settings** → **Sender Authentication**
- Click **"Verify a Single Sender"** ⬅️ This is the key!
- Fill in:
  - **From Email:** Your personal email (e.g., `yourname@gmail.com`)
  - **From Name:** Your name
  - **Address, City, etc.:** Your details (any address works for testing)
- Click **"Create"**
- **Check your email** and click the verification link
- ✅ Done! (Takes 2-3 minutes)

### 4. Add to `backend/.env`

```env
EMAIL_PROVIDER=sendgrid
SENDGRID_API_KEY=SG.your-copied-api-key-here
SENDGRID_FROM_EMAIL=yourname@gmail.com
```

**Use the exact same email you verified in step 3!**

### 5. Restart Backend

```bash
cd backend
go run main.go
```

## That's It! 🎉

You can now send emails to **any email address** from localhost!

## Common Questions

**Q: Do I need a domain?**  
A: No! Use your personal Gmail/Outlook/etc. email.

**Q: Can I send to any email?**  
A: Yes! Unlike Resend, SendGrid allows sending to any email address.

**Q: How many emails can I send?**  
A: 100 emails/day for free (perfect for testing).

**Q: What if verification fails?**  
A: Make sure you clicked the link in the verification email. Check spam folder.

## Troubleshooting

**"SENDGRID_FROM_EMAIL must be set"**

- Add `SENDGRID_FROM_EMAIL=your-verified-email@gmail.com` to `.env`

**"Sender email not verified"**

- Make sure you verified the email in SendGrid dashboard
- Use the exact same email address you verified

**"API key invalid"**

- Make sure you copied the full API key (starts with `SG.`)
- Check that API key has "Mail Send" permissions
