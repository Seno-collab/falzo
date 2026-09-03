export type ApiResponse<T> = {
  code: string;
  message: string;
  data: T;
};

export type ApiErrorPayload = {
  code?: string;
  message?: string;
  details?: unknown;
};

export type AuthTokens = {
  access_token: string;
  refresh_token: string;
  token_type: string;
  expires_in: number;
};

export type GoogleLoginResult = AuthTokens & {
  username: string;
};

export type PasswordAuthResult = AuthTokens & {
  username: string;
};

export type AuthSession = AuthTokens & {
  username: string;
  expires_at?: number;
};
