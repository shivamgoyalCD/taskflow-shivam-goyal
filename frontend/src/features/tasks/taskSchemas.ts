import { z } from "zod";

const datePattern = /^\d{4}-\d{2}-\d{2}$/;

export const taskFormSchema = z.object({
  title: z
    .string()
    .trim()
    .min(1, "Title is required.")
    .max(160, "Title must be 160 characters or fewer."),
  description: z
    .string()
    .max(2000, "Description must be 2000 characters or fewer.")
    .optional()
    .or(z.literal("")),
  status: z.enum(["todo", "in_progress", "done"]),
  priority: z.enum(["low", "medium", "high"]),
  assignee_id: z.string(),
  due_date: z
    .string()
    .refine((value) => value === "" || datePattern.test(value), "Enter a valid due date.")
    .optional()
    .or(z.literal("")),
});

export type TaskFormValues = z.infer<typeof taskFormSchema>;

export const defaultTaskFormValues: TaskFormValues = {
  title: "",
  description: "",
  status: "todo",
  priority: "medium",
  assignee_id: "unassigned",
  due_date: "",
};
