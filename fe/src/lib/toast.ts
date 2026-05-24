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
  void showToast("error", message, options);
}

export function notifySuccess(message: string, options?: ToastOptions) {
  void showToast("success", message, options);
}
