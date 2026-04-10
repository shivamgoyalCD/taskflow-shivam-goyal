import { apiClient } from "@/api/client";

export type Task = {
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
};

export type TaskListResponse = {
  tasks: Task[];
  page: number;
  limit: number;
  status?: string;
  assignee_id?: string;
};

export type ListProjectTasksParams = {
  status?: Task["status"];
  assignee?: string;
  page?: number;
  limit?: number;
};

export type CreateTaskPayload = {
  title: string;
  description?: string | null;
  assignee_id?: string | null;
  due_date?: string | null;
};

export type UpdateTaskPayload = {
  title?: string | null;
  description?: string | null;
  status?: Task["status"];
  priority?: Task["priority"];
  assignee_id?: string | null;
  due_date?: string | null;
};

export function listProjectTasks(projectId: string, params: ListProjectTasksParams = {}) {
  const query = new URLSearchParams();
  if (params.status) {
    query.set("status", params.status);
  }
  if (params.assignee) {
    query.set("assignee", params.assignee);
  }
  if (params.page) {
    query.set("page", String(params.page));
  }
  if (params.limit) {
    query.set("limit", String(params.limit));
  }

  const suffix = query.size > 0 ? `?${query.toString()}` : "";
  return apiClient.get<TaskListResponse>(`/projects/${projectId}/tasks${suffix}`);
}

export function createProjectTask(projectId: string, payload: CreateTaskPayload) {
  return apiClient.post<Task>(`/projects/${projectId}/tasks`, payload);
}

export function updateTask(taskId: string, payload: UpdateTaskPayload) {
  return apiClient.patch<Task>(`/tasks/${taskId}`, payload);
}

export function deleteTask(taskId: string) {
	return apiClient.delete<null>(`/tasks/${taskId}`);
}
