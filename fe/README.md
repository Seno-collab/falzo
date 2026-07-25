# Falzo frontend

Frontend tối giản cho Falzo, dùng Next.js App Router và CSS thuần.

## Chạy local

```bash
pnpm install
pnpm dev
```

Mặc định frontend chạy ở `http://localhost:3000` và proxy `/api/*` tới backend
ở `http://localhost:8080`. Đổi backend bằng biến `BACKEND_URL` trong `.env.local`.

Đặt cùng Google OAuth Web client ID vào `NEXT_PUBLIC_GOOGLE_CLIENT_ID` của frontend
và `GOOGLE_CLIENT_ID` của backend. Sau đó mở `http://localhost:3000/login`.
Frontend nhận Google ID credential rồi gửi tới
`POST /api/v1/auth/google/credential`; backend xác thực credential và phát Falzo
access/refresh token. FE không hiển thị chức năng đăng ký tài khoản.

## Quy ước thư mục

- `src/app`: mỗi folder là một route/page.
- `src/components`: component UI dùng lại; chỉ tạo khi có ít nhất hai nơi dùng.
- `src/lib/api.ts`: toàn bộ request tới backend.
- `src/lib/auth.ts`: đọc/ghi/xóa session trên browser.
- `src/types`: kiểu dữ liệu dùng chung.

Khi thêm feature mới, ưu tiên tạo một route trong `src/app` và một file API
trong `src/lib` hoặc `src/features/<feature>` nếu feature bắt đầu có nhiều file.
Tránh tạo global store, hook hoặc component chung chỉ cho một màn hình.

## Kiểm tra

```bash
pnpm typecheck
pnpm build
```
