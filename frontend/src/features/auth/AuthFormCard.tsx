import { zodResolver } from "@hookform/resolvers/zod";
import {
  Alert,
  Box,
  Button,
  Card,
  CardContent,
  CircularProgress,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import {
  type DefaultValues,
  type FieldValues,
  type Path,
  useForm,
} from "react-hook-form";
import { useEffect } from "react";
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
  onSubmit: (values: TValues) => Promise<void> | void;
  helperMessage?: string;
  apiError?: string | null;
  submitInProgress?: boolean;
  serverFieldErrors?: Partial<Record<Extract<keyof TValues, string>, string>>;
};

export function AuthFormCard<TValues extends FieldValues>({
  title,
  description,
  submitLabel,
  fields,
  schema,
  onSubmit,
  helperMessage,
  apiError,
  submitInProgress = false,
  serverFieldErrors,
}: AuthFormCardProps<TValues>) {
  const defaultValues = Object.fromEntries(
    fields.map((field) => [field.name, ""]),
  ) as DefaultValues<TValues>;

  const {
    register,
    handleSubmit,
    setError,
    clearErrors,
    formState: { errors, isSubmitting },
  } = useForm<TValues>({
    resolver: zodResolver(schema),
    defaultValues,
  });

  const isBusy = isSubmitting || submitInProgress;

  useEffect(() => {
    if (!serverFieldErrors) {
      return;
    }

    for (const [fieldName, message] of Object.entries(serverFieldErrors)) {
      if (!message) {
        continue;
      }

      setError(fieldName as Path<TValues>, {
        type: "server",
        message,
      });
    }
  }, [serverFieldErrors, setError]);

  return (
    <Card sx={{ maxWidth: 520, mx: "auto", width: "100%" }}>
      <CardContent sx={{ p: { xs: 3, md: 4 } }}>
        <Stack spacing={3}>
          <Box>
            <Typography variant="h4" gutterBottom>
              {title}
            </Typography>
            <Typography color="text.secondary">{description}</Typography>
          </Box>

          <Alert severity="info">
            {helperMessage ??
              "Frontend scaffold only. Form validation is wired, but API submission is not implemented yet."}
          </Alert>

          {apiError ? <Alert severity="error">{apiError}</Alert> : null}

          <Stack
            component="form"
            spacing={2}
            noValidate
            onSubmit={handleSubmit(async (values) => {
              clearErrors();
              await onSubmit(values);
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

            <Button
              type="submit"
              variant="contained"
              size="large"
              disabled={isBusy}
              endIcon={isBusy ? <CircularProgress size={18} color="inherit" /> : null}
            >
              {submitLabel}
            </Button>
          </Stack>
        </Stack>
      </CardContent>
    </Card>
  );
}
