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

export type AuthSession = {
  accessToken: string;
  refreshToken?: string;
};

export type AuthUser = {
  id?: string | number;
  email?: string;
  fullName?: string;
  name?: string;
} & Record<string, unknown>;
