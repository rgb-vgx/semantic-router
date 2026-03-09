# Kế Hoạch Kiểm Tra Lô Lớn (Đồng Thời Cao C=50+)

**Dự Án**: Tích Hợp ModernBERT-base-32k #995
**Yêu Cầu**: GPU NVIDIA A100 (40GB+ VRAM)

## Tổng Quan

Kế hoạch kiểm tra này bao gồm xác thực ModernBERT-base-32k dưới các kịch bản đồng thời cao (lô lớn). Các bài kiểm tra này không thể hoàn thành với môi trường hiện tại (GPU NVIDIA L4, 23GB VRAM) do các vấn đề phân mảnh bộ nhớ và yêu cầu GPU A100 có 40GB+ VRAM.

## Các Trường Hợp Kiểm Tra

### 1. Kiểm Tra Đồng Thời Cao (C=50, C=100)

#### 1.1 Độ Dài Bối Cảnh Thấp (1K-4K token)

| Độ Dài Bối Cảnh | Đồng Thời | Tỷ Lệ Thành Công Dự Kiến | Độ Trễ Dự Kiến |
|---|---|---|---|
| 1024 token | C=50 | ≥ 90% | < 2000ms (p95) |
| 1024 token | C=100 | ≥ 80% | < 3000ms (p95) |
| 4096 token | C=50 | ≥ 80% | < 15000ms (p95) |
| 4096 token | C=100 | ≥ 70% | < 20000ms (p95) |

#### 1.2 Độ Dài Bối Cảnh Trung Bình (8K token)

| Độ Dài Bối Cảnh | Đồng Thời | Tỷ Lệ Thành Công Dự Kiến | Độ Trễ Dự Kiến |
|---|---|---|---|
| 8192 token | C=50 | ≥ 70% | < 25000ms (p95) |
| 8192 token | C=100 | ≥ 60% | < 35000ms (p95) |

## Kết Quả Dự Kiến

### Tiêu Chuẩn Thành Công

1. **C=50**:
   - 1K token: ≥ 90% tỷ lệ thành công
   - 4K token: ≥ 80% tỷ lệ thành công
   - 8K token: ≥ 70% tỷ lệ thành công

2. **C=100**:
   - 1K token: ≥ 80% tỷ lệ thành công
   - 4K token: ≥ 70% tỷ lệ thành công
   - 8K token: ≥ 60% tỷ lệ thành công

3. **Độ Trễ**:
   - Độ trễ P95 trong các giới hạn chấp nhận được
   - Không có độ trễ đuôi quá mức

4. **Bộ Nhớ**:
   - Sử dụng bộ nhớ trong giới hạn GPU
   - Không có rò rỉ bộ nhớ
   - Phân mảnh chấp nhận được

---
