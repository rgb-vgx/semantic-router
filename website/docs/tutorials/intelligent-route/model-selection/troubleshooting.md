# Khắc Phục Sự Cố & Câu Hỏi Thường Gặp

Các vấn đề phổ biến và các câu hỏi thường gặp về lựa chọn mô hình.

## Câu Hỏi Thường Gặp

### Tôi nên bắt đầu với thuật toán nào?

**Bắt đầu với lựa chọn Tĩnh** nếu bạn chưa làm quen với lựa chọn mô hình. Nó mang tính xác định và dễ gỡ lỗi. Khi bạn hiểu các mô hình lưu lượng của mình, hãy chuyển sang các thuật toán thích ứng.

### Tôi có cần cấu hình tất cả các thuật toán không?

Không. Chỉ cấu hình thuật toán bạn đang sử dụng. Mỗi thuật toán có các giá trị mặc định hợp lý, vì vậy bạn chỉ cần chỉ định các trường bạn muốn tùy chỉnh.

### Tôi có thể chuyển đổi thuật toán mà không ngừng dịch vụ không?

Có. Các thay đổi thuật toán có hiệu lực khi tải lại cấu hình. Các yêu cầu đang thực hiện hoàn tất bằng thuật toán trước đó.

## Vấn Đề Phổ Biến

### Lựa Chọn Elo

**Vấn đề: Xếp hạng không thay đổi**

Các nguyên nhân có thể:

1. Phản hồi chưa được gửi - xác minh các yêu cầu POST đến `/api/v1/feedback` trả về 200
2. Hệ số K quá thấp - tăng từ 32 lên 64 để thích ứng nhanh hơn
3. Lưu lượng không đủ - Elo cần lượng phản hồi nhất quán

```bash
# Xác minh điểm cuối phản hồi đang hoạt động
curl -X POST http://localhost:8080/api/v1/feedback \
  -H "Content-Type: application/json" \
  -d '{"request_id": "test", "model": "gpt-4", "rating": 1}'
```

**Vấn đề: Một mô hình luôn được chọn**

Điều này là dự kiến nếu một mô hình có xếp hạng Elo cao hơn đáng kể. Các tùy chọn:

- Đặt lại xếp hạng bằng cách xóa tệp `storage_path`
- Tăng `k_factor` để cho phép thay đổi xếp hạng nhanh hơn
- Sử dụng `decay_factor` để giảm trọng số của các so sánh cũ

### Lựa Chọn RouterDC

**Vấn đề: Chọn sai mô hình cho truy vấn**

1. Kiểm tra mô tả mô hình có đủ cụ thể không:

```yaml
# Xấu - quá chung chung
description: "Một mô hình AI tốt"

# Tốt - khả năng cụ thể
description: "Suy luận toán học, chứng minh định lý, giải quyết vấn đề từng bước"
```

2. Xác minh các nhúng đang được tính:

```bash
# Kiểm tra số liệu cho độ trễ nhúng
curl http://localhost:8080/metrics | grep embedding
```

**Vấn đề: Lỗi khởi động với "mô tả bị thiếu"**

Nếu `require_descriptions: true`, tất cả các mô hình phải có mô tả:

```yaml
models:
  - name: gpt-4
    description: "Cần thiết khi require_descriptions là true"
```

### Lựa Chọn AutoMix

**Vấn đề: Luôn chọn các mô hình đắt tiền**

`Cost_quality_tradeoff` của bạn quá thấp (ưu tiên chất lượng). Tăng nó:

```yaml
automix:
  cost_quality_tradeoff: 0.5  # Cân bằng chi phí và chất lượng
```

**Vấn đề: Luôn chọn các mô hình rẻ**

`Cost_quality_tradeoff` của bạn quá cao. Giảm nó:

```yaml
automix:
  cost_quality_tradeoff: 0.2  # Ưu tiên chất lượng
```

**Vấn đề: Thiếu dữ liệu giá**

AutoMix yêu cầu thông tin giá:

```yaml
models:
  - name: gpt-4
    pricing:
      input_cost_per_1k: 0.03
      output_cost_per_1k: 0.06
```

### Lựa Chọn Hybrid

**Vấn đề: Lỗi xác thực trọng số**

Trọng số phải tính đến 1.0 (±0.01 hдопусћ):

```yaml
hybrid:
  elo_weight: 0.3
  router_dc_weight: 0.3
  automix_weight: 0.2
  cost_weight: 0.2
  # Tổng cộng: 1.0 ✓
```

**Vấn đề: Thành phần không đóng góp**

Đảm bảo thành phần có dữ liệu cần thiết:

- Elo: cần lịch sử phản hồi
- RouterDC: cần mô tả mô hình
- AutoMix: cần dữ liệu giá

## Mẹo Gỡ Lỗi

### Bật ghi nhật ký chi tiết

```yaml
logging:
  level: debug
```

### Kiểm tra số liệu lựa chọn

```bash
curl http://localhost:8080/metrics | grep selection
```

Các số liệu chính:

- `model_selection_duration_seconds` - độ trễ lựa chọn
- `model_selection_total` - đếm lựa chọn theo thuật toán
- `model_elo_rating` - xếp hạng Elo hiện tại (nếu sử dụng Elo)

### Theo dõi các yêu cầu riêng lẻ

Thêm tiêu đề ID yêu cầu và kiểm tra nhật ký:

```bash
curl -H "X-Request-ID: debug-123" http://localhost:8080/v1/chat/completions ...
```

Sau đó tìm kiếm nhật ký:

```bash
vllm-sr logs router | grep debug-123
```

## Nhận Trợ Giúp

Nếu bạn vẫn còn bị mắc kẹt:

1. Kiểm tra [GitHub Issues](https://github.com/vllm-project/semantic-router/issues) để có các vấn đề tương tự
2. Bật ghi nhật ký gỡ lỗi và capture kết quả liên quan
3. Mở một vấn đề mới với:
   - Cấu hình (bịdài bí mật)
   - Các bước để tái tạo
   - Hành vi dự kiến vs thực tế
   - Kết quả nhật ký liên quan
