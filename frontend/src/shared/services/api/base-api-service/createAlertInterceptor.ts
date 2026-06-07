import type { IApiConfig } from "@shared/services/api/base-api-service/types";
import type { TApiError } from "@shared/types/apiError.types";

export const createAlertErrorInterceptor = (alertError?: (message: string) => void) => {
  return (error: TApiError) => {
    const { status, config, response } = error;
    const { suppressErrorHandling } = (config || {}) as IApiConfig;

    if (suppressErrorHandling || error.code === "ERR_CANCELED") {
      return Promise.reject(response);
    }

    if (response && status) {
      const message = response.data?.error?.message;

      if (status >= 500) {
        alertError?.("Server error");
      } else {
        alertError?.(message || "Unexpected error");
      }
    } else {
      alertError?.("Check your network connection");
    }

    return Promise.reject(response);
  };
};
