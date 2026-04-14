export type ApiMeta = {
  request_id: string;
  timestamp: string;
};

export type ApiErrorDetail = {
  field?: string;
  code: string;
  message: string;
};

export type ApiEnvelope<T = unknown> = {
  success: boolean;
  message: string;
  data: T;
  errors?: ApiErrorDetail[];
  meta: ApiMeta;
};
