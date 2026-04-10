import { zodResolver } from "@hookform/resolvers/zod";
import {
  Alert,
  Box,
  Button,
  Card,
  CardContent,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import { type DefaultValues, type FieldValues, type Path, useForm } from "react-hook-form";
import { type ZodType } from "zod";

type FieldConfig<T extends string> = {
  name: T;
  label: string;
  type?: string;
};

type AuthFormCardProps<TValues extends FieldValues> = {
  title: string;
  description: string;
  submitLabel: string;
  fields: Array<FieldConfig<Extract<keyof TValues, string>>>;
  schema: ZodType<TValues>;
};

export function AuthFormCard<TValues extends FieldValues>({
  title,
  description,
  submitLabel,
  fields,
  schema,
}: AuthFormCardProps<TValues>) {
  const defaultValues = Object.fromEntries(
    fields.map((field) => [field.name, ""]),
  ) as DefaultValues<TValues>;

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<TValues>({
    resolver: zodResolver(schema),
    defaultValues,
  });

  return (
    <Card sx={{ maxWidth: 520, mx: "auto" }}>
      <CardContent sx={{ p: { xs: 3, md: 4 } }}>
        <Stack spacing={3}>
          <Box>
            <Typography variant="h4" gutterBottom>
              {title}
            </Typography>
            <Typography color="text.secondary">{description}</Typography>
          </Box>

          <Alert severity="info">
            Frontend scaffold only. Form validation is wired, but API submission is not
            implemented yet.
          </Alert>

          <Stack
            component="form"
            spacing={2}
            onSubmit={handleSubmit(async () => {
              await Promise.resolve();
            })}
          >
            {fields.map((field) => (
              <TextField
                key={field.name}
                label={field.label}
                type={field.type ?? "text"}
                fullWidth
                error={Boolean(errors[field.name])}
                helperText={errors[field.name]?.message as string | undefined}
                {...register(field.name as unknown as Path<TValues>)}
              />
            ))}

            <Button type="submit" variant="contained" size="large" disabled={isSubmitting}>
              {submitLabel}
            </Button>
          </Stack>
        </Stack>
      </CardContent>
    </Card>
  );
}
