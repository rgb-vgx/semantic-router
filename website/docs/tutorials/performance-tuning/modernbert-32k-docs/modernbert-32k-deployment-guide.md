# Hướng Dẫn Triển Khai Cho Người Dùng Cuối

**Dự Án**: Tích Hợp ModernBERT-base-32k #995
**Dựa Trên**: Kết Quả Xác Thực Hiệu Suất & Chức Năng (Token 1K-8K)

## Tổng Quan

Hướng dẫn này cung cấp các khuyến nghị triển khai cho ModernBERT-base-32k dựa trên kết quả kiểm tra thực tế. Tất cả các khuyến nghị dựa trên kết quả kiểm tra được xác thực cho độ dài bối cảnh từ 512 token đến 8K token.

**Để kiểm tra bối cảnh dài (16K-32K) và kiểm tra lô lớn, hãy xem các kế hoạch kiểm tra riêng:**

- [Kế Hoạch Kiểm Tra Bối Cảnh Dài](./modernbert-32k-long-context-test-plan.md)
- [Kế Hoạch Kiểm Tra Lô Lớn](./modernbert-32k-big-batch-test-plan.md)

## 1. Khuyến Nghị Hiệu Suất

### 1.1. Kỳ Vọng Độ Trễ

| Độ Dài Bối Cảnh | Độ Trễ GPU (C=1) | Độ Trễ GPU (C=10) | Độ Trễ CPU |
|---|---|---|---|
| 512 token | 163ms | N/A | 7367ms |
| 1K token | 785ms | 996ms | 806ms |
| 4K token | 896ms | 9066ms (88% thành công) | N/A |
| 8K token | 3294ms | N/A (bị lỗi) | N/A |

### 1.2. Yêu Cầu Bộ Nhớ

| Độ Dài Bối Cảnh | Bộ Nhớ GPU mỗi Yêu Cầu | Tổng Bộ Nhớ (C=10) |
|---|---|---|
| 512 token | ~5MB | ~50MB |
| 1K token | ~11MB | ~110MB |
| 4K token | ~ước tính | ~ước tính |
| 8K token | ~ước tính | ~ước tính |

**Khuyến Nghị**:
- Bộ nhớ sử dụng rất hiệu quả (~5-11MB mỗi yêu cầu)
- Đối với 8K token với C=1, đảm bảo ít nhất 2GB bộ nhớ GPU trống

## 2. Giới Hạn Đồng Thời

### 2.1. Đồng Thời Được Khuyến Nghị theo Độ Dài Bối Cảnh

| Độ Dài Bối Cảnh | Đồng Thời Tối Đa | Ghi Chú |
|---|---|---|
| 1024 token | C=10 | Được kiểm tra và đáng tin cậy |
| 4096 token | C=10 | Tỷ lệ thành công 88% (12 lỗi OOM) |
| 8192 token | C=1 | Chỉ C=1 hoạt động đáng tin cậy |
| 16384+ token | C=1 (với phân đoạn) | Yêu cầu A100 hoặc phân đoạn |

## 3. Lựa Chọn Thiết Bị

### 3.1. Ma Trận Quyết Định GPU so với CPU

| Độ Dài Bối Cảnh | Thiết Bị Được Khuyến Nghị | Lý Do |
|---|---|---|
| 512 token | GPU | Nhanh 45 lần so với CPU |
| 1K token | GPU | Tương tự như CPU, nhưng có thể mở rộng hơn |
| 4K token | GPU | CPU không được kiểm tra, GPU được khuyến nghị |
| 8K token | GPU | CPU không được kiểm tra, GPU được khuyến nghị |

## 4. Chiến Lược Phân Đoạn

### 4.1. Khi Nào Sử Dụng Phân Đoạn

| Độ Dài Bối Cảnh | Phân Đoạn Bắt Buộc | Lý Do |
|---|---|---|
| ≤ 8K token | Không | Có thể xử lý trong một lần |
| > 8K token | Có | Hạn chế bộ nhớ hoặc tối ưu hóa độ trễ |
| > 32K token | Có | Giới hạn mô hình là 32K, phải phân đoạn |

## 5. Danh Sách Kiểm Tra Triển Khai Sản Xuất

### 5.1. Trước Triển Khai
- [ ] Xác minh tính khả dụng GPU (GPU NVIDIA với hỗ trợ CUDA)
- [ ] Kích hoạt Flash Attention 2 (được khuyến nghị)
- [ ] Đặt giới hạn đồng thời thích hợp dựa trên độ dài bối cảnh
- [ ] Cấu hình phân đoạn cho các chuỗi > 8K token
- [ ] Kiểm tra với tải production dự kiến

### 5.2. Giám Sát
- [ ] Giám sát sử dụng bộ nhớ GPU
- [ ] Theo dõi độ trễ (p50, p95, p99)
- [ ] Giám sát tỷ lệ lỗi (lỗi OOM)
- [ ] Theo dõi mức độ đồng thời
- [ ] Giám sát độ chính xác (trích xuất tín hiệu)

## 6. Khắc Phục Sự Cố

### 6.1. Vấn Đề Chung

#### Lỗi OOM (Hết Bộ Nhớ)

**Triệu Chứng**:
- Lỗi `CUDA_ERROR_OUT_OF_MEMORY`
- Yêu cầu bị lỗi ở đồng thời cao

**Giải Pháp**:
1. Giảm đồng thời (C=10 → C=1)
2. Sử dụng phân đoạn cho các chuỗi > 8K token
3. Tăng thời gian chờ giữa các yêu cầu
4. Sử dụng GPU với VRAM nhiều hơn (A100 40GB+)

#### Độ Trễ Cao

**Triệu Chứng**:
- Độ trễ cao hơn dự kiến
- Các bước nhảy di độ trễ p95/p99

**Giải Pháp**:
1. Kích hoạt Flash Attention 2
2. Giảm đồng thời
3. Sử dụng phân đoạn cho các chuỗi dài
4. Giám sát sử dụng GPU

---
