# MoM Model Family là gì?

**MoM (Mixture of Models) Model Family** là một tập hợp được curation của các mô hình chuyên biệt, nhẹ được thiết kế cho định tuyến thông minh, bảo mật nội dung và hiểu biết ngữ nghĩa. Các mô hình này hỗ trợ các khả năng cốt lõi của Semantic Router, cho phép hoạt động AI nhanh, chính xác và bảo vệ quyền riêng tư.

## Tổng quan

Họ MoM bao gồm các mô hình được xây dựng với mục đích xử lý các tác vụ cụ thể trong đường ống định tuyến:

- **Classification Models**: Phát hiện miền, xác định PII, phát hiện jailbreak
- **Embedding Models**: Tương tự ngữ nghĩa, bộ nhớ đệm, truy xuất
- **Safety Models**: Phát hiện ảo tưởng, kiểm duyệt nội dung
- **Feedback Models**: Hiểu ý định người dùng, phân tích cuộc trò chuyện

Tất cả các mô hình MoM đều có:

- **Lightweight**: 33M-600M tham số để suy diễn nhanh
- **Specialized**: Tinh chỉnh cho các tác vụ định tuyến cụ thể
- **Efficient**: Nhiều mô hình sử dụng bộ điều hợp LoRA để mức lưu trữ tối thiểu
- **Open Source**: Có sẵn trên HuggingFace để minh bạch và tùy chỉnh

## Các danh mục mô hình

### 1. Classification Models

#### Domain/Intent Classifier

- **Model ID**: `models/mom-domain-classifier`
- **HuggingFace**: `LLM-Semantic-Router/lora_intent_classifier_bert-base-uncased_model`
- **Purpose**: Phân loại truy vấn người dùng thành 14 danh mục MMLU (toán học, khoa học, lịch sử, v.v.)
- **Architecture**: BERT-base (110M) + bộ điều hợp LoRA
- **Use Case**: Định tuyến truy vấn tới các mô hình hoặc chuyên gia cụ thể theo miền

#### PII Detector

- **Model ID**: `models/mom-pii-classifier`
- **HuggingFace**: `LLM-Semantic-Router/lora_pii_detector_bert-base-uncased_model`
- **Purpose**: Phát hiện 35 loại thông tin nhận dạng cá nhân
- **Architecture**: BERT-base (110M) + bộ điều hợp LoRA
- **Use Case**: Bảo vệ quyền riêng tư, tuân thủ, che giấu các nhân

#### Jailbreak Detector

- **Model ID**: `models/mom-jailbreak-classifier`
- **HuggingFace**: `LLM-Semantic-Router/lora_jailbreak_classifier_bert-base-uncased_model`
- **Purpose**: Phát hiện tiêm nhắc lệnh và các nỗ lực jailbreak
- **Architecture**: BERT-base (110M) + bộ điều hợp LoRA
- **Use Case**: Bảo mật nội dung, bảo mật nhắc lệnh

#### Feedback Detector

- **Model ID**: `models/mom-feedback-detector`
- **HuggingFace**: `llm-semantic-router/feedback-detector`
- **Purpose**: Phân loại phản hồi người dùng thành 4 loại (hài lòng, cần làm rõ, câu trả lời sai, muốn cái khác)
- **Architecture**: ModernBERT-base (149M)
- **Use Case**: Định tuyến thích ứng, cải thiện cuộc trò chuyện

### 2. Embedding Models

#### Embedding Pro (Chất lượng cao)

- **Model ID**: `models/mom-embedding-pro`
- **HuggingFace**: `Qwen/Qwen3-Embedding-0.6B`
- **Purpose**: Nhúng chất lượng cao với hỗ trợ bağlam 32K
- **Architecture**: Qwen3 (600M tham số)
- **Embedding Dimension**: 1024
- **Use Case**: Tìm kiếm ngữ nghĩa trong bảng ngữ cảnh dài, bộ nhớ đệm độ chính xác cao

#### Embedding Flash (Cân bằng)

- **Model ID**: `models/mom-embedding-flash`
- **HuggingFace**: `google/embeddinggemma-300m`
- **Purpose**: Nhúng nhanh với hỗ trợ Matryoshka
- **Architecture**: Gemma (300M tham số)
- **Embedding Dimension**: 768 (hỗ trợ 512/256/128 qua Matryoshka)
- **Use Case**: Cân bằng tốc độ/chất lượng, hỗ trợ đa ngôn ngữ

#### Embedding Light (Nhanh)

- **Model ID**: `models/mom-embedding-light`
- **HuggingFace**: `sentence-transformers/all-MiniLM-L12-v2`
- **Purpose**: Tương tự ngữ nghĩa nhẹ
- **Architecture**: MiniLM (33M tham số)
- **Embedding Dimension**: 384
- **Use Case**: Bộ nhớ đệm ngữ nghĩa nhanh, truy xuất độ trễ thấp

### 3. Hallucination Detection Models

#### Halugate Sentinel

- **Model ID**: `models/mom-halugate-sentinel`
- **HuggingFace**: `LLM-Semantic-Router/halugate-sentinel`
- **Purpose**: Sàng lọc ảo tưởng giai đoạn đầu tiên
- **Architecture**: BERT-base (110M)
- **Use Case**: Phát hiện ảo tưởng nhanh, lọc trước

#### Halugate Detector

- **Model ID**: `models/mom-halugate-detector`
- **HuggingFace**: `KRLabsOrg/lettucedect-base-modernbert-en-v1`
- **Purpose**: Xác minh ảo tưởng chính xác
- **Architecture**: ModernBERT-base (149M)
- **Context Length**: 8192 token
- **Use Case**: Xác minh độ chính xác thực tế, kiểm tra nền tảng

#### Halugate Explainer

- **Model ID**: `models/mom-halugate-explainer`
- **HuggingFace**: `tasksource/ModernBERT-base-nli`
- **Purpose**: Giải thích lý do ảo tưởng qua NLI
- **Architecture**: ModernBERT-base (149M)
- **Classes**: 3 (entailment/neutral/contradiction)
- **Use Case**: AI có thể giải thích được, phân tích ảo tưởng

## Hướng dẫn lựa chọn mô hình

### Theo Trường hợp sử dụng

| Trường hợp sử dụng | Mô hình được đề xuất | Tại sao |
|----------|------------------|-----|
| Định tuyến miền | mom-domain-classifier | 14 danh mục MMLU, hiệu quả LoRA |
| Bảo vệ quyền riêng tư | mom-pii-classifier | 35 loại PII, phát hiện cấp token |
| Bảo mật nội dung | mom-jailbreak-classifier | Phát hiện tiêm nhắc lệnh |
| Bộ nhớ đệm ngữ nghĩa | mom-embedding-light | Nhanh, 384-chiều, độ trễ thấp |
| Tìm kiếm bảng ngữ cảnh dài | mom-embedding-pro | Bảng ngữ cảnh 32K, 1024-chiều |
| Kiểm tra ảo tưởng | mom-halugate-detector | ModernBERT, bảng ngữ cảnh 8K |
| Phản hồi người dùng | mom-feedback-detector | 4 loại phản hồi, ModernBERT |

### Theo yêu cầu hiệu suất

| Yêu cầu | Tầng mô hình | Ví dụ |
|-------------|-----------|----------|
| Siêu nhanh (<10ms) | Light | mom-embedding-light, mom-jailbreak-classifier |
| Cân bằng (10-50ms) | Flash | mom-embedding-flash, mom-domain-classifier |
| Chất lượng cao (50-200ms) | Pro | mom-embedding-pro, mom-halugate-detector |

## Cấu hình

### Sử dụng mô hình MoM trong Bộ định tuyến

Các mô hình MoM được cấu hình sẵn trong `router-defaults.yaml`:

```yaml
# Phân loại miền
classifier:
  category_model:
    model_id: "models/mom-domain-classifier"
    threshold: 0.6
    use_cpu: true

# Phát hiện PII
classifier:
  pii_model:
    model_id: "models/mom-pii-classifier"
    threshold: 0.9
    use_cpu: true

# Bảo vệ jailbreak
prompt_guard:
  model_id: "models/mom-jailbreak-classifier"
  threshold: 0.7
  use_cpu: true
```

### Đăng ký mô hình tùy chỉnh

Ghi đè sổ đăng ký mặc định trong `config.yaml` của bạn:

```yaml
mom_registry:
  "models/mom-domain-classifier": "your-org/custom-domain-classifier"
  "models/mom-pii-classifier": "your-org/custom-pii-detector"
  "models/mom-embedding-pro": "your-org/custom-embeddings"
```

## Kiến trúc mô hình

### Mô hình dựa trên LoRA

Nhiều mô hình MoM sử dụng LoRA (Low-Rank Adaptation) để nâng cao hiệu quả:

- **Base Model**: BERT-base-uncased (110M tham số)
- **LoRA Adapters**: <1M tham số cho mỗi tác vụ
- **Memory Footprint**: ~440MB cơ sở + ~4MB cho mỗi bộ điều hợp
- **Inference Speed**: Giống với mô hình cơ sở (~10-20ms trên CPU)

### Mô hình ModernBERT

Các mô hình mới hơn sử dụng ModernBERT để hiệu suất tốt hơn:

- **Architecture**: ModernBERT-base (149M tham số)
- **Context Length**: 8192 token (so với 512 cho BERT)
- **Performance**: Độ chính xác tốt hơn trên các tác vụ bảng ngữ cảnh dài
- **Use Cases**: Phát hiện ảo tưởng, phân loại phản hồi

## Bước tiếp theo

- **[Signal-Driven Decisions](./signal-driven-decisions.md)** - Tìm hiểu cách các mô hình MoM hỗ trợ quyết định định tuyến
- **[Domain Routing](../tutorials/intelligent-route/domain-routing.md)** - Sử dụng mom-domain-classifier để định tuyến
- **[PII Detection](../tutorials/content-safety/pii-detection.md)** - Cấu hình mom-pii-classifier
- **[Semantic Cache](../tutorials/semantic-cache/in-memory-cache.md)** - Sử dụng các mô hình nhúng MoM
