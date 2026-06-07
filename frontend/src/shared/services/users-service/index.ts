import type { IUserDto } from "@shared/services/api/auth-api-service/types";
import type { UsersApiService } from "@shared/services/api/users-api-service";
import { makeAutoObservable, runInAction } from "mobx";


export class UsersService {
  users: IUserDto[] = [];
  isLoading = false;

  constructor(private usersApiService: UsersApiService) {
    makeAutoObservable(this);
  }

  loadUsers = async () => {
    try {
      runInAction(() => {
        this.isLoading = true;
      });

      const { data } = await this.usersApiService.getUsers({ skip: 0, limit: 20 });

      runInAction(() => {
        this.users = data.data;
      });
    } finally {
      runInAction(() => {
        this.isLoading = false;
      });
    }
  };
}
