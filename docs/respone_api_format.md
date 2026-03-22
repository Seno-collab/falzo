# API Response Convention

Tài liệu này mô tả quy ước response thống nhất cho toàn bộ hệ thống API.

**Mục tiêu:**
- Thống nhất cấu trúc response giữa các module
- Giúp frontend dễ parse dữ liệu
- Giúp backend dễ maintain và mở rộng
- Dễ chuẩn hóa exception handler, logging, tracing, swagger
- Giảm tình trạng mỗi endpoint trả về một kiểu khác nhau

---

## 1. Design Principles

Toàn bộ API tuân theo các nguyên tắc sau:

- **Cấu trúc thống nhất** — mọi response đều có cùng shape, dù thành công hay thất bại
- **Dễ phân biệt** — field `success: bool` rõ ràng, không cần đoán theo HTTP status
- **Data tách biệt** — dữ liệu chính trong `data`, metadata trong `meta`, lỗi chi tiết trong `errors`
- **Không lộ internals** — không trả stack trace, SQL error, hay internal path ra client ở production
- **HTTP status phản ánh đúng** — `200` cho thành công, `4xx` cho lỗi client, `5xx` cho lỗi server

---

## 2. Standard Response Structure

### 2.1. Success Response

```json
{
  "success": true,
  "message": "Request processed successfully",
  "data": {},
  "meta": {
    "request_id": "req_abc123",
    "timestamp": "2026-03-22T11:40:00Z"
  }
}
```

| Field        | Type     | Bắt buộc | Mô tả                                          |
|--------------|----------|-----------|------------------------------------------------|
| `success`    | `bool`   | ✅        | Luôn là `true` khi thành công                 |
| `message`    | `string` | ✅        | Mô tả ngắn kết quả                            |
| `data`       | `any`    | ✅        | Dữ liệu chính — object, array, hoặc `null`    |
| `meta`       | `object` | ✅        | Metadata của request                          |
| `meta.request_id` | `string` | ✅   | ID để trace log                               |
| `meta.timestamp`  | `string` | ✅   | Thời điểm xử lý (ISO 8601 UTC)               |

### 2.2. Error Response

```json
{
  "success": false,
  "message": "Validation failed",
  "data": null,
  "errors": [
    {
      "field": "email",
      "code": "INVALID_FORMAT",
      "message": "Email không đúng định dạng"
    },
    {
      "field": "password",
      "code": "TOO_SHORT",
      "message": "Mật khẩu phải có ít nhất 8 ký tự"
    }
  ],
  "meta": {
    "request_id": "req_abc123",
    "timestamp": "2026-03-22T11:40:00Z"
  }
}
```

| Field            | Type     | Bắt buộc | Mô tả                                        |
|------------------|----------|-----------|----------------------------------------------|
| `success`        | `bool`   | ✅        | Luôn là `false` khi có lỗi                  |
| `message`        | `string` | ✅        | Mô tả lỗi tổng quát (human-readable)        |
| `data`           | `null`   | ✅        | Luôn là `null` khi lỗi                      |
| `errors`         | `array`  | ❌        | Danh sách lỗi chi tiết (validation, v.v.)   |
| `errors[].field` | `string` | ❌        | Tên field bị lỗi (nếu có)                  |
| `errors[].code`  | `string` | ✅        | Error code chuẩn hóa để frontend xử lý      |
| `errors[].message` | `string` | ✅     | Mô tả lỗi cụ thể                           |
| `meta`           | `object` | ✅        | Metadata giống như success response          |

---

## 3. Pagination

Khi response trả về danh sách có phân trang, thêm `pagination` vào `meta`.

```json
{
  "success": true,
  "message": "Users fetched successfully",
  "data": [
    { "id": 1, "name": "Alice" },
    { "id": 2, "name": "Bob" }
  ],
  "meta": {
    "request_id": "req_abc123",
    "timestamp": "2026-03-22T11:40:00Z",
    "pagination": {
      "page": 1,
      "per_page": 20,
      "total": 100,
      "total_pages": 5,
      "has_next": true,
      "has_prev": false
    }
  }
}
```

| Field                    | Type   | Mô tả                              |
|--------------------------|--------|------------------------------------|
| `meta.pagination.page`        | `int`  | Trang hiện tại (bắt đầu từ 1)    |
| `meta.pagination.per_page`    | `int`  | Số item mỗi trang                 |
| `meta.pagination.total`       | `int`  | Tổng số item                       |
| `meta.pagination.total_pages` | `int`  | Tổng số trang                     |
| `meta.pagination.has_next`    | `bool` | Có trang tiếp theo không          |
| `meta.pagination.has_prev`    | `bool` | Có trang trước không              |

---

## 4. HTTP Status Code Convention

| Status | Tình huống                                                          |
|--------|---------------------------------------------------------------------|
| `200`  | Request thành công (GET, PUT, PATCH)                               |
| `201`  | Tạo resource mới thành công (POST)                                 |
| `204`  | Xóa thành công — không có body                                     |
| `400`  | Request không hợp lệ (validation error, bad input)                 |
| `401`  | Chưa xác thực (token thiếu hoặc hết hạn)                          |
| `403`  | Đã xác thực nhưng không có quyền                                   |
| `404`  | Resource không tồn tại                                             |
| `409`  | Conflict (duplicate entry, state conflict)                         |
| `422`  | Dữ liệu không thể xử lý (semantic error, business logic reject)   |
| `429`  | Rate limit — quá nhiều request                                     |
| `500`  | Internal server error                                              |
| `503`  | Service tạm không khả dụng                                        |

> **Lưu ý:** `success: false` luôn đi kèm `4xx` hoặc `5xx`. Không dùng `200` cho lỗi.

---

## 5. Error Code Convention

Error code là string dạng `SCREAMING_SNAKE_CASE`, dùng để frontend xử lý logic (hiển thị message, redirect, retry...).

### 5.1. Authentication & Authorization

| Code                    | HTTP  | Mô tả                              |
|-------------------------|-------|------------------------------------|
| `UNAUTHORIZED`          | `401` | Token không hợp lệ hoặc thiếu     |
| `TOKEN_EXPIRED`         | `401` | Token đã hết hạn                  |
| `FORBIDDEN`             | `403` | Không có quyền thực hiện hành động|
| `ACCOUNT_DISABLED`      | `403` | Tài khoản bị vô hiệu hóa          |

### 5.2. Validation

| Code                    | HTTP  | Mô tả                              |
|-------------------------|-------|------------------------------------|
| `VALIDATION_FAILED`     | `400` | Một hoặc nhiều field không hợp lệ |
| `REQUIRED_FIELD`        | `400` | Field bắt buộc bị thiếu           |
| `INVALID_FORMAT`        | `400` | Sai định dạng (email, phone...)   |
| `TOO_SHORT`             | `400` | Giá trị quá ngắn                  |
| `TOO_LONG`              | `400` | Giá trị quá dài                   |
| `OUT_OF_RANGE`          | `400` | Giá trị ngoài khoảng cho phép     |

### 5.3. Resource

| Code                    | HTTP  | Mô tả                              |
|-------------------------|-------|------------------------------------|
| `NOT_FOUND`             | `404` | Resource không tồn tại            |
| `ALREADY_EXISTS`        | `409` | Resource đã tồn tại (duplicate)   |
| `CONFLICT`              | `409` | Xung đột trạng thái               |

### 5.4. Server

| Code                    | HTTP  | Mô tả                              |
|-------------------------|-------|------------------------------------|
| `INTERNAL_ERROR`        | `500` | Lỗi nội bộ không xác định         |
| `SERVICE_UNAVAILABLE`   | `503` | Service tạm ngừng hoạt động       |
| `RATE_LIMITED`          | `429` | Quá số lượng request cho phép     |

---

## 6. Ví dụ Thực tế

### 6.1. GET /users/:id — Thành công

```http
GET /api/v1/users/42
HTTP/1.1 200 OK
```

```json
{
  "success": true,
  "message": "User fetched successfully",
  "data": {
    "id": 42,
    "name": "Nguyen Van A",
    "email": "vana@example.com",
    "created_at": "2026-01-15T08:00:00Z"
  },
  "meta": {
    "request_id": "req_xyz789",
    "timestamp": "2026-03-22T11:40:00Z"
  }
}
```

### 6.2. POST /users — Validation Error

```http
POST /api/v1/users
HTTP/1.1 400 Bad Request
```

```json
{
  "success": false,
  "message": "Validation failed",
  "data": null,
  "errors": [
    {
      "field": "email",
      "code": "INVALID_FORMAT",
      "message": "Email không đúng định dạng"
    },
    {
      "field": "password",
      "code": "TOO_SHORT",
      "message": "Mật khẩu phải có ít nhất 8 ký tự"
    }
  ],
  "meta": {
    "request_id": "req_xyz789",
    "timestamp": "2026-03-22T11:40:00Z"
  }
}
```

### 6.3. POST /users — Duplicate Email

```http
POST /api/v1/users
HTTP/1.1 409 Conflict
```

```json
{
  "success": false,
  "message": "Email đã được sử dụng",
  "data": null,
  "errors": [
    {
      "field": "email",
      "code": "ALREADY_EXISTS",
      "message": "Email này đã được đăng ký"
    }
  ],
  "meta": {
    "request_id": "req_xyz789",
    "timestamp": "2026-03-22T11:40:00Z"
  }
}
```

### 6.4. DELETE /users/:id — Thành công (No Content)

```http
DELETE /api/v1/users/42
HTTP/1.1 204 No Content
```

_(Không có body)_

### 6.5. GET /users — Danh sách có phân trang

```http
GET /api/v1/users?page=2&per_page=20
HTTP/1.1 200 OK
```

```json
{
  "success": true,
  "message": "Users fetched successfully",
  "data": [
    { "id": 21, "name": "Alice" },
    { "id": 22, "name": "Bob" }
  ],
  "meta": {
    "request_id": "req_xyz789",
    "timestamp": "2026-03-22T11:40:00Z",
    "pagination": {
      "page": 2,
      "per_page": 20,
      "total": 95,
      "total_pages": 5,
      "has_next": true,
      "has_prev": true
    }
  }
}
```

### 6.6. Internal Server Error

```http
HTTP/1.1 500 Internal Server Error
```

```json
{
  "success": false,
  "message": "An unexpected error occurred. Please try again later.",
  "data": null,
  "meta": {
    "request_id": "req_xyz789",
    "timestamp": "2026-03-22T11:40:00Z"
  }
}
```

> **Không bao giờ** trả stack trace, SQL message, hay internal error detail ở production.

---

## 7. Go Implementation (Gợi ý)

### 7.1. Response Types

```go
type Response struct {
    Success bool        `json:"success"`
    Message string      `json:"message"`
    Data    interface{} `json:"data"`
    Errors  []FieldError `json:"errors,omitempty"`
    Meta    Meta        `json:"meta"`
}

type Meta struct {
    RequestID  string      `json:"request_id"`
    Timestamp  time.Time   `json:"timestamp"`
    Pagination *Pagination `json:"pagination,omitempty"`
}

type Pagination struct {
    Page       int  `json:"page"`
    PerPage    int  `json:"per_page"`
    Total      int  `json:"total"`
    TotalPages int  `json:"total_pages"`
    HasNext    bool `json:"has_next"`
    HasPrev    bool `json:"has_prev"`
}

type FieldError struct {
    Field   string `json:"field,omitempty"`
    Code    string `json:"code"`
    Message string `json:"message"`
}
```

### 7.2. Helper Functions

```go
func Success(w http.ResponseWriter, r *http.Request, statusCode int, message string, data interface{}) {
    respond(w, r, statusCode, Response{
        Success: true,
        Message: message,
        Data:    data,
        Meta:    buildMeta(r),
    })
}

func Error(w http.ResponseWriter, r *http.Request, statusCode int, message string, errs ...FieldError) {
    respond(w, r, statusCode, Response{
        Success: false,
        Message: message,
        Data:    nil,
        Errors:  errs,
        Meta:    buildMeta(r),
    })
}

func respond(w http.ResponseWriter, r *http.Request, statusCode int, payload Response) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(payload)
}

func buildMeta(r *http.Request) Meta {
    return Meta{
        RequestID: middleware.GetReqID(r.Context()),
        Timestamp: time.Now().UTC(),
    }
}
```

### 7.3. Sử dụng trong Handler

```go
func GetUserHandler(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    user, err := userService.GetByID(id)
    if err != nil {
        if errors.Is(err, ErrNotFound) {
            response.Error(w, r, http.StatusNotFound, "User not found",
                response.FieldError{Code: "NOT_FOUND", Message: "User không tồn tại"})
            return
        }
        response.Error(w, r, http.StatusInternalServerError, "An unexpected error occurred")
        return
    }

    response.Success(w, r, http.StatusOK, "User fetched successfully", user)
}
```

---

## 8. Checklist

Trước khi merge PR, kiểm tra:

- [ ] Response luôn có `success`, `message`, `data`, `meta`
- [ ] `data` là `null` khi lỗi, không phải `{}` hay bỏ qua
- [ ] HTTP status code đúng với tình huống
- [ ] Validation error trả về `errors[]` với `field` và `code`
- [ ] Không lộ stack trace hay internal message ở production
- [ ] Danh sách có phân trang thì có `meta.pagination`
- [ ] `meta.request_id` được set từ middleware