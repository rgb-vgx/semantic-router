# Định tuyến Dựa trên Nhúng

Hướng dẫn này hướng dẫn bạn cách định tuyến các yêu cầu bằng cách sử dụng tương đồng ngữ nghĩa với các mô hình nhúng. Định tuyến nhúng khớp các truy vấn người dùng với các danh mục được xác định trước dựa trên ý nghĩa thay vì các từ khóa chính xác, làm cho nó lý tưởng để xử lý các cách diễn đạt đa dạng và các danh mục đang phát triển nhanh chóng.

## Lợi ích chính

- **Có thể mở rộng**: Xử lý các danh mục không giới hạn mà không cần huấn luyện lại mô hình
- **Nhanh**: Suy luận 10-50ms với các mô hình nhúng hiệu quả (Qwen3, Gemma)
- **Linh hoạt**: Thêm/xóa danh mục bằng cách cập nhật danh sách từ khóa, không cần huấn luyện lại mô hình
- **Ngữ nghĩa**: Nắm bắt ý nghĩa vượt quá khớp từ khóa chính xác

## Vấn đề nó giải quyết là gì?

Khớp từ khóa thất bại khi người dùng diễn đạt câu hỏi khác nhau. Định tuyến nhúng giải quyết:

- **Xử lý Paraphrase**: "Cách cài đặt?" khớp "hướng dẫn cài đặt" mà không có từ chính xác
- **Phát hiện Ý định**: Định tuyến dựa trên ý nghĩa ngữ nghĩa, không phải các mẫu bề mặt
- **Khớp Mờ**: Xử lý lỗi chính tả, viết tắt, ngôn ngữ không chính thức
- **Danh mục Động**: Thêm danh mục mới mà không cần huấn luyện lại các mô hình phân loại
- **Hỗ trợ Đa ngôn ngữ**: Nhúng nắm bắt ngữ nghĩa liên ngôn ngữ

## Khi nào sử dụng

- **Hỗ trợ khách hàng** với các cách diễn đạt truy vấn đa dạng
- **Truy vấn sản phẩm** nơi người dùng hỏi điều tương tự theo nhiều cách khác nhau
- **Hỗ trợ kỹ thuật** cần hiểu ngữ nghĩa của mô tả lỗi
- **Danh mục đang phát triển nhanh** nơi bạn cần thêm/cập nhật danh mục thường xuyên
- **Dung sai độ trễ Trung bình** (10-50ms chấp nhận được để có độ chính xác ngữ nghĩa tốt hơn)

## Cấu hình

Thêm quy tắc nhúng vào `config.yaml` của bạn:

```yaml
# Define embedding signals
signals:
  embeddings:
    - name: "technical_support"
      threshold: 0.75
      candidates:
        - "how to configure the system"
        - "installation guide"
        - "troubleshooting steps"
        - "error message explanation"
      aggregation_method: "max"

    - name: "product_inquiry"
      threshold: 0.70
      candidates:
        - "product features and specifications"
        - "pricing information"
        - "availability and stock"
      aggregation_method: "avg"

    - name: "account_management"
      threshold: 0.72
      candidates:
        - "password reset"
        - "account settings"
        - "subscription management"
      aggregation_method: "max"

# Define decisions using embedding signals
decisions:
  - name: technical_support
    description: "Route technical support queries"
    priority: 100
    rules:
      operator: "OR"
      conditions:
        - type: "embedding"
          name: "technical_support"
    modelRefs:
      - model: "openai/gpt-oss-120b"
        use_reasoning: true
    plugins:
      - type: "system_prompt"
        configuration:
          system_prompt: "You are a technical support specialist with deep knowledge of system configuration and troubleshooting."

  - name: product_inquiry
    description: "Route product inquiry queries"
    priority: 100
    rules:
      operator: "OR"
      conditions:
        - type: "embedding"
          name: "product_inquiry"
    modelRefs:
      - model: "openai/gpt-oss-120b"
        use_reasoning: false
    plugins:
      - type: "system_prompt"
        configuration:
          system_prompt: "You are a product specialist with comprehensive knowledge of features, pricing, and availability."
      - type: "semantic-cache"
        configuration:
          enabled: true
          similarity_threshold: 0.85

  - name: account_management
    description: "Route account management queries"
    priority: 100
    rules:
      operator: "OR"
      conditions:
        - type: "embedding"
          name: "account_management"
    modelRefs:
      - model: "openai/gpt-oss-120b"
        use_reasoning: false
    plugins:
      - type: "system_prompt"
        configuration:
          system_prompt: "You are an account management specialist. Handle user account queries with care and security."
```

## Mô hình Nhúng

- **qwen3**: Chất lượng cao, 1024-dim, ngữ cảnh 32K
- **gemma**: Cân bằng, 768-dim, ngữ cảnh 8K, hỗ trợ Matryoshka (128/256/512/768)
- **auto**: Tự động chọn dựa trên ưu tiên chất lượng/độ trễ

## Phương pháp Tổng hợp

- **max**: Sử dụng điểm tương đồng cao nhất
- **avg**: Sử dụng mức trung bình tương đồng trên các từ khóa
- **any**: Khớp nếu bất kỳ từ khóa nào vượt quá ngưỡng

## Ví dụ Yêu cầu

```bash
# Technical support query
curl -X POST http://localhost:8801/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "MoM",
    "messages": [{"role": "user", "content": "How do I troubleshoot connection errors?"}]
  }'

# Product inquiry
curl -X POST http://localhost:8801/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "MoM",
    "messages": [{"role": "user", "content": "What are the pricing options?"}]
  }'
```

## Các trường hợp sử dụng thực tế

### 1. Hỗ trợ khách hàng (Danh mục Có thể mở rộng)

**Vấn đề**: Cần thêm danh mục hỗ trợ mới hàng tuần mà không cần huấn luyện lại mô hình
**Giải pháp**: Thêm danh mục mới bằng cách cập nhật danh sách từ khóa, nhúng xử lý khớp ngữ nghĩa
**Tác động**: Triển khai danh mục mới trong vài phút so với tuần cho việc huấn luyện lại mô hình

### 2. Hỗ trợ Thương mại điện tử (Khớp Ngữ nghĩa Nhanh)

**Vấn đề**: "Đơn hàng của tôi ở đâu?" so với "theo dõi gói" so với "trạng thái vận chuyển" có cùng ý nghĩa
**Giải pháp**: Nhúng Gemma (10-20ms) định tuyến tất cả các biến đổi đến danh mục theo dõi đơn hàng
**Tác động**: Độ chính xác 95% với độ trễ 10-20ms, xử lý 5K+ truy vấn/giây

### 3. Truy vấn Sản phẩm SaaS (Định tuyến Linh hoạt)

**Vấn đề**: Người dùng hỏi về giá theo 100+ cách khác nhau
**Giải pháp**: Tương đồng ngữ nghĩa khớp tất cả các biến đổi với từ khóa "thông tin giá"
**Tác động**: Danh mục duy nhất xử lý tất cả các truy vấn giá mà không cần quy tắc rõ ràng

### 4. Lặp lại Khởi nghiệp (Cập nhật Danh mục Nhanh chóng)

**Vấn đề**: Sản phẩm phát triển nhanh chóng, cần điều chỉnh danh mục hàng ngày
**Giải pháp**: Cập nhật từ khóa nhúng trong cấu hình, không cần huấn luyện lại mô hình
**Tác động**: Cập nhật danh mục trong vài giây so với ngày cho việc tinh chỉnh

### 5. Nền tảng Đa ngôn ngữ (Hiểu biết Ngữ nghĩa)

**Vấn đề**: Câu hỏi tương tự bằng tiếng Anh, Tây Ban Nha, Trung Quốc cần định tuyến giống nhau
**Giải pháp**: Nhúng nắm bắt ngữ nghĩa liên ngôn ngữ tự động
**Tác động**: Định nghĩa danh mục duy nhất hoạt động trên các ngôn ngữ

## Chiến lược Lựa chọn Mô hình

### Chế độ Tự động (Được khuyến nghị)

```yaml
model: "auto"
quality_priority: 0.7  # Favor accuracy
latency_priority: 0.3  # Accept some latency
```

- Tự động chọn Qwen3 (chất lượng cao) hoặc Gemma (nhanh) dựa trên ưu tiên
- Cân bằng độ chính xác so với tốc độ cho mỗi yêu cầu

### Qwen3 (Chất lượng Cao)

```yaml
model: "qwen3"
dimension: 1024
```

- Tốt nhất cho: Truy vấn phức tạp, sự phân biệt tinh tế, tương tác giá trị cao
- Độ trễ: ~30-50ms cho mỗi truy vấn
- Trường hợp sử dụng: Quản lý tài khoản, truy vấn tài chính

### Gemma (Nhanh)

```yaml
model: "gemma"
dimension: 768  # or 512, 256, 128 for Matryoshka
```

- Tốt nhất cho: Thông lượng cao, phân loại đơn giản, nhạy cảm chi phí
- Độ trễ: ~10-20ms cho mỗi truy vấn
- Trường hợp sử dụng: Truy vấn sản phẩm, hỗ trợ chung

## Đặc điểm Hiệu suất

| Mô hình | Chiều | Độ trễ | Độ chính xác | Bộ nhớ |
|-------|-----------|---------|----------|--------|
| Qwen3 | 1024 | 30-50ms | Cao nhất | ~2.4GB |
| Gemma | 768 | 10-20ms | Cao | ~1.2GB |
| Gemma | 512 | 8-15ms | Trung bình | ~1.2GB |
| Gemma | 256 | 5-10ms | Thấp hơn | ~1.2GB |

## Tham chiếu

Xem [embedding.yaml](https://github.com/vllm-project/semantic-router/blob/main/config/intelligent-routing/in-tree/embedding.yaml) để xem cấu hình hoàn chỉnh.
