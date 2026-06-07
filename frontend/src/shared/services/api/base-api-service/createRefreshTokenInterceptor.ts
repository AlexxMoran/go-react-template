import type { TApiError } from "@shared/types/apiError.types";
import type { AxiosInstance } from "axios";

export const createRefreshTokenInterceptor = (
  axiosInstance: AxiosInstance,
  refreshToken?: () => Promise<string | undefined>,
  redirectToLogin?: () => void
) => {
  return async (error: TApiError) => {
    const { status, config, response } = error;

    if (status === 401 && response?.data?.error?.message_key === "invalid_token") {
      try {
        const token = await refreshToken?.();

        if (config && token) {
          config.headers.Authorization = `Bearer ${token}`;
          return axiosInstance(config);
        }

        redirectToLogin?.();
      } catch (_) {
        redirectToLogin?.();
      }
    }

    return Promise.reject(error);
  };
};
