import type { IUserDto } from "@shared/services/api/auth-api-service/types";
import type { BaseApiService } from "@shared/services/api/base-api-service";
import type { TApiResponseWrapper } from "@shared/services/api/base-api-service/types";
import type { IPaginationMeta, IPaginationParams } from "@shared/types/pagination.types";

export class UsersApiService {
  constructor(private baseApiService: BaseApiService) {}

  getUsers = (params: IPaginationParams) => {
    return this.baseApiService.get<TApiResponseWrapper<IUserDto[]> & IPaginationMeta>("/v1/users", {
      params,
      withCredentials: true
    });
  };
}
