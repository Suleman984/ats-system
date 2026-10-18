import axios, { isAxiosError } from "axios";

export const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api";

// The session lives in an HttpOnly cookie, so requests just need credentials. The client header
// is required by the API on writes (CSRF guard) and can't be sent cross-site without CORS approval.
const api = axios.create({
  baseURL: API_BASE_URL,
  withCredentials: true,
  headers: {
    "Content-Type": "application/json",
    "X-ATS-Client": "web",
  },
});

// Local copies of the signed-in user (display data only; never the token).
export const ADMIN_USER_KEY = "user";
export const SUPER_ADMIN_USER_KEY = "super_admin";

// A 401 means the session is over: forget the user and return to the matching login.
api.interceptors.response.use(
  (response) => response,
  (error) => {
    const url: string = error.config?.url ?? "";
    const isAuthCall = url.startsWith("/auth/") || url === "/super-admin/login";
    if (typeof window !== "undefined" && !isAuthCall && error.response?.status === 401) {
      const { pathname, search } = window.location;
      let loginPath = "/admin/login";
      if (url.startsWith("/super-admin")) {
        localStorage.removeItem(SUPER_ADMIN_USER_KEY);
        loginPath = "/super-admin/login";
      } else {
        localStorage.removeItem(ADMIN_USER_KEY);
        if (pathname.startsWith("/embed")) {
          const companyId = new URLSearchParams(search).get("company_id");
          loginPath = companyId
            ? `/embed/login?company_id=${encodeURIComponent(companyId)}`
            : "/embed/login";
        }
      }
      if (!pathname.startsWith(loginPath.split("?")[0])) window.location.assign(loginPath);
    }
    return Promise.reject(error);
  }
);

// Public endpoints use a separate instance without credentials, so session cookies are never sent.
const publicApi = axios.create({
  baseURL: API_BASE_URL,
});

// Machine-readable error code from the API, e.g. "email_unverified" or "plan_limit".
export function errorCode(error: unknown): string | undefined {
  return isAxiosError<{ code?: string }>(error) ? error.response?.data?.code : undefined;
}

export function errorMessage(error: unknown, fallback: string): string {
  if (isAxiosError<{ error?: string; message?: string }>(error)) {
    return error.response?.data?.message || error.response?.data?.error || fallback;
  }
  return fallback;
}

// Deadlines are calendar dates; parsing them as local time keeps them from shifting a day west of UTC.
export function formatDateOnly(date: string): string {
  const [y, m, d] = date.slice(0, 10).split("-").map(Number);
  return new Date(y, m - 1, d).toLocaleDateString();
}

// Types
export interface RegisterRequest {
  company_name: string;
  email: string;
  password: string;
  name: string;
  embedded_mode?: boolean;
  embed_domain?: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface Company {
  id: string;
  company_name: string;
  email: string;
  company_website?: string;
  embedded_mode?: boolean;
  embed_domain?: string;
  subscription_status: string;
  subscription_tier: string;
  created_at: string;
  updated_at: string;
}

export type AdminRole = "owner" | "recruiter";

export interface Admin {
  id: string;
  name: string;
  email: string;
  company_id: string;
  role: AdminRole;
}

export interface AuthResponse {
  admin: Admin;
  message?: string;
}

export interface Job {
  id: string;
  company_id: string;
  title: string;
  description: string;
  requirements?: string;
  location?: string;
  job_type?: string;
  salary_range?: string;
  deadline: string;
  status: string;
  auto_shortlist: boolean;
  shortlist_criteria?: string | Record<string, unknown>;
  created_at: string;
  updated_at: string;
}

// The candidate-facing subset of Job returned by the public endpoints.
export type PublicJob = Pick<
  Job,
  "id" | "title" | "description" | "requirements" | "location" | "job_type" | "salary_range" | "deadline" | "created_at"
>;

export interface Application {
  id: string;
  job_id?: string;
  company_id?: string;
  full_name: string;
  email: string;
  phone?: string;
  // An external link; uploaded files are private and opened via openApplicationFile.
  resume_url: string;
  has_resume_file?: boolean;
  has_portfolio_file?: boolean;
  unread_messages?: number;
  cover_letter?: string;
  years_of_experience: number;
  current_position?: string;
  linkedin_url?: string;
  portfolio_url?: string;
  status: string;
  score: number;
  analysis_result?: CVAnalysisResult | string;
  applied_at: string;
  reviewed_at?: string;
  cv_viewed_at?: string;
  referral_source?: string;
  referred_by_name?: string;
  referred_by_email?: string;
  referred_by_phone?: string;
  in_talent_pool?: boolean;
  talent_pool_added_at?: string;
  job?: Job;
}

// Auth APIs
export const authAPI = {
  // Registration never signs in: the account activates from the emailed verification link.
  register: (data: RegisterRequest) =>
    publicApi.post<{ message: string }>("/auth/register", data),
  verifyEmail: (token: string) => api.post<AuthResponse>("/auth/verify-email", { token }),
  resendVerification: (email: string) =>
    publicApi.post<{ message: string }>("/auth/verify-email/resend", { email }),
  login: (data: LoginRequest) => api.post<AuthResponse>("/auth/login", data),
  logout: () => api.post<{ message: string }>("/auth/logout"),
  me: () => api.get<{ admin: Admin }>("/auth/me"),
  getInvite: (token: string) =>
    publicApi.get<{ email: string; role: AdminRole; company_name: string }>("/auth/invites", {
      params: { token },
    }),
  acceptInvite: (token: string, name: string, password: string) =>
    api.post<AuthResponse>("/auth/invites/accept", { token, name, password }),
  forgotPassword: (email: string) =>
    publicApi.post<{ message: string }>("/auth/forgot-password", { email }),
  resetPassword: (token: string, password: string) =>
    publicApi.post<{ message: string }>("/auth/reset-password", { token, password }),
};

// Super Admin Types
export interface SuperAdmin {
  id: string;
  name: string;
  email: string;
}

export interface SuperAdminAuthResponse {
  super_admin: SuperAdmin;
  message?: string;
}

export interface SuperAdminStats {
  total_companies: number;
  active_companies: number;
  total_jobs: number;
  open_jobs: number;
  total_applications: number;
  pending_applications: number;
  shortlisted_applications: number;
  total_admins: number;
}

export interface CompanyWithStats extends Company {
  job_count: number;
  application_count: number;
}

// Activity Log Types
export interface ActivityLog {
  id: string;
  company_id?: string;
  admin_id?: string;
  action_type: string;
  entity_type: string;
  entity_id?: string;
  description: string;
  metadata?: string | Record<string, unknown>;
  created_at: string;
  admin?: Admin;
  company?: Company;
}

// Super Admin APIs
export const superAdminAPI = {
  login: (data: LoginRequest) =>
    api.post<SuperAdminAuthResponse>("/super-admin/login", data),
  logout: () => api.post<{ message: string }>("/super-admin/logout"),
  getStats: () => api.get<{ stats: SuperAdminStats }>("/super-admin/stats"),
  getAllCompanies: () =>
    api.get<{ companies: CompanyWithStats[] }>("/super-admin/companies"),
  updateCompany: (
    id: string,
    data: { subscription_status?: string; subscription_tier?: string }
  ) =>
    api.put<{ message: string; company: Company }>(`/super-admin/companies/${id}`, data),
};

// Activity Log APIs
export const activityLogAPI = {
  getAll: (params?: {
    action_type?: string;
    entity_type?: string;
    date_from?: string;
    date_to?: string;
  }) =>
    api.get<{ logs: ActivityLog[]; count: number }>("/activity-logs", {
      params,
    }),
};

// Candidate Search Types
export interface CandidateSearchRequest {
  query?: string;
  skills?: string[];
  min_experience?: number;
  max_experience?: number;
  current_position?: string;
  languages?: string[];
  has_portfolio?: boolean;
  has_linkedin?: boolean;
  status?: string;
  limit?: number;
}

export interface CandidateSearchResult {
  application: Application;
  match_score: number;
  matched_skills: string[];
  matched_reasons: string[];
}

export interface CandidateDetails {
  candidate: Application;
  cv_text: string;
  skills: string[];
  experience: number;
}

// Candidate Search APIs
export const candidateSearchAPI = {
  search: (data: CandidateSearchRequest) =>
    api.post<{
      candidates: CandidateSearchResult[];
      count: number;
      total: number;
    }>("/candidates/search", data),
  getDetails: (id: string) => api.get<CandidateDetails>(`/candidates/${id}`),
  reparseAll: () =>
    api.post<{ message: string; total: number }>("/candidates/reparse-all"),
  reparse: (id: string) =>
    api.post<{ message: string; characters: number }>(`/candidates/${id}/reparse`),
  addManual: (data: {
    job_id: string;
    full_name: string;
    email: string;
    phone?: string;
    resume_url?: string;
    resume_upload_id?: string;
    portfolio_upload_id?: string;
    cover_letter?: string;
    years_of_experience?: number;
    current_position?: string;
    linkedin_url?: string;
    portfolio_url?: string;
    status?: string;
    notes?: string;
  }) =>
    api.post<{ message: string; application: Application }>(
      "/candidates/manual",
      data
    ),
};

export const superAdminActivityLogAPI = {
  getAll: (params?: {
    company_id?: string;
    action_type?: string;
    entity_type?: string;
    date_from?: string;
    date_to?: string;
  }) =>
    api.get<{ logs: ActivityLog[]; count: number }>(
      "/super-admin/activity-logs",
      {
        params,
      }
    ),
};

// Job APIs
export const jobAPI = {
  create: (data: Partial<Job>) =>
    api.post<{ message: string; job: Job }>("/jobs", data),
  getAll: (params?: { status?: string }) =>
    api.get<{ jobs: Job[] }>("/jobs", { params }),
  getById: (id: string) => api.get<{ job: Job }>(`/jobs/${id}`),
  getPublic: (companyId: string) =>
    publicApi.get<{ jobs: PublicJob[]; company_name: string }>(`/jobs/public/${companyId}`),
  getPublicById: (id: string) =>
    publicApi.get<{
      job: PublicJob;
      company_id: string;
      company_name: string;
      accepting_applications: boolean;
    }>(`/public/jobs/${id}`),
  update: (id: string, data: Partial<Job>) =>
    api.put<{ message: string; job: Job }>(`/jobs/${id}`, data),
  delete: (id: string) => api.delete<{ message: string }>(`/jobs/${id}`),
};

export interface DashboardStats {
  total_jobs: number;
  open_jobs: number;
  total_applications: number;
  shortlisted: number;
}

export const messagesAPI = {
  unreadSummary: () =>
    api.get<{ total: number; by_job: Record<string, number> }>("/messages/unread-summary"),
};

export const dashboardAPI = {
  getStats: () =>
    api.get<{ stats: DashboardStats; recent_applications: Application[] }>(
      "/dashboard/stats"
    ),
};

// Application APIs
export const applicationAPI = {
  submit: (data: CandidateSubmission & { job_id: string }) =>
    publicApi.post<{ message: string; application: Application }>(
      "/applications",
      data
    ),
  getAll: (params?: {
    job_id?: string;
    status?: string;
    date_from?: string;
    date_to?: string;
    unread?: boolean;
    limit?: number;
    offset?: number;
  }) =>
    api.get<{ applications: Application[]; total: number }>("/applications", { params }),
  shortlist: (id: string) =>
    api.put<{ message: string; application: Application }>(
      `/applications/${id}/shortlist`
    ),
  reject: (id: string) =>
    api.put<{ message: string }>(`/applications/${id}/reject`),
  delete: (id: string) =>
    api.delete<{ message: string }>(`/applications/${id}`),
  bulkDelete: (status: string, jobId?: string) =>
    api.post<{ message: string; deleted_count: number; total_found: number }>(
      "/applications/bulk-delete",
      { status, job_id: jobId || undefined }
    ),
  getFileUrl: (id: string, kind: "resume" | "portfolio") =>
    api.get<{ url: string; expires_in: number }>(`/applications/${id}/files/${kind}`),
  advance: (id: string) =>
    api.put<{ message: string; application: Application }>(
      `/applications/${id}/advance`
    ),
  reopen: (id: string) =>
    api.put<{ message: string; application: Application }>(
      `/applications/${id}/reopen`
    ),
  trackCVView: (id: string) =>
    api.post<{ message: string; application: Application }>(
      `/applications/${id}/track-cv-view`
    ),
  getMessages: (id: string) =>
    api.get<{ messages: Message[]; count: number }>(
      `/applications/${id}/messages`
    ),
  sendMessage: (id: string, message: string) =>
    api.post<{ message: string; data: Message }>(
      `/applications/${id}/messages`,
      { message }
    ),
};

// CRM Types
export interface CandidateNote {
  id: string;
  application_id: string;
  admin_id: string;
  note: string;
  is_private: boolean;
  created_at: string;
  updated_at: string;
  admin?: Admin;
}

export interface RelationshipTimelineItem {
  type: string;
  title: string;
  description: string;
  timestamp: string;
  icon: string;
  author?: string;
  admin?: string;
  sender?: string;
  is_private?: boolean;
}

// CRM APIs
export const crmAPI = {
  addNote: (applicationId: string, note: string, isPrivate: boolean) =>
    api.post<{ message: string; note: CandidateNote }>("/crm/notes", {
      application_id: applicationId,
      note: note,
      is_private: isPrivate,
    }),
  getNotes: (applicationId: string) =>
    api.get<{ notes: CandidateNote[]; count: number }>(
      `/crm/applications/${applicationId}/notes`
    ),
  updateNote: (noteId: string, note: string, isPrivate: boolean) =>
    api.put<{ message: string; note: CandidateNote }>(`/crm/notes/${noteId}`, {
      note: note,
      is_private: isPrivate,
    }),
  deleteNote: (noteId: string) =>
    api.delete<{ message: string }>(`/crm/notes/${noteId}`),

  addToTalentPool: (applicationId: string) =>
    api.post<{ message: string; application: Application }>(
      "/crm/talent-pool",
      { application_id: applicationId }
    ),
  removeFromTalentPool: (applicationId: string) =>
    api.delete<{ message: string; application: Application }>(
      `/crm/talent-pool/${applicationId}`
    ),
  getTalentPool: () =>
    api.get<{ applications: Application[]; count: number }>("/crm/talent-pool"),

  updateReferral: (
    applicationId: string,
    referralSource: string,
    referredByName: string,
    referredByEmail: string,
    referredByPhone: string
  ) =>
    api.put<{ message: string; application: Application }>(
      `/crm/applications/${applicationId}/referral`,
      {
        referral_source: referralSource,
        referred_by_name: referredByName,
        referred_by_email: referredByEmail,
        referred_by_phone: referredByPhone,
      }
    ),

  getTimeline: (applicationId: string) =>
    api.get<{ timeline: RelationshipTimelineItem[]; count: number }>(
      `/crm/applications/${applicationId}/timeline`
    ),
};

// Candidate Portal Types
export interface ApplicationStatus {
  id: string;
  full_name: string;
  email: string;
  status: string;
  status_label?: string;
  applied_at: string;
  reviewed_at?: string;
  cv_viewed_at?: string;
  last_status_update?: string;
  expected_response_date?: string;
  expected_response_days?: number;
  unread_messages?: number;
  status_history?: Array<{
    status: string;
    label: string;
    timestamp: string | null;
    completed: boolean;
  }>;
  job: {
    id: string;
    title: string;
    company_name?: string;
  };
}

export interface Message {
  id: string;
  application_id: string;
  sender_type: "candidate" | "recruiter";
  sender_id?: string;
  sender_email: string;
  message: string;
  is_read: boolean;
  read_at?: string;
  created_at: string;
}

export const candidatePortalAPI = {
  checkStatus: (email: string, applicationId: string) =>
    publicApi.post<{ application: ApplicationStatus }>("/candidate/status", {
      email,
      application_id: applicationId,
    }),
  sendMessage: (applicationId: string, email: string, message: string) =>
    publicApi.post<{ message: string; data: Message }>(
      "/candidate/messages/send",
      {
        application_id: applicationId,
        sender_email: email,
        message: message,
      }
    ),
  streamToken: (applicationId: string, email: string) =>
    publicApi.post<{ token: string }>("/candidate/stream-token", {
      application_id: applicationId,
      email,
    }),
  getMessages: (applicationId: string, email: string) =>
    publicApi.post<{ messages: Message[]; count: number }>(
      "/candidate/messages",
      { application_id: applicationId, email }
    ),
};

// CV scoring is local keyword matching; no external AI service is called.
export interface CVAnalysisResult {
  match_score: number;
  skills: string[];
  experience: number;
  education?: string;
  languages: string[];
  summary: string;
  match_reason: string;
  missing_skills: string[];
  strengths: string[];
  skills_match?: number;
  experience_match?: number;
  language_match?: number;
}

export interface AIShortlistRequest {
  application_id: string;
  required_skills?: string[];
  min_experience?: number;
  required_languages?: string[];
  match_job_description?: boolean;
}

export interface BatchAIShortlistRequest {
  job_id: string;
  required_skills?: string[];
  min_experience?: number;
  required_languages?: string[];
  match_job_description?: boolean;
}

// AI Shortlisting APIs
export const aiShortlistAPI = {
  analyze: (data: AIShortlistRequest) =>
    api.post<{
      message: string;
      application: Application;
      analysis: CVAnalysisResult;
    }>("/applications/ai-shortlist", data),
  analyzeBatch: (data: BatchAIShortlistRequest) =>
    api.post<{
      message: string;
      total_analyzed: number;
      results: Array<{
        application_id: string;
        candidate_name: string;
        match_score?: number;
        analysis?: CVAnalysisResult;
        error?: string;
      }>;
    }>("/applications/ai-shortlist-batch", data),
};

// Team
export interface TeamInvite {
  id: string;
  email: string;
  role: AdminRole;
  expires_at: string;
  created_at: string;
}

export interface TeamMember extends Admin {
  created_at: string;
}

export const teamAPI = {
  get: () => api.get<{ members: TeamMember[]; invites: TeamInvite[] }>("/team"),
  invite: (email: string, role: AdminRole) =>
    api.post<{ message: string; invite: TeamInvite }>("/team/invites", { email, role }),
  revokeInvite: (id: string) => api.delete<{ message: string }>(`/team/invites/${id}`),
  setRole: (id: string, role: AdminRole) =>
    api.put<{ message: string }>(`/team/members/${id}/role`, { role }),
  remove: (id: string) => api.delete<{ message: string }>(`/team/members/${id}`),
};

// Billing
export interface SubscriptionPlan {
  id: string;
  name: string;
  price_pkr: number;
  features: string | string[];
  max_jobs: number;
  max_applications: number;
  max_admins: number;
  ai_shortlisting: boolean;
}

export interface PlanUsage {
  open_jobs: number;
  applications_this_month: number;
  admins: number;
  pending_invites: number;
}

export interface Payment {
  id: string;
  amount: number;
  currency: string;
  status: "pending" | "completed" | "failed" | "refunded";
  created_at: string;
}

export const billingAPI = {
  get: () =>
    api.get<{
      plans: SubscriptionPlan[];
      current_plan: SubscriptionPlan;
      usage: PlanUsage;
      payments: Payment[];
      subscription?: { current_period_end: string; plan_id: string };
      checkout_available: boolean;
    }>("/billing"),
  checkout: (plan: string) =>
    api.post<{ checkout_url: string; payment_id: string }>("/billing/checkout", { plan }),
  confirm: (paymentId: string) =>
    api.post<{ payment: Payment }>(`/billing/payments/${paymentId}/confirm`),
};

// Uploaded files are private; the API returns an ID to submit with the application, never a URL.
export interface UploadResult {
  upload_id: string;
  file_name: string;
  size: number;
}

// What a candidate (or a recruiter adding one) sends; exactly one CV source is required.
export interface CandidateSubmission {
  full_name: string;
  email: string;
  phone?: string;
  resume_url?: string;
  resume_upload_id?: string;
  portfolio_url?: string;
  portfolio_upload_id?: string;
  cover_letter?: string;
  years_of_experience?: number;
  current_position?: string;
  linkedin_url?: string;
}

// Opens a private CV/portfolio through a short-lived signed link. The tab is opened synchronously
// (inside the click) so popup blockers allow it, then pointed at the link once it arrives.
export async function openApplicationFile(id: string, kind: "resume" | "portfolio") {
  const win = window.open("", "_blank");
  try {
    const { data } = await applicationAPI.getFileUrl(id, kind);
    if (win) {
      win.opener = null;
      win.location.href = data.url;
    } else {
      window.location.assign(data.url);
    }
  } catch (error) {
    win?.close();
    throw error;
  }
}

// Uploads use their own instance so axios sets the multipart Content-Type.
const uploadApi = axios.create({
  baseURL: API_BASE_URL,
});

export const uploadAPI = {
  uploadCV: (file: File) => {
    const formData = new FormData();
    formData.append("file", file);
    return uploadApi.post<UploadResult>(
      "/upload/cv",
      formData
    );
  },
  uploadPortfolio: (file: File) => {
    const formData = new FormData();
    formData.append("file", file);
    return uploadApi.post<UploadResult>(
      "/upload/portfolio",
      formData
    );
  },
};

export default api;
