# Lựa Chọn Mô Hình Dựa Trên ML

Tài liệu này bao gồm cấu hình và kết quả thử nghiệm cho các kỹ thuật lựa chọn mô hình dựa trên ML được triển khai trong Bộ Định Tuyến Ngữ Nghĩa.

## Tổng Quan

Lựa chọn mô hình dựa trên ML sử dụng các thuật toán học máy để định tuyến thông minh các truy vấn đến LLM phù hợp nhất dựa trên đặc điểm truy vấn và dữ liệu hiệu năng lịch sử. Phương pháp này cung cấp những cải tiến đáng kể so với các chiến lược định tuyến ngẫu nhiên hoặc mô hình đơn.

### Các Thuật Toán Được Hỗ Trợ

| Thuật Toán | Mô Tả | Tốt Nhất Cho |
|-----------|-------------|----------|
| **KNN** (K-Nearest Neighbors) | Bỏ phiếu có trọng số chất lượng giữa các truy vấn tương tự | Độ chính xác cao, các loại truy vấn đa dạng |
| **KMeans** | Định tuyến dựa trên cụm với tối ưu hóa hiệu quả | Suy luận nhanh, tải cân bằng |
| **SVM** (Support Vector Machine) | Ranh giới quyết định nhân RBF | Tách biệt miền rõ ràng |
| **MLP** (Multi-Layer Perceptron) | Mạng nơ-ron với tăng tốc GPU | Thông lượng cao, môi trường có GPU |

### Các Bài Báo Tham Khảo

- [FusionFactory (arXiv:2507.10540)](https://arxiv.org/abs/2507.10540) - Hợp nhất cấp truy vấn thông qua bộ định tuyến LLM
- [Avengers-Pro (arXiv:2508.12631)](https://arxiv.org/abs/2508.12631) - Tối ưu hóa hiệu suất-hiệu quả định tuyến

## Thiết Lập GUI Bảng Điều Khiển

**Bảng Điều Khiển Bộ Định Tuyến Ngữ Nghĩa** bao gồm trình hướng dẫn 3 bước có hướng dẫn cho toàn bộ quy trình lựa chọn mô hình ML. Đây là cách dễ nhất để bắt đầu - không cần CLI.

### Truy Cập Trình Hướng Dẫn

Điều hướng đến `http://localhost:8700/ml-setup` (hoặc URL bảng điều khiển của bạn).

### Bước 1: Chuẩn Bị

Tải lên **YAML mô hình** của bạn (liệt kê các điểm cuối LLM của bạn) và tệp **queries JSONL** (kiểm tra các truy vấn có lẫn thì đúng). Định cấu hình đồng thời và mã thông báo tối đa, sau đó nhấp **Run Benchmark**. Tiến độ luồng theo thời gian thực với độ chi tiết từng truy vấn.

### Bước 2: Huấn Luyện

Chọn một hoặc nhiều thuật toán:

| Thuật Toán | GPU Bắt Buộc | Ghi Chú |
|-----------|-------------|-------|
| KNN | Không | Chỉ CPU (scikit-learn) |
| K-Means | Không | Chỉ CPU (scikit-learn) |
| SVM | Không | Chỉ CPU (scikit-learn) |
| MLP | Tùy chọn | PyTorch - Bộ chọn thiết bị (CPU/CUDA) xuất hiện khi được chọn |

Điều chỉnh siêu tham số qua bảng **Advanced Settings**, sau đó nhấp **Train Models**. Các tệp mô hình được huấn luyện (`knn_model.json`, `kmeans_model.json`, v.v.) được lưu vào thư mục `ml-train/` cố định.

### Bước 3: Tạo Cấu Hình

Xác định các quyết định định tuyến - mỗi quyết định có tên, ưu tiên, thuật toán, điều kiện miền và tên mô hình mục tiêu. Nhấp **Generate Config** để tạo `ml-model-selection-values.yaml` sẵn sàng triển khai.

YAML được tạo tuân theo sơ đồ cấu hình ngữ nghĩa-router. Hợp nhất các phần `model_selection` và `decisions` vào `config.yaml` chính của bạn hoặc sử dụng nó làm tệp cấu hình độc lập cho bộ định tuyến.

### Ví Dụ Cấu Hình Được Tạo

```yaml
config:
  model_selection:
    enabled: true
    ml:
      models_path: /data/ml-pipeline/ml-train
      embedding_dim: 1024
      knn:
        k: 5
        pretrained_path: /data/ml-pipeline/ml-train/knn_model.json
  strategy: priority
  decisions:
    - name: math-decision
      priority: 100
      rules:
        operator: OR
        conditions:
          - type: domain
            name: math
      algorithm:
        type: knn
      modelRefs:
        - model: llama3.2:3b
          use_reasoning: false
```

## Cấu Hình

### Cấu Hình Cơ Bản

Kích hoạt lựa chọn mô hình dựa trên ML trong `config.yaml` của bạn:

```yaml
# Kích hoạt lựa chọn mô hình ML
model_selection:
  ml:
    enabled: true
    models_path: ".cache/ml-models"  # Đường dẫn đến các tệp mô hình được huấn luyện

# Mô hình nhúng cho biểu diễn truy vấn
embedding_models:
  qwen3_model_path: "models/mom-embedding-pro"  # Qwen3-Embedding-0.6B
```

### Cấu Hình Thuật Toán Theo Quyết Định

Định cấu hình các thuật toán khác nhau cho các loại quyết định khác nhau:

```yaml
decisions:
  # Truy vấn toán học - sử dụng KNN cho lựa chọn có trọng số chất lượng
  - name: "math_decision"
    description: "Toán học và suy luận định lượng"
    priority: 100
    rules:
      operator: "AND"
      conditions:
        - type: "domain"
          name: "math"
    algorithm:
      type: "knn"
      knn:
        k: 5
    modelRefs:
      - model: "llama-3.2-1b"
      - model: "llama-3.2-3b"
      - model: "mistral-7b"

  # Truy vấn mã - sử dụng SVM cho ranh giới rõ ràng
  - name: "code_decision"
    description: "Lập trình và phát triển phần mềm"
    priority: 100
    rules:
      operator: "AND"
      conditions:
        - type: "domain"
          name: "computer science"
    algorithm:
      type: "svm"
      svm:
        kernel: "rbf"
        gamma: 1.0
    modelRefs:
      - model: "codellama-7b"
      - model: "llama-3.2-3b"
      - model: "mistral-7b"

  # Truy vấn chung - sử dụng KMeans để có hiệu quả
  - name: "general_decision"
    description: "Truy vấn kiến thức chung"
    priority: 50
    rules:
      operator: "AND"
      conditions:
        - type: "domain"
          name: "other"
    algorithm:
      type: "kmeans"
      kmeans:
        num_clusters: 8
    modelRefs:
      - model: "llama-3.2-1b"
      - model: "llama-3.2-3b"
      - model: "mistral-7b"

  # Truy vấn thông lượng cao - sử dụng MLP với tăng tốc GPU
  - name: "gpu_accelerated_decision"
    description: "Suy luận khối lượng cao với GPU"
    priority: 100
    rules:
      operator: "AND"
      conditions:
        - type: "domain"
          name: "engineering"
    algorithm:
      type: "mlp"
      mlp:
        device: "cuda"  # hoặc "cpu", "metal"
    modelRefs:
      - model: "llama-3.2-1b"
      - model: "llama-3.2-3b"
      - model: "mistral-7b"
      - model: "codellama-7b"
```

### Tham Số Thuật Toán

#### Tham Số KNN

```yaml
algorithm:
  type: "knn"
  knn:
    k: 5  # Số lượng hàng xóm (mặc định: 5)
```

#### Tham Số KMeans

```yaml
algorithm:
  type: "kmeans"
  kmeans:
    num_clusters: 8  # Số lượng cụm (mặc định: 8)
```

#### Tham Số SVM

```yaml
algorithm:
  type: "svm"
  svm:
    kernel: "rbf"   # Loại nhân: rbf, linear (mặc định: rbf)
    gamma: 1.0      # Gamma nhân RBF (mặc định: 1.0)
```

#### Tham Số MLP

```yaml
algorithm:
  type: "mlp"
  mlp:
    device: "cuda"  # Thiết bị: cpu, cuda, metal (mặc định: cpu)
```

Thuật toán MLP (Multi-Layer Perceptron) sử dụng bộ phân loại mạng nơ-ron với tăng tốc GPU thông qua khung Rust [Candle](https://github.com/huggingface/candle). Nó cung cấp suy luận thông lượng cao phù hợp cho triển khai sản xuất với tài nguyên GPU.

**Các Tùy Chọn Thiết Bị:**

| Thiết Bị | Mô Tả | Yêu Cầu |
|--------|-------------|--------------|
| `cpu` | Suy luận CPU (mặc định) | Không cần phần cứng đặc biệt |
| `cuda` | Tăng tốc GPU NVIDIA | GPU hỗ trợ CUDA, bộ công cụ CUDA |
| `metal` | GPU Apple Silicon | macOS với chip M1/M2/M3 |

## Kết Quả Thử Nghiệm

### Thiết Lập Chuẩn Bị

- **Truy vấn kiểm tra**: 109 truy vấn trên nhiều miền
- **Mô hình được đánh giá**: 4 LLM (codellama-7b, llama-3.2-1b, llama-3.2-3b, mistral-7b)
- **Mô hình nhúng**: Qwen3-Embedding-0.6B (1024 chiều)
- **Dữ liệu xác thực**: Truy vấn chuẩn bị thực tế với điểm hiệu năng sự thật cơ sở

### So Sánh Hiệu Năng

| Chiến Lược | Chất Lượng TB | Độ Trễ TB | Mô Hình Tốt Nhất % |
|----------|-------------|-------------|--------------|
| **Oracle (tốt nhất có thể)** | 0.495 | 10.57s | 100.0% |
| **KMEANS Selection** | 0.252 | 20.23s | 23.9% |
| Luôn llama-3.2-3b | 0.242 | 25.08s | 15.6% |
| **SVM Selection** | 0.233 | 25.83s | 14.7% |
| Luôn mistral-7b | 0.215 | 70.08s | 13.8% |
| Luôn llama-3.2-1b | 0.212 | 3.65s | 26.6% |
| **KNN Selection** | 0.196 | 36.62s | 13.8% |
| Lựa Chọn Ngẫu Nhiên | 0.174 | 40.12s | 9.2% |
| Luôn codellama-7b | 0.161 | 53.78s | 4.6% |

### Lợi Ích Định Tuyến ML So Với Lựa Chọn Ngẫu Nhiên

| Thuật Toán | Cải Tiến Chất Lượng | Lựa Chọn Mô Hình Tốt Nhất |
|-----------|---------------------|---------------------|
| **KMEANS** | **+45.5%** | 2.6x thường xuyên hơn |
| **SVM** | **+34.4%** | 1.6x thường xuyên hơn |
| **KNN** | **+13.1%** | 1.5x thường xuyên hơn |

### Những Phát Hiện Chính

1. **Tất cả các phương pháp ML vượt trội so với lựa chọn ngẫu nhiên** - Những cải tiến chất lượng đáng kể trên tất cả các thuật toán
2. **KMEANS cung cấp chất lượng tốt nhất** - Cải tiến 45% so với ngẫu nhiên với độ trễ tốt
3. **SVM mang lại hiệu suất cân bằng** - Cải tiến 34% với ranh giới quyết định rõ ràng
4. **KNN cung cấp lựa chọn mô hình đa dạng** - Sử dụng tất cả các mô hình khả dụng dựa trên sự tương tự truy vấn
5. **MLP cho phép tăng tốc GPU** - Suy luận mạng nơ-ron với hỗ trợ CUDA/Metal cho thông lượng cao

### Tăng Tốc GPU MLP

Thuật toán MLP tận dụng khung Rust [Candle](https://github.com/huggingface/candle) cho suy luận có tăng tốc GPU:

| Thiết Bị | Độ Trễ Suy Luận | Thông Lượng |
|--------|------------------|------------|
| CPU | ~5-10ms | ~100-200 QPS |
| CUDA (NVIDIA) | ~0.5-1ms | ~1000+ QPS |
| Metal (Apple) | ~1-2ms | ~500+ QPS |

**Khi nào sử dụng MLP:**

- Triển khai sản xuất khối lượng cao với tài nguyên GPU
- Các ứng dụng nhạy cảm độ trễ yêu cầu suy luận dưới một mili giây
- Các môi trường nơi chi phí lựa chọn mô hình phải được giảm thiểu

## Kiến Trúc

```text
┌─────────────────────────────────────────────────────────────────────┐
│                         ONLINE INFERENCE                            │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  Request (model="auto")                                             │
│       ↓                                                             │
│  Generate Query Embedding (Qwen3, 1024-dim)                         │
│       ↓                                                             │
│  Add Category One-Hot (14-dim) → 1038-dim feature vector            │
│       ↓                                                             │
│  Decision Engine → Match decision by domain                         │
│       ↓                                                             │
│  Load ML Selector (KNN/KMeans/SVM/MLP from JSON)                    │
│       ↓                                                             │
│  Run Inference → Select best model                                  │
│       ↓                                                             │
│  Route to selected LLM endpoint                                     │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

## Huấn Luyện Các Mô Hình Của Riêng Bạn

**Huấn Luyện Ngoại Tuyến vs Suy Luận Trực Tuyến:**

- **Huấn Luyện Ngoại Tuyến**: Được thực hiện trong **Python** sử dụng scikit-learn cho KNN, KMeans, SVM, PyTorch cho MLP
- **Suy Luận Trực Tuyến**: Được thực hiện trong **Rust** sử dụng [Linfa](https://github.com/rust-ml/linfa) cho KNN/KMeans/SVM qua `ml-binding` và [Candle](https://github.com/huggingface/candle) cho MLP qua `candle-binding`

Sự phân tách này cho phép huấn luyện linh hoạt với hệ sinh thái ML phong phú của Python trong khi duy trì suy luận hiệu suất cao trong sản xuất với Rust.

### Điều Kiện Tiên Quyết

```bash
cd src/training/model_selection/ml_model_selection
pip install -r requirements.txt
```

### Tùy Chọn 1: Tải Xuống Các Mô Hình Được Huấn Luyện Trước

```bash
python download_model.py \
  --output-dir ../../../.cache/ml-models \
  --repo-id abdallah1008/semantic-router-ml-models
```

### Tùy Chọn 2: Huấn Luyện Sử Dụng Dữ Liệu Được Chuẩn Bị Trước Từ HuggingFace

Chúng tôi cung cấp dữ liệu chuẩn bị sẵn sàng ở HuggingFace mà bạn có thể sử dụng trực tiếp để huấn luyện:

**Dataset HuggingFace:** [abdallah1008/ml-selection-benchmark-data](https://huggingface.co/datasets/abdallah1008/ml-selection-benchmark-data)

| Tệp | Mô Tả |
|------|-------------|
| `benchmark_training_data.jsonl` | Dữ liệu được chuẩn bị trước với 4 mô hình (codellama-7b, llama-3.2-1b, llama-3.2-3b, mistral-7b) |
| `validation_benchmark_with_gt.jsonl` | Dữ liệu xác thực với sự thật cơ sở để kiểm tra |

```bash
# Tải xuống dữ liệu chuẩn bị
huggingface-cli download abdallah1008/ml-selection-benchmark-data \
  --repo-type dataset \
  --local-dir .cache/ml-models

# Huấn luyện trực tiếp sử dụng dữ liệu được chuẩn bị trước
python train.py \
  --data-file .cache/ml-models/benchmark_training_data.jsonl \
  --output-dir models/
```

Đây là cách nhanh nhất để bắt đầu - không cần chạy các điểm chuẩn LLM của riêng bạn!

### Tùy Chọn 3: Huấn Luyện Với Dữ Liệu Của Riêng Bạn

#### Bước 1: Chuẩn Bị Dữ Liệu Đầu Vào (Định Dạng JSONL)

Tạo tệp JSONL với các truy vấn của bạn. Mỗi dòng phải chứa các trường `query` và `category`:

```jsonl
{"query": "Đạo hàm của x^2 là gì?", "category": "math", "ground_truth": "2x"}
{"query": "Viết hàm Python để sắp xếp danh sách", "category": "computer science", "ground_truth": "def sort(lst): return sorted(lst)"}
{"query": "Giải thích quang hợp", "category": "biology", "ground_truth": "Quá trình mà các cây chuyển đổi ánh sáng mặt trời thành năng lượng"}
{"query": "Các yêu cầu pháp lý cho hợp đồng là gì?", "category": "law"}
```

**Các trường bắt buộc:**

| Trường | Loại | Mô Tả |
|-------|------|-------------|
| `query` | string | Văn bản truy vấn đầu vào |
| `category` | string | Danh mục miền (xem [Danh Mục VSR](#danh-mục-vsr)) |
| `ground_truth` | string | Câu trả lời dự kiến (bắt buộc để tính điểm hiệu năng/chất lượng) |

**Các trường được khuyến nghị (để tính điểm hiệu năng chính xác):**

| Trường | Loại | Mô Tả |
|-------|------|-------------|
| `metric` | string | Phương pháp đánh giá - xác định cách tính hiệu năng |
| `choices` | string | Để trả lời nhiều lựa chọn - tín hiệu đánh giá MC |

**Các trường tùy chọn:**

| Trường | Loại | Mô Tả |
|-------|------|-------------|
| `task_name` | string | Định danh tác vụ để ghi nhật ký/theo dõi (ví dụ: "mmlu", "gsm8k") |

**Quan Trọng: Trường Số Liệu**

Không có `metric`, điểm chuẩn sử dụng **CEM (Conditional Exact Match)** như mặc định, có thể không chính xác tính điểm:

- Vấn đề toán học (sử dụng `metric: "GSM8K"` hoặc `metric: "MATH"`)
- Nhiều lựa chọn (sử dụng `metric: "em_mc"` hoặc bao gồm `choices`)
- Tạo mã (sử dụng `metric: "code_eval"`)

Để có kết quả tốt nhất, luôn chỉ định `metric` thích hợp cho loại câu hỏi của bạn.

**Câu Hỏi Nhiều Lựa Chọn**

Để trả lời nhiều lựa chọn, bao gồm `choices` (có thể là các tùy chọn dưới dạng chuỗi) và đặt `ground_truth` thành chữ cái đúng:

```jsonl
{"query": "Thủ đô của Pháp là gì?\nA) Luân Đôn\nB) Paris\nC) Berlin\nD) Rome", "category": "other", "ground_truth": "B", "choices": "London,Paris,Berlin,Rome"}
```

Tập lệnh chuẩn bị:

1. Phát hiện nhiều lựa chọn thông qua trường `choices` hoặc `metric: "em_mc"`
2. Trích xuất chữ cái trả lời (A/B/C/D) từ phản hồi mô hình
3. So sánh với `ground_truth` (chữ cái đúng)

**Số Liệu Đánh Giá**

Trường `metric` kiểm soát cách tính hiệu năng:

| Số Liệu | Mô Tả | Thí Dụ ground_truth |
|--------|-------------|----------------------|
| `em_mc` | Nhiều lựa chọn - trích xuất chữ cái | `"B"` |
| `GSM8K` | Toán - trích xuất số sau `####` | `"giải thích #### 42"` |
| `MATH` | Toán LaTeX - trích xuất từ `\boxed{}` | `"\\boxed{2x+1}"` |
| `f1_score` | Điểm F1 trùng lặp văn bản | `"Paris là thủ đô"` |
| `code_eval` | Chạy khẳng định mã | `"['assert func(1)==2']"` |
| (default) | CEM - khớp vùng chứa | `"Paris"` |

**Ground Truth Là Bắt Buộc Để Huấn Luyện**

Trường `ground_truth` là cần thiết để huấn luyện lựa chọn mô hình ML. Không có nó, hệ thống không thể tính toán mô hình nào thực hiện tốt hơn trên mỗi truy vấn. Quá trình huấn luyện so sánh phản hồi của mỗi LLM với `ground_truth` để tính điểm hiệu năng.

#### Bước 2: Định Cấu Hình Các Điểm Cuối LLM Của Bạn (models.yaml)

Tạo tệp `models.yaml` để định cấu hình các điểm cuối LLM của bạn với xác thực:

```yaml
models:
  # Mô hình Ollama cục bộ (không cần xác thực)
  - name: llama-3.2-1b
    endpoint: http://localhost:11434/v1

  - name: llama-3.2-3b
    endpoint: http://localhost:11434/v1

  # OpenAI với khóa API từ biến môi trường
  - name: gpt-4
    endpoint: https://api.openai.com/v1
    api_key: ${OPENAI_API_KEY}
    max_tokens: 2048
    temperature: 0.0

  # HuggingFace với token
  - name: mistral-7b-hf
    endpoint: https://api-inference.huggingface.co/models/mistralai/Mistral-7B-Instruct-v0.2
    api_key: ${HF_TOKEN}
    headers:
      Authorization: "Bearer ${HF_TOKEN}"

  # API tùy chỉnh với token bearer
  - name: custom-llm
    endpoint: https://api.custom.com/v1
    api_key: ${CUSTOM_API_KEY}
    headers:
      Authorization: "Bearer ${CUSTOM_API_KEY}"
      X-Custom-Header: "value"
    max_tokens: 1024
    temperature: 0.1

  # vLLM tự lưu trữ
  - name: codellama-7b
    endpoint: http://vllm-server:8000/v1
    # Không cần xác thực cho vLLM cục bộ
```

#### Bước 3: Chạy Chuẩn Bị

Tập lệnh chuẩn bị gửi từng truy vấn đến tất cả các LLM được định cấu hình và đo:

**Hiệu Năng (Điểm Chất Lượng 0-1):**

| Loại Truy Vấn | Phương Pháp Tính Điểm |
|------------|----------------|
| **Nhiều Lựa Chọn** (A/B/C/D) | Khớp chính xác của tùy chọn được chọn vs `ground_truth` |
| **Số/Toán** | Phân tích và so sánh các số (dựa trên dung sai) |
| **Văn Bản/Mã** | Điểm F1 giữa phản hồi mô hình và `ground_truth` |
| **Khớp Chính Xác** | Nhị phân 1.0 nếu khớp chính xác, 0.0 nếu không |

**Độ Trễ (Thời Gian Phản Hồi):**

- Được đo từ yêu cầu gửi đến phản hồi nhận được (tính bằng giây)
- Bao gồm độ trễ mạng + thời gian suy luận mô hình
- Được sử dụng để cân nhân hiệu quả: `speed_factor = 1 / (1 + độ trễ)`

**Định Dạng Đầu Ra:**

Điểm chuẩn tạo JSONL với một bản ghi cho mỗi cặp (truy vấn, mô hình):

```jsonl
{"query": "2+2 là gì?", "category": "math", "model_name": "llama-3.2-1b", "response": "4", "ground_truth": "4", "performance": 1.0, "response_time": 0.523}
{"query": "2+2 là gì?", "category": "math", "model_name": "llama-3.2-3b", "response": "Câu trả lời là 4", "ground_truth": "4", "performance": 0.85, "response_time": 1.234}
{"query": "2+2 là gì?", "category": "math", "model_name": "mistral-7b", "response": "2+2=4", "ground_truth": "4", "performance": 0.92, "response_time": 2.156}
```

**Chạy Chuẩn Bị:**

```bash
# Sử dụng tệp cấu hình mô hình (được khuyến nghị)
python benchmark.py \
  --queries your_queries.jsonl \
  --model-config models.yaml \
  --output benchmark_output.jsonl \
  --concurrency 4 \
  --limit 500  # Tùy chọn: giới hạn số lượng truy vấn để kiểm tra

# Hoặc sử dụng danh sách mô hình đơn giản (tất cả cùng điểm cuối)
python benchmark.py \
  --queries your_queries.jsonl \
  --models llama-3.2-1b,llama-3.2-3b,mistral-7b \
  --endpoint http://localhost:11434/v1 \
  --output benchmark_output.jsonl
```

**Tham Số benchmark.py:**

| Tham Số | Mặc Định | Mô Tả |
|-----------|---------|-------------|
| `--queries` | (bắt buộc) | Đường dẫn đến tệp JSONL đầu vào với các truy vấn |
| `--model-config` | Không | Đường dẫn đến models.yaml với cấu hình điểm cuối |
| `--models` | Không | Danh sách tên mô hình được phân tách bằng dấu phẩy (thay thế cho --model-config) |
| `--endpoint` | `http://localhost:8000/v1` | Điểm cuối API (được sử dụng với --models) |
| `--output` | `benchmark_output.jsonl` | Đường dẫn tệp đầu ra |
| `--concurrency` | `4` | Số lượng yêu cầu song song đến các LLM |
| `--limit` | Không | Giới hạn số lượng truy vấn được xử lý |
| `--max-tokens` | `1024` | Số lượng token tối đa trong phản hồi LLM |
| `--temperature` | `0.0` | Nhiệt độ để tạo (0.0 = xác định) |

**Tham Số Đồng Thời**

Tham số `--concurrency` kiểm soát bao nhiêu yêu cầu được gửi đến các LLM song song:

- **Các giá trị cao hơn** (8-16): Chuẩn bị nhanh hơn, nhưng có thể làm quá tải các mô hình cục bộ
- **Các giá trị thấp hơn** (1-2): Chậm hơn nhưng an toàn hơn cho các môi trường hạn chế tài nguyên
- **Được khuyến nghị**: Bắt đầu với 4, tăng nếu máy chủ LLM của bạn có thể xử lý nhiều hơn

Đối với Ollama trên một GPU duy nhất, sử dụng `--concurrency 2-4`. Đối với API đám mây (OpenAI, HuggingFace), bạn có thể sử dụng `--concurrency 8-16`.

#### Bước 4: Huấn Luyện Các Mô Hình ML

```bash
python train.py \
  --data-file benchmark_output.jsonl \
  --output-dir models/
```

### Tham Số train.py

| Tham Số | Mặc Định | Mô Tả |
|-----------|---------|-------------|
| `--data-file` | (bắt buộc) | Đường dẫn đến dữ liệu JSONL chuẩn bị |
| `--output-dir` | `models/` | Thư mục để lưu các tệp JSON mô hình được huấn luyện |
| `--embedding-model` | `qwen3` | Mô hình nhúng: `qwen3`, `gte`, `mpnet`, `e5`, `bge` |
| `--cache-dir` | `.cache/` | Thư mục lưu trữ cho nhúng |
| `--knn-k` | `5` | Số lượng hàng xóm cho KNN |
| `--kmeans-clusters` | `8` | Số lượng cụm cho KMeans |
| `--svm-kernel` | `rbf` | Nhân SVM: `rbf`, `linear` |
| `--svm-gamma` | `1.0` | Gamma SVM cho nhân RBF |
| `--mlp-hidden-dims` | `512,256` | Kích thước lớp ẩn MLP |
| `--mlp-dropout` | `0.1` | Tỷ lệ thả MLP |
| `--mlp-epochs` | `100` | Kỷ nguyên huấn luyện MLP |
| `--mlp-lr` | `0.001` | Tốc độ học tập MLP |
| `--quality-weight` | `0.9` | Trọng số chất lượng vs tốc độ (0=tốc độ, 1=chất lượng) |
| `--batch-size` | `32` | Kích thước lô để tạo nhúng |
| `--device` | `cpu` | Thiết bị: `cpu`, `cuda`, `mps` |
| `--limit` | Không | Giới hạn số lượng mẫu huấn luyện |

**Ví Dụ:**

```bash
# Huấn luyện với giá trị k KNN tùy chỉnh
python train.py \
  --data-file benchmark.jsonl \
  --output-dir models/ \
  --knn-k 7

# Huấn luyện với mẫu hạn chế (để kiểm tra)
python train.py \
  --data-file benchmark.jsonl \
  --output-dir models/ \
  --limit 1000

# Huấn luyện với tăng tốc GPU
python train.py \
  --data-file benchmark.jsonl \
  --output-dir models/ \
  --device cuda \
  --batch-size 64

# Huấn luyện với tham số thuật toán tùy chỉnh
python train.py \
  --data-file benchmark.jsonl \
  --output-dir models/ \
  --knn-k 10 \
  --kmeans-clusters 12 \
  --svm-kernel rbf \
  --svm-gamma 0.5 \
  --quality-weight 0.85

# Huấn luyện MLP với kiến trúc tùy chỉnh
python train.py \
  --data-file benchmark.jsonl \
  --output-dir models/ \
  --mlp-hidden-dims 1024,512,256 \
  --mlp-dropout 0.2 \
  --mlp-epochs 150 \
  --mlp-lr 0.0005 \
  --device cuda
```

### Danh Mục VSR

Hệ thống hỗ trợ 14 danh mục miền. Sử dụng tên chính xác (có khoảng trắng, không dấu gạch dưới):

```text
biology, business, chemistry, computer science, economics, engineering,
health, history, law, math, other, philosophy, physics, psychology
```

### Xác Thực Các Mô Hình Được Huấn Luyện

Chạy tập lệnh xác thực Go để xác minh lợi ích định tuyến ML:

```bash
cd src/training/model_selection/ml_model_selection

# Đặt đường dẫn thư viện (WSL/Linux)
export LD_LIBRARY_PATH=$PWD/../../../candle-binding/target/release:$PWD/../../../ml-binding/target/release:$LD_LIBRARY_PATH

# Chạy xác thực
go run validate.go --qwen3-model /path/to/Qwen3-Embedding-0.6B
```

## Tệp Mô Hình

Các mô hình được huấn luyện được lưu trữ dưới dạng tệp JSON:

| Tệp | Thuật Toán | Kích Thước |
|------|-----------|------|
| `knn_model.json` | K-Nearest Neighbors | ~2-10 MB |
| `kmeans_model.json` | KMeans Clustering | ~50 KB |
| `svm_model.json` | Support Vector Machine | ~1-5 MB |
| `mlp_model.json` | Multi-Layer Perceptron | ~1-10 MB |

Các tệp này được tải xuống từ HuggingFace hoặc tạo ra trong quá trình huấn luyện:

- **Mô hình**: [abdallah1008/semantic-router-ml-models](https://huggingface.co/abdallah1008/semantic-router-ml-models)
- **Dữ Liệu Chuẩn Bị**: [abdallah1008/ml-selection-benchmark-data](https://huggingface.co/datasets/abdallah1008/ml-selection-benchmark-data)

## Các Thực Hành Tốt Nhất

### Hướng Dẫn Lựa Chọn Thuật Toán

| Trường Hợp Sử Dụng | Thuật Toán Được Khuyến Nghị | Lý Do |
|----------|----------------------|--------|
| **Tác vụ quan trọng về chất lượng** | KNN (k=5) | Bỏ phiếu có trọng số chất lượng cung cấp độ chính xác tốt nhất |
| **Hệ thống thông lượng cao** | KMeans | Tìm kiếm cụm nhanh, độ trễ tốt |
| **Định tuyến miền cụ thể** | SVM | Ranh giới quyết định rõ ràng giữa các miền |
| **Môi trường có GPU** | MLP | Mạng nơ-ron với tăng tốc CUDA/Metal |
| **Mục đích chung** | KMEANS | Cân bằng tốt giữa chất lượng và tốc độ |

### Điều Chỉnh Siêu Tham Số

1. **Giá trị KNN k**: Bắt đầu với k=5, tăng lên để quyết định mịn hơn
2. **Cụm KMeans**: Khớp với số lượng mô hình truy vấn khác biệt (8-16 điển hình)
3. **Gamma SVM**: Sử dụng 1.0 cho nhúng được bình thường hóa, điều chỉnh dựa trên sự lan truyền dữ liệu
4. **Kiến trúc MLP**: Bắt đầu với các kích thước ẩn 512,256; tăng cho các tập dữ liệu phức tạp

### Thành Phần Vector Tính Năng

Các mô hình ML sử dụng vectơ tính năng 1038 chiều:

- **1024 chiều**: Nhúng ngữ nghĩa Qwen3
- **14 chiều**: Mã hóa một nóng danh mục (danh mục miền VSR)

```text
Feature Vector = [embedding(1024)] ⊕ [category_one_hot(14)]
```

## Khắc Phục Sự Cố

### Mô Hình Không Tải

```text
Error: pretrained model not found
```

Tải xuống mô hình từ HuggingFace:

```bash
cd src/training/model_selection/ml_model_selection
python download_model.py --output-dir ../../../.cache/ml-models
```

### Độ Chính Xác Lựa Chọn Thấp

1. Đảm bảo mô hình nhúng khớp với huấn luyện (Qwen3-Embedding-0.6B)
2. Xác minh phân loại danh mục hoạt động
3. Kiểm tra tên mô hình trong cấu hình khớp với dữ liệu huấn luyện

### Không Khớp Chiều

```text
Error: embedding dimension mismatch
```

Đảm bảo bạn đang sử dụng cùng một mô hình nhúng để huấn luyện và suy luận (Qwen3 tạo ra 1024 chiều).

## Bước Tiếp Theo

- [Tổng Quan Đào Tạo](/docs/training/training-overview) - Tài liệu đào tạo chung
- [Đánh Giá Hiệu Năng Mô Hình](/docs/training/model-performance-eval) - Số liệu hiệu năng chi tiết
