# Định tuyến Dựa trên Miền

Hướng dẫn này hướng dẫn bạn cách sử dụng các mô hình phân loại tinh chỉnh cho định tuyến thông minh dựa trên các miền học thuật và chuyên nghiệp. Định tuyến miền sử dụng các mô hình chuyên biệt (ModernBERT, Qwen3-Embedding, EmbeddingGemma) với các bộ điều chỉnh LoRA để phân loại truy vấn thành các danh mục như toán học, vật lý, pháp lý, kinh doanh, v.v.

## Lợi ích chính

- **Hiệu quả**: Các mô hình tinh chỉnh với các bộ điều chỉnh LoRA cung cấp suy luận nhanh (5-20ms) với độ chính xác cao
- **Chuyên biệt**: Tùy chọn mô hình đa (ModernBERT cho tiếng Anh, Qwen3 cho đa ngôn ngữ/ngữ cảnh dài, Gemma cho dấu chân nhỏ)
- **Đa nhiệm vụ**: LoRA cho phép chạy nhiều tác vụ phân loại (miền + PII + jailbreak) với mô hình cơ sở được chia sẻ
- **Hiệu quả chi phí**: Độ trễ thấp hơn so với phân loại dựa trên LLM, không có chi phí API

## Vấn đề nó giải quyết là gì?

Các cách tiếp cận phân loại chung gặp khó khăn với thuật ngữ cụ thể cho miền và sự khác biệt tinh tế giữa các trường học tập/chuyên nghiệp. Định tuyến miền cung cấp:

- **Phát hiện miền chính xác**: Các mô hình tinh chỉnh phân biệt giữa toán học, vật lý, hóa học, pháp lý, kinh doanh, v.v.
- **Hiệu quả đa nhiệm vụ**: Các bộ điều chỉnh LoRA cho phép đồng thời phân loại miền, phát hiện PII và phát hiện jailbreak với một lượt mô hình cơ sở
- **Hỗ trợ ngữ cảnh dài**: Qwen3-Embedding xử lý tới 32K mã thông báo (so với giới hạn 8K của ModernBERT)
- **Định tuyến đa ngôn ngữ**: Qwen3 được huấn luyện trên 100+ ngôn ngữ, ModernBERT được tối ưu hóa cho tiếng Anh
- **Tối ưu hóa tài nguyên**: Suy luận tốn kém chỉ được bật cho các miền được hưởng lợi (toán học, vật lý, hóa học)

## Khi nào sử dụng

- **Nền tảng giáo dục** có các khu vực đối tượng đa dạng (STEM, nhân văn, khoa học xã hội)
- **Dịch vụ chuyên nghiệp** yêu cầu chuyên môn về lĩnh vực (pháp lý, y tế, tài chính)
- **Các cơ sở kiến thức doanh nghiệp** kéo dài trên nhiều phòng ban
- **Công cụ hỗ trợ nghiên cứu** cần nhận thức về miền học tập
- **Các sản phẩm đa miền** nơi độ chính xác phân loại rất quan trọng

## Cấu hình

Cấu hình bộ phân loại miền trong `config.yaml` của bạn:

```yaml
# Define domain signals
signals:
  domains:
    - name: "math"
      description: "Mathematics and quantitative reasoning"
      mmlu_categories: ["math"]

    - name: "physics"
      description: "Physics and physical sciences"
      mmlu_categories: ["physics"]

    - name: "computer_science"
      description: "Programming and computer science"
      mmlu_categories: ["computer_science"]

    - name: "business"
      description: "Business and management"
      mmlu_categories: ["business"]

    - name: "health"
      description: "Health and medical information"
      mmlu_categories: ["health"]

    - name: "law"
      description: "Legal and regulatory information"
      mmlu_categories: ["law"]

    - name: "other"
      description: "General queries"
      mmlu_categories: ["other"]

# Define decisions using domain signals
decisions:
  - name: math
    description: "Route mathematical queries"
    priority: 100
    rules:
      operator: "OR"
      conditions:
        - type: "domain"
          name: "math"
    modelRefs:
      - model: "openai/gpt-oss-120b"
        use_reasoning: true
    plugins:
      - type: "system_prompt"
        configuration:
          system_prompt: "You are a mathematics expert. Provide step-by-step solutions with clear explanations."

  - name: physics
    description: "Route physics queries"
    priority: 100
    rules:
      operator: "OR"
      conditions:
        - type: "domain"
          name: "physics"
    modelRefs:
      - model: "openai/gpt-oss-120b"
        use_reasoning: true
    plugins:
      - type: "system_prompt"
        configuration:
          system_prompt: "You are a physics expert with deep understanding of physical laws. Explain concepts clearly with mathematical derivations."

  - name: computer_science
    description: "Route computer science queries"
    priority: 100
    rules:
      operator: "OR"
      conditions:
        - type: "domain"
          name: "computer_science"
    modelRefs:
      - model: "openai/gpt-oss-120b"
        use_reasoning: false
    plugins:
      - type: "system_prompt"
        configuration:
          system_prompt: "You are a computer science expert with knowledge of algorithms and data structures. Provide clear code examples."

  - name: business
    description: "Route business queries"
    priority: 100
    rules:
      operator: "OR"
      conditions:
        - type: "domain"
          name: "business"
    modelRefs:
      - model: "openai/gpt-oss-120b"
        use_reasoning: false
    plugins:
      - type: "system_prompt"
        configuration:
          system_prompt: "You are a senior business consultant and strategic advisor."

  - name: health
    description: "Route health queries"
    priority: 100
    rules:
      operator: "OR"
      conditions:
        - type: "domain"
          name: "health"
    modelRefs:
      - model: "openai/gpt-oss-120b"
        use_reasoning: false
    plugins:
      - type: "system_prompt"
        configuration:
          system_prompt: "You are a health and medical information expert."
      - type: "semantic-cache"
        configuration:
          enabled: true
          similarity_threshold: 0.95

  - name: law
    description: "Route legal queries"
    priority: 100
    rules:
      operator: "OR"
      conditions:
        - type: "domain"
          name: "law"
    modelRefs:
      - model: "openai/gpt-oss-120b"
        use_reasoning: false
    plugins:
      - type: "system_prompt"
        configuration:
          system_prompt: "You are a knowledgeable legal expert."

  - name: general_route
    description: "Default fallback route for general queries"
    priority: 50
    rules:
      operator: "OR"
      conditions:
        - type: "domain"
          name: "other"
    modelRefs:
      - model: "openai/gpt-oss-120b"
        use_reasoning: false
    plugins:
      - type: "semantic-cache"
        configuration:
          enabled: true
          similarity_threshold: 0.85
```

## Các Miền Được Hỗ trợ

Học tập: toán học, vật lý, hóa học, sinh học, khoa học máy tính, kỹ thuật

Chuyên nghiệp: kinh doanh, pháp luật, kinh tế, sức khỏe, tâm lý học

Chung: triết học, lịch sử, khác

## Tính năng

- **Phát hiện PII**: Tự động phát hiện và xử lý thông tin nhạy cảm
- **Bộ nhớ đệm ngữ nghĩa**: Lưu các truy vấn tương tự để phản hồi nhanh hơn
- **Kiểm soát suy luận**: Bật/tắt suy luận cho mỗi miền
- **Ngưỡng tùy chỉnh**: Điều chỉnh độ nhạy bộ nhớ đệm theo danh mục

## Ví dụ Yêu cầu

```bash
# Math query (reasoning enabled)
curl -X POST http://localhost:8801/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "MoM",
    "messages": [{"role": "user", "content": "Solve: x^2 + 5x + 6 = 0"}]
  }'

# Business query (reasoning disabled)
curl -X POST http://localhost:8801/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "MoM",
    "messages": [{"role": "user", "content": "What is a SWOT analysis?"}]
  }'

# Health query (high cache threshold)
curl -X POST http://localhost:8801/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "MoM",
    "messages": [{"role": "user", "content": "What are symptoms of diabetes?"}]
  }'
```

## Các trường hợp sử dụng thực tế

### 1. Phân loại Đa nhiệm vụ với LoRA (Hiệu quả)

**Vấn đề**: Cần phân loại miền + phát hiện PII + phát hiện jailbreak trên mỗi yêu cầu
**Giải pháp**: Các bộ điều chỉnh LoRA chạy cả 3 tác vụ với một lượt mô hình cơ sở thay vì 3 mô hình riêng biệt
**Tác động**: nhanh hơn 3x so với chạy 3 mô hình đầy đủ, <1% chi phí tham số cho mỗi tác vụ

### 2. Phân tích Tài liệu Dài (Qwen3 Chuyên biệt)

**Vấn đề**: Tài liệu/tài liệu pháp lý nghiên cứu vượt quá giới hạn 8K mã thông báo của ModernBERT
**Giải pháp**: Qwen3-Embedding hỗ trợ tới 32K mã thông báo mà không cần cắt cụt
**Tác động**: Phân loại chính xác trên tài liệu đầy đủ, không mất thông tin do cắt cụt

### 3. Nền tảng Giáo dục Đa ngôn ngữ (Qwen3 Chuyên biệt)

**Vấn đề**: Sinh viên hỏi câu hỏi bằng 100+ ngôn ngữ, ModernBERT giới hạn tiếng Anh
**Giải pháp**: Qwen3-Embedding được huấn luyện trên 100+ ngôn ngữ xử lý định tuyến đa ngôn ngữ
**Tác động**: Mô hình duy nhất phục vụ người dùng toàn cầu, chất lượng nhất quán trên các ngôn ngữ

### 4. Triển khai Edge (Gemma Chuyên biệt)

**Vấn đề**: Các thiết bị di động/IoT không thể chạy các mô hình phân loại lớn
**Giải pháp**: EmbeddingGemma-300M với nhúng Matryoshka (128-768 dims)
**Tác động**: Mô hình nhỏ hơn 5 lần, chạy trên các thiết bị edge với <100MB bộ nhớ

### 5. Nền tảng Gia sư STEM (Kiểm soát Suy luận Hiệu quả)

**Vấn đề**: Toán/vật lý cần suy luận, nhưng lịch sử/văn học thì không
**Giải pháp**: Bộ phân loại miền định tuyến STEM → mô hình suy luận, nhân văn → mô hình nhanh
**Tác động**: Độ chính xác STEM tốt hơn 2x, tiết kiệm chi phí 60% trên truy vấn không STEM

## Tối ưu hóa Cụ thể cho Miền

### Miền STEM (Suy luận Được bật)

```yaml
decisions:
- name: math
  description: "Route mathematical queries"
  priority: 100
  rules:
    operator: "OR"
    conditions:
      - type: "domain"
        name: "math"
  modelRefs:
    - model: "openai/gpt-oss-120b"
      use_reasoning: true  # Step-by-step solutions
  plugins:
    - type: "system_prompt"
      configuration:
        system_prompt: "You are a mathematics expert. Provide step-by-step solutions."

- name: physics
  description: "Route physics queries"
  priority: 100
  rules:
    operator: "OR"
    conditions:
      - type: "domain"
        name: "physics"
  modelRefs:
    - model: "openai/gpt-oss-120b"
      use_reasoning: true  # Derivations and proofs
  plugins:
    - type: "system_prompt"
      configuration:
        system_prompt: "You are a physics expert. Explain concepts clearly with mathematical derivations."

- name: chemistry
  description: "Route chemistry queries"
  priority: 100
  rules:
    operator: "OR"
    conditions:
      - type: "domain"
        name: "chemistry"
  modelRefs:
    - model: "openai/gpt-oss-120b"
      use_reasoning: true  # Reaction mechanisms
  plugins:
    - type: "system_prompt"
      configuration:
        system_prompt: "You are a chemistry expert. Explain reaction mechanisms clearly."
```

### Miền Chuyên nghiệp (PII + Bộ nhớ đệm)

```yaml
decisions:
- name: health
  description: "Route health queries"
  priority: 100
  rules:
    operator: "OR"
    conditions:
      - type: "domain"
        name: "health"
  modelRefs:
    - model: "openai/gpt-oss-120b"
      use_reasoning: false
  plugins:
    - type: "system_prompt"
      configuration:
        system_prompt: "You are a health and medical information expert."
    - type: "semantic-cache"
      configuration:
        enabled: true
        similarity_threshold: 0.95  # Very strict

- name: law
  description: "Route legal queries"
  priority: 100
  rules:
    operator: "OR"
    conditions:
      - type: "domain"
        name: "law"
  modelRefs:
    - model: "openai/gpt-oss-120b"
      use_reasoning: false
  plugins:
    - type: "system_prompt"
      configuration:
        system_prompt: "You are a knowledgeable legal expert."
```

### Miền Chung (Nhanh + Bộ nhớ đệm)

```yaml
decisions:
- name: business
  description: "Route business queries"
  priority: 100
  rules:
    operator: "OR"
    conditions:
      - type: "domain"
        name: "business"
  modelRefs:
    - model: "openai/gpt-oss-120b"
      use_reasoning: false  # Fast responses
  plugins:
    - type: "system_prompt"
      configuration:
        system_prompt: "You are a senior business consultant and strategic advisor."

- name: general_route
  description: "Default fallback route for general queries"
  priority: 50
  rules:
    operator: "OR"
    conditions:
      - type: "domain"
        name: "other"
  modelRefs:
    - model: "openai/gpt-oss-120b"
      use_reasoning: false
  plugins:
    - type: "semantic-cache"
      configuration:
        enabled: true
        similarity_threshold: 0.75  # Relaxed
```

## Đặc điểm Hiệu suất

| Miền | Suy luận | Ngưỡng Bộ nhớ đệm | Độ trễ Trung bình | Trường hợp sử dụng |
|--------|-----------|-----------------|-------------|----------|
| Toán học | ✅ | 0.85 | 2-5 giây | Giải pháp từng bước |
| Vật lý | ✅ | 0.85 | 2-5 giây | Đạo hàm |
| Hóa học | ✅ | 0.85 | 2-5 giây | Cơ chế |
| Sức khỏe | ❌ | 0.95 | 500ms | Quan trọng về an toàn |
| Pháp luật | ❌ | 0.85 | 500ms | Tuân thủ |
| Kinh doanh | ❌ | 0.80 | 300ms | Thông tin chi tiết nhanh |
| Khác | ❌ | 0.75 | 200ms | Truy vấn chung |

## Chiến lược Tối ưu hóa Chi phí

1. **Ngân sách Suy luận**: Chỉ bật cho STEM (30% truy vấn) → giảm chi phí 60%
2. **Chiến lược Bộ nhớ đệm**: Ngưỡng cao cho các miền nhạy cảm → tỷ lệ hit 70%
3. **Lựa chọn Mô hình**: Điểm thấp hơn cho các miền giá trị thấp → mô hình rẻ hơn
4. **Phát hiện PII**: Chỉ cho sức khỏe/pháp lý → chi phí xử lý giảm

## Tham chiếu

Xem [bert_classification.yaml](https://github.com/vllm-project/semantic-router/blob/main/config/intelligent-routing/in-tree/bert_classification.yaml) để xem cấu hình hoàn chỉnh.
