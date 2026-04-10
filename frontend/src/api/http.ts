export type ApiValidationFields = Record<string, string>;

export class ApiError extends Error {
  status: number;
  fields?: ApiValidationFields;

  constructor(message: string, status: number, fields?: ApiValidationFields) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.fields = fields;
  }
}

type RequestOptions = Omit<RequestInit, "body"> & {
  body?: unknown;
};

const apiBaseUrl = (import.meta.env.VITE_API_BASE_URL as string | undefined)?.trim() || "";

export async function apiRequest<TResponse>(
  path: string,
  options: RequestOptions = {},
): Promise<TResponse> {
  const { body, headers, ...restOptions } = options;

  const response = await fetch(buildUrl(path), {
    ...restOptions,
    headers: {
      "Content-Type": "application/json",
      Accept: "application/json",
      ...headers,
    },
    body: body === undefined ? undefined : JSON.stringify(body),
  });

  const payload = await parseResponse(response);

  if (!response.ok) {
    throw new ApiError(
      extractErrorMessage(payload, response.status),
      response.status,
      extractValidationFields(payload),
    );
  }

  return payload as TResponse;
}

function buildUrl(path: string) {
  const normalizedPath = path.startsWith("/") ? path : `/${path}`;
  return `${apiBaseUrl}${normalizedPath}`;
}

async function parseResponse(response: Response): Promise<unknown> {
  const contentType = response.headers.get("Content-Type") ?? "";

  if (!contentType.includes("application/json")) {
    return null;
  }

  try {
    return await response.json();
  } catch {
    return null;
  }
}

function extractErrorMessage(payload: unknown, status: number) {
  if (isObject(payload) && typeof payload.error === "string" && payload.error.trim() !== "") {
    return payload.error;
  }

  return status >= 500 ? "Something went wrong. Please try again." : "Request failed.";
}

function extractValidationFields(payload: unknown): ApiValidationFields | undefined {
  if (!isObject(payload) || !isObject(payload.fields)) {
    return undefined;
  }

  const fields: ApiValidationFields = {};
  for (const [key, value] of Object.entries(payload.fields)) {
    if (typeof value === "string") {
      fields[key] = value;
    }
  }

  return Object.keys(fields).length > 0 ? fields : undefined;
}

function isObject(value: unknown): value is Record<string, any> {
  return typeof value === "object" && value !== null;
}
