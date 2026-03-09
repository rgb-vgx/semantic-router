---
sidebar_position: 8
---

# Hướng dẫn Định tuyến Ngữ cảnh

Hướng dẫn này hướng dẫn bạn cách sử dụng **Tín hiệu Ngữ cảnh** (Số lượng Mã thông báo) để định tuyến các yêu cầu dựa trên độ dài của chúng.

Điều này hữu ích cho:

- Định tuyến các truy vấn ngắn đến các mô hình nhanh hơn, nhỏ hơn
- Định tuyến các tài liệu/dấu nhắc dài đến các mô hình có cửa sổ ngữ cảnh lớn
- Tối ưu hóa chi phí bằng cách sử dụng các mô hình rẻ hơn cho các tác vụ ngắn

## Kịch bản

Chúng tôi muốn:

1. Định tuyến các yêu cầu ngắn (< 4K mã thông báo) đến một mô hình nhanh (`llama-3-8b`)
2. Định tuyến các yêu cầu trung bình (4K - 32K mã thông báo) đến một mô hình tiêu chuẩn (`llama-3-70b`)
3. Định tuyến các yêu cầu dài (32K - 128K mã thông báo) đến một mô hình ngữ cảnh-lớn (`claude-3-opus`)

## Bước 1: Xác định Tín hiệu Ngữ cảnh

Thêm `context_rules` vào cấu hình `signals` của bạn:

```yaml
signals:
  context:
    - name: "short_context"
      min_tokens: "0"
      max_tokens: "4K"
      description: "Short queries suitable for fast models"

    - name: "medium_context"
      min_tokens: "4K"
      max_tokens: "32K"
      description: "Medium length context"

    - name: "long_context"
      min_tokens: "32K"
      max_tokens: "128K"
      description: "Long context requiring specialized handling"
```

## Bước 2: Xác định Quyết định

Tạo các quyết định được kích hoạt dựa trên các tín hiệu ngữ cảnh này:

```yaml
decisions:
  - name: "fast_route"
    priority: 10
    rules:
      operator: "AND"
      conditions:
        - type: "context"
          name: "short_context"
    modelRefs:
      - model: "llama-3-8b"

  - name: "standard_route"
    priority: 10
    rules:
      operator: "AND"
      conditions:
        - type: "context"
          name: "medium_context"
    modelRefs:
      - model: "llama-3-70b"

  - name: "long_context_route"
    priority: 10
    rules:
      operator: "AND"
      conditions:
        - type: "context"
          name: "long_context"
    modelRefs:
      - model: "claude-3-opus"
```

## Bước 3: Logic Kết hợp (Nâng cao)

Bạn có thể kết hợp các tín hiệu ngữ cảnh với các tín hiệu khác (như miền hoặc từ khóa).

**Ví dụ**: Định tuyến các tác vụ **code** dài đến một mô hình code dài-ngữ cảnh chuyên dụng:

```yaml
decisions:
  - name: "long_code_analysis"
    priority: 20  # Higher priority
    rules:
      operator: "AND"
      conditions:
        - type: "context"
          name: "long_context"
        - type: "domain"
          name: "computer_science"
    modelRefs:
      - model: "deepseek-coder-v2"
```

## Cách Tính toán Mã thông báo Hoạt động

- Router tính toán số mã thông báo **trước** khi đưa ra quyết định định tuyến.
- Nó sử dụng một bộ mã hóa mã thông báo nhanh tương thích với hầu hết các LLM.
- Các hậu tố như "K" (1000) và "M" (1.000.000) được hỗ trợ để dễ đọc.
- Nếu một yêu cầu khớp với nhiều phạm vi (ví dụ: các quy tắc chồng lấn), tất cả các tín hiệu phù hợp đều hoạt động.

## Giám sát

Bạn có thể giám sát phân phối mã thông báo bằng cách sử dụng mục quan sát Prometheus:
`llm_context_token_count`

Điều này giúp bạn điều chỉnh các phạm vi của mình dựa trên các mẫu lưu lượng thực tế.
