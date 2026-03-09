# Định tuyến Dựa trên Từ khóa

Hướng dẫn này hướng dẫn bạn cách định tuyến các yêu cầu bằng cách sử dụng các quy tắc từ khóa rõ ràng và các mẫu regex. Định tuyến từ khóa cung cấp các quyết định định tuyến minh bạch, có thể kiểm toán được và essential cho tuân thủ, an ninh và các tình huống cần AI có thể giải thích.

## Lợi ích chính

- **Minh bạch**: Các quyết định định tuyến có thể giải thích hoàn toàn và có thể kiểm toán được
- **Tuân thủ**: Hành vi quy định gặp các yêu cầu quy định (GDPR, HIPAA, SOC2)
- **Nhanh**: Độ trễ dưới một mili giây, không có chi phí suy luận ML
- **Có thể giải thích**: Các quy tắc rõ ràng làm cho gỡ lỗi và xác thực đơn giản

## Vấn đề nó giải quyết là gì?

Phân loại dựa trên ML là một hộp đen khó kiểm toán và giải thích. Định tuyến từ khóa cung cấp:

- **Quyết định có thể giải thích**: Biết chính xác tại sao một truy vấn được định tuyến đến một danh mục cụ thể
- **Tuân thủ quy định**: Các nhân viên kiểm toán có thể xác minh logic định tuyến đáp ứng các yêu cầu
- **Hành vi quy định**: Cùng một đầu vào luôn tạo ra cùng một đầu ra
- **Độ trễ không**: Không có suy luận mô hình, phân loại tức thời
- **Kiểm soát chính xác**: Các quy tắc rõ ràng cho an ninh, tuân thủ và logic kinh doanh

## Khi nào sử dụng

- **Các ngành được quản lý** (tài chính, chăm sóc sức khỏe, pháp lý) yêu cầu dấu vết kiểm toán
- **An ninh/Tuân thủ** các tình huống cần phát hiện PII quy định
- **Các hệ thống thông lượng cao** nơi độ trễ dưới một mili giây là tối quan trọng
- **Định tuyến khẩn cấp/ưu tiên** với các chỉ thị từ khóa rõ ràng
- **Dữ liệu có cấu trúc** (email, ID, đường dẫn tệp) khớp các mẫu regex

## Cấu hình

Thêm tín hiệu từ khóa vào `config.yaml` của bạn:

```yaml
# Define keyword signals
signals:
  keywords:
    - name: "urgent_keywords"
      operator: "OR"  # Match ANY keyword
      keywords: ["urgent", "immediate", "asap", "emergency"]
      case_sensitive: false

    - name: "sensitive_data_keywords"
      operator: "OR"
      keywords: ["SSN", "social security", "credit card", "password"]
      case_sensitive: false

    - name: "spam_keywords"
      operator: "OR"
      keywords: ["buy now", "free money", "click here"]
      case_sensitive: false

# Define decisions using keyword signals
decisions:
  - name: urgent_request
    description: "Route urgent requests"
    priority: 100  # High priority
    rules:
      operator: "OR"
      conditions:
        - type: "keyword"
          name: "urgent_keywords"
    modelRefs:
      - model: "openai/gpt-oss-120b"
        use_reasoning: false
    plugins:
      - type: "system_prompt"
        configuration:
          system_prompt: "You are a highly responsive assistant specialized in handling urgent requests."

  - name: sensitive_data
    description: "Route sensitive data queries"
    priority: 90
    rules:
      operator: "OR"
      conditions:
        - type: "keyword"
          name: "sensitive_data_keywords"
    modelRefs:
      - model: "openai/gpt-oss-120b"
        use_reasoning: false
    plugins:
      - type: "system_prompt"
        configuration:
          system_prompt: "You are a security-conscious assistant specialized in handling sensitive data."

  - name: filter_spam
    description: "Block spam queries"
    priority: 95
    rules:
      operator: "OR"
      conditions:
        - type: "keyword"
          name: "spam_keywords"
    modelRefs:
      - model: "openai/gpt-oss-120b"
        use_reasoning: false
    plugins:
      - type: "system_prompt"
        configuration:
          system_prompt: "This query appears to be spam. Please provide a polite response."
```

## Phương pháp

Mỗi quy tắc từ khóa có thể sử dụng một phương pháp khớp khác nhau thông qua trường `method`:

| Phương pháp | Tốt nhất cho | Dung sai Typo | Các trường Cấu hình |
|--------|----------|:---:|---------------|
| `regex` (mặc định) | Các mẫu chính xác, ranh giới từ | Không | — |
| `bm25` | Phát hiện chủ đề với danh sách từ khóa lớn | Không | `bm25_threshold` |
| `ngram` | Khớp mờ, dung sai typo | **Có** | `ngram_threshold`, `ngram_arity` |

```yaml
keyword_rules:
  # BM25: scores keywords using TF-IDF relevance ranking
  - name: "code_keywords"
    operator: "OR"
    method: "bm25"
    keywords: ["code", "function", "debug", "algorithm", "refactor"]
    bm25_threshold: 0.1
    case_sensitive: false

  # N-gram: fuzzy matching — catches typos like "urgnt" → "urgent"
  - name: "urgent_keywords"
    operator: "OR"
    method: "ngram"
    keywords: ["urgent", "immediate", "asap", "emergency"]
    ngram_threshold: 0.4
    ngram_arity: 3
    case_sensitive: false

  # Regex (default): exact substring / pattern matching
  - name: "sensitive_data_keywords"
    operator: "OR"
    keywords: ["SSN", "credit card"]
    case_sensitive: false
```

Phân loại BM25 và N-gram được hỗ trợ bởi các gói Rust (`bm25`, `ngrammatic`) thông qua FFI `nlp-binding`, vì vậy hiệu suất vẫn dưới một mili giây.

## Toán tử

- **OR**: Khớp nếu tìm thấy bất kỳ từ khóa nào
- **AND**: Khớp chỉ khi tìm thấy tất cả các từ khóa
- **NOR**: Khớp chỉ khi không tìm thấy bất kỳ từ khóa nào (loại trừ)

## Ví dụ Yêu cầu

```bash
# Urgent request (matches "urgent")
curl -X POST http://localhost:8801/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "MoM",
    "messages": [{"role": "user", "content": "I need urgent help with my account"}]
  }'

# Sensitive data (matches all keywords)
curl -X POST http://localhost:8801/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "MoM",
    "messages": [{"role": "user", "content": "My SSN and credit card were stolen"}]
  }'
```

## Các trường hợp sử dụng thực tế

### 1. Dịch vụ Tài chính (Tuân thủ Minh bạch)

**Vấn đề**: Những người kiểm toán yêu cầu các quyết định định tuyến có thể giải thích cho dấu vết kiểm toán
**Giải pháp**: Quy tắc từ khóa cung cấp "tại sao" rõ ràng cho mỗi quyết định định tuyến (ví dụ: từ khóa "SSN" → trình xử lý bảo mật)
**Tác động**: Kiểm toán SOC2 đã vượt qua, minh bạch quyết định hoàn toàn

### 2. Nền tảng Chăm sóc sức khỏe (Phát hiện PII Tuân thủ)

**Vấn đề**: HIPAA yêu cầu phát hiện PII có thể xác định, có thể kiểm toán
**Giải pháp**: Toán tử AND phát hiện nhiều chỉ báo PII với các quy tắc được ghi chép
**Tác động**: Quy định 100%, dấu vết kiểm toán đầy đủ để tuân thủ

### 3. Giao dịch Tần số cao (Định tuyến Dưới một mili giây)

**Vấn đề**: Cần phân loại <1ms cho định tuyến dữ liệu thị trường thời gian thực
**Giải pháp**: Khớp từ khóa cung cấp phân loại tức thời mà không có chi phí ML
**Tác động**: Độ trễ 0,1ms, xử lý 100K+ yêu cầu/giây

### 4. Dịch vụ Chính phủ (Quy tắc Có thể giải thích)

**Vấn đề**: Những người dân cần hiểu tại sao các yêu cầu được định tuyến/từ chối
**Giải pháp**: Quy tắc từ khóa rõ ràng có thể được giải thích bằng ngôn ngữ thường
**Tác động**: Khiếu nại giảm, ra quyết định minh bạch

### 5. An ninh Doanh nghiệp (Phát hiện Mối đe dọa Minh bạch)

**Vấn đề**: Nhóm bảo mật cần hiểu tại sao các truy vấn được gắn cờ
**Giải pháp**: Quy tắc từ khóa/regex rõ ràng cho các mẫu mối đe dọa với tài liệu rõ ràng
**Tác động**: Nhóm an ninh có thể xác thực và cập nhật quy tắc một cách tự tin

## Lợi ích Hiệu suất

- **Độ trễ dưới một mili giây**: Không có chi phí suy luận ML
- **Thông lượng cao**: 100K+ yêu cầu/giây trên một lõi
- **Chi phí Dự đoán được**: Không cần GPU/mô hình nhúng
- **Bắt đầu lạnh bằng không**: Phân loại tức thời trên yêu cầu đầu tiên

## Tham chiếu

- [keyword.yaml](https://github.com/vllm-project/semantic-router/blob/main/config/intelligent-routing/in-tree/keyword.yaml) — cấu hình chỉ regex
- [keyword-nlp.yaml](https://github.com/vllm-project/semantic-router/blob/main/config/intelligent-routing/in-tree/keyword-nlp.yaml) — cấu hình BM25 + N-gram + regex
