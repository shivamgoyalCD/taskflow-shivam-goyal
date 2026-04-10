import { useQuery } from "@tanstack/react-query";
import { getProject, getProjectStats } from "@/api/projects";

export const projectDetailQueryKeys = {
  detail: (projectId: string) => ["project", projectId] as const,
  stats: (projectId: string) => ["project", projectId, "stats"] as const,
};

export function useProjectDetailQuery(projectId: string) {
  return useQuery({
    queryKey: projectDetailQueryKeys.detail(projectId),
    queryFn: () => getProject(projectId),
    enabled: Boolean(projectId),
  });
}

export function useProjectStatsQuery(projectId: string) {
  return useQuery({
    queryKey: projectDetailQueryKeys.stats(projectId),
    queryFn: () => getProjectStats(projectId),
    enabled: Boolean(projectId),
  });
}
