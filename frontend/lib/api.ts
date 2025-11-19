import axios from "axios";

const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api";

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    "Content-Type": "application/json",
  },
});

// Add token to requests if available
api.interceptors.request.use((config) => {
  if (typeof window !== "undefined") {
    // Check for super admin token first, then regular admin token
    const superAdminToken = localStorage.getItem("super_admin_token");
    const adminToken = localStorage.getItem("token");
    const token = superAdminToken || adminToken;
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
  }
  return config;
});

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

export interface Admin {
  id: string;
  name: string;
  email: string;
  company_id: string;
}

export interface AuthResponse {
  token: string;
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
  created_at: string;
  updated_at: string;
}

export interface Application {
  id: string;
  job_id: string;
  full_name: string;
  email: string;
  phone?: string;
  resume_url: string;
  cover_letter?: string;
  years_of_experience: number;
  current_position?: string;
  linkedin_url?: string;
  portfolio_url?: string;
  status: string;
  score: number; // AI match score 0-100
  analysis_result?: CVAnalysisResult | string; // AI analysis details (can be string JSON or parsed object)
  applied_at: string;
  reviewed_at?: string;
  job?: Job;
}

// Auth APIs
export const authAPI = {
  register: (data: RegisterRequest) =>
    api.post<AuthResponse>("/auth/register", data),
  login: (data: LoginRequest) => api.post<AuthResponse>("/auth/login", data),
};

// Super Admin Types
export interface SuperAdmin {
  id: string;
  name: string;
  email: string;
}

export interface SuperAdminAuthResponse {
  token: string;
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
  action_type: string; // company_registered, job_created, job_updated, job_deleted, application_shortlisted, etc.
  entity_type: string; // company, job, application, etc.
  entity_id?: string;
  description: string;
  metadata?: any; // JSON object with additional details
  created_at: string;
  admin?: Admin;
  company?: Company;
}

// Super Admin APIs
export const superAdminAPI = {
  login: (data: LoginRequest) =>
    api.post<SuperAdminAuthResponse>("/super-admin/login", data),
  getStats: () => api.get<{ stats: SuperAdminStats }>("/super-admin/stats"),
  getAllCompanies: () =>
    api.get<{ companies: CompanyWithStats[] }>("/super-admin/companies"),
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
    api.get<{ jobs: Job[] }>(`/jobs/public/${companyId}`),
  update: (id: string, data: Partial<Job>) =>
    api.put<{ message: string; job: Job }>(`/jobs/${id}`, data),
  delete: (id: string) => api.delete<{ message: string }>(`/jobs/${id}`),
};

// Application APIs
export const applicationAPI = {
  submit: (data: Partial<Application>) =>
    api.post<{ message: string; application: Application }>(
      "/applications",
      data
    ),
  getAll: (params?: {
    job_id?: string;
    status?: string;
    date_from?: string;
    date_to?: string;
  }) => api.get<{ applications: Application[] }>("/applications", { params }),
  shortlist: (id: string) =>
    api.put<{ message: string; application: Application }>(
      `/applications/${id}/shortlist`
    ),
  reject: (id: string) =>
    api.put<{ message: string }>(`/applications/${id}/reject`),
};

// Candidate Portal Types
export interface ApplicationStatus {
  id: string;
  full_name: string;
  email: string;
  status: string;
  applied_at: string;
  reviewed_at?: string;
  score: number;
  job: {
    id: string;
    title: string;
    company_name?: string;
  };
}

// Candidate Portal APIs (public, no auth - use separate axios instance)
const publicApi = axios.create({
  baseURL: API_BASE_URL,
  // No auth interceptor for public endpoints
});

// Candidate Portal APIs (public, no auth)
export const candidatePortalAPI = {
  checkStatus: (email: string, applicationId: string) =>
    publicApi.post<{ application: ApplicationStatus }>("/candidate/status", {
      email,
      application_id: applicationId,
    }),
  getByEmail: (email: string) =>
    publicApi.get<{ applications: ApplicationStatus[]; count: number }>(
      "/candidate/applications",
      { params: { email } }
    ),
};

// CV Matching Types (Local matching, no AI required)
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
  threshold?: number;
}

// AI Shortlisting APIs
export const aiShortlistAPI = {
  analyze: (data: AIShortlistRequest) =>
    api.post<{
      message: string;
      application: Application;
      analysis: CVAnalysisResult;
      auto_shortlisted: boolean;
    }>("/applications/ai-shortlist", data),
  analyzeBatch: (data: BatchAIShortlistRequest) =>
    api.post<{
      message: string;
      total_analyzed: number;
      shortlisted_count: number;
      results: Array<{
        application_id: string;
        candidate_name: string;
        match_score: number;
        shortlisted: boolean;
        analysis: CVAnalysisResult;
      }>;
    }>("/applications/ai-shortlist-batch", data),
};

// File Upload APIs (using separate axios instance without default Content-Type)
const uploadApi = axios.create({
  baseURL: API_BASE_URL,
});

export const uploadAPI = {
  uploadCV: (file: File) => {
    const formData = new FormData();
    formData.append("file", file);
    return uploadApi.post<{ message: string; file_url: string }>(
      "/upload/cv",
      formData
    );
  },
  uploadPortfolio: (file: File) => {
    const formData = new FormData();
    formData.append("file", file);
    return uploadApi.post<{ message: string; file_url: string }>(
      "/upload/portfolio",
      formData
    );
  },
};

export default api;
