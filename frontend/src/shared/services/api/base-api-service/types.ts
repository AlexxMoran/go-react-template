import type { AxiosRequestConfig } from "axios";

export interface TApiResponseWrapper<TResponse> {
  data: TResponse;
}

export interface IApiConfig extends AxiosRequestConfig {
  suppressErrorHandling?: boolean;
}
