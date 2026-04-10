import {
  closestCorners,
  DndContext,
  DragOverlay,
  PointerSensor,
  TouchSensor,
  useSensor,
  useSensors,
  type DragEndEvent,
  type DragStartEvent,
} from "@dnd-kit/core";
import { Grid2 } from "@mui/material";
import { useMemo, useState } from "react";
import type { Task } from "@/api/tasks";
import { DraggableTaskCard } from "@/features/tasks/DraggableTaskCard";
import { StatusColumn } from "@/features/tasks/StatusColumn";

type TaskBoardProps = {
  tasksByStatus: Record<Task["status"], Task[]>;
  activeStatusTaskId: string | null;
  getAssigneeLabel: (task: Task) => string;
  onTaskStatusChange: (task: Task, status: Task["status"]) => void | Promise<void>;
  onEditTask: (task: Task) => void;
  onDeleteTask: (task: Task) => void;
};

export function TaskBoard({
  tasksByStatus,
  activeStatusTaskId,
  getAssigneeLabel,
  onTaskStatusChange,
  onEditTask,
  onDeleteTask,
}: TaskBoardProps) {
  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: {
        distance: 8,
      },
    }),
    useSensor(TouchSensor, {
      activationConstraint: {
        delay: 150,
        tolerance: 6,
      },
    }),
  );

  const [activeTaskId, setActiveTaskId] = useState<string | null>(null);

  const allTasks = useMemo(
    () =>
      boardStatuses.flatMap((status) => {
        return tasksByStatus[status];
      }),
    [tasksByStatus],
  );

  const activeTask = activeTaskId ? allTasks.find((task) => task.id === activeTaskId) ?? null : null;

  function handleDragStart(event: DragStartEvent) {
    const task = event.active.data.current?.task as Task | undefined;
    setActiveTaskId(task?.id ?? null);
  }

  function handleDragEnd(event: DragEndEvent) {
    const task = event.active.data.current?.task as Task | undefined;
    const targetStatus = isTaskStatus(event.over?.id) ? event.over.id : null;

    setActiveTaskId(null);

    if (!task || !targetStatus || targetStatus === task.status) {
      return;
    }

    void onTaskStatusChange(task, targetStatus);
  }

  return (
    <DndContext
      sensors={sensors}
      collisionDetection={closestCorners}
      onDragStart={handleDragStart}
      onDragCancel={() => setActiveTaskId(null)}
      onDragEnd={handleDragEnd}
    >
      <Grid2 container spacing={3}>
        {boardStatuses.map((status) => (
          <Grid2 key={status} size={{ xs: 12, xl: 4 }}>
            <StatusColumn status={status} count={tasksByStatus[status].length}>
              {tasksByStatus[status].map((task) => (
                <DraggableTaskCard
                  key={task.id}
                  task={task}
                  assigneeLabel={getAssigneeLabel(task)}
                  isStatusUpdating={activeStatusTaskId === task.id}
                  onEdit={() => onEditTask(task)}
                  onDelete={() => onDeleteTask(task)}
                />
              ))}
            </StatusColumn>
          </Grid2>
        ))}
      </Grid2>

      <DragOverlay>
        {activeTask ? (
          <DraggableTaskCard
            task={activeTask}
            assigneeLabel={getAssigneeLabel(activeTask)}
            isStatusUpdating={false}
            overlay
            onEdit={() => undefined}
            onDelete={() => undefined}
          />
        ) : null}
      </DragOverlay>
    </DndContext>
  );
}

const boardStatuses: Task["status"][] = ["todo", "in_progress", "done"];

function isTaskStatus(value: unknown): value is Task["status"] {
  return value === "todo" || value === "in_progress" || value === "done";
}
