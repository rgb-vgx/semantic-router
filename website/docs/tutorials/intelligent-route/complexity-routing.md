---
sidebar_position: 9
---

# Hướng dẫn Định tuyến Độ phức tạp

Hướng dẫn này hướng dẫn bạn cách sử dụng **Tín hiệu Độ phức tạp** để định tuyến các yêu cầu dựa trên mức độ khó của chúng.

Điều này hữu ích cho:

- Định tuyến các truy vấn phức tạp đến những mô hình mạnh mẽ, chuyên biệt
- Định tuyến các truy vấn đơn giản đến các mô hình nhanh, hiệu quả
- Tối ưu hóa chi phí bằng cách sử dụng các mô hình rẻ hơn cho các tác vụ dễ
- Cải thiện chất lượng phản hồi bằng cách khớp độ khó truy vấn với khả năng mô hình

## Kịch bản

Chúng tôi muốn:

1. Định tuyến các câu hỏi lập trình đơn giản (ví dụ: "in hello world") đến một mô hình nhanh (`llama-3-8b`)
2. Định tuyến các câu hỏi độ phức tạp trung bình đến một mô hình tiêu chuẩn (`llama-3-70b`)
3. Định tuyến các câu hỏi phức tạp (ví dụ: "thiết kế hệ thống phân tán") đến một mô hình chuyên biệt (`deepseek-coder-v2`)

## Bước 1: Xác định Tín hiệu Miền (Điều kiện tiên quyết)

Đầu tiên, xác định các tín hiệu miền để phân loại loại truy vấn. Điều này **bắt buộc** để sử dụng các composer với tín hiệu độ phức tạp:

```yaml
signals:
  domains:
    - name: "computer_science"
      description: "Programming, algorithms, software engineering, system design"
      mmlu_categories:
        - "computer_science"
        - "machine_learning"

    - name: "math"
      description: "Mathematics, calculus, algebra, statistics"
      mmlu_categories:
        - "mathematics"
        - "statistics"
```

## Bước 2: Xác định Tín hiệu Độ phức tạp với Composer

Thêm các quy tắc `complexity` vào cấu hình `signals` của bạn. **QUAN TRỌNG**: Luôn sử dụng `composer` để lọc dựa trên miền để ngăn chặn phân loại sai lĩnh vực:

```yaml
signals:
  complexity:
    - name: "code_complexity"
      composer:
        operator: "AND"
        conditions:
          - type: "domain"
            name: "computer_science"
      threshold: 0.1
      description: "Detects code complexity level based on task difficulty"
      hard:
        candidates:
          - "design distributed system"
          - "implement consensus algorithm"
          - "optimize for scale"
          - "architect microservices"
          - "fix race condition"
          - "implement garbage collector"
      easy:
        candidates:
          - "print hello world"
          - "loop through array"
          - "read file"
          - "sort list"
          - "string concatenation"
          - "simple function"

    - name: "math_complexity"
      composer:
        operator: "AND"
        conditions:
          - type: "domain"
            name: "math"
      threshold: 0.1
      description: "Detects mathematical problem complexity"
      hard:
        candidates:
          - "prove mathematically"
          - "derive the equation"
          - "formal proof"
          - "solve differential equation"
          - "prove by induction"
          - "analyze convergence"
      easy:
        candidates:
          - "add two numbers"
          - "calculate percentage"
          - "simple arithmetic"
          - "basic algebra"
          - "count items"
          - "find average"
```

**Tại sao sử dụng composer?** Nếu không có composer, một truy vấn toán học như "chứng minh bằng cảm ứng" có thể phù hợp không chính xác với `code_complexity` nếu nó có tương đồng cao hơn với các ứng viên code. Composer đảm bảo `code_complexity` chỉ được kích hoạt khi miền là "computer_science".

## Bước 3: Xác định Quyết định

Tạo các quyết định được kích hoạt dựa trên mức độ phức tạp:

```yaml
decisions:
  - name: "simple_code_route"
    priority: 10
    rules:
      operator: "AND"
      conditions:
        - type: "complexity"
          name: "code_complexity:easy"
    modelRefs:
      - model: "llama-3-8b"

  - name: "medium_code_route"
    priority: 10
    rules:
      operator: "AND"
      conditions:
        - type: "complexity"
          name: "code_complexity:medium"
    modelRefs:
      - model: "llama-3-70b"

  - name: "complex_code_route"
    priority: 10
    rules:
      operator: "AND"
      conditions:
        - type: "complexity"
          name: "code_complexity:hard"
    modelRefs:
      - model: "deepseek-coder-v2"
```

## Bước 4: Logic Kết hợp (Nâng cao)

Bạn có thể kết hợp các tín hiệu độ phức tạp với các tín hiệu khác (như miền hoặc ngôn ngữ).

**Ví dụ**: Định tuyến các tác vụ lập trình **tiếng Trung** phức tạp đến một mô hình chuyên biệt:

```yaml
decisions:
  - name: "complex_chinese_code"
    priority: 20  # Higher priority
    rules:
      operator: "AND"
      conditions:
        - type: "complexity"
          name: "code_complexity:hard"
        - type: "language"
          name: "zh"
    modelRefs:
      - model: "qwen-coder-plus"
```

## Cách Phân loại Độ phức tạp Hoạt động

Tín hiệu độ phức tạp sử dụng **quá trình đánh giá hai pha**:

### Pha 1: Đánh giá Tín hiệu Song song

Tất cả các quy tắc độ phức tạp được đánh giá **độc lập và song song** với các tín hiệu khác:

1. Truy vấn được so sánh với tất cả các **ứng viên khó** → max_hard_similarity
2. Truy vấn được so sánh với tất cả các **ứng viên dễ** → max_easy_similarity
3. Tín hiệu khó khăn = max_hard_similarity - max_easy_similarity
4. Phân loại:
   - Nếu tín hiệu > ngưỡng (0,1): **khó**
   - Nếu tín hiệu < -ngưỡng (-0.1): **dễ**
   - Nếu không thì: **trung bình**

Ở giai đoạn này, **tất cả các quy tắc** phù hợp sẽ được giữ lại (ví dụ: "code_complexity:hard" và "math_complexity:hard" có thể khớp).

### Pha 2: Lọc Composer

Sau khi tất cả các tín hiệu được tính toán, các điều kiện composer sẽ được đánh giá:

1. Đối với mỗi quy tắc độ phức tạp phù hợp, kiểm tra xem nó có `composer`
2. Nếu composer tồn tại, đánh giá các điều kiện của nó đối với các kết quả tín hiệu khác
3. Chỉ giữ lại các quy tắc có điều kiện composer được thỏa mãn
4. Điều này ngăn chặn phân loại sai lĩnh vực

### Ví dụ Luồng

**Truy vấn**: "Làm cách nào để triển khai một thuật toán đồng thuận phân tán?"

**Pha 1 - Đánh giá Song song:**

1. **Tín hiệu Miền**: Khớp "computer_science" (được đánh giá song song)
2. **code_complexity**:
   - max_hard_similarity = 0,85 (khớp "triển khai thuật toán đồng thuận")
   - max_easy_similarity = 0,15 (khớp thấp với ứng viên dễ)
   - difficulty_signal = 0,85 - 0,15 = 0,70
   - Kết quả: 0,70 > 0,1 → **"code_complexity:hard"** (tạm thời)
3. **math_complexity**:
   - max_hard_similarity = 0,25 (độc lập cho "thuật toán")
   - max_easy_similarity = 0,10
   - difficulty_signal = 0,25 - 0,10 = 0,15
   - Kết quả: 0,15 > 0,1 → **"math_complexity:hard"** (tạm thời)

**Pha 2 - Lọc Composer:**

1. **code_complexity** kiểm tra composer:
   - Yêu cầu: domain = "computer_science"
   - Tín hiệu miền khớp "computer_science" ✅
   - **GIỮ LẠI**: "code_complexity:hard"
2. **math_complexity** kiểm tra composer:
   - Yêu cầu: domain = "math"
   - Tín hiệu miền khớp "computer_science" (không phải "math") ❌
   - **LỌC BỎ**

**Kết quả Cuối cùng**: "code_complexity:hard"

**Định tuyến**: Khớp quyết định → Định tuyến đến `deepseek-coder-v2`

## Các thực hành tốt nhất

1. **Luôn Sử dụng Composer**: Cấu hình một composer cho mỗi quy tắc độ phức tạp để lọc dựa trên miền. Điều này **tính chất quan trọng** để ngăn chặn phân loại sai lĩnh vực.
2. **Xác định Tín hiệu Miền Trước tiên**: Các composer độ phức tạp phụ thuộc vào tín hiệu miền, vì vậy hãy xác định các miền trước các quy tắc độ phức tạp.
3. **Ứng viên Đa dạng**: Bao gồm các ví dụ đa dạng trong ứng viên khó/dễ để bao gồm các mẫu truy vấn khác nhau.
4. **Ngưỡng Điều chỉnh**: Điều chỉnh ngưỡng (mặc định 0,1) dựa trên trường hợp sử dụng của bạn. Ngưỡng thấp hơn = phân loại nhạy cảm hơn.
5. **Kết quả Giám sát**: Kiểm tra các quyết định định tuyến trong tiêu đề phản hồi để tinh chỉnh ứng viên và ngưỡng.
6. **Điều kiện Composer Nhiều**: Bạn có thể sử dụng nhiều điều kiện với các toán tử AND/OR:

   ```yaml
   composer:
     operator: "OR"
     conditions:
       - type: "domain"
         name: "computer_science"
       - type: "keyword"
         name: "coding_keywords"
   ```

7. **Mô tả là Tùy chọn**: Trường `description` bây giờ là tùy chọn và chỉ được sử dụng cho tài liệu. Nó không ảnh hưởng đến phân loại.

## Giám sát

Tín hiệu độ phức tạp trả về kết quả ở định dạng: `"rule_name:difficulty"`

Bạn có thể kiểm tra quyết định định tuyến trong tiêu đề phản hồi:

```http
x-vsr-matched-complexity: code_complexity:hard
```

Điều này giúp bạn giám sát và gỡ lỗi các quyết định định tuyến dựa trên độ phức tạp.
