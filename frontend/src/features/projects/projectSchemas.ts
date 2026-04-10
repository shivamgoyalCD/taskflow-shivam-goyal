import { z } from "zod";

export const createProjectSchema = z.object({
  name: z.string().trim().min(1, "Project name is required"),
  description: z
    .string()
    .transform((value) => value.trim())
    .optional(),
});

export type CreateProjectFormValues = z.infer<typeof createProjectSchema>;
