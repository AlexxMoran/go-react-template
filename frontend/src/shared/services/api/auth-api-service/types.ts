export interface ILoginDto {
  email: string;
  password: string;
}

export interface IRegisterDto {
  email: string;
  password: string;
  first_name?: string;
}

export interface IAccessToken {
  access_token: string;
  token_type: string;
}

export interface IUserDto {
  id: number;
  email: string;
  role: string;
  first_name: string;
  last_name: string;
  is_active: boolean;
  is_verified: boolean;
  created_at: string;
  permissions?: Record<string, boolean>;
}

export interface IEditUserDto {
  first_name?: string;
  last_name?: string;
}
