# Định tuyến Dựa trên MCP

Hướng dẫn này hướng dẫn bạn cách triển khai logic phân loại tùy chỉnh bằng Model Context Protocol (MCP). Định tuyến MCP cho phép bạn tích hợp các dịch vụ bên ngoài, LLM hoặc logic kinh doanh tùy chỉnh cho các quyết định phân loại trong khi giữ dữ liệu của bạn riêng tư và logic định tuyến có thể mở rộng.

## Lợi ích chính

- **Độ chính xác cơ sở/cao**: Sử dụng LLM mạnh mẽ (GPT-4, Claude) cho phân loại với học tập trong ngữ cảnh
- **Có thể mở rộng**: Dễ dàng tích hợp logic phân loại tùy chỉnh mà không cần sửa đổi mã router
- **Riêng tư**: Giữ logic phân loại và dữ liệu trong cơ sở hạ tầng của riêng bạn
- **Linh hoạt**: Kết hợp suy luận LLM với các quy tắc kinh doanh, ngữ cảnh người dùng và dữ liệu bên ngoài

## Vấn đề nó giải quyết là gì?

Các bộ phân loại tích hợp bị giới hạn trong các mô hình và logic được xác định trước. Định tuyến MCP cho phép:

- **Phân loại do LLM sử dụng**: Sử dụng GPT-4/Claude cho phân loại phức tạp, tinh tế
- **Học tập trong ngữ cảnh**: Cung cấp các ví dụ và ngữ cảnh để cải thiện độ chính xác phân loại
- **Logic kinh doanh tùy chỉnh**: Triển khai các quy tắc định tuyến dựa trên cấp độ người dùng, thời gian, vị trí, lịch sử
- **Tích hợp dữ liệu bên ngoài**: Truy vấn cơ sở dữ liệu, API, cờ tính năng trong quá trình phân loại
- **Thử nghiệm nhanh chóng**: Cập nhật logic phân loại mà không cần triển khai lại router

## Khi nào sử dụng

- **Yêu cầu độ chính xác cao** nơi phân loại do LLM sử dụng vượt trội hơn BERT/embeddings
- **Các miền phức tạp** cần sự hiểu biết tinh tế vượt quá khớp từ khóa/nhúng
- **Quy tắc kinh doanh tùy chỉnh** (cấp độ người dùng, A/B kiểm tra, định tuyến dựa trên thời gian)
- **Dữ liệu riêng tư/nhạy cảm** nơi phân loại phải ở trong cơ sở hạ tầng của bạn
- **Lặp lại nhanh chóng** trên logic phân loại mà không thay đổi code

## Cấu hình

Cấu hình bộ phân loại MCP trong `config.yaml` của bạn:

```yaml
classifier:
  # Disable in-tree classifier
  category_model:
    model_id: ""

  # Enable MCP classifier
  mcp_category_model:
    enabled: true
    transport_type: "http"
    url: "http://localhost:8090/mcp"
    threshold: 0.6
    timeout_seconds: 30
    # tool_name: "classify_text"  # Optional: auto-discovers if not specified

categories: []  # Categories loaded from MCP server

default_model: openai/gpt-oss-20b

vllm_endpoints:
  - name: endpoint1
    address: 127.0.0.1
    port: 8000
    weight: 1

model_config:
  openai/gpt-oss-20b:
    reasoning_family: gpt-oss
    preferred_endpoints: [endpoint1]
```

## Cách nó hoạt động

1. **Khởi động**: Router kết nối với máy chủ MCP và gọi công cụ `list_categories`
2. **Tải Danh mục**: MCP trả về danh mục, lời nhắc hệ thống và mô tả
3. **Phân loại**: Đối với mỗi yêu cầu, router gọi công cụ `classify_text`
4. **Định tuyến**: Phản hồi MCP bao gồm danh mục, mô hình và cài đặt suy luận

### Định dạng Phản hồi MCP

**list_categories**:

```json
{
  "categories": ["math", "science", "technology"],
  "category_system_prompts": {
    "math": "You are a mathematics expert...",
    "science": "You are a science expert..."
  },
  "category_descriptions": {
    "math": "Mathematical and computational queries",
    "science": "Scientific concepts and queries"
  }
}
```

**classify_text**:

```json
{
  "class": 3,
  "confidence": 0.85,
  "model": "openai/gpt-oss-20b",
  "use_reasoning": true
}
```

## Ví dụ Máy chủ MCP

```python
from fastapi import FastAPI
from pydantic import BaseModel

app = FastAPI()

class ClassifyRequest(BaseModel):
    text: str

@app.post("/mcp/list_categories")
def list_categories():
    return {
        "categories": ["math", "science", "general"],
        "category_system_prompts": {
            "math": "You are a mathematics expert.",
            "science": "You are a science expert.",
            "general": "You are a helpful assistant."
        }
    }

@app.post("/mcp/classify_text")
def classify_text(request: ClassifyRequest):
    # Custom classification logic
    if "equation" in request.text or "solve" in request.text:
        return {
            "class": 0,  # math
            "confidence": 0.9,
            "model": "openai/gpt-oss-20b",
            "use_reasoning": True
        }
    return {
        "class": 2,  # general
        "confidence": 0.7,
        "model": "openai/gpt-oss-20b",
        "use_reasoning": False
    }
```

## Ví dụ Yêu cầu

```bash
# Math query (MCP decides routing)
curl -X POST http://localhost:8801/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "MoM",
    "messages": [{"role": "user", "content": "Solve the equation: 2x + 5 = 15"}]
  }'
```

## Lợi ích

- **Logic tùy chỉnh**: Triển khai các quy tắc phân loại cụ thể cho miền
- **Định tuyến Động**: MCP quyết định mô hình và suy luận cho mỗi truy vấn
- **Kiểm soát Tập trung**: Quản lý logic định tuyến trong dịch vụ bên ngoài
- **Khả năng mở rộng**: Phân loại quy mô độc lập với router
- **Tích hợp**: Kết nối với cơ sở hạ tầng ML hiện tại

## Các trường hợp sử dụng thực tế

### 1. Phân loại Miền Phức tạp (Độ chính xác cao)

**Vấn đề**: Các truy vấn pháp lý/y tế tinh tế cần độ chính xác tốt hơn BERT/nhúng
**Giải pháp**: MCP sử dụng GPT-4 với các ví dụ trong ngữ cảnh cho phân loại
**Tác động**: Độ chính xác 98% so với 85% với BERT, đường cơ sở để so sánh chất lượng

### 2. Logic Phân loại Độc quyền (Riêng tư)

**Vấn đề**: Logic phân loại chứa bí mật thương mại, không thể sử dụng dịch vụ bên ngoài
**Giải pháp**: Máy chủ MCP chạy trong VPC riêng, giữ tất cả logic và dữ liệu nội bộ
**Tác động**: Quyền riêng tư dữ liệu đầy đủ, không có lệnh gọi API bên ngoài

### 3. Quy tắc Kinh doanh Tùy chỉnh (Có thể mở rộng)

**Vấn đề**: Cần định tuyến dựa trên cấp độ người dùng, vị trí, thời gian, A/B test
**Giải pháp**: MCP kết hợp phân loại LLM với các truy vấn cơ sở dữ liệu và logic kinh doanh
**Tác động**: Định tuyến linh hoạt mà không cần sửa đổi mã router

### 4. Thử nghiệm Nhanh chóng (Có thể mở rộng)

**Vấn đề**: Nhóm khoa học dữ liệu cần kiểm tra các cách tiếp cận phân loại mới hàng ngày
**Giải pháp**: Máy chủ MCP được cập nhật độc lập, router không thay đổi
**Tác động**: Triển khai logic phân loại mới trong vài phút so với ngày

### 5. Nền tảng Đa người thuê (Có thể mở rộng + Riêng tư)

**Vấn đề**: Mỗi khách hàng cần phân loại tùy chỉnh, dữ liệu phải cách ly
**Giải pháp**: MCP tải các mô hình/quy tắc cụ thể cho người thuê, thực thi cách li dữ liệu
**Tác động**: 1000+ người thuê với logic tùy chỉnh, quyền riêng tư dữ liệu đầy đủ

### 6. Cách tiếp cận Lai (Độ chính xác cao + Có thể mở rộng)

**Vấn đề**: Cần độ chính xác LLM cho trường hợp cạnh, định tuyến nhanh cho truy vấn thông thường
**Giải pháp**: MCP sử dụng phản hồi được lưu trong bộ nhớ cho các mẫu thông thường, LLM cho truy vấn mới
**Tác động**: Tỷ lệ hit bộ nhớ 95%, độ chính xác LLM trên đuôi dài

## Ví dụ Máy chủ MCP Nâng cao

### Phân loại Nhận thức Ngữ cảnh

```python
@app.post("/mcp/classify_text")
def classify_text(request: ClassifyRequest, user_id: str = Header(None)):
    # Check user history
    user_history = get_user_history(user_id)

    # Adjust classification based on context
    if user_history.is_premium:
        return {
            "class": 0,
            "confidence": 0.95,
            "model": "openai/gpt-4",  # Premium model
            "use_reasoning": True
        }

    # Free tier gets fast model
    return {
        "class": 0,
        "confidence": 0.85,
        "model": "openai/gpt-oss-20b",
        "use_reasoning": False
    }
```

### Định tuyến Dựa trên Thời gian

```python
@app.post("/mcp/classify_text")
def classify_text(request: ClassifyRequest):
    current_hour = datetime.now().hour

    # Peak hours: use cached responses
    if 9 <= current_hour <= 17:
        return {
            "class": get_cached_category(request.text),
            "confidence": 0.9,
            "model": "fast-model",
            "use_reasoning": False
        }

    # Off-peak: enable reasoning
    return {
        "class": classify_with_ml(request.text),
        "confidence": 0.95,
        "model": "reasoning-model",
        "use_reasoning": True
    }
```

### Định tuyến Dựa trên Rủi ro

```python
@app.post("/mcp/classify_text")
def classify_text(request: ClassifyRequest):
    # Calculate risk score
    risk_score = calculate_risk(request.text)

    if risk_score > 0.8:
        # High risk: human review
        return {
            "class": 999,  # Special category
            "confidence": 1.0,
            "model": "human-review-queue",
            "use_reasoning": False
        }

    # Normal routing
    return standard_classification(request.text)
```

## Lợi ích so với Các bộ phân loại Tích hợp

| Tính năng | Tích hợp | MCP |
|---------|----------|-----|
| Các mô hình tùy chỉnh | ❌ | ✅ |
| Logic kinh doanh | ❌ | ✅ |
| Cập nhật động | ❌ | ✅ |
| Ngữ cảnh người dùng | ❌ | ✅ |
| A/B Kiểm tra | ❌ | ✅ |
| API bên ngoài | ❌ | ✅ |
| Độ trễ | 5-50ms | 50-200ms |
| Độ phức tạp | Thấp | Cao |

## Xem xét Hiệu suất

- **Độ trễ**: MCP thêm 50-200ms mỗi yêu cầu (mạng + phân loại)
- **Bộ nhớ đệm**: Bộ nhớ đệm phản hồi MCP cho các truy vấn lặp lại
- **Timeout**: Đặt timeout thích hợp (mặc định 30 giây)
- **Fallback**: Cấu hình mô hình mặc định khi MCP không có sẵn
- **Giám sát**: Theo dõi độ trễ MCP và tỷ lệ lỗi

## Tham chiếu

Xem [config-mcp-classifier.yaml](https://github.com/vllm-project/semantic-router/blob/main/config/intelligent-routing/out-tree/config-mcp-classifier.yaml) để xem cấu hình hoàn chỉnh.
