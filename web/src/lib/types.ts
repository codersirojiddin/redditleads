export interface User {
  id: string;
  email: string;
}

export interface AuthResponse {
  token: string;
  user: User;
}

export interface Project {
  id: string;
  name: string;
  product_description: string;
  target_url: string;
  created_at: string;
}

export interface Keyword {
  id: string;
  keyword: string;
}

export type ReportStatus = "pending" | "running" | "completed" | "failed";

export interface Report {
  id: string;
  project_id: string;
  status: ReportStatus;
  error?: string;
  created_at: string;
  completed_at?: string;
}

export interface Opportunity {
  id: string;
  post_title: string;
  post_url: string;
  subreddit: string;
  author: string;
  relevance_score: number;
  intent_score: number;
  total_score: number;
  reasoning: string;
  suggested_reply: string;
}

export interface ReportDetail extends Report {
  opportunities: Opportunity[];
}

export interface ApiErrorBody {
  code: string;
  message: string;
}
