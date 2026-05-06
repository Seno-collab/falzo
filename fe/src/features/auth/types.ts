export type LoginRequest = {
  email: string;
  password: string;
  remember: boolean;
};

export type RegisterRequest = {
  fullName: string;
  email: string;
  password: string;
};

export type ChangePasswordRequest = {
  currentPassword: string;
  newPassword: string;
};

export type AuthSession = {
  accessToken: string;
  refreshToken?: string;
};

export type AuthUser = {
  id?: string | number;
  email?: string;
  fullName?: string;
  name?: string;
  user_name?: string;
  userName?: string;
  subject?: string;
  expires?: number | string | null;
} & Record<string, unknown>;
