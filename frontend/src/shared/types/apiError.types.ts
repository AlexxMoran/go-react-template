import type { components } from "@shared/api/schema";
import type { AxiosError } from "axios";

/** Error envelope returned by the Go backend: { "error": { ... } }. */
export type IApiErrorBody = components["schemas"]["ErrorBody"];

export type IApiBaseError = components["schemas"]["ErrorResponse"];

export type TApiError = AxiosError<IApiBaseError>;
