# Theo Dõi Phân Tán Với OpenTelemetry

Hướng dẫn này giải thích cách cấu hình và sử dụng theo dõi phân tán trong Bộ Định Tuyến Ngữ Nghĩa vLLM để nâng cao khả năng quan sát được và khắc phục sự cố.

## Tổng Quan

Bộ Định Tuyến Ngữ Nghĩa vLLM triển khai theo dõi phân tán toàn diện bằng OpenTelemetry, cung cấp khả năng hiển thị chi tiết vào đường ống xử lý yêu cầu. Theo dõi giúp bạn:

- **Khắc Phục Sự Cố Sản Xuất**: Theo dõi các yêu cầu riêng lẻ thông qua toàn bộ đường ống định tuyến
- **Tối Ưu Hóa Hiệu Suất**: Xác định các nút thắt cổ chai trong phân loại, lưu trữ đệm và định tuyến
- **Giám Sát Bảo Mật**: Theo dõi phát hiện PII và hoạt động ngăn chặn jailbreak
- **Phân Tích Quyết Định**: Hiểu logic định tuyến và lựa chọn chế độ suy luận
- **Tương Quan Dịch Vụ**: Kết nối các vết tích trên bộ định tuyến và các máy chủ vLLM

## Kiến Trúc

### Hệ Thống Phân Cấp Vết

Một vết tích yêu cầu điển hình tuân theo cấu trúc này:

```
semantic_router.request.received [root span]
├─ semantic_router.classification
├─ semantic_router.security.pii_detection
├─ semantic_router.security.jailbreak_detection
├─ semantic_router.cache.lookup
├─ semantic_router.routing.decision
├─ semantic_router.backend.selection
├─ semantic_router.system_prompt.injection
└─ semantic_router.upstream.request
```

### Thuộc Tính Span

Mỗi span bao gồm các thuộc tính phong phú theo quy ước OpenInference cho khả năng quan sát được LLM:

**Siêu Dữ Liệu Yêu Cầu:**

- `request.id` - Định danh yêu cầu duy nhất
- `user.id` - Định danh người dùng (nếu có sẵn)
- `http.method` - Phương thức HTTP
- `http.path` - Đường dẫn yêu cầu

**Thông Tin Mô Hình:**

- `model.name` - Tên mô hình được chọn
- `routing.original_model` - Mô hình được yêu cầu ban đầu
- `routing.selected_model` - Mô hình được bộ định tuyến chọn

**Phân Loại:**

- `category.name` - Danh mục được phân loại
- `classifier.type` - Triển khai bộ phân loại
- `classification.time_ms` - Thời gian phân loại

**Bảo Mật:**

- `pii.detected` - Liệu PII có được tìm thấy hay không
- `pii.types` - Các loại PII được phát hiện
- `jailbreak.detected` - Liệu có cố gắng jailbreak được phát hiện hay không
- `security.action` - Hành động được thực hiện (chặn, cho phép)

**Định Tuyến:**

- `routing.strategy` - Chiến lược định tuyến (tự động, được chỉ định)
- `routing.reason` - Lý do cho quyết định định tuyến
- `reasoning.enabled` - Liệu chế độ suy luận có được kích hoạt hay không
- `reasoning.effort` - Mức độ nỗ lực suy luận

**Hiệu Suất:**

- `cache.hit` - Trạng thái cache hit/miss
- `cache.lookup_time_ms` - Thời gian tra cứu cache
- `processing.time_ms` - Tổng thời gian xử lý

## Cấu Hình

### Cấu Hình Cơ Bản

Thêm phần `observability.tracing` vào `config.yaml` của bạn:

```yaml
observability:
  tracing:
    enabled: true
    provider: "opentelemetry"
    exporter:
      type: "stdout"  # hoặc "otlp"
      endpoint: "localhost:4317"
      insecure: true
    sampling:
      type: "always_on"  # hoặc "probabilistic"
      rate: 1.0
    resource:
      service_name: "vllm-semantic-router"
      service_version: "v0.1.0"
      deployment_environment: "production"
```

### Tùy Chọn Cấu Hình

#### Loại Exporter

**stdout** - In vết tích vào bảng điều khiển (phát triển)

```yaml
exporter:
  type: "stdout"
```

**otlp** - Xuất sang máy chủ tương thích OTLP (sản xuất)

```yaml
exporter:
  type: "otlp"
  endpoint: "jaeger:4317"  # Jaeger, Tempo, Datadog, v.v.
  insecure: true  # Sử dụng false với TLS trong sản xuất
```

#### Chiến Lược Lấy Mẫu

**always_on** - Lấy mẫu tất cả các yêu cầu (phát triển/gỡ lỗi)

```yaml
sampling:
  type: "always_on"
```

**always_off** - Vô hiệu hóa lấy mẫu (hiệu suất acciency)

```yaml
sampling:
  type: "always_off"
```

**probabilistic** - Lấy mẫu một phần trăm yêu cầu (sản xuất)

```yaml
sampling:
  type: "probabilistic"
  rate: 0.1  # Lấy mẫu 10% yêu cầu
```

### Cấu Hình Cụ Thể Môi Trường

#### Phát Triển

```yaml
observability:
  tracing:
    enabled: true
    provider: "opentelemetry"
    exporter:
      type: "stdout"
    sampling:
      type: "always_on"
    resource:
      service_name: "vllm-semantic-router-dev"
      deployment_environment: "development"
```

#### Sản Xuất

```yaml
observability:
  tracing:
    enabled: true
    provider: "opentelemetry"
    exporter:
      type: "otlp"
      endpoint: "tempo:4317"
      insecure: false  # Sử dụng TLS
    sampling:
      type: "probabilistic"
      rate: 0.1  # 10% lấy mẫu
    resource:
      service_name: "vllm-semantic-router"
      service_version: "v0.1.0"
      deployment_environment: "production"
```

## Triển Khai

### Với Jaeger

1. **Bắt Đầu Jaeger** (tất cả trong một để kiểm tra):

```bash
docker run -d --name jaeger \
  -p 4317:4317 \
  -p 16686:16686 \
  jaegertracing/all-in-one:latest
```

2. **Cấu Hình Bộ Định Tuyến**:

```yaml
observability:
  tracing:
    enabled: true
    exporter:
      type: "otlp"
      endpoint: "localhost:4317"
      insecure: true
    sampling:
      type: "probabilistic"
      rate: 0.1
```

3. **Truy Cập Jaeger UI**: http://localhost:16686

### Với Grafana Tempo

1. **Cấu Hình Tempo** (tempo.yaml):

```yaml
server:
  http_listen_port: 3200

distributor:
  receivers:
    otlp:
      protocols:
        grpc:
          endpoint: 0.0.0.0:4317

storage:
  trace:
    backend: local
    local:
      path: /tmp/tempo/traces
```

2. **Bắt Đầu Tempo**:

```bash
docker run -d --name tempo \
  -p 4317:4317 \
  -p 3200:3200 \
  -v $(pwd)/tempo.yaml:/etc/tempo.yaml \
  grafana/tempo:latest \
  -config.file=/etc/tempo.yaml
```

3. **Cấu Hình Bộ Định Tuyến**:

```yaml
observability:
  tracing:
    enabled: true
    exporter:
      type: "otlp"
      endpoint: "tempo:4317"
      insecure: true
```

### Triển Khai Kubernetes

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: router-config
data:
  config.yaml: |
    observability:
      tracing:
        enabled: true
        exporter:
          type: "otlp"
          endpoint: "jaeger-collector.observability.svc:4317"
          insecure: false
        sampling:
          type: "probabilistic"
          rate: 0.1
        resource:
          service_name: "vllm-semantic-router"
          deployment_environment: "production"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: semantic-router
spec:
  template:
    spec:
      containers:
      - name: router
        image: vllm-semantic-router:latest
        env:
        - name: CONFIG_PATH
          value: /config/config.yaml
        volumeMounts:
        - name: config
          mountPath: /config
      volumes:
      - name: config
        configMap:
          name: router-config
```

## Các Ví Dụ Sử Dụng

### Xem Vết Tích

#### Đầu Ra Bảng Điều Khiển (stdout exporter)

```json
{
  "Name": "semantic_router.classification",
  "SpanContext": {
    "TraceID": "abc123...",
    "SpanID": "def456..."
  },
  "Attributes": [
    {
      "Key": "category.name",
      "Value": "math"
    },
    {
      "Key": "classification.time_ms",
      "Value": 45
    }
  ],
  "Duration": 45000000
}
```

#### Jaeger UI

1. Điều hướng đến http://localhost:16686
2. Chọn dịch vụ: `vllm-semantic-router`
3. Nhấp vào "Tìm Vết Tích"
4. Xem chi tiết vết tích và dòng thời gian

### Phân Tích Hiệu Suất

**Tìm các yêu cầu chậm:**

```
Dịch vụ: vllm-semantic-router
Thời gian tối thiểu: 1s
Giới hạn: 20
```

**Phân tích các nút thắt cổ chai phân loại:**
Lọc theo hoạt động: `semantic_router.classification`
Sắp xếp theo thời gian (giảm dần)

**Theo dõi hiệu quả cache:**
Lọc theo thẻ: `cache.hit = true`
So sánh thời gian với cache miss

### Gỡ Lỗi Vấn Đề

**Tìm các yêu cầu bị lỗi:**
Lọc theo thẻ: `error = true`

**Vết tích yêu cầu cụ thể:**
Lọc theo thẻ: `request.id = req-abc-123`

**Tìm vi phạm PII:**
Lọc theo thẻ: `security.action = blocked`

## Truyền Bối Cảnh Vết Tích

Bộ định tuyến tự động truyền bối cảnh vết tích bằng tiêu đề W3C Trace Context:

**Tiêu đề Yêu Cầu** (trích xuất bởi bộ định tuyến):

```
traceparent: 00-abc123-def456-01
tracestate: vendor=value
```

**Tiêu đề Hạ Lưu** (tiêm bởi bộ định tuyến):

```
traceparent: 00-abc123-ghi789-01
x-vsr-destination-endpoint: endpoint1
x-selected-model: gpt-4
```

Điều này cho phép theo dõi từ đầu đến cuối từ máy khách → bộ định tuyến → máy chủ vLLM.

## Những Xem Xét Về Hiệu Suất

### Overhead

Theo dõi thêm chi phí tối thiểu khi cấu hình đúng:

- **Lấy mẫu luôn bật**: ~1-2% tăng độ trễ
- **10% lấy mẫu xác suất**: ~0,1-0,2% tăng độ trễ
- **Xuất không đồng bộ**: Không chặn trên xuất span

### Mẹo Tối Ưu Hóa

1. **Sử dụng lấy mẫu xác suất trong sản xuất**

   ```yaml
   sampling:
     type: "probabilistic"
     rate: 0.1  # Điều chỉnh dựa trên lưu lượng
   ```

2. **Điều chỉnh tỷ lệ lấy mẫu động**
   - Lưu lượng cao: 0,01-0,1 (1-10%)
   - Lưu lượng trung bình: 0,1-0,5 (10-50%)
   - Lưu lượng thấp: 0,5-1,0 (50-100%)

3. **Sử dụng batch exporters** (mặc định)
   - Các span được lô trước khi xuất
   - Giảm chi phí mạng

4. **Giám sát sức khỏe exporter**
   - Xem lỗi xuất trong nhật ký
   - Cấu hình chính sách thử lại

## Khắc Phục Sự Cố

### Vết Tích Không Xuất Hiện

1. **Kiểm tra theo dõi được bật**:

```yaml
observability:
  tracing:
    enabled: true
```

2. **Xác minh điểm cuối exporter**:

```bash
# Kiểm tra kết nối điểm cuối OTLP
telnet jaeger 4317
```

3. **Kiểm tra nhật ký lỗi**:

```
Không thể xuất span: kết nối bị từ chối
```

### Span Còn Thiếu

1. **Kiểm tra tỷ lệ lấy mẫu**:

```yaml
sampling:
  type: "probabilistic"
  rate: 1.0  # Tăng để xem nhiều vết tích hơn
```

2. **Xác minh tạo span trong mã**:

- Các span được tạo tại các điểm xử lý chính
- Kiểm tra bối cảnh nil

### Sử Dụng Bộ Nhớ Cao

1. **Giảm tỷ lệ lấy mẫu**:

```yaml
sampling:
  rate: 0.01  # 1% lấy mẫu
```

2. **Xác minh exporter batch đang hoạt động**:

- Kiểm tra khoảng thời gian xuất
- Giám sát độ dài hàng đợi

## Các Thực Hành Tốt Nhất

1. **Bắt đầu với stdout trong phát triển**
   - Dễ dàng xác minh theo dõi hoạt động
   - Không có phụ thuộc bên ngoài

2. **Sử dụng lấy mẫu xác suất trong sản xuất**
   - Cân bằng khả năng hiển thị và hiệu suất
   - Bắt đầu với 10% và điều chỉnh

3. **Đặt tên dịch vụ có ý nghĩa**
   - Sử dụng tên cụ thể về môi trường
   - Bao gồm thông tin phiên bản

4. **Thêm thuộc tính tùy chỉnh cho trường hợp sử dụng của bạn**
   - ID khách hàng
   - Khu vực triển khai
   - Cờ tính năng

5. **Giám sát sức khỏe exporter**
   - Theo dõi tỷ lệ thành công xuất
   - Cảnh báo tỷ lệ lỗi cao

6. **Tương quan với số liệu**
   - Sử dụng tên dịch vụ giống nhau
   - Tham chiếu chéo ID vết tích trong nhật ký

## Tích Hợp Với Ngăn Xếp vLLM

### Cải Tiến Trong Tương Lai

Triển khai theo dõi được thiết kế để hỗ trợ tích hợp tương lai với các máy chủ vLLM:

1. **Truyền bối cảnh vết tích** đến vLLM
2. **Các span tương quan** trên bộ định tuyến và engine
3. **Phân tích độ trễ từ đầu đến cuối**
4. **Thời gian cấp token** từ vLLM

Hãy chú ý các bản cập nhật về tích hợp vLLM!

## Tài Liệu Tham Khảo

- [OpenTelemetry Go SDK](https://github.com/open-telemetry/opentelemetry-go)
- [OpenInference Semantic Conventions](https://github.com/Arize-ai/openinference)
- [Jaeger Documentation](https://www.jaegertracing.io/docs/)
- [Grafana Tempo](https://grafana.com/oss/tempo/)
- [W3C Trace Context](https://www.w3.org/TR/trace-context/)
