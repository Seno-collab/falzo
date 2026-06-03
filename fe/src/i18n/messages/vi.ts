import { enMessages } from "@/i18n/messages/en";

export const viMessages = {
  ...enMessages,
  common: {
    ...enMessages.common,
    login: "Đăng nhập",
    register: "Đăng ký",
    logout: "Đăng xuất",
    dashboard: "Bảng điều khiển",
    explore: "Khám phá",
    places: "Điểm đến",
    upload: "Đăng ảnh",
    saved: "Đã lưu",
    account: "Tài khoản",
    profile: "Hồ sơ",
    language: "Ngôn ngữ",
    english: "English",
    vietnamese: "Tiếng Việt",
  },
  homePage: {
    ...enMessages.homePage,
    brand: "FALZO TRAVEL",
  },
  explorePage: {
    ...enMessages.explorePage,
    documentTitle: "Falzo Travel | Khám phá điểm đến",
    heroBadge: "Tuyển chọn cho người mê du lịch",
    heroTitle: "Khám phá những điểm đến đẹp và nơi đáng ghé thăm.",
    heroDescription:
      "Falzo kết nối hình ảnh, địa điểm và câu chuyện du lịch thực tế để bạn nhanh chóng chọn đúng nơi cho kỳ nghỉ, chuyến đi ngắn hoặc hành trình dài hơn.",
  },
  dashboardPage: {
    ...enMessages.dashboardPage,
    documentTitle: "Bảng điều khiển | Falzo",
    label: "FALZO PLATFORM",
    title: "Bảng điều khiển sau khi đăng nhập",
    subtitle:
      "Đây là khu vực nội bộ sau khi đăng nhập. Bạn có thể mở rộng module đặt tour, CRM và báo cáo kinh doanh tại đây.",
    open3DDemoCta: "Mở trang khám phá",
    backToLandingCta: "Quay lại trang chủ",
    logoutCta: "Đăng xuất",
  },
  profilePage: {
    ...enMessages.profilePage,
    documentTitle: "Hồ sơ | Falzo",
    brand: "Hồ sơ",
    subtitle: "Tài khoản Falzo của bạn",
    mobileMenuTitle: "Menu hồ sơ",
    loadingTitle: "Đang tải hồ sơ",
    loadingDescription: "Đang lấy tài khoản từ phiên đăng nhập.",
    signedIn: "Đã đăng nhập",
    fallbackAccount: "Tài khoản Falzo đã xác thực",
    uploadTitle: "Tải ảnh hồ sơ",
    uploadingPhoto: "Đang tải ảnh hồ sơ...",
    preparingLook: "Đang chuẩn bị phong cách mới...",
    lookReady: "đã sẵn sàng để lưu.",
    tryLook: "Thử phong cách du lịch",
    tryAnotherLook: "Thử phong cách khác",
    saveLook: "Lưu phong cách này",
    keepCurrentPhoto: "Giữ ảnh hiện tại",
    explore: "Khám phá",
    upload: "Đăng ảnh",
    profilePhotoUpdated: "Đã cập nhật ảnh hồ sơ.",
    newLookReady: "Phong cách mới đã sẵn sàng để xem trước.",
    unableToPrepareLook: "Không thể chuẩn bị phong cách hồ sơ",
    unableToUpdatePhoto: "Không thể cập nhật ảnh hồ sơ",
    passwordChanged:
      "Đã đổi mật khẩu thành công. Vui lòng đăng nhập lại.",
    unableToChangePassword: "Không thể đổi mật khẩu",
    overviewTitle: "Tổng quan tài khoản",
    overviewDescription:
      "Thông tin định danh tài khoản và phiên đăng nhập hiện tại.",
    notProvided: "Chưa cung cấp",
    notAvailable: "Không khả dụng",
    security: "Bảo mật",
    changePasswordTitle: "Đổi mật khẩu",
    changePasswordDescription:
      "Sử dụng ít nhất tám ký tự, bao gồm chữ và số.",
    currentPassword: "Mật khẩu hiện tại",
    newPassword: "Mật khẩu mới",
    confirmPassword: "Xác nhận mật khẩu",
    update: "Cập nhật",
    logout: "Đăng xuất",
    fields: {
      email: "Email",
      username: "Tên người dùng",
      subject: "Định danh phiên",
      tokenExpires: "Token hết hạn",
    },
  },
  loginPage: {
    ...enMessages.loginPage,
    documentTitle: "Đăng nhập | Falzo",
    title: "Đăng nhập",
    subtitle: "Nhập thông tin tài khoản để tiếp tục.",
    emailLabel: "Email",
    emailInvalid: "Email không hợp lệ.",
    passwordLabel: "Mật khẩu",
    passwordMin: "Mật khẩu phải có ít nhất 6 ký tự.",
    rememberLabel: "Ghi nhớ đăng nhập",
    submit: "Đăng nhập",
    submitting: "Đang đăng nhập...",
    noAccountText: "Chưa có tài khoản?",
    registerCta: "Đăng ký",
    successTitle: "Đăng nhập thành công",
    errorTitle: "Đăng nhập thất bại",
  },
  registerPage: {
    ...enMessages.registerPage,
    documentTitle: "Đăng ký | Falzo",
    title: "Tạo tài khoản",
    subtitle: "Tạo tài khoản mới để bắt đầu.",
    fullNameLabel: "Tên người dùng",
    fullNamePlaceholder: "nguyen-van-a",
    fullNameMin: "Tên người dùng phải có ít nhất 3 ký tự.",
    emailLabel: "Email",
    emailInvalid: "Email không hợp lệ.",
    passwordLabel: "Mật khẩu",
    passwordMin: "Mật khẩu phải có ít nhất 6 ký tự.",
    confirmPasswordLabel: "Xác nhận mật khẩu",
    confirmPasswordMin: "Vui lòng xác nhận mật khẩu.",
    confirmPasswordMismatch: "Mật khẩu xác nhận không khớp.",
    submit: "Đăng ký",
    submitting: "Đang tạo tài khoản...",
    hasAccountText: "Đã có tài khoản?",
    loginCta: "Đăng nhập",
    successTitle: "Đăng ký thành công",
    successRedirectDashboard: "Đang chuyển đến bảng điều khiển.",
    successPromptLogin: "Bạn có thể đăng nhập ngay.",
    errorTitle: "Đăng ký thất bại",
  },
  auth: {
    ...enMessages.auth,
    loginSuccess: "Đăng nhập thành công",
    loginFailed: "Đăng nhập thất bại",
    registerSuccess: "Đăng ký thành công",
    registerFailed: "Đăng ký thất bại",
  },
  apiErrors: {
    ...enMessages.apiErrors,
    unauthorized: "Email hoặc mật khẩu không đúng.",
    conflict: "Email này đã được đăng ký.",
    server: "Lỗi máy chủ. Vui lòng thử lại sau.",
    unreachable: "Không thể kết nối đến API xác thực.",
    generic: "Đã có lỗi xảy ra. Vui lòng thử lại.",
    backendMessages: {
      "Account already exists": "Tài khoản đã tồn tại",
      "Authentication service is temporarily unavailable":
        "Dịch vụ xác thực tạm thời không khả dụng",
      "Authentication service unavailable": "Dịch vụ xác thực không khả dụng",
      "Category already exists": "Danh mục đã tồn tại",
      "Category name or slug is already in use":
        "Tên hoặc slug danh mục đã được sử dụng",
      "Category service is temporarily unavailable":
        "Dịch vụ danh mục tạm thời không khả dụng",
      "Category service unavailable": "Dịch vụ danh mục không khả dụng",
      "Comment not found": "Không tìm thấy bình luận",
      "Email or password is incorrect": "Email hoặc mật khẩu không đúng",
      "Forbidden": "Không có quyền truy cập",
      "Image not found": "Không tìm thấy ảnh",
      "Image upload is temporarily unavailable":
        "Tải ảnh tạm thời không khả dụng",
      "Image upload unavailable": "Không thể tải ảnh",
      "Invalid JSON payload": "Dữ liệu JSON không hợp lệ",
      "Invalid authorization header": "Header xác thực không hợp lệ",
      "Invalid credentials": "Thông tin đăng nhập không hợp lệ",
      "Location service is temporarily unavailable":
        "Dịch vụ địa điểm tạm thời không khả dụng",
      "Location service unavailable": "Dịch vụ địa điểm không khả dụng",
      "Missing auth context": "Thiếu ngữ cảnh xác thực",
      "Missing bearer token": "Thiếu bearer token",
      "Post not found": "Không tìm thấy bài viết",
      "Post service is temporarily unavailable":
        "Dịch vụ bài viết tạm thời không khả dụng",
      "Post service unavailable": "Dịch vụ bài viết không khả dụng",
      "Refresh token is required": "Refresh token là bắt buộc",
      "Requested category does not exist": "Danh mục được yêu cầu không tồn tại",
      "Requested comment does not exist": "Bình luận được yêu cầu không tồn tại",
      "Requested image does not exist": "Ảnh được yêu cầu không tồn tại",
      "Requested post does not exist": "Bài viết được yêu cầu không tồn tại",
      "Requested saved collection does not exist":
        "Bộ sưu tập đã lưu không tồn tại",
      "Requested user does not exist": "Người dùng được yêu cầu không tồn tại",
      "Session has been revoked or expired":
        "Phiên đăng nhập đã bị thu hồi hoặc hết hạn",
      "Social service is temporarily unavailable":
        "Dịch vụ xã hội tạm thời không khả dụng",
      "Social service unavailable": "Dịch vụ xã hội không khả dụng",
      "Token is invalid": "Token không hợp lệ",
      "Too many login attempts, please try again later":
        "Đăng nhập quá nhiều lần, vui lòng thử lại sau",
      "Too many refresh attempts, please try again later":
        "Làm mới token quá nhiều lần, vui lòng thử lại sau",
      "Too many registration attempts, please try again later":
        "Đăng ký quá nhiều lần, vui lòng thử lại sau",
      "Too many requests": "Quá nhiều yêu cầu",
      "Unauthorized": "Chưa được xác thực",
      "Unexpected error": "Lỗi không mong muốn",
      "User not found": "Không tìm thấy người dùng",
      "Username or email is already in use":
        "Tên người dùng hoặc email đã được sử dụng",
      "Validation field": "Dữ liệu không hợp lệ",
      "You can only edit your own comments":
        "Bạn chỉ có thể chỉnh sửa bình luận của mình",
      "You can only edit your own posts":
        "Bạn chỉ có thể chỉnh sửa bài viết của mình",
      "You cannot block yourself": "Bạn không thể chặn chính mình",
      "You cannot follow yourself": "Bạn không thể theo dõi chính mình",
      "You cannot moderate this content":
        "Bạn không thể kiểm duyệt nội dung này",
      "avatar_url is required": "avatar_url là bắt buộc",
      "avatar_url must be a valid URL": "avatar_url phải là URL hợp lệ",
      "caption exceeds max length": "Chú thích vượt quá độ dài tối đa",
      "category does not exist": "Danh mục không tồn tại",
      "category name already exists": "Tên danh mục đã tồn tại",
      "category_ids must not contain more than 20 items":
        "category_ids không được chứa quá 20 mục",
      "collection id is required": "ID bộ sưu tập là bắt buộc",
      "collection name already exists": "Tên bộ sưu tập đã tồn tại",
      "collection name is required": "Tên bộ sưu tập là bắt buộc",
      "collection name must not exceed 120 characters":
        "Tên bộ sưu tập không được vượt quá 120 ký tự",
      "collection share slug is required":
        "Slug chia sẻ bộ sưu tập là bắt buộc",
      "comment content exceeds max length":
        "Nội dung bình luận vượt quá độ dài tối đa",
      "comment content is required": "Nội dung bình luận là bắt buộc",
      "current_password and new_password are required":
        "Mật khẩu hiện tại và mật khẩu mới là bắt buộc",
      "cursor is invalid": "Cursor không hợp lệ",
      "email and password are required": "Email và mật khẩu là bắt buộc",
      "email must be a valid email": "Email phải hợp lệ",
      "feed must be following when provided":
        "Feed phải là following nếu được cung cấp",
      "file content is not a valid image": "Nội dung file không phải ảnh hợp lệ",
      "file is required": "File là bắt buộc",
      "file mime type is invalid": "Định dạng file không hợp lệ",
      "file size exceeds the maximum allowed limit":
        "Dung lượng file vượt quá giới hạn cho phép",
      "file size is invalid": "Dung lượng file không hợp lệ",
      "id must be a valid integer": "ID phải là số nguyên hợp lệ",
      "id must be a valid positive integer": "ID phải là số nguyên dương hợp lệ",
      "image URL is invalid": "URL ảnh không hợp lệ",
      "image id is required": "ID ảnh là bắt buộc",
      "image_url is required": "image_url là bắt buộc",
      "image_url must be a valid URL": "image_url phải là URL hợp lệ",
      "image_urls must not contain more than 10 items":
        "image_urls không được chứa quá 10 mục",
      "lat must be a valid float64": "lat phải là số hợp lệ",
      "lat must be between -90 and 90": "lat phải nằm trong khoảng -90 đến 90",
      "lat, lng and radius are required": "lat, lng và radius là bắt buộc",
      "latitude must be between -90 and 90":
        "Vĩ độ phải nằm trong khoảng -90 đến 90",
      "limit must be an integer": "limit phải là số nguyên",
      "limit must be greater than 0": "limit phải lớn hơn 0",
      "limit must not exceed 50": "limit không được vượt quá 50",
      "lng must be a valid float64": "lng phải là số hợp lệ",
      "lng must be between -180 and 180":
        "lng phải nằm trong khoảng -180 đến 180",
      "location id is required": "ID địa điểm là bắt buộc",
      "location_name exceeds max length":
        "Tên địa điểm vượt quá độ dài tối đa",
      "location_name is required": "Tên địa điểm là bắt buộc",
      "longitude must be between -180 and 180":
        "Kinh độ phải nằm trong khoảng -180 đến 180",
      "name cannot exceed 255 characters":
        "Tên không được vượt quá 255 ký tự",
      "name is required": "Tên là bắt buộc",
      "nearby radius must not exceed 1000 km":
        "Bán kính gần đây không được vượt quá 1000 km",
      "nearby sort requires valid lat and lng query params":
        "Sắp xếp gần đây cần tham số lat và lng hợp lệ",
      "new_password must be at least 8 characters and contain letters and digits":
        "Mật khẩu mới phải có ít nhất 8 ký tự và gồm cả chữ lẫn số",
      "owner_id is required": "owner_id là bắt buộc",
      "page must be an integer": "page phải là số nguyên",
      "page must be greater than 0": "page phải lớn hơn 0",
      "post id is required": "ID bài viết là bắt buộc",
      "q is required": "Từ khóa tìm kiếm là bắt buộc",
      "radius must be a valid float64": "radius phải là số hợp lệ",
      "radius must be greater than 0": "radius phải lớn hơn 0",
      "reason is required": "Lý do là bắt buộc",
      "reason must not exceed 500 characters":
        "Lý do không được vượt quá 500 ký tự",
      "register fields required": "Các trường đăng ký là bắt buộc",
      "reply comment does not exist on this post":
        "Bình luận trả lời không tồn tại trong bài viết này",
      "share_slug is required": "share_slug là bắt buộc",
      "slug cannot exceed 255 characters":
        "Slug không được vượt quá 255 ký tự",
      "slug is required": "Slug là bắt buộc",
      "sort must be newest, popular, trending, or nearby":
        "sort phải là newest, popular, trending hoặc nearby",
      "target user id is required": "ID người dùng mục tiêu là bắt buộc",
      "type must be credible, suspicious, ai_generated, wrong_context, or unsure":
        "type phải là credible, suspicious, ai_generated, wrong_context hoặc unsure",
      "type is required": "Loại đánh giá là bắt buộc",
      "user id is required": "ID người dùng là bắt buộc",
      "user_name must be between 3 and 50 characters":
        "Tên người dùng phải từ 3 đến 50 ký tự",
      "user_name, email and password are required":
        "Tên người dùng, email và mật khẩu là bắt buộc",
      "Validation field: Invalid JSON payload":
        "Dữ liệu không hợp lệ: JSON không hợp lệ",
    },
  },
} as const;

export type ViMessages = typeof viMessages;
