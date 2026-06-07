import { createAddTokenInterceptor } from "@shared/services/api/base-api-service/createAddTokenInterceptor";
import { createAlertErrorInterceptor } from "@shared/services/api/base-api-service/createAlertInterceptor";
import { createRefreshTokenInterceptor } from "@shared/services/api/base-api-service/createRefreshTokenInterceptor";
import type { IApiConfig } from "@shared/services/api/base-api-service/types";
import type { AuthService } from "@shared/services/auth-service";
import type { IRootServiceData } from "@shared/services/root-service/types";
import axios from "axios";

export class BaseApiService {
  private axiosInstance = axios.create({ baseURL: "/api" });

  createInterceptors = (data: IRootServiceData, authService: AuthService) => {
    const { alertError, redirectToLoginPage } = data;

    this.axiosInstance.interceptors.response.use(
      null,
      createRefreshTokenInterceptor(this.axiosInstance, authService.refreshToken, redirectToLoginPage)
    );

    this.axiosInstance.interceptors.response.use(null, createAlertErrorInterceptor(alertError));
    this.axiosInstance.interceptors.request.use(createAddTokenInterceptor(authService));
  };

  get<TResponse>(url: string, config?: IApiConfig) {
    return this.axiosInstance.get<TResponse>(url, config);
  }

  post<TResponse>(url: string, body?: unknown, config?: IApiConfig) {
    return this.axiosInstance.post<TResponse>(url, body, config);
  }

  patch<TResponse>(url: string, body?: unknown, config?: IApiConfig) {
    return this.axiosInstance.patch<TResponse>(url, body, config);
  }

  delete<TResponse>(url: string, config?: IApiConfig) {
    return this.axiosInstance.delete<TResponse>(url, config);
  }
}
