export const viMessages = {
  common: {
    appName: "Falzo Travel",
    login: "Đăng nhập",
    register: "Đăng ký",
    dashboard: "Bảng điều khiển",
    bookNow: "Đặt ngay",
    viewDetails: "Xem chi tiết",
  },
  home: {
    heroTitle: "Chọn điểm đến nhanh, xem giá rõ ràng, đặt tour liền mạch",
    heroCta: "Khám phá tour 3D",
    searchTitle: "Tìm tour phù hợp ngay",
    featuredTitle: "Tour nổi bật",
    trustTitle: "Vì sao khách đặt qua Falzo",
  },
  travel3d: {
    title: "Chọn điểm đến trên globe 3D và đặt tour nhanh hơn",
    flowStep1: "Chọn điểm đến",
    flowStep2: "Xem lịch trình",
    flowStep3: "Đặt tour",
    detailsTitle: "Thông tin tour theo điểm đến đã chọn",
  },
  auth: {
    loginSuccess: "Đăng nhập thành công",
    loginFailed: "Đăng nhập thất bại",
    registerSuccess: "Đăng ký thành công",
    registerFailed: "Đăng ký thất bại",
  },
  apiErrors: {
    unauthorized: "Email hoặc mật khẩu không đúng.",
    conflict: "Email đã được đăng ký.",
    server: "Máy chủ đang lỗi, vui lòng thử lại sau.",
    unreachable: "Không thể kết nối API xác thực.",
    generic: "Có lỗi xảy ra, vui lòng thử lại.",
  },
} as const;

export type ViMessages = typeof viMessages;
