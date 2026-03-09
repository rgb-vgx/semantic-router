# Số Liệu & Giám Sát

Thu thập và trực quan hóa số liệu cho Bộ Định Tuyến Ngữ Nghĩa bằng Prometheus và Grafana.

---

## 1. Số Liệu & Điểm Cuối

| Thành phần | Điểm cuối | Ghi chú |
| ------------------------ | ------------------------- | ------------------------------------------ |
| Số liệu định tuyến | `:9190/metrics` | Định dạng Prometheus (cờ: `--metrics-port`) |
| Sức khỏe định tuyến | `:8080/health` | HTTP sẵn sàng/sống |
| Số liệu Envoy (tùy chọn) | `:19000/stats/prometheus` | Nếu Envoy được bật |

**Vị trí cấu hình**: `tools/observability/`
**Bảng Điều Khiển**: `tools/observability/llm-router-dashboard.json`

---

## 2. Chế Độ Cục Bộ (Định Tuyến Trên Máy Chủ)

Chạy router gốc trên máy chủ, quan sát được trong Docker.

### Khởi Động Nhanh

```bash
# Máy chủ bắt đầu
make run-router

# Bắt đầu quan sát được
make o11y-local
```

**Tiếp cận:**

- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000 (admin/admin)

**Xác minh mục tiêu:**

```bash
# Kiểm tra Prometheus Scrapes localhost:9190
open http://localhost:9090/targets
```

**Dừng:**

```bash
make stop-observability
```

### Cấu Hình

Tất cả các cấu hình trong `tools/observability/`:

- `prometheus.yaml` - Scrapes mục tiêu từ biến env `ROUTER_TARGET` (mặc định: `localhost:9190`)
- `grafana-datasource.yaml` - Trỏ đến `localhost:9090`
- `grafana-dashboard.yaml` - Cấp phát bảng điều khiển
- `llm-router-dashboard.json` - Định nghĩa bảng điều khiển

### Khắc Phục Sự Cố

| Vấn đề | Sửa chữa |
| ------------- | --------------------------------------- |
| Mục tiêu XUỐNG | Máy chủ bắt đầu: `make run-router` |
| Không có số liệu | Tạo lưu lượng, kiểm tra `:9190/metrics` |
| Xung đột cổng | Thay đổi cổng hoặc dừng dịch vụ xung đột |

---

## 3. Chế Độ Docker Compose

Tất cả các dịch vụ trong các vùng chứa Docker.

### Khởi Động Nhanh

```bash
# Bắt đầu toàn bộ ngăn xếp (bao gồm quan sát được)
docker compose -f deploy/docker-compose/docker-compose.yml up --build

# Hoặc với hồ sơ kiểm tra
docker compose -f deploy/docker-compose/docker-compose.yml --profile testing up --build
```

**Tiếp cận:**

- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000 (admin/admin)

**Mục tiêu dự kiến:**

- `semantic-router:9190`
- `envoy-proxy:19000` (tùy chọn)

### Cấu Hình

Các cấu hình giống như chế độ cục bộ (`tools/observability/`), nhưng:

- `ROUTER_TARGET=semantic-router:9190`
- `PROMETHEUS_URL=prometheus:9090`
- Sử dụng mạng cầu `semantic-network`

---

## 4. Chế Độ Kubernetes

Prometheus + Grafana sản xuất cho các cụm K8s.

> **Không gian tên**: `vllm-semantic-router-system`

### Thành Phần

| Thành phần | Mục đích | Vị trí |
| ---------- | ------------------------------------- | ---------------------------------------------- |
| Prometheus | Scrapes số liệu định tuyến, giữ lại 15 ngày | `deploy/kubernetes/observability/prometheus/` |
| Grafana | Trực quan hóa bảng điều khiển | `deploy/kubernetes/observability/grafana/` |
| Ingress | Tiếp cận bên ngoài tùy chọn | `deploy/kubernetes/observability/ingress.yaml` |

### Triển Khai

```bash
# Áp dụng bản kê khai
kubectl apply -k deploy/kubernetes/observability/

# Xác minh
kubectl get pods -n vllm-semantic-router-system
```

### Tiếp Cận

**Port-forward:**

```bash
kubectl port-forward svc/prometheus 9090:9090 -n vllm-semantic-router-system
kubectl port-forward svc/grafana 3000:3000 -n vllm-semantic-router-system
```

**Ingress:** Tùy chỉnh `ingress.yaml` bằng miền và TLS của bạn

### Cấu Hình Chính

**Prometheus** sử dụng khám phá dịch vụ Kubernetes:

```yaml
scrape_configs:
  - job_name: semantic-router
    kubernetes_sd_configs:
      - role: endpoints
        namespaces:
          names: [vllm-semantic-router-system]
```

**Grafana** thông tin xác thực (thay đổi trong sản xuất):

```bash
kubectl create secret generic grafana-admin \
  --namespace vllm-semantic-router-system \
  --from-literal=admin-user=admin \
  --from-literal=admin-password='your-password'
```

---

## 5. Số Liệu Chính

| Số liệu | Loại | Mô tả |
| --------------------------------------- | --------- | ------------------------ |
| `llm_category_classifications_count` | counter | Phân loại danh mục |
| `llm_model_completion_tokens_total` | counter | Token trên model |
| `llm_model_routing_modifications_total` | counter | Thay đổi định tuyến mô hình |
| `llm_model_completion_latency_seconds` | histogram | Độ trễ hoàn thành |

**Truy vấn mẫu:**

```promql
rate(llm_model_completion_tokens_total[5m])
histogram_quantile(0.95, rate(llm_model_completion_latency_seconds_bucket[5m]))
```

---

## 6. Số Liệu Mô Hình Được Cửa Sổ (Cân Bằng Tải)

Số liệu được cải tiến theo cửa sổ thời gian để theo dõi hiệu suất mô hình, hữu ích cho quyết định cân bằng tải trong môi trường Kubernetes.

### Cấu Hình

Kích hoạt số liệu theo cửa sổ trong `config.yaml`:

```yaml
observability:
  metrics:
    windowed_metrics:
      enabled: true
      time_windows: ["1m", "5m", "15m", "1h", "24h"]
      update_interval: "10s"
      model_metrics: true
      queue_depth_estimation: true
      max_models: 100
```

### Số Liệu Cấp Mô Hình

| Số liệu | Loại | Nhãn | Mô tả |
| ------------------------------------------- | ----- | ---------------------------- | ------------------------------------- |
| `llm_model_latency_windowed_seconds` | gauge | model, time_window | Độ trễ trung bình mỗi cửa sổ thời gian |
| `llm_model_requests_windowed_total` | gauge | model, time_window | Số lượng yêu cầu mỗi cửa sổ thời gian |
| `llm_model_tokens_windowed_total` | gauge | model, token_type, time_window | Thông lượng token mỗi cửa sổ |
| `llm_model_utilization_percentage` | gauge | model, time_window | Phần trăm sử dụng ước tính |
| `llm_model_queue_depth_estimated` | gauge | model | Độ sâu hàng đợi ước tính hiện tại |
| `llm_model_error_rate_windowed` | gauge | model, time_window | Tỷ lệ lỗi mỗi cửa sổ thời gian |
| `llm_model_latency_p50_windowed_seconds` | gauge | model, time_window | Độ trễ P50 mỗi cửa sổ thời gian |
| `llm_model_latency_p95_windowed_seconds` | gauge | model, time_window | Độ trễ P95 mỗi cửa sổ thời gian |
| `llm_model_latency_p99_windowed_seconds` | gauge | model, time_window | Độ trễ P99 mỗi cửa sổ thời gian |

### Truy Vấn Ví Dụ

```promql
# Độ trễ trung bình cho model trong 5 phút cuối cùng
llm_model_latency_windowed_seconds{model="gpt-4", time_window="5m"}

# So sánh độ trễ P95 trên các mô hình
llm_model_latency_p95_windowed_seconds{time_window="15m"}

# Thông lượng token trên mỗi model
llm_model_tokens_windowed_total{token_type="completion", time_window="1h"}

# Độ sâu hàng đợi hiện tại cho quyết định cân bằng tải
llm_model_queue_depth_estimated{model="gpt-4"}

# Giám sát tỷ lệ lỗi
llm_model_error_rate_windowed{time_window="5m"} > 0.05
```

### Trường Hợp Sử Dụng

1. **Cân Bằng Tải**: Sử dụng độ sâu hàng đợi và số liệu độ trễ để định tuyến yêu cầu đến các mô hình ít tải hơn
2. **Giám Sát Hiệu Suất**: Theo dõi xu hướng độ trễ P95/P99 trên các cửa sổ thời gian
3. **Lập Kế Hoạch Dung Lượng**: Giám sát phần trăm sử dụng để xác định khi nào cần mở rộng quy mô các mô hình
4. **Cảnh Báo**: Đặt cảnh báo trên tỷ lệ lỗi hoặc độ trễ tăng đột ngột trong các cửa sổ thời gian cụ thể

---

## 7. Khắc Phục Sự Cố

| Vấn đề | Kiểm Tra | Sửa Chữa |
| --------------- | ------------------- | ----------------------------------------------------- |
| Mục tiêu XUỐNG | Prometheus /targets | Xác minh router chạy và tiếp xúc `:9190/metrics` |
| Không có số liệu | Tạo lưu lượng | Gửi yêu cầu thông qua định tuyến |
| Bảng điều khiển trống | Nguồn dữ liệu Grafana | Kiểm tra cấu hình URL Prometheus |

---
