import type {
  ApiErrorBody,
  AuthResponse,
  Keyword,
  Project,
  Report,
  ReportDetail,
} from "./types";

const TOKEN_KEY = "reddit_leads_token";

// In local dev and single-domain deploys (docker-compose + Caddy), the
// frontend is served behind the same origin as the API, so a relative
// "/api/v1" path works via the proxy. On split-host free-tier deploys
// (e.g. frontend on Render Static Site / Netlify, API on a different Render
// service URL) set VITE_API_BASE_URL at build time to the API's full origin,
// e.g. "https://reddit-leads-api.onrender.com/api/v1".
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "/api/v1";

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY);
}

export function setToken(token: string) {
  localStorage.setItem(TOKEN_KEY, token);
}

export function clearToken() {
  localStorage.removeItem(TOKEN_KEY);
}

export class ApiError extends Error {
  status: number;
  code: string;

  constructor(status: number, body: ApiErrorBody) {
    super(body.message || "Request failed");
    this.status = status;
    this.code = body.code || "unknown_error";
  }
}

async function request<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const token = getToken();
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(options.headers as Record<string, string> | undefined),
  };
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  const res = await fetch(`${API_BASE_URL}${path}`, { ...options, headers });

  if (res.status === 204) {
    return undefined as T;
  }

  const body = await res.json().catch(() => ({}));

  if (!res.ok) {
    if (res.status === 401) {
      clearToken();
    }
    throw new ApiError(res.status, body as ApiErrorBody);
  }

  return body as T;
}

export const api = {
  register: (email: string, password: string) =>
    request<AuthResponse>("/auth/register", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    }),

  login: (email: string, password: string) =>
    request<AuthResponse>("/auth/login", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    }),

  me: () => request<{ id: string; email: string }>("/me"),

  listProjects: () => request<Project[]>("/projects/"),

  createProject: (input: {
    name: string;
    product_description: string;
    target_url?: string;
  }) =>
    request<Project>("/projects/", {
      method: "POST",
      body: JSON.stringify(input),
    }),

  getProject: (id: string) => request<Project>(`/projects/${id}/`),

  deleteProject: (id: string) =>
    request<void>(`/projects/${id}/`, { method: "DELETE" }),

  listKeywords: (projectId: string) =>
    request<Keyword[]>(`/projects/${projectId}/keywords/`),

  addKeyword: (projectId: string, keyword: string) =>
    request<Keyword>(`/projects/${projectId}/keywords/`, {
      method: "POST",
      body: JSON.stringify({ keyword }),
    }),

  deleteKeyword: (projectId: string, keywordId: string) =>
    request<void>(`/projects/${projectId}/keywords/${keywordId}`, {
      method: "DELETE",
    }),

  listReports: (projectId: string) =>
    request<Report[]>(`/projects/${projectId}/reports/`),

  requestReport: (projectId: string) =>
    request<Report>(`/projects/${projectId}/reports/`, { method: "POST" }),

  getReport: (reportId: string) =>
    request<ReportDetail>(`/reports/${reportId}`),
};
