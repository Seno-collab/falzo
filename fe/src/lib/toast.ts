type ToastOptions = {
  description?: string;
};

async function showToast(
  type: "error" | "success",
  message: string,
  options?: ToastOptions,
) {
  const { toast } = await import("sonner");
  toast[type](message, options);
}

export function notifyError(message: string, options?: ToastOptions) {
  showToast("error", message, options).catch(() => undefined);
}

export function notifySuccess(message: string, options?: ToastOptions) {
  showToast("success", message, options).catch(() => undefined);
}
