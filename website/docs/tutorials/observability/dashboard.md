# Bảng Điều Khiển Bộ Định Tuyến Ngữ Nghĩa

Bảng Điều Khiển Bộ Định Tuyến Ngữ Nghĩa là giao diện toán tử hợp nhất mang lại Quản Lý Cấu Hình, Sân Chơi Tương Tác và Giám Sát & Quan Sát Theo Thời Gian Thực. Nó cung cấp một điểm nhập duy nhất trên các triển khai phát triển cục bộ, Docker Compose và Kubernetes.

- Một nơi để xem và chỉnh sửa cấu hình (với bảo vệ)
- Một tab để kiểm tra lời nhắc thông qua sân chơi trò chuyện tích hợp
- Một tab để xem số liệu/bảng điều khiển (Grafana/Prometheus)
- Proxy phụ trợ duy nhất chuẩn hóa xác thực, CORS và CSP trên các dịch vụ

## Trong Cái Gì

### Giao Diện Trước (React + TypeScript + Vite)

Ứng dụng một trang hiện đại với:

- React 18 + TypeScript + Vite
- React Router cho định tuyến phía máy khách
- CSS Modules, chủ đề tối/sáng với tính bền vững
- Thanh bên có thể thu gọn để nhảy qua các phần
- Trực quan hóa tôpô được cấp bởi React Flow

Các trang:

- Trang đích: Giới thiệu và liên kết nhanh

![Trang đích Bảng điều khiển](/img/dashboard/landing.png)

- Sân chơi: Sân chơi trò chuyện tích hợp để kiểm tra nhanh

- Cấu hình: Trình xem/trình chỉnh sửa cấu hình theo thời gian thực với bảng điều khiển có cấu trúc và chế độ xem thô

![Trang Cấu Hình](/img/dashboard/config.png)

- Tôpô: Luồng trực quan từ yêu cầu người dùng đến lựa chọn mô hình

![Chế độ xem Tôpô](/img/dashboard/topology.png)

- Giám sát: Bảng điều khiển Grafana nhúng

![Grafana Nhúng](/img/dashboard/grafana.png)

### Phụ Trợ (Máy Chủ HTTP Go)

- Phục vụ bản dựng giao diện trước (định tuyến SPA)
- Các máy chủ Proxy ngược với chuẩn hóa tiêu đề cho nhúng iframe
- Tiếp xúc một bộ API bảng điều khiển nhỏ cho cấu hình và cơ sở dữ liệu công cụ

Các tuyến đường chính:

- Sức khỏe: `GET /healthz`
- Cấu hình (đọc): `GET /api/router/config/all` (đọc YAML, trả lại JSON)
- Cấu hình (viết): `POST /api/router/config/update` (viết YAML trở lại tệp)
- Cơ sở dữ liệu Công cụ: `GET /api/tools-db` (phục vụ tools_db.json bên cạnh cấu hình)
- API Định Tuyến: `GET/POST /api/router/*` (Tiêu đề Ủy quyền được chuyển tiếp)
- Grafana (nhúng): `GET /embedded/grafana/*`
- Prometheus (nhúng): `GET /embedded/prometheus/*`
- Số liệu định tuyến Passthrough: `GET /metrics/router` → chuyển hướng đến số liệu định tuyến

Proxy loại bỏ/ghi đè `X-Frame-Options` và điều chỉnh `Content-Security-Policy` để cho phép `frame-ancestors 'self'`, cho phép nhúng an toàn dưới nguồn gốc bảng điều khiển.

## Các Biến Môi Trường

Cung cấp các mục tiêu hạ lưu và cài đặt thời gian chạy thông qua các biến env (mặc định trong ngoặc):

- `DASHBOARD_PORT` (8700)
- `TARGET_GRAFANA_URL`
- `TARGET_PROMETHEUS_URL`
- `TARGET_ROUTER_API_URL` (http://localhost:8080)
- `TARGET_ROUTER_METRICS_URL` (http://localhost:9190/metrics)
- `ROUTER_CONFIG_PATH` (../../config/config.yaml)
- `DASHBOARD_STATIC_DIR` (../frontend)

Lưu ý: API cập nhật cấu hình ghi vào `ROUTER_CONFIG_PATH`. Trong các vùng chứa/Kubernetes, đường dẫn này phải có thể ghi được (không phải ConfigMap chỉ đọc). Gắn một tập lưu trữ có thể ghi nếu bạn cần các chỉnh sửa thời gian chạy để duy trì.

## Khởi Động Nhanh

### Docker Compose (Được Khuyến Khích)

Bảng điều khiển được tích hợp vào tệp Compose chính.

```bash
# Từ gốc kho lưu trữ
make docker-compose-up
```

Sau đó mở trong trình duyệt:

- Bảng điều khiển: http://localhost:8700
- Grafana: http://localhost:3000
- Prometheus: http://localhost:9090

## Tài Liệu Liên Quan

- [Cấu Hình Cài Đặt](../../installation/configuration.md)
- [Số Liệu Quan Sát được](./metrics.md)
- [Theo Dõi Phân Tán](./distributed-tracing.md)
