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

// Super Admin APIs
export const superAdminAPI = {
  login: (data: LoginRequest) =>
    api.post<SuperAdminAuthResponse>("/super-admin/login", data),
  getStats: () => api.get<{ stats: SuperAdminStats }>("/super-admin/stats"),
  getAllCompanies: () =>
    api.get<{ companies: CompanyWithStats[] }>("/super-admin/companies"),
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
