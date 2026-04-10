import { useMutation, useQueryClient } from "@tanstack/react-query";
import type { ProjectDetail, ProjectStats } from "@/api/projects";
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

type UpdateTaskStatusMutationInput = {
  task: Task;
  status: Task["status"];
};

type UseUpdateTaskStatusMutationOptions = {
  onRollback?: (message: string) => void;
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

export function useUpdateTaskStatusMutation(
  projectId: string,
  options?: UseUpdateTaskStatusMutationOptions,
) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ task, status }: UpdateTaskStatusMutationInput) =>
      updateTask(task.id, { status }),
    onMutate: async ({ task, status }) => {
      await Promise.all([
        queryClient.cancelQueries({
          queryKey: projectDetailQueryKeys.detail(projectId),
        }),
        queryClient.cancelQueries({
          queryKey: projectDetailQueryKeys.stats(projectId),
        }),
      ]);

      const previousProject = queryClient.getQueryData<ProjectDetail>(
        projectDetailQueryKeys.detail(projectId),
      );
      const previousStats = queryClient.getQueryData<ProjectStats>(
        projectDetailQueryKeys.stats(projectId),
      );

      if (previousProject) {
        queryClient.setQueryData<ProjectDetail>(
          projectDetailQueryKeys.detail(projectId),
          {
            ...previousProject,
            tasks: previousProject.tasks.map((currentTask) =>
              currentTask.id === task.id
                ? {
                    ...currentTask,
                    status,
                    updated_at: new Date().toISOString(),
                  }
                : currentTask,
            ),
          },
        );
      }

      if (previousStats && task.status !== status) {
        queryClient.setQueryData<ProjectStats>(
          projectDetailQueryKeys.stats(projectId),
          {
            ...previousStats,
            status_counts: {
              ...previousStats.status_counts,
              [task.status]: Math.max(0, previousStats.status_counts[task.status] - 1),
              [status]: previousStats.status_counts[status] + 1,
            },
          },
        );
      }

      return { previousProject, previousStats };
    },
    onError: (_error, _variables, context) => {
      if (context?.previousProject) {
        queryClient.setQueryData(projectDetailQueryKeys.detail(projectId), context.previousProject);
      }

      if (context?.previousStats) {
        queryClient.setQueryData(projectDetailQueryKeys.stats(projectId), context.previousStats);
      }

      options?.onRollback?.("Task status update failed. Changes were reverted.");
    },
    onSuccess: (updatedTask) => {
      const currentProject = queryClient.getQueryData<ProjectDetail>(
        projectDetailQueryKeys.detail(projectId),
      );

      if (!currentProject) {
        return;
      }

      queryClient.setQueryData<ProjectDetail>(projectDetailQueryKeys.detail(projectId), {
        ...currentProject,
        tasks: currentProject.tasks.map((currentTask) =>
          currentTask.id === updatedTask.id ? updatedTask : currentTask,
        ),
      });
    },
    onSettled: async () => {
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: projectDetailQueryKeys.detail(projectId),
        }),
        queryClient.invalidateQueries({
          queryKey: projectDetailQueryKeys.stats(projectId),
        }),
      ]);
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
