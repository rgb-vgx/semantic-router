# Phát hiện Ảo tưởng (Hallucination Detection)

Semantic Router cung cấp các tính năng phát hiện ảo tưởng tiên tiến để xác minh rằng các phản hồi của LLM được dựa trên bối cảnh được cung cấp. Hệ thống sử dụng các bộ phân loại token ModernBERT được tinh chỉnh để xác định các yêu cầu không được hỗ trợ bởi kết quả truy xuất hoặc đầu ra của công cụ.

## Tổng quan

Hệ thống phát hiện ảo tưởng:

- **Xác minh** các phản hồi của LLM so với bối cảnh được cung cấp (kết quả RAG, đầu ra công cụ)
- **Xác định** các yêu cầu không được hỗ trợ ở cấp độ token
- **Cung cấp** các giải thích chi tiết bằng NLI (Natural Language Inference)
- **Cảnh báo hoặc chặn** khi phát hiện ảo tưởng
- **Tích hợp** liền mạch với các quy trình RAG và gọi công cụ

## Cách thức hoạt động

Phát hiện ảo tưởng hoạt động theo một pipeline ba giai đoạn:

1. **Phân loại Kiểm chứng Thực tế**: Xác định xem một truy vấn có yêu cầu xác minh thực tế không (câu hỏi thực tế so với sáng tạo/dựa trên ý kiến)
2. **Phát hiện Cấp độ Token**: Phân tích các phản hồi của LLM để xác định các yêu cầu không được hỗ trợ
3. **Giải thích NLI** (tùy chọn): Cung cấp suy luận chi tiết cho các phần ảo tưởng

## Cấu hình

### Cấu hình Mô hình Global

Đầu tiên, cấu hình các mô hình phát hiện ảo tưởng trong `router-defaults.yaml`:

```yaml
# router-defaults.yaml
# Cấu hình giảm nhẹ ảo tưởng
# Tắt theo mặc định - bật trong các quyết định thông qua plugin ảo tưởng
hallucination_mitigation:
  enabled: false

  # Bộ phân loại kiểm chứng thực tế: xác định xem một prompt có cần xác minh thực tế không
  fact_check_model:
    model_id: "models/mom-halugate-sentinel"
    threshold: 0.6
    use_cpu: true

  # Bộ phát hiện ảo tưởng: xác minh xem phản hồi LLM có được dựa trên bối cảnh không
  hallucination_model:
    model_id: "models/mom-halugate-detector"
    threshold: 0.8
    use_cpu: true

  # Mô hình NLI: cung cấp giải thích cho các phần ảo tưởng
  nli_model:
    model_id: "models/mom-halugate-explainer"
    threshold: 0.9
    use_cpu: true
```

### Bật Phát hiện Ảo tưởng trong Các Quyết định

Bật phát hiện ảo tưởng cho mỗi quyết định bằng cách sử dụng plugin `hallucination`:

```yaml
# config.yaml
decisions:
  - name: "general_decision"
    description: "Các câu hỏi chung có kiểm chứng thực tế"
    priority: 100
    rules:
      operator: "OR"
      conditions:
        - type: "domain"
          name: "general"
        - type: "fact_check"
          name: "needs_fact_check"
    modelRefs:
      - model: "openai/gpt-oss-120b"
        use_reasoning: false
    plugins:
      - type: "hallucination"
        configuration:
          enabled: true
          use_nli: true  # Bật NLI để có giải thích chi tiết
          # Hành động khi phát hiện ảo tưởng: "header", "body", "block", hoặc "none"
          hallucination_action: "header"
          # Hành động khi cần kiểm chứng thực tế nhưng không có bối cảnh công cụ: "header", "body", hoặc "none"
          unverified_factual_action: "header"
          # Bao gồm thông tin chi tiết (độ tin cậy, phần) trong cảnh báo phần nội dung
          include_hallucination_details: true
```

### Tùy chọn Cấu hình Plugin

| Tùy chọn                       | Giá trị                         | Mô tả                                                    |
|---------------------------------|-------------------------------------|----------------------------------------------------------|
| `enabled`                       | `true`, `false`                     | Bật/tắt phát hiện ảo tưởng cho quyết định này           |
| `use_nli`                       | `true`, `false`                     | Sử dụng mô hình NLI để có giải thích chi tiết           |
| `hallucination_action`          | `header`, `body`, `block`, `none`   | Hành động khi phát hiện ảo tưởng                        |
| `unverified_factual_action`     | `header`, `body`, `none`            | Hành động khi cần kiểm chứng nhưng không có bối cảnh    |
| `include_hallucination_details` | `true`, `false`                     | Bao gồm độ tin cậy và phần trong phần nội dung phản hồi  |

### Chế độ Hành động

| Hành động | Hành vi                                               | Trường hợp sử dụng                     |
|----------|-------------------------------------------------------|---------------------------------------|
| `header` | Thêm các tiêu đề cảnh báo, cho phép phản hồi         | Phát triển, giám sát                  |
| `body`   | Thêm cảnh báo trong phần nội dung phản hồi, cho phép | Cảnh báo hướng đến người dùng          |
| `block`  | Trả về lỗi, chặn phản hồi                            | Sản xuất, các ứng dụng có mức cao     |
| `none`   | Không hành động, chỉ ghi nhật ký                     | Giám sát âm thầm                      |

## Cách thức Phát hiện Ảo tưởng Hoạt động

Khi một yêu cầu được xử lý:

1. **Phân loại Kiểm chứng Thực tế**: Mô hình tuần vệ xác định xem truy vấn có cần xác minh thực tế không
2. **Trích xuất Bối cảnh**: Kết quả công cụ hoặc bối cảnh RAG được thu thập từ phản hồi của LLM
3. **Phát hiện Ảo tưởng**: Nếu có bối cảnh, bộ phát hiện phân tích phản hồi
4. **Hành động**: Dựa trên cấu hình, hệ thống thêm tiêu đề, sửa đổi phần nội dung, hoặc chặn phản hồi

Các tiêu đề phản hồi khi phát hiện ảo tưởng:

```http
X-Hallucination-Detected: true
X-Hallucination-Confidence: 0.85
X-Unsupported-Spans: "Paris was founded in 1492"
```

## Các trường hợp sử dụng

### RAG (Retrieval Augmented Generation)

Xác minh rằng các phản hồi của LLM được dựa trên các tài liệu được truy xuất:

```yaml
plugins:
  - type: "hallucination"
    configuration:
      enabled: true
      use_nli: false
      hallucination_action: "header"
      unverified_factual_action: "header"
```

**Ví dụ**: Một bot hỗ trợ khách hàng truy xuất tài liệu và tạo câu trả lời. Phát hiện ảo tưởng đảm bảo các phản hồi không bao gồm thông tin không có trong tài liệu.

### Quy trình Gọi Công cụ

Xác thực rằng các phản hồi của LLM chính xác phản ánh đầu ra của công cụ:

```yaml
plugins:
  - type: "hallucination"
    configuration:
      enabled: true
      use_nli: true
      hallucination_action: "block"
      unverified_factual_action: "header"
      include_hallucination_details: true
```

**Ví dụ**: Một tác nhân AI gọi công cụ truy vấn cơ sở dữ liệu. Phát hiện ảo tưởng ngăn chặn tác nhân từ việc tạo dữ liệu không được trả về bởi truy vấn.
