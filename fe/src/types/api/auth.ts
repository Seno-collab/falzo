export type ApiLanguage = "vi" | "en";

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
