# Định tuyến Tín hiệu Ưu tiên

Hướng dẫn này hướng dẫn bạn cách định tuyến các yêu cầu bằng cách sử dụng khớp ưu tiên dựa trên LLM. Tín hiệu ưu tiên sử dụng một LLM bên ngoài để phân tích ý định phức tạp và đưa ra các quyết định định tuyến tinh tế.

## Lợi ích chính

- **Phân tích Ý định Phức tạp**: Sử dụng suy luận LLM cho các quyết định định tuyến tinh tế
- **Logic Linh hoạt**: Xác định ưu tiên định tuyến bằng ngôn ngữ tự nhiên
- **Độ chính xác Cao**: Phát hiện ý định phức tạp với độ chính xác 90-98%
- **Có thể mở rộng**: Thêm các ưu tiên mới mà không cần huấn luyện lại mô hình

## Vấn đề nó giải quyết là gì?

Một số quyết định định tuyến quá phức tạp đối với việc khớp mẫu hoặc phân loại đơn giản:

- **Ý định Tinh tế**: "Giải thích những hàm ý triết học của cơ học lượng tử"
- **Truy vấn Đa khía cạnh**: "So sánh và đối chiếu chủ nghĩa lợi ích và lý thuyết bổn phận"
- **Phụ thuộc Ngữ cảnh**: "Cách tiếp cận tốt nhất cho vấn đề này là gì?"

Tín hiệu ưu tiên sử dụng một LLM bên ngoài để phân tích các truy vấn phức tạp này và khớp chúng với các ưu tiên định tuyến, cho phép bạn:

1. Xử lý ý định phức tạp mà các tín hiệu khác bỏ lỡ
2. Đưa ra các quyết định định tuyến tinh tế dựa trên suy luận LLM
3. Xác định logic định tuyến bằng ngôn ngữ tự nhiên
4. Thích ứng với các trường hợp sử dụng mới mà không cần huấn luyện lại

## Cấu hình

### Cấu hình cơ bản

Xác định các tín hiệu ưu tiên trong `config.yaml` của bạn:

```yaml
signals:
  preferences:
    - name: "code_generation"
      description: "Generating new code snippets, writing functions, creating classes"

    - name: "bug_fixing"
      description: "Identifying and fixing errors, debugging issues, troubleshooting problems"

    - name: "code_review"
      description: "Reviewing code quality, suggesting improvements, best practices"

    - name: "other"
      description: "Irrelevant queries or already fulfilled requests"
```

### Cấu hình LLM Bên ngoài

Cấu hình LLM bên ngoài cho khớp ưu tiên trong `router-defaults.yaml`:

```yaml
# External models configuration
# Used for advanced routing signals like preference-based routing via external LLM
external_models:
  - llm_provider: "vllm"
    model_role: "preference"
    llm_endpoint:
      address: "127.0.0.1"
      port: 8000
    llm_model_name: "openai/gpt-oss-120b"
    llm_timeout_seconds: 30
    parser_type: "json"
    access_key: ""  # Optional: for Authorization header (Bearer token)
```

### Sử dụng trong các Quy tắc Quyết định

```yaml
decisions:
  - name: preference_code_generation
    description: "Route code generation requests based on LLM preference matching"
    priority: 200
    rules:
      operator: "AND"
      conditions:
        - type: "preference"
          name: "code_generation"
    modelRefs:
      - model: "openai/gpt-oss-120b"
        use_reasoning: false
    plugins:
      - type: "system_prompt"
        configuration:
          system_prompt: "You are an expert code generator. Write clean, efficient, and well-documented code."

  - name: preference_bug_fixing
    description: "Route bug fixing requests based on LLM preference matching"
    priority: 200
    rules:
      operator: "AND"
      conditions:
        - type: "preference"
          name: "bug_fixing"
    modelRefs:
      - model: "openai/gpt-oss-120b"
        use_reasoning: true
    plugins:
      - type: "system_prompt"
        configuration:
          system_prompt: "You are an expert debugger. Analyze the issue carefully, identify the root cause, and provide a clear fix with explanation."
```

## Cách nó hoạt động

### 1. Phân tích Truy vấn

LLM bên ngoài phân tích truy vấn:

```
Truy vấn: "Giải thích những hàm ý triết học của cơ học lượng tử"

Phân tích LLM:
- Yêu cầu suy luận sâu: CÓ
- Mức độ phức tạp: CAO
- Miền: Triết học + Vật lý
- Loại suy luận: Phân tích, khái niệm
```

### 2. Khớp Ưu tiên

LLM khớp truy vấn với các ưu tiên được xác định:

```yaml
preferences:
  - name: "complex_reasoning"
    description: "Requires deep reasoning and analysis"
    # LLM evaluates: Does this query require deep reasoning?
    # Result: YES (confidence: 0.95)
```

### 3. Quyết định Định tuyến

Dựa trên kết quả khớp, truy vấn được định tuyến:

```
Ưu tiên khớp: complex_reasoning (0.95)
Quyết định: deep_reasoning
Mô hình: reasoning-specialist
```

## Các trường hợp sử dụng

### 1. Nghiên cứu Học thuật - Phân tích Phức tạp

**Vấn đề**: Các truy vấn nghiên cứu cần suy luận sâu và phân tích

```yaml
signals:
  preferences:
    - name: "research_analysis"
      description: "Academic research requiring deep analysis and critical thinking"

  domains:
    - name: "philosophy"
      description: "Philosophical queries"
      mmlu_categories: ["philosophy", "formal_logic"]

decisions:
  - name: academic_research
    description: "Route academic research queries"
    priority: 200
    rules:
      operator: "AND"
      conditions:
        - type: "domain"
          name: "philosophy"
        - type: "preference"
          name: "research_analysis"
    modelRefs:
      - model: "openai/gpt-oss-120b"
        use_reasoning: true
    plugins:
      - type: "system_prompt"
        configuration:
          system_prompt: "You are an academic research specialist with expertise in critical analysis and philosophical reasoning."
```

**Truy vấn ví dụ**:

- "Phân tích các hàm ý nhận thức của Phê phán Kant" → ✅ Phân tích phức tạp
- "Triết học là gì?" → ❌ Định nghĩa đơn giản

### 2. Chiến lược Kinh doanh - Quyết định

**Vấn đề**: Các truy vấn chiến lược cần phân tích tinh tế

```yaml
signals:
  preferences:
    - name: "strategic_thinking"
      description: "Business strategy requiring multi-faceted analysis"

  keywords:
    - name: "business_keywords"
      operator: "OR"
      keywords: ["strategy", "market", "competition", "growth"]
      case_sensitive: false

decisions:
  - name: strategic_analysis
    description: "Route strategic business queries"
    priority: 200
    rules:
      operator: "AND"
      conditions:
        - type: "keyword"
          name: "business_keywords"
        - type: "preference"
          name: "strategic_thinking"
    modelRefs:
      - model: "openai/gpt-oss-120b"
        use_reasoning: true
    plugins:
      - type: "system_prompt"
        configuration:
          system_prompt: "You are a senior business strategist with expertise in market analysis and competitive strategy."
```

**Truy vấn ví dụ**:

- "Phân tích vị trí cạnh tranh của chúng tôi và đề xuất các chiến lược tăng trưởng" → ✅ Chiến lược
- "Doanh thu của chúng tôi là bao nhiêu?" → ❌ Truy vấn đơn giản

### 3. Kiến trúc Kỹ thuật - Quyết định Thiết kế

**Vấn đề**: Các quyết định kiến trúc yêu cầu suy luận kỹ thuật sâu

```yaml
signals:
  preferences:
    - name: "architecture_design"
      description: "Technical architecture requiring design thinking and trade-off analysis"

  keywords:
    - name: "architecture_keywords"
      operator: "OR"
      keywords: ["architecture", "design", "scalability", "performance"]
      case_sensitive: false

decisions:
  - name: architecture_analysis
    description: "Route architecture design queries"
    priority: 200
    rules:
      operator: "AND"
      conditions:
        - type: "keyword"
          name: "architecture_keywords"
        - type: "preference"
          name: "architecture_design"
    modelRefs:
      - model: "openai/gpt-oss-120b"
        use_reasoning: true
    plugins:
      - type: "system_prompt"
        configuration:
          system_prompt: "You are a technical architecture specialist with expertise in system design, scalability, and performance optimization."
```

**Truy vấn ví dụ**:

- "Thiết kế một kiến trúc vi dịch vụ có thể mở rộng với những ưu nhược điểm" → ✅ Thiết kế
- "Dịch vụ vi mô là gì?" → ❌ Định nghĩa đơn giản

## Đặc điểm Hiệu suất

| Khía cạnh | Giá trị |
|--------|-------|
| Độ trễ | 100-500ms (phụ thuộc vào LLM) |
| Độ chính xác | 90-98% |
| Chi phí | Cao hơn (lệnh gọi LLM bên ngoài) |
| Khả năng mở rộng | Bị giới hạn bởi điểm cuối LLM |

## Các thực hành tốt nhất

### 1. Sử dụng Làm Phương sách Cuối cùng

Tín hiệu ưu tiên tốn kém. Sử dụng các tín hiệu khác trước:

```yaml
decisions:
  - name: simple_math
    priority: 10
    rules:
      operator: "OR"
      conditions:
        - type: "keyword"
          name: "math_keywords"  # Fast, cheap

  - name: complex_reasoning
    priority: 5
    rules:
      operator: "OR"
      conditions:
        - type: "preference"
          name: "complex_reasoning"  # Slow, expensive
```

### 2. Kết hợp với Các Tín hiệu Khác

Sử dụng toán tử AND để giảm dương tính giả:

```yaml
rules:
  operator: "AND"
  conditions:
    - type: "domain"
      name: "philosophy"  # Fast pre-filter
    - type: "preference"
      name: "complex_reasoning"  # Expensive verification
```

### 3. Bộ nhớ đệm Phản hồi LLM

Bật bộ nhớ đệm để giảm độ trễ và chi phí:

```yaml
preferences:
  - name: "complex_reasoning"
    description: "Requires deep reasoning"
    llm_endpoint: "http://localhost:11434"
    cache_enabled: true
    cache_ttl: 3600  # 1 hour
```

### 4. Đặt Timeout Thích hợp

Ngăn các lệnh gọi LLM chậm từ chặn:

```yaml
preferences:
  - name: "complex_reasoning"
    description: "Requires deep reasoning"
    llm_endpoint: "http://localhost:11434"
    timeout: 2000  # 2 seconds
    fallback_on_timeout: false  # Don't match if timeout
```

### 5. Giám sát Hiệu suất

Theo dõi độ trễ lệnh gọi LLM và độ chính xác:

```yaml
logging:
  level: info
  preference_signals: true
  llm_latency: true
```

## Cấu hình Nâng cao

### Nhiều Điểm cuối LLM

Sử dụng các LLM khác nhau cho các ưu tiên khác nhau:

```yaml
signals:
  preferences:
    - name: "complex_reasoning"
      description: "Deep reasoning"
      llm_endpoint: "http://localhost:11434"
      model: "llama3-70b"  # Large model for complex reasoning

    - name: "simple_classification"
      description: "Simple intent classification"
      llm_endpoint: "http://localhost:11435"
      model: "llama3-8b"  # Small model for simple tasks
```

### Dấu nhắc Tùy chỉnh

Tùy chỉnh dấu nhắc LLM để có độ chính xác tốt hơn:

```yaml
preferences:
  - name: "complex_reasoning"
    description: "Requires deep reasoning"
    llm_endpoint: "http://localhost:11434"
    prompt_template: |
      Analyze the following query and determine if it requires deep reasoning and analysis.
      Query: {query}
      Answer with YES or NO and explain why.
```

## Tham chiếu

Xem [Signal-Driven Decision Architecture](../../overview/signal-driven-decisions.md) để biết kiến trúc tín hiệu hoàn chỉnh.
