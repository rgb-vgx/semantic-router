# Khí Quyển Lựa Chọn Mô Hình

Lựa chọn mô hình là một tính năng nâng cao của bộ định tuyến ngữ nghĩa vLLM tự động chọn LLM tốt nhất từ nhiều ứng viên dựa trên các ưu tiên đã học, tương đồng truy vấn và tối ưu hóa chi phí-chất lượng.

Bộ định tuyến ngữ nghĩa hỗ trợ **9 thuật toán lựa chọn** trên hai loại:

- **Các thuật toán cơ bản**: Tĩnh, Nhận Thức Độ Trễ, Elo, RouterDC, AutoMix, Hybrid
- **Các thuật toán được điều khiển bởi RL**: Thompson Sampling, GMTRouter, Router-R1

## Vấn Đề Nó Giải Quyết?

Khi bạn có nhiều backend LLM (ví dụ: GPT-4, Claude, Llama, Mistral), bạn đối mặt với một thách thức: **mô hình nào sẽ xử lý từng yêu cầu?**

Các phương pháp truyền thống:

- **Định tuyến tĩnh**: Luôn sử dụng cùng một mô hình (đơn giản nhưng không tối ưu)
- **Round-robin**: Phân phối đều (bỏ qua điểm mạnh mô hình)
- **Ngẫu nhiên**: Không có trí tuệ (lãng phí tài nguyên)

Lựa chọn mô hình giải quyết vấn đề bằng cách **khớp thông minh các truy vấn với các mô hình** dựa trên:

1. **Ưu tiên chất lượng đã học** (xếp hạng Elo từ phản hồi người dùng)
2. **Tương đồng truy vấn-mô hình** (nhúng RouterDC)
3. **Sự đánh đổi chi phí-chất lượng** (tối ưu hóa AutoMix)
4. **Tín hiệu kết hợp** (phương pháp Hybrid)

## Các Thuật Toán Có Sẵn

### Các Thuật Toán Cơ Bản

| Thuật Toán | Tốt Cho | Lợi Ích Chính |
|-----------|----------|-------------|
| [**Tĩnh**](./static.md) | Triển khai đơn giản | Dự đoán được, chi phí chung bằng không |
| **Nhận Thức Độ Trễ** | Định tuyến nhạy cảm với độ trễ | Chọn theo phần trăm TPOT/TTFT |
| [**Elo**](./elo.md) | Học từ phản hồi | Thích ứng với ưu tiên người dùng |
| [**RouterDC**](./router-dc.md) | Khớp truy vấn-mô hình | Khớp chuyên môn với truy vấn |
| [**AutoMix**](./automix.md) | Tối ưu hóa chi phí | Cân bằng chất lượng và chi phí |
| [**Hybrid**](./hybrid.md) | Yêu cầu phức tạp | Kết hợp tất cả các phương pháp |

### Các Thuật Toán Được Điều Khiển Bởi RL

| Thuật Toán | Tốt Cho | Lợi Ích Chính |
|-----------|----------|-------------|
| [**Thompson Sampling**](./thompson-sampling.md) | Khám phá/Khai thác | Học thích ứng Bayes |
| [**GMTRouter**](./gmtrouter.md) | Cá Nhân Hóa | Học ưa thích cho mỗi người dùng |
| [**Router-R1**](./router-r1.md) | Suy luận phức tạp | Quyết định định tuyến do LLM cung cấp |

## Khởi Động Nhanh

### Cấu Hình Cơ Bản (Mỗi Quyết Định)

Lựa chọn mô hình được cấu hình cho mỗi quyết định, cho phép các chiến lược khác nhau cho các loại truy vấn khác nhau:

```yaml
decisions:
  - name: tech
    description: "Truy vấn kỹ thuật"
    priority: 10
    rules:
      operator: "OR"
      conditions:
        - type: "domain"
          name: "tech"
    modelRefs:
      - model: "llama3.2:3b"
      - model: "phi4"
      - model: "gemma3:27b"
    algorithm:
      type: "elo"  # Sử dụng xếp hạng Elo cho quyết định này
      elo:
        k_factor: 32
        category_weighted: true
```

### Loại Thuật Toán

#### Tĩnh (Mặc Định)
Sử dụng mô hình đầu tiên trong `modelRefs`. Không học, hoàn toàn xác định.

```yaml
algorithm:
  type: "static"
```

#### Nhận Thức Độ Trễ
Chọn mô hình nhanh nhất theo phần trăm TPOT/TTFT.

```yaml
algorithm:
  type: "latency_aware"
  latency_aware:
    tpot_percentile: 10
    ttft_percentile: 10
```

#### Xếp Hạng Elo
Tìm hiểu từ phản hồi người dùng để xếp hạng mô hình theo chất lượng.

```yaml
algorithm:
  type: "elo"
  elo:
    k_factor: 32
    storage_path: "/var/lib/vsr/elo.json"
```

#### RouterDC
Khớp các nhúng truy vấn với mô tả mô hình.

```yaml
algorithm:
  type: "router_dc"
  router_dc:
    temperature: 0.07
    require_descriptions: true
```

#### AutoMix
Tối ưu hóa sự đánh đổi chi phí-chất lượng bằng POMDP.

```yaml
algorithm:
  type: "automix"
  automix:
    cost_quality_tradeoff: 0.4
```

#### Hybrid
Kết hợp tất cả các phương pháp với trọng số có thể cấu hình.

```yaml
algorithm:
  type: "hybrid"
  hybrid:
    elo_weight: 0.3
    router_dc_weight: 0.3
    automix_weight: 0.2
    cost_weight: 0.2
```

## Cách Hoạt Động

```
┌─────────────────────────────────────────────────────────────────────┐
│                         Truy Vấn Người Dùng                         │
│                    "Giải thích máy tính lượng tử"                   │
└─────────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      Khớp Quyết Định                                │
│                 Quyết Định "tech" khớp → 3 mô hình                 │
└─────────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    Thuật Toán Lựa Chọn                              │
│                                                                      │
│  algorithm.type: "elo"                                              │
│                                                                      │
│  ┌─────────────────────────────────────────────────────┐            │
│  │ EloSelector.Select()                                │            │
│  │                                                     │            │
│  │ Xếp hạng Mô hình:                                   │            │
│  │   llama3.2:3b  → 1468 (0 thắng, 2 thua)             │            │
│  │   phi4         → 1501 (3 thắng, 2 thua)             │            │
│  │   gemma3:27b   → 1531 (5 thắng, 1 thua) ← CAO NHẤT  │            │
│  └─────────────────────────────────────────────────────┘            │
└─────────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      Được Chọn: gemma3:27b                          │
│                   (xếp hạng Elo cao nhất: 1531)                     │
└─────────────────────────────────────────────────────────────────────┘
```

## Chọn Thuật Toán

Xem [Chọn Thuật Toán Phù Hợp](./choosing-algorithm.md) để hướng dẫn chi tiết.

**Cây Ra Quyết Định Nhanh:**

1. **Mới bắt đầu?** → Sử dụng `static` (mặc định)
2. **Cần định tuyến dựa trên độ trễ?** → Sử dụng `latency_aware`
3. **Có phản hồi người dùng?** → Sử dụng `elo`
4. **Có mô tả mô hình?** → Sử dụng `router_dc`
5. **Muốn tối ưu hóa chi phí?** → Sử dụng `automix`
6. **Cần tất cả mọi thứ?** → Sử dụng `hybrid`

## Các Tính Năng Liên Quan

- **Định Tuyến Phản Hồi Người Dùng** - Collect tín hiệu phản hồi qua `/api/v1/feedback`
- **Định Tuyến Ưu Tiên** - Định tuyến dựa trên ưu tiên người dùng trong hệ thống
- **Định Tuyến Miền** - Định tuyến theo danh mục chủ đề bằng phân loại nhúng

## Bài Báo Tham Khảo

Các thuật toán lựa chọn dựa trên những bài báo nghiên cứu này:

### Các Thuật Toán Cơ Bản

- **Elo**: Được lấy cảm hứng từ các khái niệm định tuyến dựa trên ưu tiên; xem [RouteLLM](https://arxiv.org/abs/2406.18665) (Ong et al., ICLR 2025) đạt giảm 50% chi phí
- **RouterDC**: [Bộ Định Tuyến Dựa Trên Truy Vấn Bằng Học Tập Đối Lập Kép](https://arxiv.org/abs/2409.19886) (NeurIPS 2024) - cải tiến chính xác +2.76%
- **AutoMix**: [Tự Động Trộn Mô Hình Ngôn Ngữ](https://arxiv.org/abs/2310.12963) (NeurIPS 2024) - giảm chi phí >50%
- **Hybrid**: [Định Tuyến Truy Vấn Có Hiệu Suất Chi Phí Lý-Nhận Thức](https://arxiv.org/abs/2404.14618) (ICLR 2024) - 40% ít gọi mô hình đắt tiền hơn

### Các Thuật Toán Được Điều Khiển Bởi RL

- **Thompson Sampling**: Phương pháp cờ ba vũ trang đa cánh cổ điển; xem [Hướng Dẫn về Thompson Sampling](https://arxiv.org/abs/1707.02038) (Russo, Van Roy et al.)
- **GMTRouter**: [GMTRouter: Bộ Định Tuyến LLM Được Cá Nhân Hóa Qua Tương Tác Người Dùng Nhiều Lượt](https://arxiv.org/abs/2511.08590) (Wang et al.) - cải tiến chính xác 0.9-21.6%
- **Router-R1**: [Router-R1: Dạy LLM Định Tuyến và Tổng Hợp Nhiều Vòng qua RL](https://arxiv.org/abs/2506.09033) (Hu et al., NeurIPS 2025)
