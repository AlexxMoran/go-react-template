import type { AxiosError } from "axios";

/** Error envelope returned by the Go backend: { "error": { ... } }. */
export interface IApiErrorBody {
  message: string;
  message_key: string;
  fields?: Record<string, string>;
}

export interface IApiBaseError {
  error: IApiErrorBody;
}

export type TApiError = AxiosError<IApiBaseError>;
