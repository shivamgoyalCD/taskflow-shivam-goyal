import { loadStoredSession } from "@/features/auth/authStorage";

export type ApiValidationFields = Record<string, string>;

export class ApiError extends Error {
  status: number;
  fields?: ApiValidationFields;
  data?: unknown;

  constructor(message: string, status: number, options?: { fields?: ApiValidationFields; data?: unknown }) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.fields = options?.fields;
    this.data = options?.data;
  }
}

type RequestOptions = Omit<RequestInit, "body"> & {
  body?: unknown;
};

const apiBaseUrl = (import.meta.env.VITE_API_BASE_URL as string | undefined)?.trim() || "";

export const apiClient = {
  get: <TResponse>(path: string, options?: RequestOptions) =>
    request<TResponse>(path, { ...options, method: "GET" }),
  post: <TResponse>(path: string, body?: unknown, options?: RequestOptions) =>
    request<TResponse>(path, { ...options, method: "POST", body }),
  patch: <TResponse>(path: string, body?: unknown, options?: RequestOptions) =>
    request<TResponse>(path, { ...options, method: "PATCH", body }),
  delete: <TResponse>(path: string, options?: RequestOptions) =>
    request<TResponse>(path, { ...options, method: "DELETE" }),
};

export async function request<TResponse>(
  path: string,
  options: RequestOptions = {},
): Promise<TResponse> {
  const { body, headers, ...restOptions } = options;
  const response = await fetch(buildUrl(path), {
    ...restOptions,
    headers: buildHeaders(headers, body),
    body: body === undefined ? undefined : JSON.stringify(body),
  });

  const payload = await parseJSON(response);

  if (!response.ok) {
    throw new ApiError(resolveErrorMessage(payload, response.status), response.status, {
      fields: extractValidationFields(payload),
      data: payload,
    });
  }

  return payload as TResponse;
}

function buildUrl(path: string) {
  const normalizedPath = path.startsWith("/") ? path : `/${path}`;
  return `${apiBaseUrl}${normalizedPath}`;
}

function buildHeaders(headers: HeadersInit | undefined, body: unknown): Headers {
  const mergedHeaders = new Headers(headers);

  mergedHeaders.set("Accept", "application/json");
  if (body !== undefined) {
    mergedHeaders.set("Content-Type", "application/json");
  }

  const token = loadStoredSession()?.token;
  if (token && !mergedHeaders.has("Authorization")) {
    mergedHeaders.set("Authorization", `Bearer ${token}`);
  }

  return mergedHeaders;
}

async function parseJSON(response: Response): Promise<unknown> {
  if (response.status === 204) {
    return null;
  }

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

function resolveErrorMessage(payload: unknown, status: number) {
  if (isObject(payload) && typeof payload.error === "string" && payload.error.trim() !== "") {
    return payload.error;
  }

  if (status >= 500) {
    return "Something went wrong. Please try again.";
  }

  return "Request failed.";
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

function isObject(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}
