/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_BASE_URL?: string;
  readonly VITE_AUTH_LOGIN_ENDPOINT?: string;
  readonly VITE_AUTH_REGISTER_ENDPOINT?: string;
  readonly VITE_AUTH_ME_ENDPOINT?: string;
  readonly VITE_AUTH_REFRESH_ENDPOINT?: string;
  readonly VITE_AUTH_LOGOUT_ENDPOINT?: string;
  readonly VITE_API_PROXY_TARGET?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
