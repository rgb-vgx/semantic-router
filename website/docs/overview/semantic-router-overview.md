---
sidebar_position: 2
---

# Semantic Router là gì?

**Semantic Router** là một lớp định tuyến thông minh động chọn mô hình ngôn ngữ phù hợp nhất cho mỗi truy vấn dựa trên nhiều tín hiệu được trích xuất từ yêu cầu.

## Vấn đề

Triển khai LLM truyền thống sử dụng một mô hình duy nhất cho tất cả các tác vụ:

```text
Truy vấn người dùng → Single LLM → Phản hồi
```

**Vấn đề**:

- Chi phí cao cho các truy vấn đơn giản
- Hiệu suất không tối ưu cho các tác vụ chuyên biệt
- Không có kiểm soát bảo mật hoặc tuân thủ
- Sử dụng tài nguyên kém

## Giải pháp

Semantic Router sử dụng **signal-driven decision making** để định tuyến truy vấn một cách thông minh:

```text
Truy vấn người dùng → Trích xuất tín hiệu → Công cụ quyết định → Mô hình tốt nhất → Phản hồi
```

**Lợi ích**:

- Định tuyến hiệu quả về chi phí (sử dụng mô hình nhỏ hơn cho các tác vụ đơn giản)
- Chất lượng tốt hơn (sử dụng mô hình chuyên biệt cho những điểm mạnh của chúng)
- Bảo mật tích hợp sẵn (phát hiện jailbreak, lọc PII)
- Linh hoạt và mở rộng (kiến trúc plugin)

## Cách nó hoạt động

### 1. Trích xuất tín hiệu

Bộ định tuyến trích xuất nhiều loại tín hiệu từ mỗi yêu cầu:

| Loại tín hiệu | Nó phát hiện những gì | Ví dụ |
|------------|----------------|---------|
| **keyword** | Các thuật ngữ và mẫu cụ thể | "calculate", "prove", "debug" |
| **embedding** | Ý nghĩa ngữ nghĩa | Ý định toán học, ý định mã, ý định sáng tạo |
| **domain** | Lĩnh vực kiến thức | Toán học, khoa học máy tính, lịch sử |
| **fact_check** | Cần xác minh | Các tuyên bố thực tế, lời khuyên y tế |
| **user_feedback** | Sự hài lòng của người dùng | "That's wrong", "try again" |
| **preference** | Ưu tiên định tuyến | Khớp ý định phức tạp |

### 2. Quyết định

Các tín hiệu được kết hợp bằng cách sử dụng các quy tắc logic để đưa ra quyết định định tuyến:

```yaml
decisions:
  - name: math_routing
    rules:
      operator: "AND"
      conditions:
        - type: "keyword"
          name: "math_keywords"
        - type: "domain"
          name: "mathematics"
    modelRefs:
      - model: qwen-math
        weight: 1.0
```

**Cách nó hoạt động**: Nếu truy vấn chứa từ khóa toán học **AND** được phân loại là miền toán học, hãy định tuyến tới mô hình toán học.

### 3. Lựa chọn mô hình

Dựa trên quyết định, bộ định tuyến chọn mô hình tốt nhất:

- **Math queries** → Mô hình chuyên biệt về toán (ví dụ: Qwen-Math)
- **Code queries** → Mô hình chuyên biệt về mã (ví dụ: DeepSeek-Coder)
- **Creative queries** → Mô hình sáng tạo (ví dụ: Claude)
- **Simple queries** → Mô hình nhẹ (ví dụ: Llama-3-8B)

### 4. Plugin Chain

Trước và sau khi thực thi mô hình, các plugin xử lý yêu cầu/bản phản hồi:

```yaml
plugins:
  - type: "semantic-cache"    # Check cache first
  - type: "jailbreak"         # Detect adversarial prompts
  - type: "pii"               # Filter sensitive data
  - type: "system_prompt"     # Add context
  - type: "hallucination"     # Verify facts
```

## Các khái niệm chính

### Mixture of Models (MoM)

Không giống như Mixture of Experts (MoE) hoạt động trong một mô hình duy nhất, Mixture of Models hoạt động ở **cấp độ hệ thống**:

| khía cạnh | Mixture of Experts (MoE) | Mixture of Models (MoM) |
|--------|-------------------------|------------------------|
| **Scope** | Trong một mô hình duy nhất | Trên nhiều mô hình |
| **Routing** | Mạng gating nội bộ | Bộ định tuyến ngữ nghĩa bên ngoài |
| **Models** | Kiến trúc chia sẻ | Mô hình độc lập |
| **Flexibility** | Cố định tại thời điểm đào tạo | Động tại thời chạy |
| **Use Case** | Hiệu quả mô hình | Trí thông minh cấp độ hệ thống |

### Signal-Driven Decisions

Định tuyến truyền thống sử dụng các quy tắc đơn giản:

```yaml
# Truyền thống: Khớp từ khóa đơn giản
if "math" in query:
    route_to_math_model()
```

Định tuyến dựa trên tín hiệu sử dụng nhiều tín hiệu:

```yaml
# Dựa trên tín hiệu: Nhiều tín hiệu được kết hợp
if (has_math_keywords AND is_math_domain) OR has_high_math_embedding:
    route_to_math_model()
```

**Lợi ích**:

- Định tuyến chính xác hơn
- Xử lý trường hợp cạnh tốt hơn
- Thích ứng với bảng ngữ cảnh
- Giảm dương tính giả

## Ví dụ thực tế

**Truy vấn người dùng**: "Prove that the square root of 2 is irrational"

**Trích xuất tín hiệu**:

- keyword: ["prove", "square root", "irrational"] ✓
- embedding: 0.89 tương tự với truy vấn toán học ✓
- domain: "mathematics" ✓

**Quyết định**: Định tuyến tới `qwen-math` (tất cả các tín hiệu toán học đều đồng ý)

**Plugin được áp dụng**:

- semantic-cache: Cache miss, proceed
- jailbreak: Không có mẫu đối nghịch
- system_prompt: Thêm "Provide rigorous mathematical proof"
- hallucination: Bật để xác minh

**Kết quả**: Bằng chứng toán học chất lượng cao từ mô hình chuyên biệt

## Bước tiếp theo

- [Collective Intelligence là gì?](collective-intelligence.md) - Cách các tín hiệu tạo ra trí thông minh hệ thống
- [Signal-Driven Decision là gì?](signal-driven-decisions.md) - Tìm hiểu sâu về công cụ quyết định
- [Hướng dẫn cấu hình](../installation/configuration.md) - Thiết lập bộ định tuyến ngữ nghĩa của bạn
