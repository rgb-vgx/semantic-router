# Kết Quả Điểm Chuẩn Hiệu Suất ModernBERT-base-32k

Hướng dẫn này cung cấp kết quả điểm chuẩn và hướng dẫn điều chỉnh hiệu suất cho tích hợp ModernBERT-base-32k. Sử dụng những kết quả này để cấp phát phần cứng và điều chỉnh kỳ vọng về tải công việc cho triển khai của bạn.

## Tổng Quan

ModernBERT-base-32k mở rộng cửa sổ bối cảnh từ 512 token (BERT-base) thành 32.768 token, cho phép xử lý các tài liệu dài và cuộc trò chuyện. Hướng dẫn này trình bày kết quả kiểm tra thực tế từ các bài kiểm tra toàn diện trên các độ dài bối cảnh và mức độ đồng thời khác nhau.

**Môi Trường Kiểm Tra**:
- **GPU**: NVIDIA L4 (23GB VRAM)
- **Flash Attention 2**: Được bật
- **Mô Hình**: `llm-semantic-router/modernbert-base-32k`
- **Công Cụ Kiểm Tra**: `candle-binding/examples/benchmark_concurrent.rs`

---

## Kết Quả Điểm Chuẩn

### Độ Trễ Yêu Cầu Đơn Lẻ (C=1)

| Độ Dài Bối Cảnh | Độ Trễ Trung Bình | Độ Trễ p50 | Độ Trễ p95 | Độ Trễ p99 | Trạng Thái |
|---|---|---|---|---|---|
| 1.024 token | 90.98ms | 94.18ms | 94.24ms | 94.24ms | Thành công |
| 4.096 token | 899.87ms | 955.05ms | 955.93ms | 955.93ms | Thành công |
| 8.192 token | 3.299.92ms | 3.524.62ms | 3.526.34ms | 3.526.34ms | Thành công |

### Yêu Cầu Đồng Thời (C=10)

| Độ Dài Bối Cảnh | Độ Trễ Trung Bình | Độ Trễ p50 | Độ Trễ p95 | Tỷ Lệ Thành Công | Trạng Thái |
|---|---|---|---|---|---|
| 1.024 token | 1.001.22ms | 970.65ms | 1.379.32ms | 100% | Thành công |
| 4.096 token | 9.323.45ms | 9.389.28ms | 10.349.11ms | 93% | Thành công Một Phần |
| 8.192 token | N/A | N/A | N/A | 0% | Bị Lỗi |

### Đồng Thời Cao (C=50, C=100)

Tất cả các bài kiểm tra đồng thời cao (C=50+) bị lỗi do hạn chế phần cứng. Môi trường kiểm tra hiện tại (GPU NVIDIA L4 với 23GB VRAM) không cung cấp đủ bộ nhớ cho các khối công việc đồng thời cao với độ dài bối cảnh lớn hơn. Xem [Kế Hoạch Kiểm Tra Lô Lớn](./modernbert-32k-docs/modernbert-32k-big-batch-test-plan.md) để biết chi tiết về các bài kiểm tra cần thiết GPU với 40GB+ VRAM (ví dụ: NVIDIA A100).

---

## Hướng Dẫn Cấp Phát Phần Cứng

### Yêu Cầu Tối Thiểu

| Độ Dài Bối Cảnh | GPU VRAM | RAM Hệ Thống | GPU Được Khuyến Nghị |
|---|---|---|---|
| ≤ 1K token | ≥ 5GB | ≥ 16GB | NVIDIA T4, L4 |
| ≤ 4K token | ≥ 10GB | ≥ 32GB | NVIDIA L4, A10G |
| ≤ 8K token | ≥ 23GB | ≥ 32GB | NVIDIA L4, A10G |
| 16K+ token | ≥ 40GB | ≥ 64GB | NVIDIA A100 |

### Cấu Hình Được Khuyến Nghị

**Cho Sản Xuất (1K-8K token)**:
- **GPU**: NVIDIA L4 (23GB VRAM) hoặc tốt hơn
- **RAM Hệ Thống**: 32GB+
- **CUDA**: Phiên bản 12.0+
- **Flash Attention 2**: Được bật (tăng tốc độ 1,75x-11,9x)

**Cho Bối Cảnh Dài (16K-32K token)**:
- **GPU**: NVIDIA A100 (40GB+ VRAM) - **Bắt Buộc**
- **RAM Hệ Thống**: 64GB+
- Xem [Kế Hoạch Kiểm Tra Bối Cảnh Dài](./modernbert-32k-docs/modernbert-32k-long-context-test-plan.md) để biết chi tiết

---

## Kỳ Vọng Tải Công Việc

### Giới Hạn Đồng Thời theo Độ Dài Bối Cảnh

| Độ Dài Bối Cảnh | Đồng Thời Tối Đa | Thông Lượng Dự Kiến | Ghi Chú |
|---|---|---|---|
| 1.024 token | C=10 | ~10 req/s | Được kiểm tra và đáng tin cậy |
| 4.096 token | C=10 | ~1 req/s | Tỷ lệ thành công 88% |
| 8.192 token | C=1 | ~0,3 req/s | Chỉ C=1 hoạt động đáng tin cậy |
| 16.384+ token | C=1 (với phân đoạn) | Khác nhau | Yêu cầu A100 hoặc phân đoạn |

### Kỳ Vọng Độ Trễ

**Yêu Cầu Đơn Lẻ (C=1)**:
- 1K token: ~100ms (p50)
- 4K token: ~950ms (p50)
- 8K token: ~3.500ms (p50)

**Yêu Cầu Đồng Thời (C=10)**:
- 1K token: ~1.000ms (trung bình)
- 4K token: ~9.400ms (trung bình, tỷ lệ thành công 93%)
- 8K token: Không được hỗ trợ (OOM)

### Sử Dụng Bộ Nhớ

| Độ Dài Bối Cảnh | Bộ Nhớ GPU mỗi Yêu Cầu | Ghi Chú |
|---|---|---|
| 512 token | ~5MB | Rất hiệu quả |
| 1K token | ~11MB | Rất hiệu quả |
| 4K token | ~ước tính | Vừa phải |
| 8K token | ~ước tính | Cao (yêu cầu 22GB+ VRAM) |

---

## Hướng Dẫn Cấu Hình

### Kích Hoạt ModernBERT-base-32k

Để sử dụng ModernBERT-base-32k trong cấu hình bộ định tuyến ngữ nghĩa của bạn:

```yaml
classifier:
  category_model:
    model_id: "models/mom-domain-classifier"
    use_modernbert: true  # Kích hoạt ModernBERT-base-32k
    threshold: 0.6
    use_cpu: false
    category_mapping_path: "models/mom-domain-classifier/category_mapping.json"
```

### Flash Attention 2

Flash Attention 2 cung cấp cải tiến hiệu suất đáng kể (1,75x-11,9x tăng tốc độ). Hãy đảm bảo nó được bật khi xây dựng:

```bash
cargo build --release --features cuda,flash-attn
```

---

## Khuyến Nghị Điều Chỉnh Hiệu Suất

### 1. Lựa Chọn Độ Dài Bối Cảnh

Chọn độ dài bối cảnh dựa trên trường hợp sử dụng:
- **Các truy vấn ngắn (≤1K token)**: Hiệu suất tốt nhất, hỗ trợ C=10
- **Tài liệu trung bình (1K-4K token)**: Hiệu suất tốt, hỗ trợ C=10 với tỷ lệ thành công 88%
- **Tài liệu dài (4K-8K token)**: Hiệu suất chấp nhận được, chỉ C=1 được hỗ trợ
- **Tài liệu rất dài (8K+ token)**: Yêu cầu phân đoạn hoặc GPU A100

### 2. Điều Chỉnh Đồng Thời

**Bắt Đầu Bảo Thủ**:
1. Bắt đầu với C=1 cho tất cả các độ dài bối cảnh
2. Từ từ tăng đến C=10 cho token 1K-4K
3. Giám sát bộ nhớ GPU và tỷ lệ lỗi
4. Giảm đồng thời nếu xảy ra lỗi OOM

**Cài Đặt Sản Xuất**:
```yaml
concurrency_limits:
  1024: 10    # 1K token: C=10
  4096: 10    # 4K token: C=10 (giám sát OOM)
  8192: 1     # 8K token: C=1 chỉ
```

---

## Chạy Điểm Chuẩn

### Điều Kiện Tiên Quyết

1. **Cài Đặt Rust và CUDA**:

```bash
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
```

2. **Xây Dựng Với Flash Attention 2**:

```bash
cd candle-binding
cargo build --example benchmark_concurrent --release --features cuda,flash-attn
```

### Chạy Điểm Chuẩn Yêu Cầu Đồng Thời

```bash
cargo run --example benchmark_concurrent --release --features cuda,flash-attn
```

### Kiểm Tra Bối Cảnh Dài (16K-32K)

Để kiểm tra token 16K-32K (yêu cầu GPU A100):

1. **Mở phần bình luận trường hợp kiểm tra** trong `benchmark_concurrent.rs`:
```rust
let context_lengths = vec![
    1024,
    4096,
    8192,
    16384,  // Mở nhận xét này
    32768,  // Mở nhận xét này
];
```

2. **Chạy điểm chuẩn**:
```bash
cargo run --example benchmark_concurrent --release --features cuda,flash-attn
```

3. **Xem**: [Kế Hoạch Kiểm Tra Bối Cảnh Dài](./modernbert-32k-docs/modernbert-32k-long-context-test-plan.md) để biết chi tiết kế hoạch kiểm tra

---

## Khắc Phục Sự Cố

### Lỗi Hết Bộ Nhớ (OOM)

**Triệu Chứng**:
- Lỗi `CUDA_ERROR_OUT_OF_MEMORY`
- Yêu cầu bị lỗi ở đồng thời cao

**Giải Pháp**:
1. Giảm đồng thời (C=10 → C=1)
2. Sử dụng phân đoạn cho chuỗi > 8K token
3. Tăng thời gian chờ giữa các yêu cầu
4. Sử dụng GPU với VRAM nhiều hơn (A100 40GB+)

### Độ Trễ Cao

**Triệu Chứng**:
- Độ trễ cao hơn dự kiến
- Các bước nhảy độ trễ p95/p99

**Giải Pháp**:
1. Kích hoạt Flash Attention 2
2. Giảm đồng thời
3. Sử dụng phân đoạn cho chuỗi dài
4. Giám sát sử dụng GPU

---

## Những Phát Hiện Chính

### Cái Gì Hoạt Động

- **1K-4K token**: Đáng tin cậy với C=1 và C=10
- **8K token**: Đáng tin cậy với C=1
- **Flash Attention 2**: Tăng tốc độ 1,75x-11,9x
- **Hiệu quả bộ nhớ**: ~5-11MB mỗi yêu cầu

### Hạn Chế

- **4K token**: C=10 có tỷ lệ thành công 88% (12 lỗi OOM)
- **8K token**: C=10+ không được hỗ trợ (OOM)
- **16K+ token**: Không thể kiểm tra với 23GB VRAM (yêu cầu A100)

### Công Việc Tương Lai

- **16K-32K token**: Kế hoạch kiểm tra sẵn sàng, chờ môi trường A100 (40GB+ VRAM)
- **Đồng thời cao (C=50+)**: Kế hoạch kiểm tra sẵn sàng, chờ môi trường A100 (40GB+ VRAM)

Xem:
- [Kế Hoạch Kiểm Tra Bối Cảnh Dài](./modernbert-32k-docs/modernbert-32k-long-context-test-plan.md)
- [Kế Hoạch Kiểm Tra Lô Lớn](./modernbert-32k-docs/modernbert-32k-big-batch-test-plan.md)

---

## Tham Khảo

- **Công Cụ Điểm Chuẩn**: `candle-binding/examples/benchmark_concurrent.rs`
- **Công Cụ Hiệu Suất**: `candle-binding/examples/benchmark_performance.rs`
- **Kết Quả Đầy Đủ**: [Xác Thực Hiệu Suất](./modernbert-32k-docs/modernbert-32k-performance-validation.md)
- **Hướng Dẫn Triển Khai**: [Hướng Dẫn Triển Khai](./modernbert-32k-docs/modernbert-32k-deployment-guide.md)

---
