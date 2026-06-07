import type { AuthApiService } from "@shared/services/api/auth-api-service";
import type { ILoginDto, IUserDto } from "@shared/services/api/auth-api-service/types";
import type { TMaybe } from "@shared/types/main.types";
import { makeAutoObservable, runInAction } from "mobx";

export class AuthService {
  accessToken: TMaybe<string> = null;
  me: TMaybe<IUserDto> = null;

  constructor(private authApiService: AuthApiService) {
    makeAutoObservable(this);
  }

  get isAuthenticated() {
    return Boolean(this.me && this.accessToken);
  }

  setMe = (me: IUserDto) => {
    this.me = me;
  };

  login = async (params: ILoginDto) => {
    const { data } = await this.authApiService.login(params);
    runInAction(() => {
      this.accessToken = data.access_token;
    });
    await this.getMe();
  };

  logout = async () => {
    try {
      await this.authApiService.logout();
    } finally {
      runInAction(() => {
        this.me = null;
        this.accessToken = null;
      });
    }
  };

  refreshToken = async () => {
    try {
      const { data } = await this.authApiService.refreshToken({ suppressErrorHandling: true });
      runInAction(() => {
        this.accessToken = data.access_token;
      });

      if (!this.me) {
        await this.getMe();
      }

      return data.access_token;
    } catch (_) {
      runInAction(() => {
        this.accessToken = null;
        this.me = null;
      });

      return undefined;
    }
  };

  getMe = async () => {
    try {
      const { data } = await this.authApiService.getMe();
      runInAction(() => {
        this.me = data.data;
      });
    } catch (_) {
      /* empty */
    }
  };
}
