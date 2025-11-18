# Embed Code Security Implementation

## Problem Statement

Previously, the embed code was generic:

```html
<iframe src="http://localhost:3000/embed/dashboard"></iframe>
```

This created security concerns:

- Anyone could copy the embed code from another company's website
- There was no way to distinguish which company owned which embed code
- No validation to ensure the logged-in user belongs to the company

## Solution: Company-Specific Embed Codes

### 1. **Unique Embed Code Per Company**

Each company now gets a unique embed code that includes their `company_id`:

```html
<iframe
  src="http://localhost:3000/embed/dashboard?company_id=6ba9d0c4-68de-41ef-a244-87e0117aa433"
  width="100%"
  height="900px"
  frameborder="0"
>
</iframe>
```

### 2. **Security Flow**

#### Step 1: User Visits Embedded Dashboard

- User visits a website with the embed code
- The iframe loads: `/embed/dashboard?company_id=xxx`
- System checks if `company_id` is present in URL

#### Step 2: Authentication Check

- **If NOT logged in:**

  - Redirects to `/embed/login?company_id=xxx`
  - User sees login page

- **If logged in:**
  - System validates: `logged_in_user.company_id === url_company_id`
  - **If match:** User sees dashboard ✅
  - **If mismatch:** Shows security error ❌

#### Step 3: Login Validation

- User enters credentials
- System validates login
- **Additional check:** `logged_in_company_id === url_company_id`
- **If match:** Redirects to dashboard ✅
- **If mismatch:** Shows error message ❌

### 3. **Security Features**

✅ **Company ID Validation**

- Every embed URL includes `company_id`
- System validates that logged-in user belongs to that company
- Prevents cross-company access

✅ **URL Parameter Preservation**

- `company_id` is preserved across all navigation within embed
- All links include `?company_id=xxx` parameter
- Ensures consistent validation

✅ **Error Handling**

- Clear error messages if company_id mismatch
- Prevents unauthorized access attempts
- Guides users to use correct embed code

### 4. **What Happens If Someone Copies Your Embed Code?**

**Scenario:** Company A copies Company B's embed code

1. **Company A pastes embed code on their website:**

   ```html
   <iframe src="...embed/dashboard?company_id=COMPANY_B_ID"></iframe>
   ```

2. **User visits Company A's website:**

   - Iframe loads with `company_id=COMPANY_B_ID`
   - User sees login page (if not logged in)

3. **User tries to log in:**

   - If user belongs to Company A: ❌ **Login fails** - "This login does not match the embed code"
   - If user belongs to Company B: ✅ **Login succeeds** - But they see Company B's dashboard

4. **Result:**
   - Company A cannot access Company B's data
   - Only Company B's users can successfully log in
   - Even if someone copies the embed code, they can't access unauthorized data

### 5. **Implementation Details**

#### Frontend Changes:

1. **Embed Code Generation** (`/admin/dashboard/embed/page.tsx`)

   - Now includes `company_id` in URL: `?company_id=${user?.company_id}`

2. **Embed Dashboard Page** (`/embed/dashboard/page.tsx`)

   - Extracts `company_id` from URL parameters
   - Validates against logged-in user's `company_id`
   - Shows error if mismatch

3. **Embed Login Page** (`/embed/login/page.tsx`)

   - Extracts `company_id` from URL
   - Validates login against URL `company_id`
   - Prevents cross-company logins

4. **Embed Layout** (`/embed/dashboard/layout.tsx`)
   - Preserves `company_id` in all navigation links
   - Ensures consistent validation across pages

### 6. **User Experience**

**For Legitimate Users:**

- ✅ Seamless experience - no extra steps
- ✅ Login works normally
- ✅ All features accessible

**For Unauthorized Access:**

- ❌ Clear error messages
- ❌ Cannot access wrong company's data
- ✅ Guided to use correct embed code

### 7. **Future Enhancements (Optional)**

1. **Domain Validation**

   - Use `embed_domain` field from Company model
   - Validate `document.referrer` matches allowed domain
   - Additional layer of security

2. **Token-Based Embed Codes**

   - Generate unique tokens per embed code
   - Tokens can be revoked/regenerated
   - More granular control

3. **Embed Code Expiration**
   - Time-limited embed codes
   - Automatic rotation for security

## Summary

The embed system now ensures:

- ✅ Each company has a unique embed code
- ✅ Only authorized users can access their company's dashboard
- ✅ Cross-company access is prevented
- ✅ Clear error messages guide users
- ✅ Security without compromising user experience

**Key Takeaway:** Even if someone copies your embed code, they can only see the login page, and only users from your company can successfully authenticate and access the dashboard.
