# Định tuyến Tín hiệu Kiểm tra Sự thật

Hướng dẫn này hướng dẫn bạn cách định tuyến các yêu cầu dựa trên việc chúng có yêu cầu xác minh sự thật hay không. Tín hiệu fact_check giúp xác định các truy vấn thực tế cần phát hiện ảo giác hoặc kiểm tra sự thật.

## Lợi ích chính

- **Phát hiện Tự động**: Phát hiện trên cơ sở ML các truy vấn thực tế so với sáng tạo/code
- **Phòng chống Ảo giác**: Định tuyến các truy vấn thực tế đến các mô hình với xác minh
- **Tối ưu hóa Tài nguyên**: Áp dụng kiểm tra sự thật tốn kém chỉ khi cần thiết
- **Tuân thủ**: Đảm bảo độ chính xác thực tế cho các ngành được quản lý

## Vấn đề nó giải quyết là gì?

Không phải tất cả các truy vấn đều yêu cầu xác minh sự thật:

- **Truy vấn thực tế**: "Thủ đô của Pháp là gì?" → Cần xác minh
- **Truy vấn sáng tạo**: "Viết một câu chuyện về rồng" → Không cần xác minh
- **Truy vấn code**: "Viết một hàm Python" → Không cần xác minh

Tín hiệu fact_check tự động xác định truy vấn nào cần xác minh sự thật, cho phép bạn:

1. Định tuyến các truy vấn thực tế đến các mô hình với phát hiện ảo giác
2. Bật các plugin kiểm tra sự thật chỉ cho các truy vấn thực tế
3. Tối ưu hóa chi phí bằng cách tránh xác minh không cần thiết

## Cấu hình

### Cấu hình cơ bản

Xác định các tín hiệu kiểm tra sự thật trong `config.yaml` của bạn:

```yaml
signals:
  fact_check:
    - name: needs_fact_check
      description: "Query contains factual claims that should be verified against context"

    - name: no_fact_check_needed
      description: "Query is creative, code-related, or opinion-based - no fact verification needed"
```

### Sử dụng trong các Quy tắc Quyết định

```yaml
decisions:
  - name: factual_queries
    description: "Route factual queries with verification"
    priority: 150
    rules:
      operator: "AND"
      conditions:
        - type: "fact_check"
          name: "needs_fact_check"
    modelRefs:
      - model: "openai/gpt-oss-120b"
        use_reasoning: true
    plugins:
      - type: "system_prompt"
        configuration:
          system_prompt: "You are a factual information specialist. Provide accurate, verifiable information with sources when possible."
      - type: "hallucination"
        configuration:
          enabled: true
          threshold: 0.7
```

## Các trường hợp sử dụng

### 1. Chăm sóc sức khỏe - Thông tin Y tế

**Vấn đề**: Các truy vấn y tế phải chính xác thực tế để tránh gây hại

```yaml
signals:
  fact_check:
    - name: needs_fact_check
      description: "Query contains factual claims that should be verified"

  domains:
    - name: "health"
      description: "Medical and health queries"
      mmlu_categories: ["health"]

decisions:
  - name: verified_medical
    description: "Medical queries with fact verification"
    priority: 200
    rules:
      operator: "AND"
      conditions:
        - type: "domain"
          name: "health"
        - type: "fact_check"
          name: "needs_fact_check"
    modelRefs:
      - model: "openai/gpt-oss-120b"
        use_reasoning: true
    plugins:
      - type: "system_prompt"
        configuration:
          system_prompt: "You are a medical information specialist. Provide accurate, evidence-based health information."
      - type: "hallucination"
        configuration:
          enabled: true
          threshold: 0.8  # High threshold for medical
```

**Truy vấn ví dụ**:

- "Các triệu chứng của bệnh tiểu đường là gì?" → ✅ Định tuyến với xác minh
- "Viết một câu chuyện về một bác sĩ" → ❌ Sáng tạo, không cần xác minh

### 2. Dịch vụ Tài chính - Thông tin Đầu tư

**Vấn đề**: Lời khuyên tài chính phải chính xác để tuân thủ các quy định

```yaml
signals:
  fact_check:
    - name: needs_fact_check
      description: "Query contains factual claims that should be verified"

  keywords:
    - name: "financial_keywords"
      operator: "OR"
      keywords: ["stock", "investment", "portfolio", "dividend"]
      case_sensitive: false

decisions:
  - name: verified_financial
    description: "Financial queries with verification"
    priority: 200
    rules:
      operator: "AND"
      conditions:
        - type: "keyword"
          name: "financial_keywords"
        - type: "fact_check"
          name: "needs_fact_check"
    modelRefs:
      - model: "openai/gpt-oss-120b"
        use_reasoning: true
    plugins:
      - type: "system_prompt"
        configuration:
          system_prompt: "You are a financial information specialist. Provide accurate financial information with appropriate disclaimers."
      - type: "hallucination"
        configuration:
          enabled: true
          threshold: 0.8
```

**Truy vấn ví dụ**:

- "Tỷ lệ P/E hiện tại của Apple là bao nhiêu?" → ✅ Thực tế, được xác minh
- "Giải thích các chiến lược đầu tư" → ❌ Lời khuyên chung, không cần xác minh

### 3. Giáo dục - Sự thật Lịch sử

**Vấn đề**: Nội dung giáo dục phải chính xác thực tế

```yaml
signals:
  fact_check:
    - name: needs_fact_check
      description: "Query contains factual claims that should be verified"

  domains:
    - name: "history"
      description: "Historical queries"
      mmlu_categories: ["history"]

decisions:
  - name: verified_history
    description: "Historical queries with verification"
    priority: 150
    rules:
      operator: "AND"
      conditions:
        - type: "domain"
          name: "history"
        - type: "fact_check"
          name: "needs_fact_check"
    modelRefs:
      - model: "openai/gpt-oss-120b"
        use_reasoning: true
    plugins:
      - type: "system_prompt"
        configuration:
          system_prompt: "You are a history specialist. Provide accurate historical information with proper context."
      - type: "hallucination"
        configuration:
          enabled: true
          threshold: 0.7
```

**Truy vấn ví dụ**:

- "Thế chiến thứ hai kết thúc khi nào?" → ✅ Thực tế, được xác minh
- "Viết một câu chuyện lịch sử hư cấu" → ❌ Sáng tạo, không cần xác minh

## Đặc điểm Hiệu suất

| Khía cạnh | Giá trị |
|--------|-------|
| Độ trễ | 20-50ms |
| Độ chính xác | 80-90% |
| Dương tính giả | 5-10% (sáng tạo được đánh dấu thành thực tế) |
| Âm tính giả | 5-10% (thực tế được đánh dấu thành sáng tạo) |

## Các thực hành tốt nhất

### 1. Kết hợp với Tín hiệu Miền

Sử dụng cả tín hiệu fact_check và domain để có độ chính xác tốt hơn:

```yaml
rules:
  operator: "AND"
  conditions:
    - type: "domain"
      name: "science"
    - type: "fact_check"
      name: "needs_verification"
```

### 2. Đặt Ưu tiên Thích hợp

Truy vấn thực tế nên có ưu tiên cao:

```yaml
decisions:
  - name: verified_factual
    priority: 100  # High priority
    rules:
      operator: "AND"
      conditions:
        - type: "fact_check"
          name: "needs_verification"
```

### 3. Bật Phát hiện Ảo giác

Luôn bật plugin ảo giác cho các truy vấn thực tế:

```yaml
plugins:
  - type: "hallucination"
    configuration:
      enabled: true
      threshold: 0.7
```

### 4. Giám sát Dương tính giả/Âm tính giả

Theo dõi các truy vấn được phân loại sai:

```yaml
logging:
  level: debug
  fact_check: true
```

## Tham chiếu

Xem [Signal-Driven Decision Architecture](../../overview/signal-driven-decisions.md) để biết kiến trúc tín hiệu hoàn chỉnh.
