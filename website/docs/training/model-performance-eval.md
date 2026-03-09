# Đánh Giá Hiệu Năng Mô Hình

## Tại sao đánh giá?

Đánh giá làm cho định tuyến dữ liệu hướng. Bằng cách đo độ chính xác từng danh mục MMLU-Pro (và thực hiện kiểm tra sơ bộ nhanh với ARC), bạn có thể:

- Chọn mô hình phù hợp cho từng quyết định và cấu hình chúng trong decisions.modelRefs
- Chọn một default_model hợp lý dựa trên hiệu năng tổng thể
- Quyết định khi CoT prompting là giá trị đáng việc với sự đánh đổi độ trễ/chi phí
- Bắt các hồi quy khi các mô hình, lời nhắc hoặc tham số thay đổi
- Giữ các thay đổi có thể tái tạo và kiểm toán được cho CI và phát hành

Tóm lại, đánh giá chuyển đổi những suy đoán thành các tín hiệu có thể đo được cải thiện chất lượng, hiệu quả chi phí và độ tin cậy của bộ định tuyến.

---

Hướng dẫn này ghi lại quy trình làm việc tự động để đánh giá các mô hình thông qua điểm cuối OpenAI tương thích vLLM (MMLU-Pro và ARC Challenge), tạo cấu hình định tuyến dựa trên hiệu năng và cập nhật `categories.model_scores` trong cấu hình.

Xem mã trong [/src/training/model_eval](https://github.com/vllm-project/semantic-router/tree/main/src/training/model_eval)

### Những gì bạn sẽ chạy từ đầu đến cuối

#### 1) Đánh giá mô hình

- độ chính xác từng danh mục
- ARC Challenge: độ chính xác tổng thể

#### 2) Trực quan hóa kết quả

- biểu đồ thanh/heatmap của độ chính xác từng danh mục

![Bar](/img/bar.png)
![Heatmap](/img/heatmap.png)

#### 3) Tạo cấu hình cập nhật config.yaml

- Tạo quyết định cho mỗi danh mục với modelRefs
- Đặt default_model là người thực hiện tốt nhất trung bình
- Giữ hoặc áp dụng cài đặt lý do quyết định cấp độ

## 1. Điều kiện tiên quyết

- Một điểm cuối OpenAI tương thích vLLM đang chạy phục vụ các mô hình của bạn
  - URL Endpoint như http://localhost:8000/v1
  - API key tùy chọn nếu điểm cuối của bạn cần thiết

  ```bash
  # Terminal 1
  vllm serve microsoft/phi-4 --port 11434 --served_model_name phi4

  # Terminal 2
  vllm serve Qwen/Qwen3-0.6B --port 11435 --served_model_name qwen3-0.6B
  ```

- Gói Python cho các tập lệnh đánh giá:
  - Từ thư mục gốc của repo: matplotlib trong [requirements.txt](https://github.com/vllm-project/semantic-router/blob/main/requirements.txt)
  - Từ `/src/training/model_eval`: [requirements.txt](https://github.com/vllm-project/semantic-router/blob/main/src/training/model_eval/requirements.txt)

  ```bash
  # Chúng tôi sẽ làm việc ở thư mục này trong hướng dẫn này
  cd /src/training/model_eval
  pip install -r requirements.txt
  ```

**⚠️ Yêu Cầu Cấu Hình Quan Trọng:**

Tham số `--served-model-name` trong lệnh vLLM của bạn **phải khớp chính xác** với tên mô hình trong `config/config.yaml`:

```yaml
# config/config.yaml phải khớp với các giá trị --served-model-name ở trên
vllm_endpoints:
  - name: "endpoint1"
    address: "127.0.0.1"
    port: 11434
  - name: "endpoint2"
    address: "127.0.0.1"
    port: 11435

model_config:
  "phi4":                     # Khớp --served_model_name phi4
    # ... cấu hình
  "qwen3-0.6B":               # Khớp --served_model_name qwen3-0.6B
    # ... cấu hình
```

**Mẹo tùy chọn:**

- Đảm bảo `config/config.yaml` của bạn bao gồm tên mô hình được triển khai của bạn trong `vllm_endpoints[].models` và bất kỳ giá định giá/chính sách nào trong `model_config` nếu bạn dự định sử dụng cấu hình được tạo trực tiếp.

## 2. Đánh Giá MMLU-Pro

Xem tập lệnh trong [mmul_pro_vllm_eval.py](https://github.com/vllm-project/semantic-router/blob/main/src/training/model_eval/mmlu_pro_vllm_eval.py)

### Ví dụ về mô hình sử dụng

```bash
# Đánh giá một vài mô hình, vài mẫu cho mỗi danh mục, nhắc trực tiếp
python mmlu_pro_vllm_eval.py \
  --endpoint http://localhost:11434/v1 \
  --models phi4 \
  --samples-per-category 10

python mmlu_pro_vllm_eval.py \
  --endpoint http://localhost:11435/v1 \
  --models qwen3-0.6B \
  --samples-per-category 10

# Đánh giá với CoT (kết quả được lưu dưới *_cot)
python mmlu_pro_vllm_eval.py \
  --endpoint http://localhost:11435/v1 \
  --models qwen3-0.6B \
  --samples-per-category 10
  --use-cot

# Nếu bạn đã thiết lập Semantic Router đúng, bạn có thể chạy trong một lần
python mmlu_pro_vllm_eval.py \
  --endpoint http://localhost:8801/v1 \
  --models qwen3-0.6B, phi4 \
  --samples-per-category
  # --use-cot # Bỏ ghi chú dòng này nếu sử dụng CoT
```

### Các cờ chính

- **--endpoint**: URL OpenAI vLLM (mặc định http://localhost:8000/v1)
- **--models**: danh sách được phân tách bằng khoảng trắng HOẶC chuỗi được phân tách bằng dấu phẩy; nếu bỏ qua, tập lệnh truy vấn /models từ điểm cuối
- **--categories**: hạn chế đánh giá cho các danh mục cụ thể; nếu bỏ qua, sử dụng tất cả các danh mục trong dataset
- **--samples-per-category**: giới hạn câu hỏi cho mỗi danh mục (hữu ích cho các lần chạy nhanh)
- **--use-cot**: cho phép biến thể Chain-of-Thought prompting; kết quả được lưu trong thư mục con riêng biệt (_cot vs _direct)
- **--concurrent-requests**: đồng thời cho thông lượng
- **--output-dir**: nơi lưu kết quả (kết quả mặc định)
- **--max-tokens**, **--temperature**, **--seed**: tạo nút và tái tạo lại

### Nó xuất ra những gì cho mỗi mô hình

- **results/Model_Name_(direct|cot)/**
  - **detailed_results.csv**: một hàng cho mỗi câu hỏi với is_correct và danh mục
  - **analysis.json**: overall_accuracy, category_accuracy map, avg_response_time, counts
  - **summary.json**: số liệu cô đặc
- **mmlu_pro_vllm_eval.txt**: nhật ký lời nhắc và câu trả lời (gỡ lỗi/kiểm tra)

**Ghi chú**

- **Đặt tên mô hình**: dấu gạch chéo được thay thế bằng dấu gạch dưới cho tên thư mục; ví dụ: gemma3:27b -> thư mục gemma3:27b_direct.
- Độ chính xác danh mục được tính trên các truy vấn thành công; các yêu cầu không thành công bị loại trừ.

## 3. Đánh Giá trên ARC Challenge (tùy chọn, kiểm tra sơ bộ tổng thể)

Xem tập lệnh trong [arc_challenge_vllm_eval.py](https://github.com/vllm-project/semantic-router/blob/main/src/training/model_eval/arc_challenge_vllm_eval.py)

### Ví dụ về mô hình sử dụng

``` bash
python arc_challenge_vllm_eval.py \
  --endpoint http://localhost:8801/v1\
  --models qwen3-0.6B,phi4
  --output-dir arc_results
```

### Các cờ chính

- **--samples**: tổng số câu hỏi cần lấy mẫu (mặc định 20); ARC không được phân loại trong tập lệnh của chúng tôi
- Các cờ khác phản chiếu tập lệnh **MMLU-Pro**

### Nó xuất ra những gì cho mỗi mô hình

- **results/Model_Name_(direct|cot)/**
  - **detailed_results.csv**: một hàng cho mỗi câu hỏi với is_correct và danh mục
  - **analysis.json**: overall_accuracy, avg_response_time
  - **summary.json**: số liệu cô đặc
- **arc_challenge_vllm_eval.txt**: nhật ký lời nhắc và câu trả lời (gỡ lỗi/kiểm tra)

**Ghi chú**

- Kết quả ARC không trực tiếp cấp `categories[].model_scores`, nhưng chúng có thể giúp phát hiện hồi quy.

## 4. Trực Quan Hóa Hiệu Năng Từng Danh Mục

Xem tập lệnh trong [plot_category_accuracies.py](https://github.com/vllm-project/semantic-router/blob/main/src/training/model_eval/plot_category_accuracies.py)

### Ví dụ về mô hình sử dụng:

```bash
# Sử dụng results/ để tạo biểu đồ thanh
python plot_category_accuracies.py \
  --results-dir results \
  --plot-type bar \
  --output-file results/bar.png

# Sử dụng results/ để tạo biểu đồ heatmap
python plot_category_accuracies.py \
  --results-dir results \
  --plot-type heatmap \
  --output-file results/heatmap.png

# Sử dụng sample-data để tạo biểu đồ ví dụ
python src/training/model_eval/plot_category_accuracies.py \
  --sample-data \
  --plot-type heatmap \
  --output-file results/category_accuracies.png
```

### Các cờ chính

- **--results-dir**: nơi các tệp analysis.json nằm
- **--plot-type**: thanh hoặc heatmap
- **--output-file**: đường dẫn hình ảnh đầu ra (mô hình_eval/category_accuracies.png mặc định)
- **--sample-data**: nếu không có kết quả nào tồn tại, hãy tạo dữ liệu giả để xem trước cốt truyện

### Nó làm cái gì

- Tìm tất cả `results/**/analysis.json`, tổng hợp analysis["category_accuracy"] cho mỗi mô hình
- Thêm một cột Tổng thể đại diện cho mức trung bình trên các danh mục
- Tạo một hình để nhanh chóng so sánh hiệu năng mô hình/danh mục

**Ghi chú**

- Nó hợp nhất `direct` và `cot` là các biến thể mô hình khác biệt bằng cách thêm `:direct` hoặc `:cot` vào nhãn; huyền thoại ẩn `:direct` để tóm gọn.

## 5. Tạo Cấu Hình Định Tuyến Dựa Trên Hiệu Năng

Xem tập lệnh trong [result_to_config.py](https://github.com/vllm-project/semantic-router/blob/main/src/training/model_eval/result_to_config.py)

### Ví dụ về mô hình sử dụng

```bash
# Sử dụng results/ để tạo tệp cấu hình mới (không bị ghi đè)
python src/training/model_eval/result_to_config.py \
  --results-dir results \
  --output-file config/config.eval.yaml

# Sửa đổi ngưỡng tương tự
python src/training/model_eval/result_to_config.py \
  --results-dir results \
  --output-file config/config.eval.yaml \
  --similarity-threshold 0.85

# Tạo từ thư mục cụ thể
python src/training/model_eval/result_to_config.py \
  --results-dir results/mmlu_run_2025_09_10 \
  --output-file config/config.eval.yaml
```

### Các cờ chính

- **--results-dir**: trỏ đến thư mục nơi các tệp analysis.json nằm
- **--output-file**: đường dẫn cấu hình mục tiêu (config/config.yaml mặc định)
- **--similarity-threshold**: ngưỡng bộ nhớ cache ngữ nghĩa để đặt trong cấu hình được tạo

### Nó làm cái gì

- Đọc tất cả các tệp `analysis.json`, trích xuất analysis["category_accuracy"]
- Xây dựng một cấu hình mới:
  - **categories**: Tạo định nghĩa danh mục đơn giản (chỉ tên)
  - **decisions**: Cho mỗi danh mục có trong kết quả, tạo một quyết định với:
    - **rules**: Điều kiện định tuyến dựa trên miền
    - **modelRefs**: Các mô hình được xếp hạng theo độ chính xác (không có trường điểm)
    - **plugins**: Lời nhắc hệ thống và cấu hình khác
  - **default_model**: người thực hiện tốt nhất trung bình trên các danh mục
  - **Cài đặt lý do quyết định**: tự động điền từ ánh xạ tích hợp (bạn có thể điều chỉnh sau khi tạo)
    - toán, vật lý, hóa học, CS, kỹ thuật -> lý luận cao
    - khác mặc định -> thấp/trung bình
  - Bỏ qua bất kỳ mô hình trình giữ chỗ đặc biệt "auto" nào nếu có

### Căn Chỉnh Sơ Đồ

- **categories[].name**: chuỗi danh mục MMLU-Pro (đơn giản hóa, không có model_scores)
- **decisions[].name**: khớp tên danh mục
- **decisions[].modelRefs**: các mô hình được xếp hạng theo độ chính xác cho danh mục đó (không có trường điểm)
- **decisions[].rules**: điều kiện định tuyến dựa trên miền
- **decisions[].plugins**: cấu hình system_prompt và chính sách khác
- **default_model**: một người thực hiện hàng đầu trên các danh mục (tiền tố cách tiếp cận bị xóa, ví dụ: gemma3:27b từ gemma3:27b:direct)
- Giữ các phần cấu hình khác (semantic_cache, tools, classifier, prompt_guard) với các mặc định hợp lý; bạn có thể chỉnh sửa chúng sau khi tạo nếu môi trường của bạn khác

**Ghi chú**

- Tập lệnh này chỉ hoạt động với kết quả từ **MMLU_Pro** Evaluation.
- Cấu hình hiện tại config.yaml có thể bị ghi đè. Hãy cân nhắc viết vào tệp tạm thời trước và diff:
  - `--output-file config/config.eval.yaml`
- Nếu config.yaml sản xuất của bạn mang **cài đặt cụ thể về môi trường (điểm cuối, giá định giá, chính sách)**, hãy port `decisions[].modelRefs` được đánh giá và `default_model` trở lại cấu hình chính tắc của bạn.

### Ví Dụ config.eval.yaml

Xem thêm về cấu hình tại [configuration](https://vllm-semantic-router.com/docs/installation/configuration)

```yaml
bert_model:
  model_id: sentence-transformers/all-MiniLM-L12-v2
  threshold: 0.6
  use_cpu: true
semantic_cache:
  enabled: true
  similarity_threshold: 0.85
  max_entries: 1000
  ttl_seconds: 3600
tools:
  enabled: true
  top_k: 3
  similarity_threshold: 0.2
  tools_db_path: config/tools_db.json
  fallback_to_empty: true
prompt_guard:
  enabled: true
  use_modernbert: true
  model_id: models/jailbreak_classifier_modernbert-base_model
  threshold: 0.7
  use_cpu: true
  jailbreak_mapping_path: models/jailbreak_classifier_modernbert-base_model/jailbreak_type_mapping.json

# Thiếu cấu hình điểm cuối và model_config ngay tại đây, sửa đổi ở đây khi cần

classifier:
  category_model:
    model_id: models/category_classifier_modernbert-base_model
    use_modernbert: true
    threshold: 0.6
    use_cpu: true
    category_mapping_path: models/category_classifier_modernbert-base_model/category_mapping.json
  pii_model:
    model_id: models/pii_classifier_modernbert-base_presidio_token_model
    use_modernbert: true
    threshold: 0.7
    use_cpu: true
    pii_mapping_path: models/pii_classifier_modernbert-base_presidio_token_model/pii_type_mapping.json
categories:
- name: business
- name: law
- name: engineering

decisions:
- name: business
  description: "Route business queries"
  priority: 10
  reasoning_effort: low
  rules:
    operator: "OR"
    conditions:
      - type: "domain"
        name: "business"
  modelRefs:
  - model: phi4
    use_reasoning: false
  - model: qwen3-0.6B
    use_reasoning: false
  plugins:
    - type: "system_prompt"
      configuration:
        enabled: true
        system_prompt: "Business content is typically conversational"
        mode: "replace"

- name: law
  description: "Route legal queries"
  priority: 10
  reasoning_effort: medium
  rules:
    operator: "OR"
    conditions:
      - type: "domain"
        name: "law"
  modelRefs:
  - model: phi4
    use_reasoning: false
  - model: qwen3-0.6B
    use_reasoning: false
  plugins:
    - type: "system_prompt"
      configuration:
        enabled: true
        system_prompt: "Legal content is typically explanatory"
        mode: "replace"

# Bỏ qua một số danh mục ở đây

- name: engineering
  description: "Route engineering queries"
  priority: 10
  reasoning_effort: high
  rules:
    operator: "OR"
    conditions:
      - type: "domain"
        name: "engineering"
  modelRefs:
  - model: phi4
    use_reasoning: true
  - model: qwen3-0.6B
    use_reasoning: true
  plugins:
    - type: "system_prompt"
      configuration:
        enabled: true
        system_prompt: "Engineering problems require systematic problem-solving"
        mode: "replace"

default_reasoning_effort: medium
default_model: phi4
```
