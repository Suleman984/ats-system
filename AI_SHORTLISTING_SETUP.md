# AI-Powered CV Analysis & Auto-Shortlisting Setup

This guide will help you set up AI-powered CV analysis and auto-shortlisting using OpenAI.

## Overview

The AI shortlisting feature:

- **Analyzes CVs/resumes** using OpenAI GPT-4
- **Scores candidates** (0-100%) based on match criteria
- **Auto-shortlists** candidates above threshold (default 70%)
- **Shows match percentage** in applications table
- **Stores analysis results** for review

## Prerequisites

- OpenAI API account (free tier available: $5 credit)
- CV files accessible via URL (stored in Supabase Storage)

## Step 1: Get OpenAI API Key

1. Go to https://platform.openai.com/
2. Sign up or log in
3. Navigate to **API Keys** section
4. Click **"Create new secret key"**
5. Name it: `ATS System`
6. **Copy the key immediately** (starts with `sk-`)

**Note:** OpenAI provides $5 free credit for new accounts, perfect for testing!

## Step 2: Add to Backend .env

Add this to your `backend/.env` file:

```env
OPENAI_API_KEY=sk-your-api-key-here
```

## Step 3: How It Works

### Scoring Algorithm

The AI calculates match score (0-100%) based on:

- **Skills Match** (40% weight) - Required skills found in CV
- **Experience Match** (30% weight) - Years of experience vs requirement
- **Language Requirements** (20% weight) - Required languages found
- **Job Description Alignment** (10% weight) - Overall fit with job

### Auto-Shortlisting

- Candidates with **score ≥ 70%** are automatically shortlisted
- Email notification is sent automatically
- You can adjust threshold in batch analysis

## Step 4: Using AI Shortlisting

### Option 1: Analyze Single Application

1. Go to **Applications** page
2. Click **"🤖 AI Analyze"** button on any pending application
3. Set criteria:
   - Required Skills (comma-separated)
   - Minimum Experience (years)
   - Required Languages
   - Match Job Description (checkbox)
4. Click **"🤖 Analyze with AI"**
5. View the match score and analysis

### Option 2: Batch Analyze All Applications

1. Select a **Job** from the filter dropdown
2. Click **"🤖 AI Analyze All"** button (top right)
3. Set criteria (same as above)
4. All pending applications for that job will be analyzed
5. Candidates above 70% will be auto-shortlisted

### Option 3: Use Job's Criteria

If you set criteria when creating/editing a job, the AI will use those automatically.

## Step 5: Viewing Results

### Match Score Column

- **Green (80-100%)**: Excellent match
- **Yellow (60-79%)**: Good match
- **Red (0-59%)**: Poor match
- **"Not analyzed"**: CV not yet analyzed

### Analysis Details

Click the **📊** icon next to score to view detailed analysis:

- Skills found
- Experience level
- Education
- Languages
- Strengths
- Missing skills
- Match reason

## Step 6: Criteria Options

### When Creating/Editing Job

You can set `shortlist_criteria` as JSON:

```json
{
  "required_skills": ["JavaScript", "React", "Node.js"],
  "min_experience": 3,
  "required_languages": ["English"],
  "match_job_description": true
}
```

### On Applications Page

Set criteria dynamically when analyzing:

- **Required Skills**: Comma-separated list
- **Min Experience**: Number of years
- **Required Languages**: Comma-separated list
- **Match Job Description**: Checkbox to match job requirements

## Cost Estimation

**OpenAI Pricing (gpt-4o-mini):**

- ~$0.15 per 1M input tokens
- ~$0.60 per 1M output tokens
- Average CV analysis: ~$0.01-0.02 per analysis

**Free Tier:**

- $5 free credit for new accounts
- Enough for ~250-500 CV analyses

## Troubleshooting

### "OPENAI_API_KEY not set"

- Add `OPENAI_API_KEY` to `backend/.env`
- Restart backend server

### "Failed to analyze CV"

- Check OpenAI API key is valid
- Verify CV URL is accessible
- Check backend logs for detailed error

### "CV text extraction failed"

- For best results, ensure CVs are in PDF or text format
- CVs stored in Supabase Storage work best
- The AI can analyze URLs, but text extraction improves accuracy

### Low Match Scores

- Adjust criteria to be more realistic
- Check if CV format is readable
- Review analysis details to see what's missing

## Future Enhancements

For production, consider:

1. **PDF Text Extraction**: Use libraries like `pdfcpu` for better text extraction
2. **Caching**: Cache analysis results to avoid re-analyzing
3. **Rate Limiting**: Limit AI calls per company/user
4. **Custom Models**: Fine-tune models for your industry
5. **Multi-language Support**: Better language detection

## API Endpoints

- `POST /api/applications/ai-shortlist` - Analyze single application
- `POST /api/applications/ai-shortlist-batch` - Batch analyze applications

## Example Request

```json
{
  "application_id": "uuid-here",
  "required_skills": ["JavaScript", "React"],
  "min_experience": 2,
  "required_languages": ["English"],
  "match_job_description": true
}
```

## Security Notes

- Never commit API keys to version control
- Use environment variables for all keys
- Consider rate limiting for production
- Monitor API usage to control costs

## That's It! 🎉

Your AI shortlisting system is ready! Start analyzing CVs and see match scores appear automatically.
