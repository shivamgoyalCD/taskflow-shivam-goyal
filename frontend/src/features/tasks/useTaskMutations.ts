import { useMutation, useQueryClient } from "@tanstack/react-query";
import { createProjectTask, deleteTask, updateTask, type Task } from "@/api/tasks";
import { projectDetailQueryKeys } from "@/features/projects/useProjectDetail";
import { projectsQueryKeys } from "@/features/projects/useProjects";

export type SaveTaskPayload = {
  title: string;
  description: string | null;
  status: Task["status"];
  priority: Task["priority"];
  assignee_id: string | null;
  due_date: string | null;
};

type UpdateTaskMutationInput = {
  taskId: string;
  payload: SaveTaskPayload;
};

type DeleteTaskMutationInput = {
  taskId: string;
};

export function useCreateTaskMutation(projectId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (payload: SaveTaskPayload) => {
      const createdTask = await createProjectTask(projectId, {
        title: payload.title,
        description: payload.description,
        assignee_id: payload.assignee_id,
        due_date: payload.due_date,
      });

      if (payload.status === "todo" && payload.priority === "medium") {
        return createdTask;
      }

      return updateTask(createdTask.id, {
        status: payload.status,
        priority: payload.priority,
      });
    },
    onSuccess: async () => {
      await invalidateProjectTaskQueries(queryClient, projectId);
    },
  });
}

export function useUpdateTaskMutation(projectId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ taskId, payload }: UpdateTaskMutationInput) =>
      updateTask(taskId, {
        title: payload.title,
        description: payload.description,
        status: payload.status,
        priority: payload.priority,
        assignee_id: payload.assignee_id,
        due_date: payload.due_date,
      }),
    onSuccess: async () => {
      await invalidateProjectTaskQueries(queryClient, projectId);
    },
  });
}

export function useDeleteTaskMutation(projectId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ taskId }: DeleteTaskMutationInput) => deleteTask(taskId),
    onSuccess: async () => {
      await invalidateProjectTaskQueries(queryClient, projectId);
    },
  });
}

async function invalidateProjectTaskQueries(
  queryClient: ReturnType<typeof useQueryClient>,
  projectId: string,
) {
  await Promise.all([
    queryClient.invalidateQueries({
      queryKey: projectDetailQueryKeys.detail(projectId),
    }),
    queryClient.invalidateQueries({
      queryKey: projectDetailQueryKeys.stats(projectId),
    }),
    queryClient.invalidateQueries({
      queryKey: projectsQueryKeys.all,
    }),
  ]);
}
