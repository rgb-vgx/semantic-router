# Tích hợp RAG OpenAI

Hướng dẫn này minh họa cách sử dụng API File Store và Vector Store của OpenAI cho RAG (Retrieval-Augmented Generation) trong Semantic Router, theo các [hướng dẫn RAG của OpenAI](https://cookbook.openai.com/examples/rag_on_pdfs_using_file_search).

## Tổng quan

Phần backend RAG của OpenAI tích hợp với API File Store và Vector Store của OpenAI để cung cấp trải nghiệm RAG hạng nhất. Nó hỗ trợ hai chế độ quy trình làm việc:

1. **Chế độ tìm kiếm trực tiếp** (mặc định): Truy xuất đồng bộ bằng API tìm kiếm kho lưu trữ vectơ
2. **Chế độ dựa trên công cụ**: Thêm công cụ `file_search` vào yêu cầu (quy trình làm việc API Phản hồi)

## Kiến trúc

```
┌─────────────┐
│   Ứng dụng  │
└──────┬──────┘
       │
       ▼
┌─────────────────────────────────────┐
│     Semantic Router                 │
│  ┌───────────────────────────────┐  │
│  │      Plugin RAG               │  │
│  │  ┌─────────────────────────┐  │  │
│  │  │  Backend RAG OpenAI     │  │  │
│  │  └──────┬──────────────────┘  │  │
│  └─────────┼──────────────────── ┘  │
└────────────┼─────────────────────── ┘
             │
             ▼
┌─────────────────────────────────────┐
│      API OpenAI                     │
│  ┌──────────────┐  ┌──────────────┐ │
│  │ File Store   │  │Vector Store  │ │
│  │   API        │  │   API        │ │
│  └──────────────┘  └──────────────┘ │
└─────────────────────────────────────┘
```

## Điều kiện tiên quyết

1. Khóa API OpenAI có quyền truy cập vào File Store và Vector Store APIs
2. Tệp được tải lên File Store của OpenAI
3. Kho lưu trữ vectơ được tạo và điền với các tệp

## Cấu hình

### Cấu hình cơ bản

Thêm backend RAG OpenAI vào cấu hình quyết định của bạn:

```yaml
decisions:
  - name: rag-openai-decision
    signals:
      - type: keyword
        keywords: ["research", "document", "knowledge"]
    plugins:
      rag:
        enabled: true
        backend: "openai"
        backend_config:
          vector_store_id: "vs_abc123"  # ID kho lưu trữ vectơ của bạn
          api_key: "${OPENAI_API_KEY}"  # Hoặc sử dụng biến môi trường
          max_num_results: 10
          workflow_mode: "direct_search"  # hoặc "tool_based"
```

### Cấu hình nâng cao

```yaml
rag:
  enabled: true
  backend: "openai"
  similarity_threshold: 0.7
  top_k: 10
  max_context_length: 5000
  injection_mode: "tool_role"  # hoặc "system_prompt"
  on_failure: "skip"  # hoặc "warn" hoặc "block"
  cache_results: true
  cache_ttl_seconds: 3600
  backend_config:
    vector_store_id: "vs_abc123"
    api_key: "${OPENAI_API_KEY}"
    base_url: "https://api.openai.com/v1"  # Tùy chọn, mặc định là OpenAI
    max_num_results: 10
    file_ids:  # Tùy chọn: giới hạn tìm kiếm cho các tệp cụ thể
      - "file-123"
      - "file-456"
    filter:  # Tùy chọn: bộ lọc siêu dữ liệu
      category: "research"
      published_date: "2024-01-01"
    workflow_mode: "direct_search"  # hoặc "tool_based"
    timeout_seconds: 30
```

## Chế độ quy trình làm việc

### 1. Chế độ tìm kiếm trực tiếp (Mặc định)

Truy xuất đồng bộ bằng API tìm kiếm kho lưu trữ vectơ. Nội dung được truy xuất trước khi gửi yêu cầu đến LLM.

**Trường hợp sử dụng**: Khi bạn cần tiêm bối cảnh ngay lập tức và muốn kiểm soát quá trình truy xuất.

**Ví dụ**:

```yaml
backend_config:
  workflow_mode: "direct_search"
  vector_store_id: "vs_abc123"
```

**Luồng**:

1. Người dùng gửi truy vấn
2. Plugin RAG gọi API tìm kiếm kho lưu trữ vectơ
3. Bối cảnh được truy xuất được tiêm vào yêu cầu
4. Yêu cầu được gửi đến LLM với bối cảnh

### 2. Chế độ dựa trên công cụ (Responses API)

Thêm công cụ `file_search` vào yêu cầu. LLM gọi công cụ tự động, và kết quả xuất hiện trong chú thích phản hồi.

**Trường hợp sử dụng**: Khi sử dụng Responses API và muốn LLM kiểm soát khi nào tìm kiếm.

**Ví dụ**:

```yaml
backend_config:
  workflow_mode: "tool_based"
  vector_store_id: "vs_abc123"
```

**Luồng**:

1. Người dùng gửi truy vấn
2. Plugin RAG thêm công cụ `file_search` vào yêu cầu
3. Yêu cầu được gửi đến LLM
4. LLM gọi công cụ `file_search`
5. Kết quả xuất hiện trong chú thích phản hồi

## Ví dụ sử dụng

### Ví dụ 1: Truy vấn RAG cơ bản

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "X-VSR-Selected-Decision: rag-openai-decision" \
  -d '{
    "model": "gpt-4o-mini",
    "messages": [
      {
        "role": "user",
        "content": "Deep Research là gì?"
      }
    ]
  }'
```

### Ví dụ 2: Responses API có công cụ file_search

```bash
curl -X POST http://localhost:8080/v1/responses \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o-mini",
    "input": "Deep Research là gì?",
    "tools": [
      {
        "type": "file_search",
        "file_search": {
          "vector_store_ids": ["vs_abc123"],
          "max_num_results": 5
        }
      }
    ]
  }'
```

### Ví dụ 3: Python Client

```python
import requests

# Chế độ tìm kiếm trực tiếp
response = requests.post(
    "http://localhost:8080/v1/chat/completions",
    headers={
        "Content-Type": "application/json",
        "X-VSR-Selected-Decision": "rag-openai-decision"
    },
    json={
        "model": "gpt-4o-mini",
        "messages": [
            {"role": "user", "content": "Deep Research là gì?"}
        ]
    }
)

result = response.json()
print(result["choices"][0]["message"]["content"])
```

## Hoạt động File Store

Backend RAG OpenAI bao gồm một ứng dụng khách File Store để quản lý các tệp:

### Tải lên tệp

```go
import "github.com/vllm-project/semantic-router/src/semantic-router/pkg/openai"

client := openai.NewFileStoreClient("https://api.openai.com/v1", apiKey)
file, err := client.UploadFile(ctx, fileReader, "document.pdf", "assistants")
```

### Tạo kho lưu trữ vectơ

```go
vectorStoreClient := openai.NewVectorStoreClient("https://api.openai.com/v1", apiKey)
store, err := vectorStoreClient.CreateVectorStore(ctx, &openai.CreateVectorStoreRequest{
    Name:    "my-vector-store",
    FileIDs: []string{"file-123", "file-456"},
})
```

### Đính kèm tệp vào kho lưu trữ vectơ

```go
_, err := vectorStoreClient.CreateVectorStoreFile(ctx, "vs_abc123", "file-123")
```

## Kiểm thử

### Kiểm thử đơn vị

Chạy kiểm thử đơn vị cho RAG OpenAI:

```bash
cd src/semantic-router
go test ./pkg/openai/... -v
go test ./pkg/extproc/req_filter_rag_openai_test.go -v
```

### Kiểm thử E2E

Chạy kiểm thử E2E dựa trên sổ tay OpenAI:

```bash
# Kiểm thử E2E dựa trên Python
python e2e/testing/08-rag-openai-test.py --base-url http://localhost:8080

# Kiểm thử E2E dựa trên Go (yêu cầu cụm Kubernetes)
make e2e-test E2E_TESTS=rag-openai
```

### Bộ kiểm thử xác thực API OpenAI

Các bài kiểm thử xác thực đảm bảo triển khai API OpenAI (Files, Vector Stores, Search) vẫn tương thích với upstream. Được chuyển thể từ [openai-python/tests](https://github.com/openai/openai-python/tree/main/tests). Chạy khi `OPENAI_API_KEY` được đặt.

**E2E Python (xác thực hợp đồng so với API thực):**

```bash
# Từ gốc repo; bỏ qua tất cả các bài kiểm thử nếu OPENAI_API_KEY không được đặt
OPENAI_API_KEY=sk-... python e2e/testing/09-openai-api-validation-test.py --verbose

# Tùy chọn: ghi đè URL cơ sở API
OPENAI_BASE_URL=https://api.openai.com/v1 OPENAI_API_KEY=sk-... python e2e/testing/09-openai-api-validation-test.py
```

**Tích hợp Go (ứng dụng khách pkg/openai so với API thực):**

```bash
cd src/semantic-router
# Bỏ qua các bài kiểm thử nếu OPENAI_API_KEY không được đặt
OPENAI_API_KEY=sk-... go test -tags=openai_validation ./pkg/openai -v
```

Các bài kiểm thử bao gồm: Files (list, upload, get, delete), Vector Stores (list, create, get, update, delete), Vector Store Files (list), and Vector Store Search (response schema).

## Giám sát và Khả năng quan sát

Backend RAG OpenAI phơi bày các chỉ số sau:

- `rag_retrieval_attempts_total{backend="openai", decision="...", status="success|error"}`
- `rag_retrieval_latency_seconds{backend="openai", decision="..."}`
- `rag_similarity_score{backend="openai", decision="..."}`
- `rag_context_length_chars{backend="openai", decision="..."}`
- `rag_cache_hits_total{backend="openai"}`
- `rag_cache_misses_total{backend="openai"}`

### Tracing

Các khoảng OpenTelemetry được tạo cho:

- `semantic_router.rag.retrieval` - Hoạt động truy xuất RAG
- `semantic_router.rag.context_injection` - Hoạt động tiêm bối cảnh

## Xử lý lỗi

Plugin RAG hỗ trợ ba chế độ thất bại:

- **skip** (mặc định): Tiếp tục mà không có bối cảnh, ghi lại cảnh báo
- **warn**: Tiếp tục cảnh báo header
- **block**: Trả lại phản hồi lỗi (503)

```yaml
rag:
  on_failure: "skip"  # hoặc "warn" hoặc "block"
```

## Thực tiễn tốt nhất

1. **Sử dụng Tìm kiếm trực tiếp cho Quy trình làm việc đồng bộ**: Khi bạn cần tiêm bối cảnh ngay lập tức
2. **Sử dụng Dựa trên công cụ cho Responses API**: Khi sử dụng Responses API và muốn tìm kiếm do LLM kiểm soát
3. **Kết quả bộ đệm**: Bật caching cho các truy vấn thường xuyên truy cập
4. **Đặt thời gian chờ thích hợp**: Cấu hình `timeout_seconds` dựa trên kích thước kho lưu trữ vectơ của bạn
5. **Lọc kết quả**: Sử dụng `file_ids` hoặc `filter` để tìm kiếm phạm vi hẹp
6. **Tracing**: Theo dõi độ trễ truy xuất và điểm số tương tự

## Khắc phục sự cố

### Không tìm thấy kết quả

- Xác minh ID kho lưu trữ vectơ chính xác
- Kiểm tra xem các tệp có được đính kèm với kho lưu trữ vectơ không
- Đảm bảo các tệp đã hoàn thành xử lý (kiểm tra `file_counts.completed`)

### Độ trễ cao

- Giảm `max_num_results`
- Bật bộ đệm kết quả
- Sử dụng `file_ids` để giới hạn phạm vi tìm kiếm

### Lỗi xác thực

- Xác minh khóa API chính xác
- Kiểm tra xem khóa API có quyền truy cập vào API File Store và Vector Store không
- Đảm bảo URL cơ sở chính xác (nếu sử dụng điểm cuối tùy chỉnh)

## Tài liệu tham khảo

- [Sổ tay Responses API OpenAI - RAG trên PDF](https://cookbook.openai.com/examples/rag_on_pdfs_using_file_search)
- [Tài liệu API File Store OpenAI](https://platform.openai.com/docs/api-reference/files)
- [Tài liệu API Vector Store OpenAI](https://platform.openai.com/docs/api-reference/vector-stores)
