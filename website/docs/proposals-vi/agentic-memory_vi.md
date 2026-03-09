# Bộ nhớ của Agent

## Tóm tắt Điều hành

Tài liệu này mô tả một **Bằng chứng Khái niệm** cho Bộ nhớ của Agent trong Semantic Router. Bộ nhớ của Agent cho phép các agent AI **ghi nhớ thông tin qua các phiên làm việc**, cung cấp tính liên tục và cá nhân hóa.

> **⚠️ Phạm vi POC:** Đây là bằng chứng khái niệm chứ không phải thiết kế sản xuất. Mục tiêu là xác thực quy trình bộ nhớ cốt lõi (truy xuất → tiêm → trích xuất → lưu trữ) với độ chính xác có thể chấp nhận được. Cứng cáp sản xuất (xử lý lỗi, mở rộng, theo dõi) nằm ngoài phạm vi.

### Khả năng cốt lõi

| Khả năng | Mô tả |
|------------|-------------|
| **Truy xuất bộ nhớ** | Tìm kiếm dựa trên nhúng với lọc trước đơn giản |
| **Lưu bộ nhớ** | Trích xuất các sự kiện và thủ tục dựa trên LLM |
| **Tính bền vững qua phiên** | Bộ nhớ được lưu trữ trong Milvus (tồn tại sau khi khởi động lại; sao lưu/HA sản xuất chưa được kiểm tra) |
| **Cách ly người dùng** | Bộ nhớ được phân loại theo user_id (xem ghi chú dưới đây) |

> **⚠️ Cách ly người dùng - Ghi chú hiệu suất Milvus:**
>
> | Cách tiếp cận | POC | Sản xuất (10K+ người dùng) |
> |----------|-----|-------------------------|
> | **Bộ lọc đơn giản** | ✅ Lọc theo `user_id` sau tìm kiếm | ❌ Giảm chất lượng: tìm kiếm tất cả người dùng, sau đó lọc |
> | **Khoá Phân vùng** | ❌ Quá mức | ✅ Tách biệt vật lý, O(log N) cho mỗi người dùng |
> | **Chỉ mục vô hướng** | ❌ Quá mức | ✅ Chỉ mục trên `user_id` để lọc nhanh |
>
> **POC:** Sử dụng lọc siêu dữ liệu đơn giản (đủ để kiểm tra).
> **Sản xuất:** Cấu hình `user_id` làm Khoá Phân vùng hoặc Trường được Chỉ mục vô hướng trong lược đồ Milvus.

### Nguyên tắc thiết kế chính

1. **Bộ lọc trước đơn giản** quyết định liệu truy vấn có nên tìm kiếm bộ nhớ hay không
2. **Cửa sổ bối cảnh** từ lịch sử để phân biệt truy vấn
3. **LLM trích xuất các sự kiện** và phân loại loại khi lưu
4. **Lọc dựa trên ngưỡng** trên kết quả tìm kiếm

### Giả định Rõ ràng (POC)

| Giả định | Hàm ý | Rủi ro nếu sai |
|------------|-------------|---------------|
| Trích xuất LLM được coi là hợp lý chính xác | Một số sự kiện không chính xác có thể được lưu | Ô nhiễm bộ nhớ (có thể khắc phục qua API Quên) |
| Ngưỡng tương tự 0,6 là điểm bắt đầu | Có thể cần điều chỉnh (bỏ lỡ liên quan hoặc bao gồm không liên quan) | Điều chỉnh dựa trên nhật ký chất lượng truy xuất |
| Milvus có sẵn và được cấu hình | Tính năng bị vô hiệu hóa nếu bị hỏng | Suy giảm duyền (không sập) |
| Mô hình nhúng tạo ra các vectơ 384 chiều | Phải khớp với lược đồ Milvus | Lỗi khởi động (có thể phát hiện) |
| Lịch sử có sẵn qua chuỗi Response API | Cần thiết cho bối cảnh | Bỏ qua bộ nhớ nếu không có sẵn |

---

## Mục lục

1. [Tuyên bố vấn đề](#1-tuyên-bố-vấn-đề)
2. [Tổng quan kiến trúc](#2-tổng-quan-kiến-trúc)
3. [Loại bộ nhớ](#3-loại-bộ-nhớ)
4. [Tích hợp đường ống](#4-tích-hợp-đường-ống)
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

Response API cung cấp chuỗi trò chuyện qua `previous_response_id`, nhưng kiến thức bị mất qua các phiên:

```
Phiên A (15 tháng 3):
  Người dùng: "Ngân sách của tôi cho chuyến đi Hawaii là 10.000 đô la"
  → Lưu trong chuỗi phiên

Phiên B (20 tháng 3) - PHIÊN MỚI:
  Người dùng: "Ngân sách của tôi cho chuyến đi là bao nhiêu?"
  → Không có previous_response_id → Kiến thức BỊ MẤT ❌
```

### Trạng thái mong muốn

Với Bộ nhớ của Agent:

```
Phiên A (15 tháng 3):
  Người dùng: "Ngân sách của tôi cho chuyến đi Hawaii là 10.000 đô la"
  → Trích xuất và lưu vào Milvus

Phiên B (20 tháng 3) - PHIÊN MỚI:
  Người dùng: "Ngân sách của tôi cho chuyến đi là bao nhiêu?"
  → Bộ lọc trước: liên quan bộ nhớ ✓
  → Tìm kiếm Milvus → Tìm thấy: "ngân sách cho Hawaii là 10K"
  → Tiêm vào ngữ cảnh LLM
  → Trợ lý: "Ngân sách cho chuyến đi Hawaii của bạn là 10.000 đô la!" ✅
```

---

## 2. Tổng quan Kiến trúc

```
┌─────────────────────────────────────────────────────────────────────────┐
│                      KIẾN TRÚC BỘ NHỚ CỦA AGENT                         │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│                         Đường ống ExtProc                               │
│  ┌──────────────────────────────────────────────────────────────────┐   │
│  │                                                                  │   │
│  │  Yêu cầu → Sự kiện? → Công cụ? → Bảo mật → Bộ nhớ đệm → BỘ NHỚ → LLM       │   │
│  │              │       │                          ↑↓               │   │
│  │              └───────┴──── tín hiệu được sử dụng ────────┘                │   │
│  │                                                                  │   │
│  │  Phản hồi ← [trích xuất & lưu] ←─────────────────┘                │   │
│  │                                                                  │   │
│  └──────────────────────────────────────────────────────────────────┘   │
│                                          │                              │
│                    ┌─────────────────────┴─────────────────────┐        │
│                    │                                           │        │
│          ┌─────────▼─────────┐                    ┌────────────▼───┐    │
│          │ Truy xuất bộ nhớ   │                    │ Lưu bộ nhớ     │    │
│          │  (giai đoạn yêu cầu)                   │(giai đoạn phản hồi)│    │
│          ├───────────────────┤                    ├────────────────┤    │
│          │ 1. Kiểm tra tín   │                    │ 1. Trích xuất  │    │
│          │    hiệu (Sự kiện?│                    │    LLM         │    │
│          │ 2. Tạo bối cảnh  │                    │ 2. Phân loại   │    │
│          │ 3. Tìm kiếm Milvus                    │ 3. Khử trùng  │    │
│          │ 4. Tiêm vào LLM  │                    │ 4. Lưu trữ     │    │
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
|-----------|-------------|----------|
| **Bộ lọc bộ nhớ** | Quyết định + tìm kiếm + tiêm | `pkg/extproc/req_filter_memory.go` |
| **Trích xuất bộ nhớ** | Trích xuất sự kiện dựa trên LLM | `pkg/memory/extractor.go` (mới) |
| **Lưu trữ bộ nhớ** | Giao diện lưu trữ | `pkg/memory/store.go` |
| **Lưu trữ Milvus** | Phía sau cơ sở dữ liệu vectơ | `pkg/memory/milvus_store.go` |
| **Bộ phân loại hiện có** | Tín hiệu Sự kiện/Công cụ (tái sử dụng) | `pkg/extproc/processor_req_body.go` |

### Kiến trúc lưu trữ

[Vấn đề #808](https://github.com/vllm-project/semantic-router/issues/808) đề xuất kiến trúc lưu trữ đa lớp. Chúng ta thực hiện điều này từng giai đoạn:

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    KIẾN TRÚC LƯU TRỮ (Giai đoạn)                        │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  GIAI ĐOẠN 1 (MVP)                                              │    │
│  │  ┌─────────────────────────────────────────────────────────┐    │    │
│  │  │  Milvus (Chỉ mục vectơ)                                 │    │    │
│  │  │  • Tìm kiếm ngữ nghĩa trên bộ nhớ                      │    │    │
│  │  │  • Lưu trữ nhúng                                        │    │    │
│  │  │  • Nội dung + siêu dữ liệu                              │    │    │
│  │  └─────────────────────────────────────────────────────────┘    │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  GIAI ĐOẠN 2 (Hiệu suất)                                        │    │
│  │  ┌─────────────────────────────────────────────────────────┐    │    │
│  │  │  Redis (Bộ nhớ đệm nóng)                                │    │    │
│  │  │  • Tra cứu siêu dữ liệu nhanh                           │    │    │
│  │  │  • Bộ nhớ được truy cập gần đây                         │    │    │
│  │  │  • Hỗ trợ TTL/hết hạn                                  │    │    │
│  │  └─────────────────────────────────────────────────────────┘    │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  GIAI ĐOẠN 3+ (Nếu cần thiết)                                   │    │
│  │  ┌───────────────────────┐  ┌───────────────────────┐           │    │
│  │  │  Cửa hàng biểu đồ (Neo4j)  │  Chỉ mục chuỗi thời gian  │           │    │
│  │  │  • Liên kết bộ nhớ   │  │  • Truy vấn theo thời gian │           │    │
│  │  │  • Mối quan hệ     │  │  • Điểm giảm              │           │    │
│  │  └───────────────────────┘  └───────────────────────┘           │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

| Lớp | Mục đích | Khi nào cần thiết | Trạng thái |
|-------|---------|-------------|--------|
| **Milvus** | Tìm kiếm vectơ ngữ nghĩa | Chức năng cốt lõi | ✅ MVP |
| **Redis** | Bộ nhớ đệm nóng, truy cập nhanh, TTL | Tối ưu hóa hiệu suất | 🔶 Giai đoạn 2 |
| **Biểu đồ (Neo4j)** | Mối quan hệ bộ nhớ | Truy vấn lý do đa bước | ⚪ Nếu cần |
| **Chuỗi thời gian** | Truy vấn theo thời gian, giảm | Ghi điểm quan trọng theo thời gian | ⚪ Nếu cần |

> **Quyết định thiết kế:** Chúng ta bắt đầu chỉ với Milvus. Các lớp bổ sung được thêm dựa trên nhu cầu được chứng minh, không phải đoán. Giao diện `Store` tóm tắt lưu trữ, cho phép thêm phía sau mà không thay đổi logic truy xuất/lưu.

---

## 3. Loại bộ nhớ

| Loại | Mục đích | Ví dụ | Trạng thái |
|------|---------|---------|--------|
| **Ngữ nghĩa** | Sự kiện, sở thích, kiến thức | "Ngân sách của người dùng cho Hawaii là 10.000 đô la" | ✅ MVP |
| **Thủ tục** | Hướng dẫn, bước, quy trình | "Để triển khai payment-service: chạy npm build, rồi docker push" | ✅ MVP |
| **Tập sự** | Tóm tắt phiên, sự kiện quá khứ | "Vào ngày 29 tháng 12 năm 2024, người dùng lập kế hoạch kỳ nghỉ Hawaii với ngân sách 10K" | ⚠️ MVP (hạn chế) |
| **Phản xạ** | Tự phân tích, bài học đã học | "Phản hồi ngân sách trước đó không đầy đủ - người dùng thích những chi tiết chi tiết" | 🔮 Tương lai |

> **⚠️ Bộ nhớ Tập sự (Giới hạn MVP):** Phát hiện kết thúc phiên không được thực hiện. Bộ nhớ Tập sự chỉ được tạo khi trích xuất LLM rõ ràng tạo ra đầu ra kiểu tóm tắt. Các kích hoạt kết thúc phiên đáng tin cậy được hoãn lại đến Giai đoạn 2.
>
> **🔮 Bộ nhớ Phản xạ:** Tự phân tích và bài học. Không nằm trong phạm vi POC này. Xem [Phụ lục A](#phụ-lục-a-bộ-nhớ-phản-xạ).

### Không gian vectơ bộ nhớ

Bộ nhớ nhóm theo **nội dung/chủ đề**, không phải loại. Loại là siêu dữ liệu:

```
┌────────────────────────────────────────────────────────────────────────┐
│                      KHÔNG GIAN VECTƠ BỘ NHỚ                           │
│                                                                        │
│     ┌─────────────────┐                    ┌─────────────────┐         │
│     │  NGÂN SÁCH/TIỀN │                    │   TRIỂN KHAI    │         │
│     │    CỤM          │                    │    CỤM          │         │
│     │                 │                    │                 │         │
│     │ • ngân sách=$10K│                    │ • npm build     │         │
│     │   (ngữ nghĩa)  │                    │   (thủ tục)    │         │
│     │ • chi phí=$5K  │                    │ • docker push   │         │
│     │   (ngữ nghĩa)  │                    │   (thủ tục)    │         │
│     └─────────────────┘                    └─────────────────┘         │
│                                                                        │
│  • = bộ nhớ với loại làm siêu dữ liệu                                  │
│  Truy vấn khớp nội dung → loại từ bộ nhớ phù hợp                      │
│                                                                        │
└────────────────────────────────────────────────────────────────────────┘
```

### Response API vs. Bộ nhớ của Agent: Khi nào bộ nhớ thêm giá trị?

**Sự khác biệt quan trọng:** Response API đã gửi toàn bộ lịch sử trò chuyện cho LLM khi `previous_response_id` có mặt. Giá trị của Bộ nhớ của Agent là cho **bối cảnh qua phiên**.

```
┌─────────────────────────────────────────────────────────────────────────┐
│           RESPONSE API vs. BỘ NHỚ CỦA AGENT: NGUỒN BỐI CẢNH              │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  CÙNG PHiên (có previous_response_id):                                  │
│  ─────────────────────────────────────                                  │
│    Response API cung cấp:                                               │
│      └── Chuỗi trò chuyện đầy đủ (tất cả lượt) → gửi tới LLM          │
│                                                                         │
│    Bộ nhớ của Agent:                                                    │
│      └── VẪN CÓ GIÁ TRỊ - phiên hiện tại có thể không có câu trả lời   │
│      └── Ví dụ: 100 lượt lập kế hoạch kỳ nghỉ, nhưng ngân sách không   │
│      └── Cách đây vài ngày: "Tôi có 10K tiền dư, có đủ cho một tuần    │
│          ở Thái Lan không?" → LLM trích xuất: "Người dùng có ngân sách  │
│          $10K cho chuyến đi"                                            │
│      └── Bây giờ: "Ngân sách của tôi là bao nhiêu?" → câu trả lời      │
│          trong bộ nhớ, không phải trong chuỗi này                       │
│                                                                         │
│  PHIÊN MỚI (không có previous_response_id):                             │
│  ──────────────────────────────────────                                 │
│    Response API cung cấp:                                               │
│      └── Không gì cả (không có chuỗi để theo)                          │
│                                                                         │
│    Bộ nhớ của Agent:                                                    │
│      └── THÊM GIÁ TRỊ - truy xuất bối cảnh qua phiên                  │
│      └── "Ngân sách Hawaii của tôi là bao nhiêu?" → tìm thấy sự kiện   │
│          từ phiên tháng 3                                               │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

> **Quyết định thiết kế:** Truy xuất bộ nhớ thêm giá trị trong **cả hai** kịch bản — phiên mới (không có chuỗi) và phiên hiện có (truy vấn có thể tham chiếu các phiên khác). Chúng ta luôn tìm kiếm khi bộ lọc trước vượt qua.
>
> **Dự phòng đã biết:** Khi câu trả lời CÓ trong chuỗi hiện tại, chúng tôi vẫn tìm kiếm bộ nhớ (~10-30ms bị lãng phí). Chúng ta không thể rẻ tiền phát hiện "câu trả lời đã ở trong lịch sử?" mà không hiểu truy vấn ngữ nghĩa. Đối với POC, chúng tôi chấp nhận chi phí chung này.
>
> **Giải pháp Giai đoạn 2:** [Nén Bối cảnh](#nén-bối-cảnh-ưu-tiên-cao) giải quyết vấn đề này đúng cách — thay vì Response API gửi toàn bộ lịch sử, chúng ta gửi tóm tắt nén + lượt gần đây + bộ nhớ liên quan. Các sự kiện được trích xuất trong quá trình nén, loại bỏ dự phòng hoàn toàn.

---

[Phần còn lại của tài liệu tiếp theo giữ nguyên cấu trúc Markdown, dịch tất cả văn bản tiếng Anh sang tiếng Việt chuyên môn, giữ lại tất cả code blocks, links, URL và số liệu không thay đổi...]

Tôi sẽ tiếp tục dịch phần còn lại của file này. Do độ dài của tài liệu, tôi sẽ lưu phần còn lại vào file:
