import { AUTH_ENDPOINT } from "@shared/services/api/auth-api-service/constants";
import type {
  IAccessToken,
  IEditUserDto,
  ILoginDto,
  IRegisterDto,
  IUserDto
} from "@shared/services/api/auth-api-service/types";
import type { BaseApiService } from "@shared/services/api/base-api-service";
import type { IApiConfig, TApiResponseWrapper } from "@shared/services/api/base-api-service/types";
import axios from "axios";

export class AuthApiService {
  constructor(private baseApiService: BaseApiService) {}

  login = (params: ILoginDto) => {
    return this.baseApiService.post<IAccessToken>(`${AUTH_ENDPOINT}/login`, params, {
      withCredentials: true
    });
  };

  register = (params: IRegisterDto) => {
    return this.baseApiService.post<TApiResponseWrapper<IUserDto>>(`${AUTH_ENDPOINT}/register`, params);
  };

  getMe = () => {
    return this.baseApiService.get<TApiResponseWrapper<IUserDto>>(`${AUTH_ENDPOINT}/me`, {
      withCredentials: true
    });
  };

  editMe = (params: IEditUserDto) => {
    return this.baseApiService.patch<TApiResponseWrapper<IUserDto>>(`${AUTH_ENDPOINT}/me`, params, {
      withCredentials: true
    });
  };

  logout = () => {
    return this.baseApiService.post(`${AUTH_ENDPOINT}/logout`, {}, { withCredentials: true });
  };

  refreshToken = (config?: IApiConfig) => {
    return axios.post<IAccessToken>("/api/auth/refresh", {}, { withCredentials: true, ...config });
  };
}
