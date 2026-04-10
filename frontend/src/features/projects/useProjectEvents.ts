import { useQueryClient } from "@tanstack/react-query";
import { useEffect, useState } from "react";
import { buildApiUrl } from "@/api/client";
import type { ProjectDetail, ProjectStats } from "@/api/projects";
import type { Task } from "@/api/tasks";
import { projectDetailQueryKeys } from "@/features/projects/useProjectDetail";

type ProjectTaskEvent = {
  type: "task_created" | "task_updated" | "task_deleted";
  project_id: string;
  task?: EventTaskPayload;
  task_id?: string;
};

type EventTaskPayload = {
  id: string;
  title: string;
  description: string | null;
  status: string;
  priority: string;
  project_id: string;
  assignee_id: string | null;
  creator_id: string;
  due_date: string | null;
  created_at: string;
  updated_at: string;
};

export type ProjectEventsConnectionState =
  | "disconnected"
  | "connecting"
  | "connected"
  | "reconnecting";

const eventTypes: ProjectTaskEvent["type"][] = [
  "task_created",
  "task_updated",
  "task_deleted",
];

export function useProjectEvents(projectId: string, token: string | null) {
  const queryClient = useQueryClient();
  const [liveStatus, setLiveStatus] = useState<ProjectEventsConnectionState>("disconnected");

  useEffect(() => {
    if (
      !projectId ||
      !token ||
      typeof window === "undefined" ||
      typeof window.EventSource === "undefined"
    ) {
      setLiveStatus("disconnected");
      return;
    }

    const query = new URLSearchParams({
      access_token: token,
    });
    const eventSource = new window.EventSource(
      buildApiUrl(`/projects/${projectId}/events?${query.toString()}`),
    );

    let isClosed = false;
    setLiveStatus("connecting");

    const handleOpen = () => {
      if (isClosed) {
        return;
      }

      setLiveStatus("connected");
    };

    const handleError = () => {
      if (isClosed) {
        return;
      }

      setLiveStatus("reconnecting");
    };

    const handleTaskEvent = (message: MessageEvent<string>) => {
      const parsedEvent = parseProjectTaskEvent(message.data);
      if (!parsedEvent || parsedEvent.project_id !== projectId) {
        return;
      }

      applyProjectTaskEvent(queryClient, projectId, parsedEvent);
    };

    eventSource.onopen = handleOpen;
    eventSource.onerror = handleError;

    for (const eventType of eventTypes) {
      eventSource.addEventListener(eventType, handleTaskEvent as EventListener);
    }

    return () => {
      isClosed = true;
      setLiveStatus("disconnected");

      for (const eventType of eventTypes) {
        eventSource.removeEventListener(eventType, handleTaskEvent as EventListener);
      }

      eventSource.close();
    };
  }, [projectId, queryClient, token]);

  return {
    liveStatus,
    isLiveConnected: liveStatus === "connected",
  };
}

function applyProjectTaskEvent(
  queryClient: ReturnType<typeof useQueryClient>,
  projectId: string,
  event: ProjectTaskEvent,
) {
  const detailKey = projectDetailQueryKeys.detail(projectId);
  const statsKey = projectDetailQueryKeys.stats(projectId);
  const currentProject = queryClient.getQueryData<ProjectDetail>(detailKey);
  const currentStats = queryClient.getQueryData<ProjectStats>(statsKey);
  const nextTask = event.task ? mapEventTaskToTask(event.task) : null;

  if (!currentProject) {
    return;
  }

  const nextTasks = applyEventToTasks(currentProject.tasks, event, nextTask);
  queryClient.setQueryData<ProjectDetail>(detailKey, {
    ...currentProject,
    tasks: nextTasks,
  });

  if (!currentStats) {
    return;
  }

  queryClient.setQueryData<ProjectStats>(
    statsKey,
    buildStatsFromTasks(projectId, nextTasks, currentStats),
  );
}

function applyEventToTasks(
  tasks: Task[],
  event: ProjectTaskEvent,
  nextTask: Task | null,
) {
  if (event.type === "task_deleted") {
    return tasks.filter((task) => task.id !== event.task_id);
  }

  if (!nextTask) {
    return tasks;
  }

  const existingTaskIndex = tasks.findIndex((task) => task.id === nextTask.id);
  if (existingTaskIndex === -1) {
    return [...tasks, nextTask];
  }

  return tasks.map((task) => (task.id === nextTask.id ? nextTask : task));
}

function buildStatsFromTasks(
  projectId: string,
  tasks: Task[],
  previousStats: ProjectStats,
): ProjectStats {
  const assigneeNameByID = new Map<string, string | null>();
  for (const entry of previousStats.assignee_counts) {
    if (entry.assignee_id) {
      assigneeNameByID.set(entry.assignee_id, entry.assignee_name);
    }
  }

  const statusCounts: ProjectStats["status_counts"] = {
    todo: 0,
    in_progress: 0,
    done: 0,
  };
  const assigneeCounts = new Map<
    string,
    {
      assignee_id: string | null;
      assignee_name: string | null;
      count: number;
    }
  >();

  for (const task of tasks) {
    statusCounts[task.status] += 1;

    const key = task.assignee_id ?? "unassigned";
    const existingEntry = assigneeCounts.get(key);
    if (existingEntry) {
      existingEntry.count += 1;
      continue;
    }

    assigneeCounts.set(key, {
      assignee_id: task.assignee_id,
      assignee_name: task.assignee_id ? assigneeNameByID.get(task.assignee_id) ?? null : null,
      count: 1,
    });
  }

  return {
    project_id: projectId,
    total_tasks: tasks.length,
    status_counts: statusCounts,
    assignee_counts: Array.from(assigneeCounts.values()).sort((left, right) => {
      if (left.assignee_name === null && right.assignee_name !== null) {
        return 1;
      }
      if (left.assignee_name !== null && right.assignee_name === null) {
        return -1;
      }
      return (right.count - left.count) || (left.assignee_name ?? "").localeCompare(right.assignee_name ?? "");
    }),
  };
}

function parseProjectTaskEvent(rawValue: string): ProjectTaskEvent | null {
  try {
    const parsedValue = JSON.parse(rawValue) as Partial<ProjectTaskEvent>;
    if (
      !parsedValue ||
      typeof parsedValue.type !== "string" ||
      typeof parsedValue.project_id !== "string"
    ) {
      return null;
    }

    if (!eventTypes.includes(parsedValue.type as ProjectTaskEvent["type"])) {
      return null;
    }

    return parsedValue as ProjectTaskEvent;
  } catch {
    return null;
  }
}

function mapEventTaskToTask(task: EventTaskPayload): Task | null {
  if (!isTaskStatus(task.status) || !isTaskPriority(task.priority)) {
    return null;
  }

  return {
    id: task.id,
    title: task.title,
    description: task.description,
    status: task.status,
    priority: task.priority,
    project_id: task.project_id,
    assignee_id: task.assignee_id,
    creator_id: task.creator_id,
    due_date: task.due_date,
    created_at: task.created_at,
    updated_at: task.updated_at,
  };
}

function isTaskStatus(value: string): value is Task["status"] {
  return value === "todo" || value === "in_progress" || value === "done";
}

function isTaskPriority(value: string): value is Task["priority"] {
  return value === "low" || value === "medium" || value === "high";
}
