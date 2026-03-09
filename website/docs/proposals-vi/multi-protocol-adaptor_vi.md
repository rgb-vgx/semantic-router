# Tài liệu Thiết kế: Kiến trúc Bộ chuyển đổi Đa giao thức

**Tác giả:** Đội ngũ vLLM Semantic Router
**Trạng thái:** Sẽ được triển khai
**Tạo:** Tháng 2 năm 2026
**Cập nhật lần cuối:** Tháng 2 năm 2026

## Tổng quan

Tài liệu này mô tả thiết kế và triển khai kiến trúc bộ chuyển đổi đa giao thức cho vLLM Semantic Router, nhằm trừu tượng hóa lớp API để hỗ trợ nhiều giao thức front-end vượt ra ngoài Envoy ExtProc.

## Lý do

Semantic Router ban đầu được tích hợp chặt chẽ với giao thức Envoy External Processor (ExtProc) thông qua gRPC. Mặc dù điều này cung cấp khả năng tích hợp mạnh mẽ với Envoy, nhưng nó tạo ra những rào cản cho những người dùng muốn:

- Sử dụng bộ định tuyến mà không cần triển khai Envoy
- Ưa thích tích hợp API HTTP/REST trực tiếp
- Sử dụng Nginx hoặc các reverse proxy khác
- Cần kiến trúc triển khai đơn giản hơn cho phát triển hoặc kiểm thử

### Động lực

- **Tính linh hoạt:** Người dùng cần quyền truy cập API HTTP trực tiếp mà không cần cơ sở hạ tầng Envoy
- **Kiểm thử:** Nhà phát triển cần kiểm thử nhẹ mà không cần triển khai Envoy đầy đủ
- **Khả năng mở rộng:** Hỗ trợ nginx, gRPC gốc và các giao thức tùy chỉnh
- **Tái sử dụng:** محک định tuyến duy nhất được chia sẻ trên tất cả các giao thức
- **Các tùy chọn triển khai:** Cho phép các tình huống triển khai không máy chủ, cạnh và đơn giản hóa

## Mục tiêu

### Mục tiêu chính

1. **Trừu tượng hóa giao thức:** Tách logic định tuyến từ mã cụ thể giao thức
2. **Hỗ trợ đa giao thức:** Cho phép hoạt động đồng thời của nhiều giao thức
3. **Khả năng tương thích ngược:** Bảo tồn chức năng ExtProc hiện có
4. **Trạng thái được chia sẻ:** Một nguồn sự thật duy nhất cho bộ đệm, phát lại và quyết định định tuyến
5. **Mở rộng dễ dàng:** Mẫu đơn giản để thêm các bộ chuyển đổi giao thức mới

### Không phải là mục tiêu

1. Thay thế hoặc ngừng sử dụng hỗ trợ Envoy ExtProc
2. Thay đổi các thuật toán quyết định định tuyến hoặc logic phân loại
3. Sửa đổi định dạng cấu hình ngoài phần bộ chuyển đổi
4. Hỗ trợ các tính năng cụ thể giao thức phá vỡ tính trừu tượng

## Nguyên lý thiết kế

### 1. Quy trình định tuyến duy nhất

**QUAN TRỌNG:** Tất cả logic định tuyến phải chạy qua `RouterEngine.Route()`. Không có ngoại lệ.

- ✅ Các bộ chuyển đổi chuyển đổi giao thức → `RouteRequest` → gọi `RouterEngine.Route()`
- ✅ `RouterEngine.Route()` trả về `RouteResponse` → các bộ chuyển đổi chuyển đổi → giao thức
- ❌ Các bộ chuyển đổi phải không trùng lặp logic phân loại, bảo mật, bộ đệm, phát lại
- ❌ Các bộ chuyển đổi phải không gọi trực tiếp bộ phân loại, bộ đệm hoặc bộ ghi phát lại

### 2. Lớp bộ chuyển đổi mỏng

Các bộ chuyển đổi chỉ là **bản dịch giao thức**:

- Phân tích cú pháp định dạng yêu cầu cụ thể giao thức
- Chuyển đổi thành `RouteRequest`
- Gọi `RouterEngine.Route()`
- Chuyển đổi `RouteResponse` thành định dạng giao thức
- Trở lại ứng dụng khách

### 3. RouterEngine sở hữu tất cả định tuyến

`RouterEngine.Route()` là **NƠI DUY NHẤT** nơi:

- Phân loại xảy ra
- Phát hiện PII/jailbreak chạy
- Bộ đệm được kiểm tra/cập nhật
- Công cụ được chọn
- Phát lại được ghi lại
- Lựa chọn backend xảy ra
- Proxying xảy ra (hoặc trả về thông tin proxy)

## Thiết kế

### Tổng quan kiến trúc

```
┌────────────────────────────────────────────────────────────┐
│                    Lớp ứng dụng                            │
│                                                            │
│  ┌───────────────────────────────────────────────────┐     │
│  │            Trình quản lý bộ chuyển đổi            │     │
│  │  - Đọc cấu hình bộ chuyển đổi                     │     │
│  │  - Tạo bộ chuyển đổi giao thức                    │     │
│  │  - Quản lý vòng đời                               │     │
│  └──────┬────────┬────────┬───────────┬──────────────┘     │
│         │        │        │           │                    │
│  ┌──────▼──┐ ┌───▼─── ┐ ┌─▼──────┐ ┌──▼─────┐              │
│  │ ExtProc │ │ HTTP   │ │ gRPC   │ │ Nginx  │              │
│  │Adapter  │ │Adapter │ │Adapter │ │Adapter │              │
│  │ ┌─────┐ │ │ ┌─────┐│ │ ┌─────┐│ │ ┌─────┐│              │
│  │ │Parse│ │ │ │Parse││ │ │Parse││ │ │Parse││              │
│  │ │ExtP │ │ │ │HTTP ││ │ │gRPC ││ │ │NJS  ││              │
│  │ └──┬──┘ │ │ └─┬───┘│ │ └─┬───┘│ │ └──┬──┘│              │
│  │    │Conv│ │   │Con │ │   │Con │ │    │Con│              │
│  │    ▼    │ │   ▼    │ │   ▼    │ │    ▼   │              │
│  │ ┌─────┐ │ │ ┌────┐ │ │ ┌────┐ │ │ ┌─────┐│              │
│  │ │Req  │ │ │ │Req │ │ │ │Req │ │ │ │Req  ││              │
│  │ └──┬──┘ │ │ └─┬──┘ │ │ └─┬──┘ │ │ └──┬──┘│              │
│  └────┼────┘ └───┼────┘ └───┼────┘ └────┼───┘              │
│       │          │          │          │                   │
│       └──────────┴──────────┴──────────┘                   │
│                   Điểm vào duy nhất                        │
│                             │                              │
│                             ▼                              │
│        ┌──────────────────────────────────────────┐        │
│        │          RouterEngine.Route()            │        │
│        │  1. Phân loại yêu cầu                    │        │
│        │  2. Kiểm tra PII / jailbreak             │        │
│        │  3. Kiểm tra bộ đệm                      │        │
│        │  4. Chọn công cụ                         │        │
│        │  5. Chọn mô hình/backend                 │        │
│        │  6. Ghi lại phát lại                     │        │
│        │  7. Proxy đến backend (qua Backend Layer)│        │
│        │  8. Cập nhật bộ đệm                      │        │
│        └──────────────┬───────────────────────────┘        │
│                       │                                    │
│                       ▼                                    │
│                  RouteResponse                             │
│                       │                                    │
│        ┌──────────────┼──────────────┬───────────┐         │
│        │              │              │           │         │
│  ┌─────▼─────┐ ┌──────▼────┐ ┌───────▼───┐ ┌─────▼─────┐   │
│  │ ExtProc   │ │ HTTP      │ │ gRPC      │ │ Nginx     │   │
│  │ Adapter   │ │ Adapter   │ │ Adapter   │ │ Adapter   │   │
│  │ ┌───────┐ │ │ ┌───────┐ │ │ ┌───────┐ │ │ ┌───────┐ │   │
│  │ │Convert│ │ │ │Convert│ │ │ │Convert│ │ │ │Convert│ │   │
│  │ │to gRPC│ │ │ │to HTTP│ │ │ │gRPC   │ │ │ │to NJS │ │   │
│  │ └───────┘ │ │ └───────┘ │ │ └───────┘ │ │ └───────┘ │   │
│  └─────┬─────┘ └─────┬─────┘ └─────┬─────┘ └─────┬─────┘   │
│        │             │             │             │         │
└────────┼─────────────┼─────────────┼─────────────┼─────────┘
         │             │             │             │
         └─────────────┴─────────────┴─────────────┘
                            │
                            ▼
         ┌─────────────────────────────────────────┐
         │      Lớp trừu tượng hóa Backend         │
         └──────┬──────────────────┬───────────────┘
                │                  │
       ┌────────▼────────┐  ┌──────▼──────────┐
       │ Envoy Proxy     │  │ Direct Proxy    │
       │ (ExtProc mode)  │  │ (HTTP/gRPC)     │
       │ - Dynamic fwd   │  │ - HTTP client   │
       │ - Headers only  │  │ - Full response │
       └────────┬────────┘  └──────┬──────────┘
                │                  │
                └──────────┬───────┘
                           ▼
             ┌────────────────────────────┐
             │      Inference Backends    │
             │  ┌────────┐  ┌────────┐    │
             │  │ vLLM   │  │Ollama  │    │
             │  │Server  │  │Server  │    │
             │  └────────┘  └────────┘    │
             └────────────────────────────┘
```

**Cái nhìn sâu sắc:** Các bộ chuyển đổi là **lớp bản dịch mỏng**. Tất cả sự thông minh sống trong RouterEngine.

### Thiết kế thành phần

#### 1. RouterEngine (Inti)

**Vị trí:** `pkg/router/engine/`

**Trách nhiệm:**

- Logic định tuyến không phụ thuộc giao thức
- Phân loại yêu cầu và đánh giá quyết định
- Hoạt động bộ đệm ngữ nghĩa
- Lựa chọn công cụ và nhúng
- Ghi lại phát lại định tuyến
- Phát hiện PII và jailbreak
- Lựa chọn mô hình

**Phương thức chính:**

```go
type RouterEngine struct {
    Config               *config.RouterConfig
    Classifier           *classification.Classifier
    PIIChecker           *pii.PolicyChecker
    Cache                cache.CacheBackend
    ToolsDatabase        *tools.ToolsDatabase
    ModelSelector        *selection.Registry
    ReplayRecorders      map[string]*routerreplay.Recorder
}

func (e *RouterEngine) Route(ctx context.Context, req *RouteRequest) (*RouteResponse, error)
func (e *RouterEngine) ClassifyRequest(ctx context.Context, messages []Message) (*ClassificationResult, error)
func (e *RouterEngine) CheckCache(ctx context.Context, model, query, decisionName string) (string, bool, error)
func (e *RouterEngine) UpdateCache(ctx context.Context, model, query, response, decisionName string) error
func (e *RouterEngine) SelectTools(ctx context.Context, query string, topK int) ([]openai.ChatCompletionToolParam, error)
func (e *RouterEngine) RecordReplay(ctx context.Context, decisionName string, record *routerreplay.RoutingRecord) error
```

**Quyết định thiết kế:**

- Một thể hiện duy nhất được chia sẻ trên tất cả các bộ chuyển đổi
- Có trạng thái (duy trì bộ đệm, bộ ghi phát lại)
- Không có logic cụ thể giao thức
- Trả về cấu trúc dữ liệu không phụ thuộc giao thức

#### 2. Giao diện bộ chuyển đổi

**Vị trí:** `pkg/adapter/manager.go`

```go
type Adapter interface {
    Start() error                      // Khởi động bộ chuyển đổi (chặn)
    Stop() error                       // Tắt nhẹ nhàng
    GetEngine() *engine.RouterEngine  // Truy cập đến công cụ được chia sẻ
}
```

**Quyết định thiết kế:**

- Giao diện tối thiểu để có tính linh hoạt tối đa
- Mỗi bộ chuyển đổi sở hữu vòng đời của nó
- Không có phương thức cụ thể giao thức trong giao diện
- Các bộ chuyển đổi chạy trong goroutine riêng

#### 3. Trình quản lý bộ chuyển đổi

**Vị trí:** `pkg/adapter/manager.go`

**Trách nhiệm:**

- Phân tích cú pháp cấu hình bộ chuyển đổi
- Khởi tạo bộ chuyển đổi dựa trên cấu hình
- Bắt đầu các bộ chuyển đổi trong các goroutine riêng
- Tọa độ tắc nhẹ nhàng

**Phương thức chính:**

```go
func (m *Manager) CreateAdapters(cfg *config.RouterConfig, eng *engine.RouterEngine, configPath string) error
func (m *Manager) StartAll() error
func (m *Manager) StopAll() error
func (m *Manager) Wait()
```

#### 4. Bộ chuyển đổi ExtProc

**Vị trí:** `pkg/adapter/extproc/`

**Trách nhiệm:**

- Bao bọc triển khai ExtProc hiện có của Envoy
- Duy trì khả năng tương thích ngược
- Xử lý các đặc tính giao thức gRPC/Envoy
- Hỗ trợ cấu hình TLS

**Tính năng chính:**

- Sử dụng `extproc.OpenAIRouter` hiện có trong
- Chuyển đổi yêu cầu Envoy thành cuộc gọi RouterEngine
- Bảo tồn tất cả chức năng ExtProc hiện có
- Hỗ trợ TLS có thể cấu hình

#### 5. Bộ chuyển đổi HTTP

**Vị trí:** `pkg/adapter/http/`

**Trách nhiệm:**

- Cung cấp API REST tương thích OpenAI
- Truy cập trực tiếp mà không cần Envoy
- Xử lý các mối quan tâm cụ thể HTTP (CORS, headers, v.v.)

**Các điểm cuối:**

- `POST /v1/chat/completions` - Chat completions
- `POST /v1/completions` - Text completions (tương lai)
- `GET /v1/models` - Danh sách mô hình có sẵn
- `POST /v1/classify` - Điểm cuối phân loại
- `POST /v1/route` - Điểm cuối quyết định định tuyến
- `GET /v1/router_replay` - Danh sách bản ghi phát lại
- `GET /v1/router_replay/{id}` - Nhận bản ghi phát lại
- `GET /health` - Kiểm tra sức khỏe
- `GET /ready` - Kiểm tra sẵn sàng

#### 6. Bộ chuyển đổi gRPC

**Vị trí:** `pkg/adapter/grpc/`

**Trách nhiệm:**

- Cung cấp API gRPC gốc cho định tuyến
- Hiệu quả hơn ExtProc cho các ứng dụng khách gRPC trực tiếp
- Định nghĩa dịch vụ tùy chỉnh được tối ưu hóa cho định tuyến
- Hỗ trợ kiểu RPC streaming và unary

**Tính năng chính:**

- Định nghĩa dịch vụ `.proto` tùy chỉnh
- Được tối ưu hóa cho quyết định định tuyến độ trễ thấp
- Hỗ trợ cả định tuyến đồng bộ và không đồng bộ
- Cân bằng tải tích hợp và kết hợp kết nối
- Tương thích với hệ sinh thái gRPC (grpc-gateway, v.v.)

**Định nghĩa dịch vụ:**

```protobuf
service SemanticRouter {
  rpc Route(RouteRequest) returns (RouteResponse);
  rpc Classify(ClassifyRequest) returns (ClassifyResponse);
  rpc StreamRoute(stream RouteRequest) returns (stream RouteResponse);
}
```

#### 7. Bộ chuyển đổi Nginx

**Vị trí:** `pkg/adapter/nginx/`

**Trách nhiệm:**

- Tích hợp với Nginx thông qua mô-đun NJS (JavaScript)
- Hỗ trợ kịch bản Lua cho triển khai OpenResty
- Định tuyến dựa trên tiêu đề tương tự như ExtProc
- Cấu hình upstream trực tiếp Nginx

**Tính năng chính:**

- Mô-đun NJS để xử lý yêu cầu/phản hồi
- Giao tiếp với RouterEngine thông qua HTTP hoặc gRPC
- Đặt lựa chọn upstream dựa trên quyết định định tuyến
- Overhead tối thiểu so với ExtProc
- Đặc điểm hiệu suất Nginx gốc

**Phương thức tích hợp:**

1. **Mô-đun NJS:** Xử lý yêu cầu dựa trên JavaScript
2. **Lua/OpenResty:** Cho triển khai OpenResty
3. **HTTP Subrequest:** Gọi bộ chuyển đổi HTTP trong
4. **Bộ nhớ được chia sẻ:** IPC trực tiếp với quá trình RouterEngine

### Thiết kế cấu hình

#### Cấu hình bộ chuyển đổi

```yaml
adapters:
  - type: "envoy" # Bộ chuyển đổi ExtProc
    enabled: true
    port: 50051
    tls:
      enabled: true
      cert_file: "/path/to/cert.pem"
      key_file: "/path/to/key.pem"

  - type: "http" # HTTP REST API
    enabled: true
    port: 9000

  - type: "grpc" # API gRPC gốc
    enabled: true
    port: 50052
    tls:
      enabled: true
      cert_file: "/path/to/cert.pem"
      key_file: "/path/to/key.pem"

  - type: "nginx" # Tích hợp Nginx
    enabled: true
    port: 9001
    mode: "njs" # Tùy chọn: njs, lua, http
    config:
      upstream_variable: "backend_upstream"
      header_prefix: "x-vsr-"
```

**Quyết định thiết kế:**

- Mảng cho phép nhiều bộ chuyển đổi cùng loại (tương lai: nhiều HTTP trên các cổng khác nhau)
- Cấu hình TLS trên mỗi bộ chuyển đổi
- Bật/tắt đơn giản mà không cần loại bỏ cấu hình
- Cấu hình cổng ở mức bộ chuyển đổi

### Luồng dữ liệu

#### Luồng yêu cầu (Ví dụ về bộ chuyển đổi HTTP)

```
1. Yêu cầu ứng dụng khách
   ↓
2. Bộ chuyển đổi HTTP nhận POST /v1/chat/completions
   ↓
3. Phân tích cú pháp định dạng yêu cầu OpenAI
   ↓
4. Gọi RouterEngine.Route(RouteRequest)
   ↓
5. RouterEngine thực hiện:
   - Phân loại (quyết định nào khớp?)
   - Kiểm tra bộ đệm (tương tự ngữ nghĩa)
   - Lựa chọn công cụ (nếu được bật)
   - Ghi lại phát lại (nếu được cấu hình)
   ↓
6. RouterEngine trả về RouteResponse
   ↓
7. Bộ chuyển đổi HTTP proxy đến backend được chọn
   ↓
8. Trả lại phản hồi cho ứng dụng khách
```

#### Luồng trạng thái được chia sẻ

```
Yêu cầu HTTP A → Bộ chuyển đổi HTTP
                     ↓
                RouterEngine → Bộ đệm (hit/miss)
                     ↑
Yêu cầu ExtProc B → Bộ chuyển đổi ExtProc
```

Cả hai bộ chuyển đổi chia sẻ:

- Các mục nhập bộ đệm tương tự
- Các bộ ghi phát lại tương tự
- Các quyết định phân loại tương tự
- Trạng thái lựa chọn mô hình tương tự

## Chi tiết triển khai

### Chuỗi khởi tạo

```go
// main.go
1. Tải cấu hình
2. Khởi tạo mô hình nhúng
3. Tạo RouterEngine (NewRouterEngine)
   - Khởi tạo bộ phân loại
   - Tạo bộ đệm ngữ nghĩa
   - Cấu hình cơ sở dữ liệu công cụ
   - Khởi tạo bộ ghi phát lại trên mỗi quyết định
   - Bộ chọn mô hình thiết lập
4. Tạo Trình quản lý bộ chuyển đổi
5. Trình quản lý tạo các bộ chuyển đổi (CreateAdapters)
   - Mỗi bộ chuyển đổi nhận tham chiếu đến RouterEngine
   - Cấu hình trên mỗi bộ chuyển đổi (cổng, TLS)
6. Trình quản lý bắt đầu tất cả các bộ chuyển đổi (StartAll)
   - Mỗi bộ chuyển đổi trong goroutine riêng
7. Khối chính trên Manager.Wait()
```

### Xử lý lỗi

- **Lỗi tạo bộ chuyển đổi:** Lỗi R, ứng dụng thoát
- **Lỗi bắt đầu bộ chuyển đổi:** Lỗi fatal, ứng dụng thoát
- **Lỗi thời gian chạy:** Ghi nhật ký, bộ chuyển đổi tiếp tục nếu có thể
- **Lỗi RouterEngine:** Trả về bộ chuyển đổi để xử lý cụ thể giao thức

### Mô hình đồng thời

- **RouterEngine:** An toàn với luồng, nhiều bộ chuyển đổi có thể gọi đồng thời
- **Bộ đệm:** Backend xử lý đồng thời (Redis, Milvus, v.v.)
- **Bộ ghi phát lại:** Bản đồ an toàn với luồng với khóa trên mỗi quyết định
- **Các bộ chuyển đổi:** Goroutine độc lập, không có trạng thái bộ chuyển đổi được chia sẻ

## Cân nhắc và các lựa chọn thay thế

### Quyết định thiết kế

#### 1. RouterEngine được chia sẻ duy nhất so với ngôn các bộ chuyển đổi

**Đã chọn:** RouterEngine được chia sẻ duy nhất

**Lý do:**

- Quyết định định tuyến nhất quán trên các giao thức
- Bộ đệm được chia sẻ cải thiện tỷ lệ hit
- Một nguồn sự thật duy nhất cho các bản ghi phát lại
- Kích thước bộ nhớ giảm

**Đánh đổi:** Điểm tranh chấp tiềm năng (giảm bớt bằng thiết kế an toàn với luồng)

#### 2. Thiết kế giao diện bộ chuyển đổi

**Các lựa chọn được xem xét:**

A. **Giao diện đa dạng:**

```go
type Adapter interface {
    Start() error
    Stop() error
    HandleRequest(req *Request) (*Response, error)
    GetMetrics() *Metrics
    Configure(cfg *Config) error
}
```

B. **Giao diện tối thiểu (được chọn):**

```go
type Adapter interface {
    Start() error
    Stop() error
    GetEngine() *engine.RouterEngine
}
```

**Lý do:** Giao diện tối thiểu cho phép tính linh hoạt giao thức tối đa. Các giao thức khác nhau có các mô hình yêu cầu/phản hồi rất khác nhau.

#### 3. Phương pháp cấu hình

**Các lựa chọn:**

A. Các tệp riêng cho mỗi bộ chuyển đổi
B. Biến môi trường
C. Cấu hình duy nhất với phần bộ chuyển đổi (được chọn)

**Lý do:** Tệp cấu hình duy nhất giữ tất cả cấu hình ở một nơi, dễ quản lý và kiểm soát phiên bản.

#### 4. Khả năng tương thích ngược

**Phương pháp:** Bao bọc triển khai ExtProc hiện có thay vì viết lại

**Lý do:**

- Không có thay đổi phá vỡ
- Đường dẫn di cư dần dần
- Mã được chứng minh, đã kiểm thử vẫn được sử dụng
- Rủi ro giảm

### Hạn chế đã biết

1. **Không tối ưu hóa cụ thể giao thức:** Tính trừu tượng hóa ngăn các tối ưu hóa cụ thể giao thức
2. **Cách ly bộ chuyển đổi:** Các bộ chuyển đổi không thể giao tiếp trực tiếp (theo thiết kế)
3. **Thách thức trạng thái được chia sẻ:** Điều kiện chạy nếu RouterEngine không an toàn với luồng
4. **Phức tạp cấu hình:** Nhiều tùy chọn hơn để người dùng cấu hình

## Chiến lược kiểm thử

### Kiểm thử đơn vị

- Phương thức RouterEngine với các bộ chuyển đổi giả lập
- Logic bộ chuyển đổi cá nhân
- Phân tích cú pháp cấu hình

### Kiểm thử tích hợp

- Nhiều bộ chuyển đổi chạy đồng thời
- Tính nhất quán trạng thái được chia sẻ (bộ đệm hit trên các bộ chuyển đổi)
- Ghi lại phát lại từ cả hai giao thức

### Kiểm thử E2E

- ExtProc thông qua Envoy trên cổng 8801
- HTTP trực tiếp trên cổng 9000
- Xác minh quyết định định tuyến giống nhau
- Xác minh bản ghi phát lại hiển thị từ cả hai

## Công việc trong tương lai

### Ngắn hạn

1. **Tắc nhẹ nhàng**
   - Yêu cầu xả trong các yêu cầu
   - Đóng kết nối sạch sẽ
   - Yêu cầu xả bản ghi phát lại

2. **Chỉ số bộ chuyển đổi**
   - Bộ đếm yêu cầu trên mỗi bộ chuyển đổi
   - Biểu đồ trễ
   - Tỷ lệ lỗi

3. **Tích hợp Nginx nâng cao**
   - Mô-đun Lua OpenResty
   - API upstream động Nginx Plus
   - IPC bộ nhớ được chia sẻ để sao chép không

### Dài hạn

1. **Nâng cao gRPC Streaming**
   - Hỗ trợ streaming hai chiều
   - Streaming phía máy chủ cho các yêu cầu hàng loạt
   - Streaming phía ứng dụng khách cho các đầu vào lớn

2. **Bộ chuyển đổi WebSocket**
   - Streaming thời gian thực
   - Giao tiếp hai chiều

3. **Hệ thống plugin**
   - Tải bộ chuyển đổi động
   - Các bộ chuyển đổi của bên thứ ba

4. **Cấu hình trên mỗi bộ chuyển đổi**
   - Giới hạn tỷ lệ
   - Xác thực
   - Phần mềm trung gian tùy chỉnh

## Hướng dẫn di cư

### Từ ExtProc-Only cũ đến kiến trúc bộ chuyển đổi

**Trước:**

```go
server := extproc.NewServer(configPath, port, secure, certPath)
server.Start()
```

**Sau:**

```go
engine := engine.NewRouterEngine(configPath)
manager := adapter.NewManager()
manager.CreateAdapters(cfg, engine, configPath)
manager.StartAll()
manager.Wait()
```

**Cấu hình:**

```yaml
# Thêm vào config.yaml
adapters:
  - type: "envoy"
    enabled: true
    port: 50051
```

### Thêm bộ chuyển đổi mới

1. Tạo gói `pkg/adapter/myprotocol/`
2. Triển khai giao diện `Adapter`:

```go
type MyAdapter struct {
    engine *engine.RouterEngine
    port   int
}

func NewAdapter(eng *engine.RouterEngine, port int) (*MyAdapter, error) {
    return &MyAdapter{engine: eng, port: port}, nil
}

func (a *MyAdapter) Start() error {
    // Cấu hình máy chủ cụ thể giao thức
    // Gọi a.engine.Route() cho logic định tuyến
}

func (a *MyAdapter) Stop() error {
    // Tắc nhẹ nhàng
}

func (a *MyAdapter) GetEngine() *engine.RouterEngine {
    return a.engine
}
```

3. Đăng ký trong trình quản lý:

```go
// pkg/adapter/manager.go
case "myprotocol":
    adapter, err = myprotocol.NewAdapter(eng, adapterCfg.Port)
```

4. Thêm hỗ trợ cấu hình:

```yaml
adapters:
  - type: "myprotocol"
    enabled: true
    port: 9001
```

## Tài liệu tham khảo

- [Đặc tả API OpenAI](https://platform.openai.com/docs/api-reference)
- [Tài liệu ExtProc của Envoy](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/ext_proc_filter)
- [Tài liệu vLLM Semantic Router](https://vllm-semantic-router.com)

## Phụ lục

### Xem xét hiệu suất

- **RouterEngine:** Một thể hiện duy nhất giảm bộ nhớ, nhưng có thể là nút thắt bình đỏ
- **Bộ đệm:** Lựa chọn backend rất quan trọng (Redis/Milvus cho sản xuất)
- **Ghi lại phát lại:** Ghi lại không đồng bộ được khuyến khích cho thông lượng cao
- **Overhead bộ chuyển đổi:** Tối thiểu, chủ yếu là tuần tự hóa mạng/giao thức

### Xem xét bảo mật

- **Hỗ trợ TLS:** Cấu hình TLS trên mỗi bộ chuyển đổi
- **Xác thực:** Được xử lý ở mức bộ chuyển đổi (công việc tương lai: tính trừu tượng authz bên ngoài)
- **Phê duyệt:** Công việc tương lai để tính trừu tượng các nhà cung cấp authz bên ngoài (OPA, tùy chỉnh)
- **Phát hiện PII:** Được chia sẻ trên tất cả các bộ chuyển đổi
- **Phát hiện Jailbreak:** Được chia sẻ trên tất cả các bộ chuyển đổi

### Giám sát và Khả năng quan sát

- **Chỉ số:** Chỉ số trên mỗi bộ chuyển đổi và RouterEngine
- **Tracing:** Các khoảng tracing phân tán bộ chuyển đổi
- **Logging:** Nhật ký có cấu trúc với ngữ cảnh bộ chuyển đổi
- **Kiểm tra sức khỏe:** Các điểm cuối kiểm tra sức khỏe trên mỗi bộ chuyển đổi
