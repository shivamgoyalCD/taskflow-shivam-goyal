import { apiClient } from "@/api/client";

export type Project = {
  id: string;
  name: string;
  description: string | null;
  owner_id: string;
  created_at: string;
};

export type ProjectListResponse = {
  projects: Project[];
  page: number;
  limit: number;
};

export type ProjectDetail = Project & {
  tasks: Array<{
    id: string;
    title: string;
    description: string | null;
    status: "todo" | "in_progress" | "done";
    priority: "low" | "medium" | "high";
    project_id: string;
    assignee_id: string | null;
    creator_id: string;
    due_date: string | null;
    created_at: string;
    updated_at: string;
  }>;
};

export type ProjectStats = {
  project_id: string;
  total_tasks: number;
  status_counts: {
    todo: number;
    in_progress: number;
    done: number;
  };
  assignee_counts: Array<{
    assignee_id: string | null;
    assignee_name: string | null;
    count: number;
  }>;
};

export type ListProjectsParams = {
  page?: number;
  limit?: number;
};

export type CreateProjectPayload = {
  name: string;
  description?: string | null;
};

export type UpdateProjectPayload = {
  name?: string;
  description?: string | null;
};

export function listProjects(params: ListProjectsParams = {}) {
  const query = new URLSearchParams();
  if (params.page) {
    query.set("page", String(params.page));
  }
  if (params.limit) {
    query.set("limit", String(params.limit));
  }

  const suffix = query.size > 0 ? `?${query.toString()}` : "";
  return apiClient.get<ProjectListResponse>(`/projects${suffix}`);
}

export function createProject(payload: CreateProjectPayload) {
  return apiClient.post<Project>("/projects", payload);
}

export function getProject(id: string) {
  return apiClient.get<ProjectDetail>(`/projects/${id}`);
}

export function getProjectStats(id: string) {
  return apiClient.get<ProjectStats>(`/projects/${id}/stats`);
}

export function updateProject(id: string, payload: UpdateProjectPayload) {
  return apiClient.patch<Project>(`/projects/${id}`, payload);
}

export function deleteProject(id: string) {
	return apiClient.delete<null>(`/projects/${id}`);
}
