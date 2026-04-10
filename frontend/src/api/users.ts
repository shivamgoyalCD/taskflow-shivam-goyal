import { apiClient } from "@/api/client";

export type UserSummary = {
  id: string;
  name: string;
  email: string;
};

export type ListUsersResponse = {
  users: UserSummary[];
};

export function listUsers() {
  return apiClient.get<ListUsersResponse>("/users");
}
