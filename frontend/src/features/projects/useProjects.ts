import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createProject, listProjects } from "@/api/projects";

export const projectsQueryKeys = {
  all: ["projects"] as const,
  list: (page: number, limit: number) => [...projectsQueryKeys.all, page, limit] as const,
};

export function useProjectsQuery(page: number, limit: number) {
  return useQuery({
    queryKey: projectsQueryKeys.list(page, limit),
    queryFn: () => listProjects({ page, limit }),
    placeholderData: (previousData) => previousData,
  });
}

export function useCreateProjectMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: createProject,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: projectsQueryKeys.all });
    },
  });
}
