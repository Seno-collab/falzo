let nextRequestId = 0;
const activeRequestIds = new Set<number>();
const listeners = new Set<() => void>();

export function beginApiRequest() {
  const requestId = nextRequestId;
  nextRequestId += 1;
  activeRequestIds.add(requestId);
  notifyListeners();

  let finished = false;
  return () => {
    if (finished) return;
    finished = true;
    activeRequestIds.delete(requestId);
    notifyListeners();
  };
}

export function subscribeToApiActivity(listener: () => void) {
  listeners.add(listener);
  return () => listeners.delete(listener);
}

export function getActiveApiRequestCount() {
  return activeRequestIds.size;
}

function notifyListeners() {
  listeners.forEach((listener) => listener());
}
