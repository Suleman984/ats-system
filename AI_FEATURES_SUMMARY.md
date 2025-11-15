# AI-Powered CV Analysis & Auto-Shortlisting - Complete Implementation

## ✅ What's Been Implemented

### 1. **Backend AI Services**

- ✅ CV parsing service (`backend/services/cv_parser.go`)
- ✅ OpenAI API integration for CV analysis
- ✅ Scoring algorithm (0-100% match score)
- ✅ Auto-shortlisting based on threshold (70% default)

### 2. **Database Updates**

- ✅ Added `analysis_result` JSONB field to `applications` table
- ✅ `score` field now stores AI match percentage (0-100)
- ✅ Criteria stored in `jobs.shortlist_criteria` JSONB field

### 3. **API Endpoints**

- ✅ `POST /api/applications/ai-shortlist` - Analyze single application
- ✅ `POST /api/applications/ai-shortlist-batch` - Batch analyze all applications for a job

### 4. **Frontend Features**

- ✅ **Match Score Column** - Shows percentage match (color-coded)
- ✅ **AI Analyze Button** - Individual application analysis
- ✅ **AI Analyze All Button** - Batch analysis for selected job
- ✅ **Criteria Modal** - Set skills, experience, languages, job match
- ✅ **Analysis Details** - Click 📊 icon to view full analysis

### 5. **Scoring System**

- ✅ **Skills Match** (40% weight)
- ✅ **Experience Match** (30% weight)
- ✅ **Language Requirements** (20% weight)
- ✅ **Job Description Alignment** (10% weight)

## 🎯 How to Use

### Single Application Analysis

1. Go to Applications page
2. Click **"🤖 AI Analyze"** on any pending application
3. Set criteria in the modal:
   - Required Skills (comma-separated)
   - Minimum Experience (years)
   - Required Languages
   - Match Job Description (checkbox)
4. Click **"🤖 Analyze with AI"**
5. View match score and analysis

### Batch Analysis

1. Select a **Job** from filter dropdown
2. Click **"🤖 AI Analyze All"** (top right)
3. Set criteria (same as above)
4. All pending applications analyzed automatically
5. Candidates ≥70% auto-shortlisted

### Using Job Criteria

When creating/editing a job, you can set `shortlist_criteria` as JSON:

```json
{
  "required_skills": ["JavaScript", "React"],
  "min_experience": 3,
  "required_languages": ["English"],
  "match_job_description": true
}
```

## 📊 Match Score Display

- **Green (80-100%)**: Excellent match ✅
- **Yellow (60-79%)**: Good match ⚠️
- **Red (0-59%)**: Poor match ❌
- **"Not analyzed"**: CV not yet analyzed

Click the **📊** icon to view:

- Skills found
- Experience level
- Education
- Languages
- Strengths
- Missing skills
- Match reason

## 🔧 Setup Required

### 1. Get OpenAI API Key

1. Go to https://platform.openai.com/api-keys
2. Create new secret key
3. Copy the key (starts with `sk-`)

### 2. Add to Backend .env

```env
OPENAI_API_KEY=sk-your-api-key-here
```

### 3. Restart Backend

```bash
cd backend
go run main.go
```

## 💰 Cost Information

**OpenAI Pricing (gpt-4o-mini):**

- ~$0.01-0.02 per CV analysis
- $5 free credit for new accounts
- Enough for ~250-500 analyses

**Model Used:** `gpt-4o-mini` (cost-effective, accurate)

## 🎨 UI Features

### Applications Table

- New **"Match Score"** column showing percentage
- Color-coded badges (green/yellow/red)
- 📊 icon to view detailed analysis
- **"🤖 AI Analyze"** button for each application

### Batch Analysis

- **"🤖 AI Analyze All"** button (appears when job selected)
- Analyzes all pending applications for selected job
- Shows progress and results

### Criteria Modal

- Skills input (comma-separated)
- Experience slider/number input
- Languages input
- "Match job description" checkbox
- Clean, user-friendly interface

## 🔄 Auto-Shortlisting

- **Threshold:** 70% (configurable in batch analysis)
- **Auto-action:** Status changed to "shortlisted"
- **Email:** Automatic notification sent
- **Logging:** All actions logged for audit

## 📝 What You Need to Do

1. **Get OpenAI API Key** (free $5 credit available)
2. **Add to .env:** `OPENAI_API_KEY=sk-...`
3. **Restart backend server**
4. **Test with a real application**

## 🚀 Future Enhancements (For Paid Tier)

- PDF text extraction library integration
- Custom AI model fine-tuning
- Advanced criteria (education level, certifications)
- Multi-language CV support
- Analysis caching
- Rate limiting per company
- Custom scoring weights

## 📚 Documentation

- **Setup Guide:** `AI_SHORTLISTING_SETUP.md`
- **API Documentation:** See backend routes
- **Frontend Usage:** See applications page UI

## 🎉 Ready to Use!

The system is fully implemented and ready. Just add your OpenAI API key and start analyzing CVs!
