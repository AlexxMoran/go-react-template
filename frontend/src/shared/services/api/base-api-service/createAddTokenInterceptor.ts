import type { AuthService } from "@shared/services/auth-service";
import type { InternalAxiosRequestConfig } from "axios";

export const createAddTokenInterceptor = (authService: AuthService) => {
  return (config: InternalAxiosRequestConfig) => {
    if (authService.accessToken) {
      config.headers.Authorization = `Bearer ${authService.accessToken}`;
    }
    return config;
  };
};
