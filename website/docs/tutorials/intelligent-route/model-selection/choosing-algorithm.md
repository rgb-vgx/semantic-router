# Chọn Thuật Toán Lựa Chọn

Hướng dẫn này giúp bạn chọn thuật toán lựa chọn mô hình phù hợp cho trường hợp sử dụng của bạn.

## Cây Ra Quyết Định Nhanh

```
Bạn có cần định tuyến xác định không?
├── Có → Lựa Chọn Tĩnh
└── Không → Tiếp tục...

Bạn có sẵn phản hồi người dùng không?
├── Có, dồi dào → Xếp Hạng Elo
├── Có phản hồi → Hybrid (với thành phần Elo)
└── Không có phản hồi → Tiếp tục...

Bạn có mô tả mô hình tốt không?
├── Có → RouterDC
└── Không → Tiếp tục...

Tối ưu hóa chi phí có quan trọng không?
├── Có → AutoMix
└── Không → Tĩnh hoặc Hybrid
```

## So Sánh Thuật Toán

### Các Thuật Toán Cơ Bản

| Thuật Toán | Cần Phản Hồi | Độ Phức Tạp Cấu Hình | Khả Năng Thích Ứng | Tối Ưu Hóa Chi Phí |
|-----------|-----------------|------------------|--------------|-------------------|
| Tĩnh | Không | Thấp | Không | Thủ công |
| Elo | Cao | Trung bình | Cao | Gián tiếp |
| RouterDC | Không | Trung bình | Trung bình | Không |
| AutoMix | Thấp | Trung bình | Trung bình | Có |
| Hybrid | Thay đổi | Cao | Cao | Có |

### Các Thuật Toán Được Điều Khiển Bởi RL

| Thuật Toán | Cần Phản Hồi | Độ Phức Tạp Cấu Hình | Khả Năng Thích Ứng | Cá Nhân Hóa |
|-----------|-----------------|------------------|--------------|-----------------|
| Thompson | Trung bình | Thấp | Cao | Tùy Chọn Cho Mỗi Người Dùng |
| GMTRouter | Trung bình | Cao | Rất Cao | Tích Hợp |
| Router-R1 | Không | Cao | Cao | Qua LLM |

## Khuyến Nghị Trường Hợp Sử Dụng

### Khởi Động / MVP
**Được Khuyến Nghị: Tĩnh**

Bắt đầu đơn giản với các quy tắc rõ ràng. Chuyển sang các phương pháp thích ứng khi bạn thu thập dữ liệu.

```yaml
algorithm:
  type: static
  static:
    default_model: gpt-3.5-turbo
```

### Sản Xuất Khối Lượng Lớn
**Được Khuyến Nghị: AutoMix hoặc Hybrid**

Tối ưu hóa chi phí trong khi duy trì chất lượng quy mô.

```yaml
algorithm:
  type: automix
  automix:
    cost_quality_tradeoff: 0.4
```

### Ứng Dụng Hướng Người Dùng
**Được Khuyến Nghị: Elo**

Cho phép phản hồi người dùng xác định lựa chọn mô hình cho chất lượng chủ quan.

```yaml
algorithm:
  type: elo
  elo:
    k_factor: 32
    storage_path: /data/elo-ratings.json
```

### Lĩnh Vực Chuyên Biệt
**Được Khuyến Nghị: RouterDC**

Khi các mô hình có các chuyên môn riêng biệt, hãy khớp các truy vấn với khả năng mô hình.

```yaml
algorithm:
  type: router_dc
  router_dc:
    require_descriptions: true
```

### Doanh Nghiệp / Nhiều Mục Tiêu
**Được Khuyến Nghị: Hybrid**

Cân bằng nhiều yếu tố: chất lượng, chi phí, sự hài lòng người dùng và chuyên môn.

```yaml
algorithm:
  type: hybrid
  hybrid:
    elo_weight: 0.3
    router_dc_weight: 0.3
    automix_weight: 0.2
    cost_weight: 0.2
```

### Nền Tảng Đa Người Dùng Được Cá Nhân Hóa
**Được Khuyến Nghị: Thompson Sampling hoặc GMTRouter**

Tìm hiểu các ưu tiên người dùng riêng lẻ theo thời gian.

```yaml
algorithm:
  type: thompson
  thompson:
    per_user: true
    min_samples: 10
```

### Nghiên Cứu / Logic Định Tuyến Phức Tạp
**Được Khuyến Nghị: Router-R1**

Khi các quyết định định tuyến yêu cầu sự hiểu biết ngữ nghĩa khó mã hóa trong các quy tắc.

```yaml
algorithm:
  type: router_r1
  router_r1:
    router_endpoint: http://localhost:8001
    use_cot: true
```

## Đường Dẫn Di Cư

Một tiến trình điển hình khi hệ thống của bạn trưởng thành:

1. **Bắt đầu**: Lựa chọn tĩnh với các quy tắc đơn giản
2. **Thêm phản hồi**: Di cư đến Elo khi bạn thu thập phản hồi người dùng
3. **Thêm mô tả**: Thêm RouterDC để khớp truy vấn-mô hình
4. **Tối ưu hóa chi phí**: Kết hợp AutoMix để hiệu quả chi phí
5. **Kết hợp**: Sử dụng Hybrid để tận dụng tất cả các phương pháp

## Những Cân Nhắc Chính

### Yêu Cầu Dữ Liệu

- **Tĩnh**: Không cần dữ liệu
- **Elo**: Cần phản hồi người dùng nhất quán (cái cộc/cái trừ)
- **RouterDC**: Cần mô tả mô hình chất lượng
- **AutoMix**: Cần giá chính xác và điểm chất lượng
- **Hybrid**: Kết hợp của những điều trên
- **Thompson**: Cần phản hồi; hoạt động trực tuyến
- **GMTRouter**: Được hưởng lợi từ lịch sử tương tác; có thể huấn luyện trước
- **Router-R1**: Cần máy chủ LLM định tuyến; mô tả mô hình giúp

### Ảnh Hưởng Độ Trễ

| Thuật Toán | Độ Trễ Điển Hình |
|-----------|-----------------|
| Tĩnh | Dưới 1ms |
| Elo | Dưới 2ms |
| RouterDC | 2-5ms (nhúng) |
| AutoMix | Dưới 3ms |
| Hybrid | 3-5ms |
| Thompson | Dưới 2ms |
| GMTRouter | 5-15ms (GNN) |
| Router-R1 | 100-500ms (LLM) |

### Bảo Trì

- **Tĩnh**: Cập nhật quy tắc thủ công
- **Elo**: Tự duy trì với phản hồi
- **RouterDC**: Cập nhật mô tả khi mô hình thay đổi
- **AutoMix**: Cập nhật giá khi chi phí thay đổi
- **Hybrid**: Điều chỉnh trọng số định kỳ được khuyến nghị
