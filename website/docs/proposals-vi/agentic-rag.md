# Tích hợp OpenAI RAG

Hướng dẫn này trình bày cách sử dụng các API File Store và Vector Store của OpenAI cho RAG (Retrieval-Augmented Generation) trong Semantic Router, theo [OpenAI Responses API cookbook](https://cookbook.openai.com/examples/rag_on_pdfs_using_file_search).

## Tổng quan

Backend OpenAI RAG tích hợp với các API File Store và Vector Store của OpenAI để cung cấp trải nghiệm RAG hạng nhất. Nó hỗ trợ hai chế độ quy trình:

1. **Direct Search Mode** (mặc định): Truy xuất đồng bộ sử dụng vector store search API
2. **Tool-Based Mode**: Thêm công cụ `file_search` vào yêu cầu (quy trình Responses API)

## Kiến trúc

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │
       ▼
┌─────────────────────────────────────┐
│     Semantic Router                 │
│  ┌───────────────────────────────┐  │
│  │      RAG Plugin               │  │
│  │  ┌─────────────────────────┐  │  │
│  │  │  OpenAI RAG Backend     │  │  │
│  │  └──────┬──────────────────┘  │  │
│  └─────────┼──────────────────── ┘  │
└────────────┼─────────────────────── ┘
             │
             ▼
┌─────────────────────────────────────┐
│      OpenAI API                     │
│  ┌──────────────┐  ┌──────────────┐ │
│  │ File Store   │  │Vector Store  │ │
│  │   API        │  │   API        │ │
│  └──────────────┘  └──────────────┘ │
└─────────────────────────────────────┘
```

## Điều kiện tiên quyết

1. Khóa API OpenAI có quyền truy cập vào các API File Store và Vector Store
2. Tệp được tải lên OpenAI File Store
3. Vector store được tạo và điền dữ liệu từ các tệp

## Cấu hình

### Cấu hình cơ bản

Thêm backend OpenAI RAG vào cấu hình quyết định của bạn:

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
          vector_store_id: "vs_abc123"  # ID vector store của bạn
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
    file_ids:  # Tùy chọn: giới hạn tìm kiếm trên các tệp cụ thể
      - "file-123"
      - "file-456"
    filter:  # Tùy chọn: bộ lọc siêu dữ liệu
      category: "research"
      published_date: "2024-01-01"
    workflow_mode: "direct_search"  # hoặc "tool_based"
    timeout_seconds: 30
```

## Các chế độ quy trình

### 1. Direct Search Mode (Mặc định)

Truy xuất đồng bộ sử dụng vector store search API. Ngữ cảnh được truy xuất trước khi gửi yêu cầu đến LLM.

**Trường hợp sử dụng**: Khi bạn cần tiêm ngữ cảnh ngay lập tức và muốn kiểm soát quy trình truy xuất.

**Ví dụ**:

```yaml
backend_config:
  workflow_mode: "direct_search"
  vector_store_id: "vs_abc123"
```

**Luồng**:

1. Người dùng gửi truy vấn
2. Plugin RAG gọi vector store search API
3. Ngữ cảnh được truy xuất được tiêm vào yêu cầu
4. Yêu cầu được gửi đến LLM cùng với ngữ cảnh

### 2. Tool-Based Mode (Responses API)

Thêm công cụ `file_search` vào yêu cầu. LLM gọi công cụ tự động, và kết quả xuất hiện trong các chú thích của phản hồi.

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
5. Kết quả xuất hiện trong các chú thích của phản hồi

## Các ví dụ sử dụng

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
        "content": "What is Deep Research?"
      }
    ]
  }'
```

### Ví dụ 2: Responses API với công cụ file_search

```bash
curl -X POST http://localhost:8080/v1/responses \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o-mini",
    "input": "What is Deep Research?",
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

# Chế độ direct search
response = requests.post(
    "http://localhost:8080/v1/chat/completions",
    headers={
        "Content-Type": "application/json",
        "X-VSR-Selected-Decision": "rag-openai-decision"
    },
    json={
        "model": "gpt-4o-mini",
        "messages": [
            {"role": "user", "content": "What is Deep Research?"}
        ]
    }
)

result = response.json()
print(result["choices"][0]["message"]["content"])
```

## Các thao tác File Store

Backend OpenAI RAG bao gồm File Store client để quản lý tệp:

### Tải tệp lên

```go
import "github.com/vllm-project/semantic-router/src/semantic-router/pkg/openai"

client := openai.NewFileStoreClient("https://api.openai.com/v1", apiKey)
file, err := client.UploadFile(ctx, fileReader, "document.pdf", "assistants")
```

### Tạo Vector Store

```go
vectorStoreClient := openai.NewVectorStoreClient("https://api.openai.com/v1", apiKey)
store, err := vectorStoreClient.CreateVectorStore(ctx, &openai.CreateVectorStoreRequest{
    Name:    "my-vector-store",
    FileIDs: []string{"file-123", "file-456"},
})
```

### Đính kèm tệp vào Vector Store

```go
_, err := vectorStoreClient.CreateVectorStoreFile(ctx, "vs_abc123", "file-123")
```

## Kiểm thử

### Bài kiểm thử đơn vị

Chạy bài kiểm thử đơn vị cho OpenAI RAG:

```bash
cd src/semantic-router
go test ./pkg/openai/... -v
go test ./pkg/extproc/req_filter_rag_openai_test.go -v
```

### Bài kiểm thử E2E

Chạy bài kiểm thử E2E dựa trên OpenAI cookbook:

```bash
# Bài kiểm thử E2E dựa trên Python
python e2e/testing/08-rag-openai-test.py --base-url http://localhost:8080

# Bài kiểm thử E2E dựa trên Go (yêu cầu cụm Kubernetes)
make e2e-test E2E_TESTS=rag-openai
```

### Bộ kiểm thử xác thực OpenAI API

Bài kiểm thử xác thực đảm bảo việc triển khai OpenAI API (Files, Vector Stores, Search) vẫn tương thích với phiên bản mới nhất. Được điều chỉnh từ [openai-python/tests](https://github.com/openai/openai-python/tree/main/tests). Chạy khi `OPENAI_API_KEY` được đặt.

**Python E2E (xác thực hợp đồng với API thực):**

```bash
# Từ thư mục gốc của repo; bỏ qua tất cả bài kiểm thử nếu OPENAI_API_KEY không được đặt
OPENAI_API_KEY=sk-... python e2e/testing/09-openai-api-validation-test.py --verbose

# Tùy chọn: ghi đè URL cơ sở API
OPENAI_BASE_URL=https://api.openai.com/v1 OPENAI_API_KEY=sk-... python e2e/testing/09-openai-api-validation-test.py
```

**Tích hợp Go (pkg/openai client với API thực):**

```bash
cd src/semantic-router
# Bỏ qua bài kiểm thử nếu OPENAI_API_KEY không được đặt
OPENAI_API_KEY=sk-... go test -tags=openai_validation ./pkg/openai -v
```

Bài kiểm thử bao gồm: Files (list, upload, get, delete), Vector Stores (list, create, get, update, delete), Vector Store Files (list), và Vector Store Search (response schema).

## Giám sát và Khả năng quan sát

Backend OpenAI RAG cách tiếp cận các metric sau:

- `rag_retrieval_attempts_total{backend="openai", decision="...", status="success|error"}`
- `rag_retrieval_latency_seconds{backend="openai", decision="..."}`
- `rag_similarity_score{backend="openai", decision="..."}`
- `rag_context_length_chars{backend="openai", decision="..."}`
- `rag_cache_hits_total{backend="openai"}`
- `rag_cache_misses_total{backend="openai"}`

### Tracing

Các span OpenTelemetry được tạo cho:

- `semantic_router.rag.retrieval` - Thao tác truy xuất RAG
- `semantic_router.rag.context_injection` - Thao tác tiêm ngữ cảnh

## Xử lý lỗi

Plugin RAG hỗ trợ ba chế độ thất bại:

- **skip** (mặc định): Tiếp tục mà không có ngữ cảnh, ghi nhật ký cảnh báo
- **warn**: Tiếp tục với tiêu đề cảnh báo
- **block**: Trả về phản hồi lỗi (503)

```yaml
rag:
  on_failure: "skip"  # hoặc "warn" hoặc "block"
```

## Các phương pháp hay nhất

1. **Sử dụng Direct Search cho Quy trình đồng bộ**: Khi bạn cần tiêm ngữ cảnh ngay lập tức
2. **Sử dụng Tool-Based cho Responses API**: Khi sử dụng Responses API và muốn tìm kiếm được kiểm soát bởi LLM
3. **Kết quả bộ nhớ cache**: Bật bộ nhớ cache cho các truy vấn được truy cập thường xuyên
4. **Đặt Thời gian chờ thích hợp**: Cấu hình `timeout_seconds` dựa trên kích thước vector store của bạn
5. **Lọc kết quả**: Sử dụng `file_ids` hoặc `filter` để thu hẹp phạm vi tìm kiếm
6. **Giám sát Metric**: Theo dõi độ trễ truy xuất và điểm số tương tự

## Khắc phục sự cố

### Không tìm thấy kết quả

- Xác minh ID vector store là chính xác
- Kiểm tra xem các tệp có được đính kèm với vector store không
- Đảm bảo các tệp đã hoàn thành xử lý (kiểm tra `file_counts.completed`)

### Độ trễ cao

- Giảm `max_num_results`
- Bật lưu cache kết quả
- Sử dụng `file_ids` để giới hạn phạm vi tìm kiếm

### Lỗi xác thực

- Xác minh khóa API là chính xác
- Kiểm tra khóa API có quyền truy cập vào các API File Store và Vector Store
- Đảm bảo URL cơ sở là chính xác (nếu sử dụng endpoint tùy chỉnh)

## Tài liệu tham khảo

- [OpenAI Responses API Cookbook - RAG on PDFs](https://cookbook.openai.com/examples/rag_on_pdfs_using_file_search)
- [OpenAI File Store API Documentation](https://platform.openai.com/docs/api-reference/files)
- [OpenAI Vector Store API Documentation](https://platform.openai.com/docs/api-reference/vector-stores)
