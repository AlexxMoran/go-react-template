import { AuthApiService } from "@shared/services/api/auth-api-service";
import { BaseApiService } from "@shared/services/api/base-api-service";
import { UsersApiService } from "@shared/services/api/users-api-service";
import { AuthService } from "@shared/services/auth-service";
import type { IRootServiceData } from "@shared/services/root-service/types";
import { UsersService } from "@shared/services/users-service";

export class RootService {
  baseApiService = new BaseApiService();
  authApiService = new AuthApiService(this.baseApiService);
  usersApiService = new UsersApiService(this.baseApiService);
  authService = new AuthService(this.authApiService);
  usersService = new UsersService(this.usersApiService);

  constructor(data: IRootServiceData) {
    this.baseApiService.createInterceptors(data, this.authService);
  }
}
