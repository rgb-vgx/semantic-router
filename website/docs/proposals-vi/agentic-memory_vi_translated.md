# Agentic Memory

## Tóm tắt điều hành

Tài liệu này mô tả một **Bằng chứng khái niệm (Proof of Concept)** cho Agentic Memory trong Semantic Router. Agentic Memory cho phép các tác nhân AI **ghi nhớ thông tin qua các phiên làm việc**, cung cấp tính liên tục và cá nhân hóa.

> **⚠️ Phạm vi POC:** Đây là bằng chứng khái niệm, không phải thiết kế sản xuất. Mục tiêu là để xác thực luồng bộ nhớ cơ bản (retrieve → inject → extract → store) với độ chính xác chấp nhận được. Cứng rắn sản xuất (xử lý lỗi, mở rộng, giám sát) nằm ngoài phạm vi.

### Khả năng cơ bản

| Khả năng | Mô tả |
|------------|-------------|
| **Memory Retrieval** | Tìm kiếm dựa trên embedding với lọc trước đơn giản |
| **Memory Saving** | Trích xuất sự kiện và quy trình dựa trên LLM |
| **Cross-Session Persistence** | Bộ nhớ được lưu trữ trong Milvus (tồn tại qua khởi động lại; không kiểm tra backup/HA sản xuất) |
| **User Isolation** | Bộ nhớ giới hạn theo user_id (xem ghi chú dưới đây) |

> **⚠️ User Isolation - Milvus Performance Note:**
>
> | Tiếp cận | POC | Sản xuất (10K+ người dùng) |
> |----------|-----|-------------------------|
> | **Simple filter** | ✅ Lọc theo `user_id` sau tìm kiếm | ❌ Suy giảm: tìm kiếm tất cả người dùng, sau đó lọc |
> | **Partition Key** | ❌ Quá mức | ✅ Tách biệt vật lý, O(log N) cho mỗi người dùng |
> | **Scalar Index** | ❌ Quá mức | ✅ Chỉ mục trên `user_id` để lọc nhanh |
>
> **POC:** Sử dụng lọc metadaten đơn giản (đủ cho thử nghiệm).
> **Sản xuất:** Cấu hình `user_id` làm Partition Key hoặc Scalar Indexed Field trong lược đồ Milvus.

### Các nguyên tắc thiết kế chính

1. **Simple pre-filter** quyết định nếu truy vấn sẽ tìm kiếm bộ nhớ
2. **Context window** từ lịch sử để làm rõ truy vấn
3. **LLM extracts facts** và phân loại loại khi lưu
4. **Threshold-based filtering** trên kết quả tìm kiếm

### Giả định rõ ràng (POC)

| Giả định | Hàm ý | Rủi ro nếu sai |
|------------|-------------|---------------|
| Trích xuất LLM là hợp lý chính xác | Một số sự kiện không chính xác có thể được lưu trữ | Ô nhiễm bộ nhớ (có thể điều chỉnh qua Forget API) |
| Ngưỡng tương tự 0.6 là điểm bắt đầu | Có thể cần điều chỉnh (bỏ lỡ liên quan hoặc bao gồm không liên quan) | Có thể điều chỉnh dựa trên nhật ký chất lượng truy xuất |
| Milvus khả dụng và được cấu hình | Tính năng bị vô hiệu hóa nếu bị lỗi | Suy giảm duyên vô (không gặp sự cố) |
| Embedding model tạo vectơ 384 chiều | Phải khớp với lược đồ Milvus | Lỗi khởi động (có thể phát hiện được) |
| Lịch sử có sẵn thông qua chuỗi Response API | Bắt buộc cho bối cảnh | Bỏ qua bộ nhớ nếu không có sẵn |

---

## Mục lục

1. [Tuyên bố vấn đề](#1-tuyên-bố-vấn-đề)
2. [Tổng quan về kiến trúc](#2-tổng-quan-về-kiến-trúc)
3. [Các loại bộ nhớ](#3-các-loại-bộ-nhớ)
4. [Tích hợp quy trình](#4-tích-hợp-quy-trình)
5. [Truy xuất bộ nhớ](#5-truy-xuất-bộ-nhớ)
6. [Lưu bộ nhớ](#6-lưu-bộ-nhớ)
7. [Hoạt động bộ nhớ](#7-hoạt-động-bộ-nhớ)
8. [Cấu trúc dữ liệu](#8-cấu-trúc-dữ-liệu)
9. [Mở rộng API](#9-mở-rộng-api)
10. [Cấu hình](#10-cấu-hình)
11. [Chế độ lỗi và dự phòng](#11-chế-độ-lỗi-và-dự-phòng-poc)
12. [Tiêu chí thành công](#12-tiêu-chí-thành-công-poc)
13. [Kế hoạch thực hiện](#13-kế-hoạch-thực-hiện)
14. [Cải tiến tương lai](#14-cải-tiến-tương-lai)

---

## 1. Tuyên bố vấn đề

### Trạng thái hiện tại

Response API cung cấp chuỗi hội thoại thông qua `previous_response_id`, nhưng kiến thức bị mất qua các phiên làm việc:

```
Phiên A (ngày 15 tháng 3):
  Người dùng: "Ngân sách của tôi cho chuyến đi Hawaii là $10,000"
  → Lưu trong chuỗi phiên

Phiên B (ngày 20 tháng 3) - PHIÊN MỚI:
  Người dùng: "Ngân sách của tôi cho chuyến đi là bao nhiêu?"
  → Không có previous_response_id → Kiến thức bị MẤT ❌
```

### Trạng thái mong muốn

Với Agentic Memory:

```
Phiên A (ngày 15 tháng 3):
  Người dùng: "Ngân sách của tôi cho chuyến đi Hawaii là $10,000"
  → Trích xuất và lưu vào Milvus

Phiên B (ngày 20 tháng 3) - PHIÊN MỚI:
  Người dùng: "Ngân sách của tôi cho chuyến đi là bao nhiêu?"
  → Pre-filter: liên quan đến bộ nhớ ✓
  → Tìm kiếm Milvus → Tìm thấy: "ngân sách cho Hawaii là $10K"
  → Tiêm vào bối cảnh LLM
  → Trợ lý: "Ngân sách của bạn cho chuyến đi Hawaii là $10,000!" ✅
```

---

## 2. Tổng quan về kiến trúc

```
┌─────────────────────────────────────────────────────────────────────────┐
│                      AGENTIC MEMORY ARCHITECTURE                        │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│                         ExtProc Pipeline                                │
│  ┌──────────────────────────────────────────────────────────────────┐   │
│  │                                                                  │   │
│  │  Request → Fact? → Tool? → Security → Cache → MEMORY → LLM       │   │
│  │              │       │                          ↑↓               │   │
│  │              └───────┴──── signals used ────────┘                │   │
│  │                                                                  │   │
│  │  Response ← [extract & store] ←─────────────────┘                │   │
│  │                                                                  │   │
│  └──────────────────────────────────────────────────────────────────┘   │
│                                          │                              │
│                    ┌─────────────────────┴─────────────────────┐        │
│                    │                                           │        │
│          ┌─────────▼─────────┐                    ┌────────────▼───┐    │
│          │ Memory Retrieval  │                    │ Memory Saving  │    │
│          │  (request phase)  │                    │(response phase)│    │
│          ├───────────────────┤                    ├────────────────┤    │
│          │ 1. Check signals  │                    │ 1. LLM extract │    │
│          │    (Fact? Tool?)  │                    │ 2. Classify    │    │
│          │ 2. Build context  │                    │ 3. Deduplicate │    │
│          │ 3. Milvus search  │                    │ 4. Store       │    │
│          │ 4. Inject to LLM  │                    │                │    │
│          └─────────┬─────────┘                    └────────┬───────┘    │
│                    │                                       │            │
│                    │         ┌──────────────┐              │            │
│                    └────────►│    Milvus    │◄─────────────┘            │
│                              └──────────────┘                           │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### Trách nhiệm thành phần

| Thành phần | Trách nhiệm | Vị trí |
|-----------|---------------|----------|
| **Memory Filter** | Quyết định + tìm kiếm + tiêm | `pkg/extproc/req_filter_memory.go` |
| **Memory Extractor** | Trích xuất sự kiện dựa trên LLM | `pkg/memory/extractor.go` (mới) |
| **Memory Store** | Giao diện lưu trữ | `pkg/memory/store.go` |
| **Milvus Store** | Backend cơ sở dữ liệu vectơ | `pkg/memory/milvus_store.go` |
| **Existing Classifiers** | Tín hiệu Fact/Tool (được tái sử dụng) | `pkg/extproc/processor_req_body.go` |

### Kiến trúc lưu trữ

[Issue #808](https://github.com/vllm-project/semantic-router/issues/808) đề xuất kiến trúc lưu trữ đa tầng. Chúng tôi thực hiện điều này từng bước:

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    STORAGE ARCHITECTURE (Phased)                        │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  PHASE 1 (MVP)                                                  │    │
│  │  ┌─────────────────────────────────────────────────────────┐    │    │
│  │  │  Milvus (Vector Index)                                  │    │    │
│  │  │  • Tìm kiếm ngữ nghĩa qua bộ nhớ                        │    │    │
│  │  │  • Lưu trữ embedding                                    │    │    │
│  │  │  • Nội dung + metadaten                                 │    │    │
│  │  └─────────────────────────────────────────────────────────┘    │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  PHASE 2 (Performance)                                          │    │
│  │  ┌─────────────────────────────────────────────────────────┐    │    │
│  │  │  Redis (Hot Cache)                                      │    │    │
│  │  │  • Tra cứu metadaten nhanh                              │    │    │
│  │  │  • Bộ nhớ được truy cập gần đây                        │    │    │
│  │  │  • Hỗ trợ TTL/hết hạn                                   │    │    │
│  │  └─────────────────────────────────────────────────────────┘    │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  PHASE 3+ (If Needed)                                           │    │
│  │  ┌───────────────────────┐  ┌───────────────────────┐           │    │
│  │  │  Graph Store (Neo4j)  │  │  Time-Series Index    │           │    │
│  │  │  • Liên kết bộ nhớ    │  │  • Truy vấn thời gian │           │    │
│  │  │  • Mối liên hệ        │  │  • Điểm số suy giảm   │           │    │
│  │  └───────────────────────┘  └───────────────────────┘           │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

| Lớp | Mục đích | Khi cần | Trạng thái |
|-------|---------|-------------|--------|
| **Milvus** | Tìm kiếm vectơ ngữ nghĩa | Chức năng cốt lõi | ✅ MVP |
| **Redis** | Hot cache, truy cập nhanh, TTL | Tối ưu hóa hiệu suất | 🔶 Phase 2 |
| **Graph (Neo4j)** | Mối liên hệ bộ nhớ | Truy vấn lập luận nhiều hop | ⚪ Nếu cần |
| **Time-Series** | Truy vấn thời gian, suy giảm | Điểm số tầm quan trọng theo thời gian | ⚪ Nếu cần |

> **Quyết định thiết kế:** Chúng tôi bắt đầu với Milvus chỉ. Các lớp bổ sung được thêm dựa trên nhu cầu đã chứng minh, không phải phỏng đoán. Giao diện `Store` trừu tượng hóa lưu trữ, cho phép các backend được thêm mà không thay đổi logic truy xuất/lưu.

---

## 3. Các loại bộ nhớ

| Loại | Mục đích | Ví dụ | Trạng thái |
|------|---------|---------|--------|
| **Semantic** | Sự kiện, sở thích, kiến thức | "Ngân sách của người dùng cho Hawaii là $10,000" | ✅ MVP |
| **Procedural** | Hướng dẫn, các bước, quy trình | "Để triển khai payment-service: chạy npm build, sau đó docker push" | ✅ MVP |
| **Episodic** | Tóm tắt phiên, sự kiện quá khứ | "Vào ngày 29 tháng 12 năm 2024, người dùng lên kế hoạch kỳ nghỉ Hawaii với ngân sách $10K" | ⚠️ MVP (hạn chế) |
| **Reflective** | Tự phân tích, bài học học được | "Phản hồi ngân sách trước đó không đầy đủ - người dùng thích phân tích chi tiết" | 🔮 Tương lai |

> **⚠️ Episodic Memory (MVP Limitation):** Phát hiện kết thúc phiên không được thực hiện. Bộ nhớ tập sự chỉ được tạo khi trích xuất LLM rõ ràng tạo ra thứ gì đó có kiểu tóm tắt. Các kích hoạt kết thúc phiên đáng tin cậy bị hoãn cho Phase 2.
>
> **🔮 Reflective Memory:** Tự phân tích và bài học học được. Nằm ngoài phạm vi POC này. Xem [Appendix A](#appendix-a-reflective-memory).

### Vùng vectơ bộ nhớ

Bộ nhớ cụm theo **nội dung/chủ đề**, không theo loại. Loại là metadaten:

```
┌────────────────────────────────────────────────────────────────────────┐
│                      MEMORY VECTOR SPACE                               │
│                                                                        │
│     ┌─────────────────┐                    ┌─────────────────┐         │
│     │  BUDGET/MONEY   │                    │   DEPLOYMENT    │         │
│     │    CLUSTER      │                    │    CLUSTER      │         │
│     │                 │                    │                 │         │
│     │ ● budget=$10K   │                    │ ● npm build     │         │
│     │   (semantic)    │                    │   (procedural)  │         │
│     │ ● cost=$5K      │                    │ ● docker push   │         │
│     │   (semantic)    │                    │   (procedural)  │         │
│     └─────────────────┘                    └─────────────────┘         │
│                                                                        │
│  ● = bộ nhớ với loại là metadaten                                      │
│  Truy vấn khớp nội dung → loại đến từ bộ nhớ phù hợp                  │
│                                                                        │
└────────────────────────────────────────────────────────────────────────┘
```

### Response API so với Agentic Memory: Khi bộ nhớ thêm giá trị?

**Phân biệt quan trọng:** Response API đã gửi toàn bộ lịch sử hội thoại đến LLM khi có `previous_response_id`. Giá trị của Agentic Memory là để **cross-session** context.

```
┌─────────────────────────────────────────────────────────────────────────┐
│           RESPONSE API so với AGENTIC MEMORY: CONTEXT SOURCES            │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  CÙNG PHIÊN (có previous_response_id):                                  │
│  ─────────────────────────────────────                                  │
│    Response API cung cấp:                                               │
│      └── Chuỗi hội thoại đầy đủ (tất cả lượt) → gửi đến LLM            │
│                                                                         │
│    Agentic Memory:                                                      │
│      └── VẪN CÓ GIÁ TRỊ - phiên hiện tại có thể không có câu trả lời   │
│      └── Ví dụ: 100 lượt quy hoạch kỳ nghỉ, nhưng ngân sách không bao giờ nói │
│      └── Vài ngày trước: "Tôi có $10K tiền dự phòng, đó có đủ cho một tuần │
│          ở Thái Lan?" → LLM trích xuất: "Người dùng có ngân sách $10K cho chuyến đi" │
│      └── Bây giờ: "Ngân sách của tôi là bao nhiêu?" → trả lời trong bộ nhớ, không trong chuỗi này │
│                                                                         │
│  PHIÊN MỚI (không có previous_response_id):                             │
│  ──────────────────────────────────────                                 │
│    Response API cung cấp:                                               │
│      └── Không gì cả (không có chuỗi để theo)                          │
│                                                                         │
│    Agentic Memory:                                                      │
│      └── THÊM GIÁ TRỊ - truy xuất bối cảnh cross-session               │
│      └── "Ngân sách Hawaii của tôi là bao nhiêu?" → tìm sự kiện từ phiên tháng 3 │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

> **Quyết định thiết kế:** Truy xuất bộ nhớ thêm giá trị trong **cả hai** kịch bản — phiên mới (không có chuỗi) và phiên hiện có (truy vấn có thể tham chiếu phiên khác). Chúng tôi luôn tìm kiếm khi pre-filter vượt qua.
>
> **Dư thừa đã biết:** Khi câu trả lời CÓ trong chuỗi hiện tại, chúng tôi vẫn tìm kiếm bộ nhớ (~10-30ms bị lãng phí). Chúng tôi không thể phát hiện rẻ "câu trả lời đã có trong lịch sử?" mà không hiểu truy vấn ngữ nghĩa. Đối với POC, chúng tôi chấp nhận chi phí này.
>
> **Giải pháp Phase 2:** [Context Compression](#context-compression-high-priority) giải quyết điều này đúng cách — thay vì Response API gửi toàn bộ lịch sử, chúng tôi gửi tóm tắt nén + lượt gần đây + bộ nhớ liên quan. Sự kiện được trích xuất trong quá trình nén, loại bỏ hoàn toàn dư thừa.

---

## 4. Tích hợp quy trình

### Quy trình hiện tại (main branch)

```
1. Response API Translation
2. Parse Request
3. Fact-Check Classification
4. Tool Detection
5. Decision & Model Selection
6. Security Checks
7. PII Detection
8. Semantic Cache Check
9. Model Routing → LLM
```

### Quy trình nâng cao với Agentic Memory

```
REQUEST PHASE:
─────────────
1.  Response API Translation
2.  Parse Request
3.  Fact-Check Classification        ──┐
4.  Tool Detection                     ├── Existing signals
5.  Decision & Model Selection       ──┘
6.  Security Checks
7.  PII Detection
8.  Semantic Cache Check ───► if HIT → return cached
9.  🆕 Memory Decision:
    └── if (NOT Fact) AND (NOT Tool) AND (NOT Greeting) → continue
    └── else → skip to step 12
10. 🆕 Build context + rewrite query          [~1-5ms]
11. 🆕 Search Milvus, inject memories         [~10-30ms]
12. Model Routing → LLM

RESPONSE PHASE:
──────────────
13. Parse LLM Response
14. Cache Update
15. 🆕 Memory Extraction (async goroutine, if auto_store enabled)
    └── Runs in background, does NOT add latency to response
16. Response API Translation
17. Return to Client
```

> **Chi tiết bước 10:** Các chiến lược viết lại truy vấn (context prepend, LLM rewrite, HyDE) được giải thích trong [Appendix C](#appendix-c-query-rewriting-for-memory-search).

---

## 5. Truy xuất bộ nhớ

### Luồng

```
┌─────────────────────────────────────────────────────────────────────────┐
│                      MEMORY RETRIEVAL FLOW                              │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  1. MEMORY DECISION (reuse existing pipeline signals)                   │
│     ──────────────────────────────────────────────────────              │
│                                                                         │
│     Pipeline already classified:                                        │
│     ├── ctx.IsFact       (Fact-Check classifier)                        │
│     ├── ctx.RequiresTool (Tool Detection)                               │
│     └── isGreeting(query) (simple pattern)                              │
│                                                                         │
│     Decision:                                                           │
│     ├── Fact query?     → SKIP (general knowledge)                      │
│     ├── Tool query?     → SKIP (tool provides answer)                   │
│     ├── Greeting?       → SKIP (no context needed)                      │
│     └── Otherwise       → SEARCH MEMORY                                 │
│                                                                         │
│  2. BUILD CONTEXT + REWRITE QUERY                                       │
│     ─────────────────────────────                                       │
│     History: ["Planning vacation", "Hawaii sounds nice"]                │
│     Query: "How much?"                                                  │
│                                                                         │
│     Option A (MVP): Context prepend                                     │
│     → "How much? Hawaii vacation planning"                              │
│                                                                         │
│     Option B (v1): LLM rewrite                                          │
│     → "What is the budget for the Hawaii vacation?"                     │
│                                                                         │
│  3. MILVUS SEARCH                                                       │
│     ─────────────                                                       │
│     Embed context → Search with user_id filter → Top-k results          │
│                                                                         │
│  4. THRESHOLD FILTER                                                    │
│     ────────────────                                                    │
│     Keep only results with similarity > 0.6                             │
│     ⚠️ Threshold is configurable; 0.6 is starting value, tune via logs  │
│                                                                         │
│  5. INJECT INTO LLM CONTEXT                                             │
│     ────────────────────────                                            │
│     Add as system message: "User's relevant context: ..."               │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### Thực hiện

#### MemoryFilter Struct

```go
// pkg/extproc/req_filter_memory.go

type MemoryFilter struct {
    store memory.Store  // Interface - can be MilvusStore or InMemoryStore
}

func NewMemoryFilter(store memory.Store) *MemoryFilter {
    return &MemoryFilter{store: store}
}
```

> **Ghi chú:** `store` là giao diện `Store` (Phần 8), không phải thực hiện cụ thể.
> Tại thời gian chạy, đây thường là `MilvusStore` cho sản xuất hoặc `InMemoryStore` để thử nghiệm.

#### Quyết định bộ nhớ (Tái sử dụng quy trình hiện có)

> **⚠️ Known Limitation:** Trình phân loại `IsFact` được thiết kế cho kiểm tra sự kiện kiến thức chung (ví dụ: "Thủ đô của Pháp là gì?"). Nó có thể phân loại sai câu hỏi sự kiện cá nhân ("Ngân sách của tôi là bao nhiêu?") là truy vấn sự kiện, khiến bộ nhớ bị bỏ qua.
>
> **POC Mitigation:** Chúng tôi thêm kiểm tra chỉ báo cá nhân. Nếu truy vấn chứa đại từ cá nhân ("my", "I", "me"), chúng tôi ghi đè `IsFact` và tìm kiếm bộ nhớ vẫn vậy.
>
> **Tương lai:** Đào tạo lại hoặc tăng cường trình phân loại kiểm tra sự kiện để phân biệt sự kiện chung vs. cá nhân.

```go
// pkg/extproc/req_filter_memory.go

// shouldSearchMemory decides if query should trigger memory search
// Reuses existing pipeline classification signals with personal-fact override
func shouldSearchMemory(ctx *RequestContext, query string) bool {
    // Check for personal indicators (overrides IsFact for personal questions)
    hasPersonalIndicator := containsPersonalPronoun(query)

    // 1. Fact query → skip UNLESS it contains personal pronouns
    if ctx.IsFact && !hasPersonalIndicator {
        logging.Debug("Memory: Skipping - general fact query")
        return false
    }

    // 2. Tool required → skip (tool provides answer)
    if ctx.RequiresTool {
        logging.Debug("Memory: Skipping - tool query")
        return false
    }

    // 3. Greeting/social → skip (no context needed)
    if isGreeting(query) {
        logging.Debug("Memory: Skipping - greeting")
        return false
    }

    // 4. Default: search memory (conservative - don't miss context)
    return true
}

func containsPersonalPronoun(query string) bool {
    // Simple check for personal context indicators
    personalPatterns := regexp.MustCompile(`(?i)\b(my|i|me|mine|i'm|i've|i'll)\b`)
    return personalPatterns.MatchString(query)
}

func isGreeting(query string) bool {
    // Match greetings that are ONLY greetings, not "Hi, what's my budget?"
    lower := strings.ToLower(strings.TrimSpace(query))

    // Short greetings only (< 20 chars and matches pattern)
    if len(lower) > 20 {
        return false
    }

    greetings := []string{
        `^(hi|hello|hey|howdy)[\s\!\.\,]*$`,
        `^(hi|hello|hey)[\s\,]*(there)?[\s\!\.\,]*$`,
        `^(thanks|thank you|thx)[\s\!\.\,]*$`,
        `^(bye|goodbye|see you)[\s\!\.\,]*$`,
        `^(ok|okay|sure|yes|no)[\s\!\.\,]*$`,
    }
    for _, p := range greetings {
        if regexp.MustCompile(p).MatchString(lower) {
            return true
        }
    }
    return false
}
```

#### Xây dựng bối cảnh

```go
// buildSearchQuery builds an effective search query from history + current query
// MVP: context prepend, v1: LLM rewrite for vague queries
func buildSearchQuery(history []Message, query string) string {
    // If query is self-contained, use as-is
    if isSelfContained(query) {
        return query
    }

    // MVP: Simple context prepend
    context := summarizeHistory(history)
    return query + " " + context

    // v1 (future): LLM rewrite for vague queries
    // if isVague(query) {
    //     return rewriteWithLLM(history, query)
    // }
}

func isSelfContained(query string) bool {
    // Self-contained: "What's my budget for the Hawaii trip?"
    // NOT self-contained: "How much?", "And that one?", "What about it?"

    vaguePatterns := []string{`^how much\??$`, `^what about`, `^and that`, `^this one`}
    for _, p := range vaguePatterns {
        if regexp.MustCompile(`(?i)`+p).MatchString(query) {
            return false
        }
    }
    return len(query) > 20 // Short queries are often vague
}

func summarizeHistory(history []Message) string {
    // Extract key terms from last 3 user messages
    var terms []string
    count := 0
    for i := len(history) - 1; i >= 0 && count < 3; i-- {
        if history[i].Role == "user" {
            terms = append(terms, extractKeyTerms(history[i].Content))
            count++
        }
    }
    return strings.Join(terms, " ")
}

// v1: LLM-based query rewriting (future enhancement)
func rewriteWithLLM(history []Message, query string) string {
    prompt := fmt.Sprintf(`Conversation context: %s

Rewrite this vague query to be self-contained: "%s"
Return ONLY the rewritten query.`, summarizeHistory(history), query)

    // Call LLM endpoint
    resp, _ := http.Post(llmEndpoint+"/v1/chat/completions", ...)
    return parseResponse(resp)
    // "how much?" → "What is the budget for the Hawaii vacation?"
}
```

#### Truy xuất đầy đủ

```go
// pkg/extproc/req_filter_memory.go

func (f *MemoryFilter) RetrieveMemories(
    ctx context.Context,
    query string,
    userID string,
    history []Message,
) ([]*memory.RetrieveResult, error) {

    // 1. Memory decision (skip if fact/tool/greeting)
    if !shouldSearchMemory(ctx, query) {
        logging.Debug("Memory: Skipping - not memory-relevant")
        return nil, nil
    }

    // 2. Build search query (context prepend or LLM rewrite)
    searchQuery := buildSearchQuery(history, query)

    // 3. Search Milvus
    results, err := f.store.Retrieve(ctx, memory.RetrieveOptions{
        Query:     searchQuery,
        UserID:    userID,
        Limit:     5,
        Threshold: 0.6,
    })
    if err != nil {
        return nil, err
    }

    logging.Infof("Memory: Retrieved %d memories", len(results))
    return results, nil
}

// InjectMemories adds memories to the LLM request
func (f *MemoryFilter) InjectMemories(
    requestBody []byte,
    memories []*memory.RetrieveResult,
) ([]byte, error) {
    if len(memories) == 0 {
        return requestBody, nil
    }

    // Format memories as context
    var sb strings.Builder
    sb.WriteString("## User's Relevant Context\n\n")
    for _, mem := range memories {
        sb.WriteString(fmt.Sprintf("- %s\n", mem.Memory.Content))
    }

    // Add as system message
    return injectSystemMessage(requestBody, sb.String())
}
```

---

## 6. Lưu bộ nhớ

### Kích hoạt

Trích xuất bộ nhớ được kích hoạt bởi ba sự kiện:

| Kích hoạt | Mô tả | Trạng thái |
|---------|-------------|--------|
| **Every N turns** | Trích xuất sau mỗi 10 lượt | ✅ MVP |
| **End of session** | Tạo tóm tắt cuối phiên khi phiên kết thúc | 🔮 Tương lai |
| **Context drift** | Trích xuất khi chủ đề thay đổi đáng kể | 🔮 Tương lai |

> **Ghi chú:** Phát hiện kết thúc phiên và phát hiện độ lệch bối cảnh yêu cầu thực hiện bổ sung.
> Đối với MVP, chúng tôi chỉ dựa vào kích hoạt "every N turns".

### Luồng

```
┌─────────────────────────────────────────────────────────────────────────┐
│                       MEMORY SAVING FLOW                                │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  TRIGGERS:                                                              │
│  ─────────                                                              │
│  ├── Every N turns (e.g., 10)      ← MVP                                │
│  ├── End of session                ← Future (needs detection)           │
│  └── Context drift detected        ← Future (needs detection)           │
│                                                                         │
│  Runs: Async (background) - no user latency                             │
│                                                                         │
│  1. GET BATCH                                                           │
│     ─────────                                                           │
│     Get last 10-15 turns from session                                   │
│                                                                         │
│  2. LLM EXTRACTION                                                      │
│     ──────────────                                                      │
│     Prompt: "Extract important facts. Include context.                  │
│              Return JSON: [{type, content}, ...]"                       │
│                                                                         │
│     LLM returns:                                                        │
│       [{"type": "semantic", "content": "budget for Hawaii is $10K"}]    │
│                                                                         │
│  3. DEDUPLICATION                                                       │
│     ─────────────                                                       │
│     For each extracted fact:                                            │
│     - Embed content                                                     │
│     - Search existing memories (same user, same type)                   │
│     - If similarity > 0.9: UPDATE existing (merge/replace)              │
│     - If similarity 0.7-0.9: CREATE new (gray zone, conservative)       │
│     - If similarity < 0.7: CREATE new                                   │
│                                                                         │
│     Example:                                                            │
│       Existing: "User's budget for Hawaii is $10,000"                   │
│       New:      "User's budget is now $15,000"                          │
│       → Similarity ~0.92 → UPDATE existing with new value               │
│                                                                         │
│  4. STORE IN MILVUS                                                     │
│     ───────────────                                                     │
│     Memory { id, type, content, embedding, user_id, created_at }        │
│                                                                         │
│  5. SESSION END (future): Create episodic summary                       │
│     ─────────────────────────────────────────────────────────────────   │
│     "On Dec 29, user planned Hawaii vacation with $10K budget"          │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

> **Ghi chú về `user_id`:** Khi chúng tôi đề cập đến `user_id` để sử dụng bộ nhớ, chúng tôi có nghĩa là **người dùng đã đăng nhập** (danh tính người dùng được xác thực), không phải người dùng phiên chúng tôi hiện có. Đây là thứ cần được cấu hình trong tác nhân định tuyến ngữ nghĩa chính nó.

### Thực hiện

```go
// pkg/memory/extractor.go

type MemoryExtractor struct {
    store       memory.Store  // Interface - can be MilvusStore or InMemoryStore
    llmEndpoint string        // LLM endpoint for fact extraction
    batchSize   int           // Extract every N turns (default: 10)
    turnCounts  map[string]int
    mu          sync.Mutex
}

// ProcessResponse extracts and stores memories (runs async)
//
// Triggers (MVP: only first one implemented):
//   - Every N turns (e.g., 10)       ← MVP
//   - End of session                 ← Future: needs session end detection
//   - Context drift detected         ← Future: needs drift detection
//
func (e *MemoryExtractor) ProcessResponse(
    ctx context.Context,
    sessionID string,
    userID string,
    history []Message,
) error {
    e.mu.Lock()
    e.turnCounts[sessionID]++
    turnCount := e.turnCounts[sessionID]
    e.mu.Unlock()

    // MVP: Only extract every N turns
    // Future: Also trigger on session end or context drift
    if turnCount % e.batchSize != 0 {
        return nil
    }

    // Get recent batch
    batchStart := max(0, len(history) - e.batchSize - 5)
    batch := history[batchStart:]

    // LLM extraction
    extracted, err := e.extractWithLLM(batch)
    if err != nil {
        return err
    }

    // Store with deduplication
    for _, fact := range extracted {
        existing, similarity := e.findSimilar(ctx, userID, fact.Content, fact.Type)

        if similarity > 0.9 && existing != nil {
            // Very similar → UPDATE existing memory
            existing.Content = fact.Content  // Use newer content
            existing.UpdatedAt = time.Now()
            if err := e.store.Update(ctx, existing.ID, existing); err != nil {
                logging.Warnf("Failed to update memory: %v", err)
            }
            continue
        }

        // similarity < 0.9 → CREATE new memory
        mem := &Memory{
            ID:        generateID("mem"),
            Type:      fact.Type,
            Content:   fact.Content,
            UserID:    userID,
            Source:    "conversation",
            CreatedAt: time.Now(),
        }

        if err := e.store.Store(ctx, mem); err != nil {
            logging.Warnf("Failed to store memory: %v", err)
        }
    }

    return nil
}

// findSimilar searches for existing similar memories
func (e *MemoryExtractor) findSimilar(
    ctx context.Context,
    userID string,
    content string,
    memType MemoryType,
) (*Memory, float32) {
    results, err := e.store.Retrieve(ctx, memory.RetrieveOptions{
        Query:     content,
        UserID:    userID,
        Types:     []MemoryType{memType},
        Limit:     1,
        Threshold: 0.7,  // Only consider reasonably similar
    })
    if err != nil || len(results) == 0 {
        return nil, 0
    }
    return results[0].Memory, results[0].Score
}

// extractWithLLM uses LLM to extract facts
//
// ⚠️ POC Limitation: LLM extraction is best-effort. Failures are logged but do not
// block the response. Incorrect extractions may occur.
//
// Future: Self-correcting memory (see Section 14 - Future Enhancements):
//   - Track memory usage (access_count, last_accessed)
//   - Score memories based on usage + age + retrieval feedback
//   - Periodically prune low-score, unused memories
//   - Detect contradictions → auto-merge or flag for resolution
//
func (e *MemoryExtractor) extractWithLLM(messages []Message) ([]ExtractedFact, error) {
    prompt := `Extract important information from these messages.

IMPORTANT: Include CONTEXT for each fact.

For each piece of information:
- Type: "semantic" (facts, preferences) or "procedural" (instructions, how-to)
- Content: The fact WITH its context

BAD:  {"type": "semantic", "content": "budget is $10,000"}
GOOD: {"type": "semantic", "content": "budget for Hawaii vacation is $10,000"}

Messages:
` + formatMessages(messages) + `

Return JSON array (empty if nothing to remember):
[{"type": "semantic|procedural", "content": "fact with context"}]`

    // Call LLM with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    reqBody := map[string]interface{}{
        "model": "qwen3",
        "messages": []map[string]string{
            {"role": "user", "content": prompt},
        },
    }
    jsonBody, _ := json.Marshal(reqBody)

    req, _ := http.NewRequestWithContext(ctx, "POST",
        e.llmEndpoint+"/v1/chat/completions",
        bytes.NewReader(jsonBody))
    req.Header.Set("Content-Type", "application/json")

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        logging.Warnf("Memory extraction LLM call failed: %v", err)
        return nil, err  // Caller handles gracefully
    }
    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        logging.Warnf("Memory extraction LLM returned %d", resp.StatusCode)
        return nil, fmt.Errorf("LLM returned %d", resp.StatusCode)
    }

    facts, err := parseExtractedFacts(resp.Body)
    if err != nil {
        // JSON parse error - LLM returned malformed output
        logging.Warnf("Memory extraction parse failed: %v", err)
        return nil, err  // Skip this batch, don't store garbage
    }

    return facts, nil
}
```

---

## 7. Hoạt động bộ nhớ

Tất cả các hoạt động có thể được thực hiện trên bộ nhớ. Được triển khai trong giao diện `Store` (xem [Phần 8](#8-cấu-trúc-dữ-liệu)).

| Hoạt động | Mô tả | Kích hoạt | Phương thức giao diện | Trạng thái |
|-----------|-------------|---------|------------------|--------|
| **Store** | Lưu bộ nhớ mới vào Milvus | Tự động (trích xuất LLM) hoặc API rõ ràng | `Store()` | ✅ MVP |
| **Retrieve** | Tìm kiếm ngữ nghĩa cho bộ nhớ liên quan | Tự động (trên truy vấn) | `Retrieve()` | ✅ MVP |
| **Update** | Sửa đổi nội dung bộ nhớ hiện có | Khử trùng hoặc API rõ ràng | `Update()` | ✅ MVP |
| **Forget** | Xóa bộ nhớ cụ thể theo ID | Lệnh gọi API rõ ràng | `Forget()` | ✅ MVP |
| **ForgetByScope** | Xóa tất cả bộ nhớ cho người dùng/dự án | Lệnh gọi API rõ ràng | `ForgetByScope()` | ✅ MVP |
| **Consolidate** | Hợp nhất các bộ nhớ liên quan thành tóm tắt | Được lên lịch / trên ngưỡng | `Consolidate()` | 🔮 Tương lai |
| **Reflect** | Tạo thông tin chi tiết từ mô hình bộ nhớ | Tác nhân bắt đầu | `Reflect()` | 🔮 Tương lai |

### Hoạt động quên

```go
// Forget single memory
DELETE /v1/memory/{memory_id}

// Forget all memories for a user
DELETE /v1/memory?user_id=user_123

// Forget all memories for a project
DELETE /v1/memory?user_id=user_123&project_id=project_abc
```

**Trường hợp sử dụng:**

- Người dùng yêu cầu "quên những gì tôi nói với bạn về X"
- Tuân thủ GDPR/quyền riêng tư (quyền bị lãng quên)
- Xóa thông tin lỗi thời

### Tương lai: Consolidate

Hợp nhất nhiều bộ nhớ liên quan thành một tóm tắt duy nhất:

```
Before:
  - "Budget for Hawaii is $10,000"
  - "Added $2,000 to Hawaii budget"
  - "Final Hawaii budget is $12,000"

After consolidation:
  - "Hawaii trip budget: $12,000 (updated from initial $10,000)"
```

**Tùy chọn kích hoạt:**

- Khi số lượng bộ nhớ vượt quá ngưỡng
- Công việc nền được lên lịch
- Khi kết thúc phiên

### Tương lai: Reflect

Tạo thông tin chi tiết bằng cách phân tích mô hình bộ nhớ:

```
Input: All memories for user_123 about "deployment"

Output (Insight):
  - "User frequently deploys payment-service (12 times)"
  - "Common issue: port conflicts"
  - "Preferred approach: docker-compose"
```

**Trường hợp sử dụng:** Tác nhân có thể chủ động cung cấp trợ giúp dựa trên mô hình.

---

## 8. Cấu trúc dữ liệu

### Memory

```go
// pkg/memory/types.go

type MemoryType string

const (
    MemoryTypeEpisodic   MemoryType = "episodic"
    MemoryTypeSemantic   MemoryType = "semantic"
    MemoryTypeProcedural MemoryType = "procedural"
)

type Memory struct {
    ID          string         `json:"id"`
    Type        MemoryType     `json:"type"`
    Content     string         `json:"content"`
    Embedding   []float32      `json:"-"`
    UserID      string         `json:"user_id"`
    ProjectID   string         `json:"project_id,omitempty"`
    Source      string         `json:"source,omitempty"`
    CreatedAt   time.Time      `json:"created_at"`
    AccessCount int            `json:"access_count"`
    Importance  float32        `json:"importance"`
}
```

### Store Interface

```go
// pkg/memory/store.go

type Store interface {
    // MVP Operations
    Store(ctx context.Context, memory *Memory) error                         // Save new memory
    Retrieve(ctx context.Context, opts RetrieveOptions) ([]*RetrieveResult, error) // Semantic search
    Get(ctx context.Context, id string) (*Memory, error)                     // Get by ID
    Update(ctx context.Context, id string, memory *Memory) error             // Modify existing
    Forget(ctx context.Context, id string) error                             // Delete by ID
    ForgetByScope(ctx context.Context, scope MemoryScope) error              // Delete by scope

    // Utility
    IsEnabled() bool
    Close() error

    // Future Operations (not yet implemented)
    // Consolidate(ctx context.Context, memoryIDs []string) (*Memory, error)  // Merge memories
    // Reflect(ctx context.Context, scope MemoryScope) ([]*Insight, error)    // Generate insights
}
```

---

## 9. Mở rộng API

### Request (hiện có)

```go
// pkg/responseapi/types.go

type ResponseAPIRequest struct {
    // ... existing fields ...
    MemoryConfig  *MemoryConfig  `json:"memory_config,omitempty"`
    MemoryContext *MemoryContext `json:"memory_context,omitempty"`
}

type MemoryConfig struct {
    Enabled             bool     `json:"enabled"`
    MemoryTypes         []string `json:"memory_types,omitempty"`
    RetrievalLimit      int      `json:"retrieval_limit,omitempty"`
    SimilarityThreshold float32  `json:"similarity_threshold,omitempty"`
    AutoStore           bool     `json:"auto_store,omitempty"`
}

type MemoryContext struct {
    UserID    string `json:"user_id"`
    ProjectID string `json:"project_id,omitempty"`
}
```

### Ví dụ Request

```json
{
    "model": "qwen3",
    "input": "What's my budget for the trip?",
    "previous_response_id": "resp_abc123",
    "memory_config": {
        "enabled": true,
        "auto_store": true
    },
    "memory_context": {
        "user_id": "user_456"
    }
}
```

---

## 10. Cấu hình

```yaml
# config.yaml
memory:
  enabled: true
  auto_store: true  # Enable automatic fact extraction

  milvus:
    address: "milvus:19530"
    collection: "agentic_memory"
    dimension: 384             # Must match embedding model output

  # Embedding model for memory
  embedding:
    model: "all-MiniLM-L6-v2"   # 384-dim, optimized for semantic similarity
    dimension: 384

  # Retrieval settings
  default_retrieval_limit: 5
  default_similarity_threshold: 0.6   # Tunable; start conservative

  # Extraction runs every N conversation turns
  extraction_batch_size: 10

# External models for memory LLM features
# Query rewriting and fact extraction are enabled by adding external_models
external_models:
  - llm_provider: "vllm"
    model_role: "memory_rewrite"      # Enables query rewriting
    llm_endpoint:
      address: "qwen"
      port: 8000
    llm_model_name: "qwen3"
    llm_timeout_seconds: 30
    max_tokens: 100
    temperature: 0.1
  - llm_provider: "vllm"
    model_role: "memory_extraction"   # Enables fact extraction
    llm_endpoint:
      address: "qwen"
      port: 8000
    llm_model_name: "qwen3"
    llm_timeout_seconds: 30
    max_tokens: 500
    temperature: 0.1
```

### Ghi chú cấu hình

| Tham số | Giá trị | Căn cứ |
|-----------|-------|-----------|
| `dimension: 384` | Cố định | Phải khớp với đầu ra all-MiniLM-L6-v2 |
| `default_similarity_threshold: 0.6` | Giá trị bắt đầu | Điều chỉnh dựa trên nhật ký chất lượng truy xuất |
| `extraction_batch_size: 10` | Mặc định | Cân bằng giữa tính mới và chi phí LLM |
| `llm_timeout_seconds: 30` | Mặc định | Ngăn trích xuất khỏi chặn vô thời hạn |

> **Lựa chọn mô hình nhúng:**
>
> | Mô hình | Kích thước | Ưu điểm | Nhược điểm |
> |-------|-----------|------|------|
> | **all-MiniLM-L6-v2** (lựa chọn POC) | 384 | Tương tự ngữ nghĩa tốt hơn, tha thứ với cách diễn đạt, lý tưởng để truy xuất bộ nhớ & khử trùng | Yêu cầu tải mô hình riêng |
> | Qwen3-Embedding-0.6B (hiện có) | 1024 | Đã tải cho bộ nhớ cache ngữ nghĩa, không có bộ nhớ bổ sung | Nhạy cảm hơn với lời diễn đạt chính xác, có thể bỏ lỡ bộ nhớ tương tự |
>
> **Tại sao 384-dim cho Memory?** Kích thước thấp hơn nắm bắt ý nghĩa ngữ nghĩa cấp cao và kém nhạy cảm với các chi tiết cụ thể (số, tên). Điều này có lợi cho:
>
> - **Retrieval**: "Ngân sách của tôi là bao nhiêu?" khớp với "Ngân sách chuyến đi Hawaii là $10K" ngay cả với cách diễn đạt khác nhau
> - **Deduplication**: "ngân sách là $10K" và "ngân sách bây giờ là $15K" được công nhận là cùng một chủ đề (cập nhật giá trị)
> - **Cross-session**: Cách diễn đạt tự nhiên khác nhau giữa các phiên
>
> **Thay thế:** Có thể tái sử dụng Qwen3-Embedding (1024-dim) để tránh tải một mô hình thứ hai. Sự đánh đổi là khớp hơi chặt hơn có thể tăng âm tính giả.

---

## 11. Chế độ lỗi và dự phòng (POC)

Phần này rõ ràng ghi lại cách hệ thống hoạt động khi các thành phần gặp sự cố. Trong phạm vi POC, chúng tôi ưu tiên **suy giảm duyên vô** hơn phục hồi phức tạp.

| Lỗi | Phát hiện | Hành vi | Ghi nhật ký |
|---------|-----------|----------|---------|
| **Milvus unavailable** | Lỗi kết nối trên Store init | Tính năng bộ nhớ bị vô hiệu hóa cho phiên | `ERROR: Milvus unavailable, memory disabled` |
| **Milvus search timeout** | Hết hạn bối cảnh vượt quá | Bỏ qua tiêm bộ nhớ, tiếp tục mà không | `WARN: Memory search timeout, skipping` |
| **Embedding generation fails** | Lỗi từ candle-binding | Bỏ qua bộ nhớ cho yêu cầu này | `WARN: Embedding failed, skipping memory` |
| **LLM extraction fails** | Lỗi HTTP hoặc timeout | Bỏ qua trích xuất, bộ nhớ không được lưu | `WARN: Extraction failed, batch skipped` |
| **LLM returns invalid JSON** | Lỗi phân tích | Bỏ qua trích xuất, bộ nhớ không được lưu | `WARN: Extraction parse failed` |
| **No history available** | `ctx.ConversationHistory` rỗng | Tìm kiếm với chỉ truy vấn (không có context prepend) | `DEBUG: No history, query-only search` |
| **Threshold too high** | 0 kết quả trả về | Không có bộ nhớ được tiêm | `DEBUG: No memories above threshold` |
| **Threshold too low** | Nhiều kết quả không phù hợp | Bối cảnh ồn ào (chấp nhận được cho POC) | `DEBUG: Retrieved N memories` |

### Nguyên tắc suy giảm duyên vô

> **Yêu cầu PHẢI thành công ngay cả khi bộ nhớ gặp sự cố.** Bộ nhớ là một cải tiến, không phải một phụ thuộc. Tất cả các hoạt động bộ nhớ được bao gói trong các trình xử lý lỗi ghi nhật ký và tiếp tục.

```go
// Example: Memory retrieval with fallback
memories, err := memoryFilter.RetrieveMemories(ctx, query, userID, history)
if err != nil {
    logging.Warnf("Memory retrieval failed: %v", err)
    memories = nil  // Continue without memories
}
// Proceed with request (memories may be nil/empty)
```

---

## 12. Tiêu chí thành công (POC)

### Tiêu chí chức năng

| Tiêu chí | Cách xác thực | Điều kiện vượt qua |
|-----------|-----------------|----------------|
| Cross-session retrieval | Lưu sự kiện trong Phiên A, truy vấn trong Phiên B | Sự kiện được truy xuất và tiêm |
| User isolation | Người dùng A lưu sự kiện, Người dùng B truy vấn | Người dùng B KHÔNG thấy sự kiện của Người dùng A |
| Graceful degradation | Dừng Milvus, gửi yêu cầu | Yêu cầu thành công (mà không có bộ nhớ) |
| Extraction runs | Kiểm tra nhật ký sau hội thoại | `Memory: Stored N facts` xuất hiện |

### Tiêu chí chất lượng (Đo lường sau POC)

| Số liệu | Mục tiêu | Cách đo |
|--------|--------|----------------|
| Retrieval relevance | Đa số bộ nhớ được tiêm có liên quan | Xem xét thủ công 50 mẫu |
| Extraction accuracy | Đa số sự kiện được trích xuất là chính xác | Xem xét thủ công 50 mẫu |
| Latency impact | &lt;50ms được thêm vào P50 | So sánh với/mà không bộ nhớ được bật |

> **Phạm vi POC:** Chúng tôi xác thực tiêu chí chức năng chỉ. Số liệu chất lượng được đo lường sau POC để thông báo cho điều chỉnh ngưỡng và cải tiến nhắc trích xuất.

---

## 13. Kế hoạch thực hiện

### Giai đoạn 1: Retrieval

| Nhiệm vụ | Tệp |
|------|-------|
| Memory decision (use existing Fact/Tool signals) | `pkg/extproc/req_filter_memory.go` |
| Context building from history | `pkg/extproc/req_filter_memory.go` |
| Milvus search + threshold filter | `pkg/memory/milvus_store.go` |
| Memory injection into request | `pkg/extproc/req_filter_memory.go` |
| Integrate in request phase | `pkg/extproc/processor_req_body.go` |

### Giai đoạn 2: Saving

| Nhiệm vụ | Tệp |
|------|-------|
| Create MemoryExtractor | `pkg/memory/extractor.go` |
| LLM-based fact extraction | `pkg/memory/extractor.go` |
| Deduplication logic | `pkg/memory/extractor.go` |
| Integrate in response phase (async) | `pkg/extproc/processor_res_body.go` |

### Giai đoạn 3: Testing & Tuning

| Nhiệm vụ | Mô tả |
|------|-------------|
| Unit tests | Memory decision, extraction, retrieval |
| Integration tests | End-to-end flow |
| Threshold tuning | Adjust similarity threshold based on results |

---

## 14. Cải tiến tương lai

### Context Compression (Ưu tiên cao)

**Vấn đề:** Response API hiện đang gửi **toàn bộ** lịch sử hội thoại đến LLM. Đối với phiên 200 lượt, điều này có nghĩa là hàng nghìn token cho mỗi yêu cầu — đắt tiền và có thể gặp giới hạn bối cảnh.

**Giải pháp:** Thay thế các tin nhắn cũ bằng hai kết quả:

| Kết quả | Mục đích | Lưu trữ | Thay thế |
|--------|---------|---------|----------|
| **Facts** | Bộ nhớ dài hạn | Milvus | (Đã có trong Phần 6) |
| **Current state** | Bối cảnh phiên | Redis | Tin nhắn cũ |

> **Insight chính:** "current state" sẽ **được cấu trúc** (không phải tóm tắt văn bản), làm cho nó sẵn sàng cho KG:
>
> ```json
> {"topic": "Hawaii vacation", "budget": "$10K", "decisions": ["fly direct"], "open": ["which hotel?"]}
> ```

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    CONTEXT COMPRESSION FLOW                             │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  BACKGROUND (every 10 turns):                                           │
│    1. Extract facts (reuse Section 6) → save to Milvus                  │
│    2. Build current state (structured JSON) → save to Redis             │
│                                                                         │
│  ON REQUEST (turn N):                                                   │
│    Context = [current state from Redis]   ← replaces old messages       │
│            + [raw last 5 turns]           ← recent context              │
│            + [relevant memories]          ← cross-session (Milvus)      │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

**Thay đổi thực hiện:**

| Tệp | Thay đổi |
|------|--------|
| `pkg/responseapi/translator.go` | Thay thế toàn bộ lịch sử bằng trạng thái hiện tại + gần đây |
| `pkg/responseapi/context_manager.go` | Mới: quản lý trạng thái hiện tại |
| Redis config | Lưu trữ trạng thái hiện tại với TTL |

**LLM nhận cái gì (thay vì toàn bộ lịch sử):**

```
Context sent to LLM:
  1. Current state (structured JSON from Redis)  ~100 tokens
  2. Last 5 raw messages                         ~400 tokens
  3. Relevant memories from Milvus               ~150 tokens
  ─────────────────────────────────────────────────────────────
  Total: ~650 tokens (vs 10K for full history)
```

**Synergy với Agentic Memory:**

- Trích xuất sự kiện (Phần 6) runs trong quá trình nén → lưu vào Milvus
- Trạng thái hiện tại thay thế tin nhắn cũ → giảm token
- Định dạng được cấu trúc → sẵn sàng cho KG

**Lợi ích:**

- Sử dụng token được kiểm soát (chi phí dự đoán)
- Chất lượng bối cảnh tốt hơn (trạng thái được cấu trúc vs. toàn bộ lịch sử)
- **KG-ready**: Trạng thái được cấu trúc ánh xạ trực tiếp đến các nút/cạnh đồ thị
- Quy mô cho các phiên rất dài (1000+ lượt)

---

### Lưu kích hoạt

| Tính năng | Mô tả | Tiếp cận |
|---------|-------------|----------|
| **Session end detection** | Kích hoạt trích xuất khi phiên kết thúc | Timeout / tín hiệu rõ ràng / lệnh gọi API |
| **Context drift detection** | Kích hoạt khi chủ đề thay đổi đáng kể | Tương tự embedding giữa các lượt |

### Lớp lưu trữ

| Tính năng | Mô tả | Ưu tiên |
|---------|-------------|----------|
| **Redis hot cache** | Lớp truy cập nhanh trước Milvus | Cao |
| **TTL & expiration** | Tự động xóa bộ nhớ cũ (Redis native) | Cao |

### Tính năng nâng cao

| Tính năng | Mô tả | Ưu tiên |
|---------|-------------|----------|
| **Self-correcting memory** | Theo dõi sử dụng, điểm bằng truy cập/tuổi, tự động cắt tỉa bộ nhớ điểm thấp | Cao |
| **Contradiction detection** | Phát hiện sự kiện xung đột, tự động hợp nhất hoặc cờ | Cao |
| **Memory type routing** | Tìm kiếm các loại cụ thể (semantic/procedural/episodic) | Vừa |
| **Per-user quotas** | Giới hạn lưu trữ trên mỗi người dùng | Vừa |
| **Graph store** | Mối liên hệ bộ nhớ cho truy vấn nhiều hop | Nếu cần |
| **Time-series index** | Truy vấn thời gian và điểm số suy giảm | Nếu cần |
| **Concurrency handling** | Locking cho các phiên đồng thời cùng người dùng | Vừa |

### Hạn chế POC đã biết (Rõ ràng hoãn lại)

| Hạn chế | Tác động | Tại sao chấp nhận được |
|------------|--------|----------------|
| **No concurrency control** | Tình trạng chạy đua nếu cùng người dùng có 2+ phiên đồng thời | Hiếm trong thử nghiệm POC; sửa chữa trong sản xuất |
| **No memory limits** | Người dùng quyền lực có thể tích lũy bộ nhớ không giới hạn | Hạn ngạch được thêm trong Giai đoạn 3 |
| **No backup/restore tested** | Lỗi đĩa Milvus = mất dữ liệu tiềm ẩn | Ngoại lệ cơ bản hoạt động; backup/HA được xác thực trong sản xuất |
| **No smart updates** | Sửa chữa tạo bản sao | Phiên bản mới thắng; Forget API có sẵn |
| **No adversarial defense** | Prompt injection có thể làm ô nhiễm bộ nhớ | Tin tưởng trầnput người dùng trong POC; thêm bộ lọc sau |

---

---

## Phụ lục

### Phụ lục A: Reflective Memory

**Trạng thái:** Phần mở rộng tương lai - nằm ngoài phạm vi POC này.

Tự phân tích và bài học học được từ các tương tác quá khứ. Được lấy cảm hứng từ [Reflexion paper](https://arxiv.org/abs/2303.11366).

**Nó lưu trữ cái gì:**

- Thông tin chi tiết từ phản hồi không chính xác hoặc không tối ưu
- Sở thích học được về phong cách phản hồi
- Mô hình cải thiện tương tác tương lai

**Ví dụ:**

- "Tôi đã đưa ra các bước triển khai không chính xác - lần tới xác minh phiên bản k8s trước"
- "Người dùng thích các dấu đầu dòng hơn đoạn văn cho nội dung kỹ thuật"
- "Câu hỏi về ngân sách sẽ bao gồm phân tích, không chỉ tổng"

**Tại sao tương lai:** Yêu cầu khả năng đánh giá chất lượng phản hồi và tạo ra sự phản ánh tự, được xây dựng dựa trên cơ sở hạ tầng bộ nhớ cốt lõi.

---

### Phụ lục B: Cây tệp

```
pkg/
├── extproc/
│   ├── processor_req_body.go     (EXTEND) Integrate retrieval
│   ├── processor_res_body.go     (EXTEND) Integrate extraction
│   └── req_filter_memory.go      (EXTEND) Pre-filter, retrieval, injection
│
├── memory/
│   ├── extractor.go              (NEW) LLM-based fact extraction
│   ├── store.go                  (existing) Store interface
│   ├── milvus_store.go           (existing) Milvus implementation
│   └── types.go                  (existing) Memory types
│
├── responseapi/
│   └── types.go                  (existing) MemoryConfig, MemoryContext
│
└── config/
    └── config.go                 (EXTEND) Add extraction config
```

---

### Phụ lục C: Viết lại truy vấn cho tìm kiếm bộ nhớ

Khi tìm kiếm bộ nhớ, các truy vấn không rõ như "how much?" cần bối cảnh để có hiệu quả. Phụ lục này bao gồm các chiến lược viết lại truy vấn.

#### Vấn đề

```
History: ["Planning Hawaii vacation", "Looking at hotels"]
Query: "How much?"
→ Tìm kiếm trực tiếp cho "How much?" sẽ không tìm thấy "Hawaii budget is $10,000"
```

#### Tùy chọn 1: Context Prepend (MVP)

Nối đơn giản - không có cuộc gọi LLM, ~0ms latency.

```go
func buildSearchQuery(history []Message, query string) string {
    context := extractKeyTerms(history)  // "Hawaii vacation planning"
    return query + " " + context         // "How much? Hawaii vacation planning"
}
```

**Ưu điểm:** Nhanh, đơn giản
**Nhược điểm:** Có thể bao gồm các thuật ngữ không liên quan

#### Tùy chọn 2: LLM Query Rewriting

Sử dụng LLM để viết lại truy vấn như câu hỏi độc lập. ~100-200ms latency.

```go
func rewriteQuery(history []Message, query string) string {
    prompt := `Given conversation about: %s
    Rewrite this query to be self-contained: "%s"
    Return ONLY the rewritten query.`
    return llm.Complete(fmt.Sprintf(prompt, summarize(history), query))
}
// "How much?" → "What is the budget for the Hawaii vacation?"
```

**Ưu điểm:** Truy vấn tự nhiên, khớp nhúng tốt hơn
**Nhược điểm:** LLM latency, chi phí

#### Tùy chọn 3: HyDE (Hypothetical Document Embeddings)

Tạo câu trả lời giả thuyết, nhúng thay vào truy vấn.

**Vấn đề HyDE giải quyết:**

```
Query: "What's the cost?"           → embeds as QUESTION style
Stored: "Budget is $10,000"         → embeds as STATEMENT style
Result: Low similarity (style mismatch)

With HyDE:
Query → LLM generates: "The cost is approximately $10,000"
This embeds as STATEMENT style → matches stored memory!
```

```go
func hydeRewrite(query string, history []Message) string {
    prompt := `Based on this conversation: %s
    Write a short factual answer to: "%s"`
    return llm.Complete(fmt.Sprintf(prompt, summarize(history), query))
}
// "How much?" → "The budget for the Hawaii trip is approximately $10,000"
```

**Ưu điểm:** Chất lượng truy xuất tốt nhất (bộ lọc câu hỏi-đến-tài liệu)
**Nhược điểm:** Latency cao nhất (~200ms), chi phí LLM

#### Khuyến nghị

| Giai đoạn | Tiếp cận | Sử dụng khi |
|-------|----------|----------|
| **MVP** | Context prepend | Tất cả truy vấn (mặc định) |
| **v1** | LLM rewrite | Truy vấn mơ hồ ("how much?", "and that?") |
| **v2** | HyDE | **Sau khi quan sát** điểm truy xuất thấp cho truy vấn kiểu câu hỏi |

> **Ghi chú:** HyDE là một tối ưu hóa dựa trên hiệu suất quan sát, không phải dự đoán.
> Áp dụng nó khi bạn thấy bộ nhớ liên quan tồn tại nhưng không được truy xuất.

#### Tham khảo

**Viết lại truy vấn:**

1. **HyDE** - [Precise Zero-Shot Dense Retrieval without Relevance Labels](https://arxiv.org/abs/2212.10496) (Gao et al., 2022) - Biên kiểu (câu hỏi → tài liệu)
2. **RRR** - [Query Rewriting for Retrieval-Augmented LLMs](https://arxiv.org/abs/2305.14283) (Ma et al., 2023) - Rewriter có thể huấn luyện với RL, xử lý bối cảnh hội thoại

**Agentic Memory (từ [Issue #808](https://github.com/vllm-project/semantic-router/issues/808)):**

5. **MemGPT** - [Towards LLMs as Operating Systems](https://arxiv.org/abs/2310.08560) (Packer et al., 2023)
6. **Generative Agents** - [Simulacra of Human Behavior](https://arxiv.org/abs/2304.03442) (Park et al., 2023)
7. **Reflexion** - [Language Agents with Verbal Reinforcement Learning](https://arxiv.org/abs/2303.11366) (Shinn et al., 2023)
8. **Voyager** - [An Open-Ended Embodied Agent with LLMs](https://arxiv.org/abs/2305.16291) (Wang et al., 2023)

---

*Tác giả tài liệu: [Yehudit Kerido, Marina Koushnir]*
*Cập nhật lần cuối: Tháng 12 năm 2025*
*Trạng thái: POC DESIGN - v3 (Review-Addressed)*
*Dựa trên: [Issue #808 - Explore Agentic Memory in Response API](https://github.com/vllm-project/semantic-router/issues/808)*
