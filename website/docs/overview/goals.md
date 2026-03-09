---
sidebar_position: 1
---

# Các mục tiêu của chúng tôi là gì?

Chúng tôi đang xây dựng **System Level Intelligence** cho Mixture-of-Models (MoM), mang **Collective Intelligence** vào **các hệ thống LLM**.

## Các câu hỏi cốt lõi

Dự án của chúng tôi giải quyết năm thách thức cơ bản trong các hệ thống LLM:

### 1. Làm cách nào để nắm bắt các tín hiệu bị thiếu?

Trong định tuyến LLM truyền thống, chúng tôi chỉ xem xét văn bản truy vấn của người dùng. Nhưng có rất nhiều thông tin mà chúng tôi đang bỏ qua:

- **Context signals**: Miền này liên quan đến những gì? (toán học, mã, viết sáng tạo?)
- **Quality signals**: Truy vấn này có cần kiểm tra thực tế không? Người dùng có đưa ra phản hồi không?
- **User signals**: Các ưu tiên của người dùng là gì? Mức độ hài lòng của họ là gì?

**Giải pháp của chúng tôi**: Một hệ thống trích xuất tín hiệu toàn diện nắm bắt 9 loại tín hiệu yêu cầu từ các yêu cầu, phản hồi và bảng ngữ cảnh.

### 2. Làm cách nào để kết hợp các tín hiệu?

Có nhiều tín hiệu thật tuyệt vời, nhưng chúng ta sử dụng chúng như thế nào cùng nhau để đưa ra quyết định tốt hơn?

- Chúng ta nên định tuyến tới mô hình toán học nếu chúng tôi phát hiện **cả** từ khóa toán học **và** miền toán học không?
- Chúng ta nên bật kiểm tra thực tế nếu chúng tôi phát hiện **trong m hoặc** một câu hỏi thực tế **hoặc** một miền nhạy cảm không?

**Giải pháp của chúng tôi**: Một công cụ quyết định linh hoạt với các toán tử AND/OR cho phép bạn kết hợp tín hiệu theo những cách mạnh mẽ.

### 3. Làm cách nào để cộng tác hiệu quả hơn?

Các mô hình khác nhau tốt ở những điều khác nhau. Làm cách nào chúng tôi có thể làm cho chúng làm việc cùng nhau như một nhóm?

- Định tuyến các câu hỏi toán học tới các mô hình chuyên biệt về toán
- Định tuyến viết sáng tạo tới các mô hình có sự sáng tạo tốt hơn
- Định tuyến các câu hỏi mã tới các mô hình được huấn luyện trên mã
- Sử dụng các mô hình nhỏ hơn cho các tác vụ đơn giản, các mô hình lớn hơn cho các tác vụ phức tạp

**Giải pháp của chúng tôi**: Định tuyến thông minh phù hợp với truy vấn tới mô hình tốt nhất dựa trên nhiều tín hiệu, không chỉ là các quy tắc đơn giản.

### 4. Làm cách nào để bảo mật hệ thống?

Các hệ thống LLM phải đối mặt với thách thức bảo mật duy nhất:

- **Jailbreak attacks**: Các lời nhắc đối nghịch cố gắng vượt qua các lá chắn bảo vệ
- **PII leaks**: Vô tình tiếp lộ thông tin cá nhân nhạy cảm
- **Hallucinations**: Các mô hình tạo ra thông tin sai lệch hoặc gây hiểu nhầm

**Giải pháp của chúng tôi**: Kiến trúc chuỗi plugin với nhiều lớp bảo mật (phát hiện jailbreak, lọc PII, phát hiện ảo tưởng).

### 5. Làm cách nào để thu thập các tín hiệu có giá trị?

Hệ thống nên học và cải thiện theo thời gian:

- Theo dõi những tín hiệu nào dẫn đến quyết định định tuyến tốt hơn
- Thu thập phản hồi người dùng để cải thiện phát hiện tín hiệu
- Xây dựng một hệ thống tự học trở nên thông minh hơn khi sử dụng

**Giải pháp của chúng tôi**: Quan sát toàn diện và thu thập phản hồi mà phản hồi lại vào hệ thống trích xuất tín hiệu và công cụ quyết định.

## Tầm nhìn

Chúng tôi hình dung một tương lai trong đó:

- **Các hệ thống LLM thông minh ở cấp độ hệ thống**, không chỉ ở cấp độ mô hình
- **Nhiều mô hình cộng tác liền mạch**, mỗi mô hình góp phần vào những điểm mạnh của chúng
- **Bảo mật được tích hợp sẵn**, không phải áp đặt
- **Hệ thống học và cải thiện** từ mỗi tương tác
- **Trí thông minh tập thể xuất hiện** từ sự kết hợp của tín hiệu, quyết định và phản hồi

## Tại sao điều này quan trọng

### Dành cho các nhà phát triển

- Xây dựng các ứng dụng LLM có khả năng cao hơn với ít nỗ lực
- Tận dụng nhiều mô hình mà không cần sắp xếp phức tạp
- Nhận được bảo mật và tuân thủ tích hợp sẵn

### Cho các tổ chức

- Giảm chi phí bằng cách định tuyến tới các mô hình thích hợp
- Cải thiện chất lượng thông qua lựa chọn mô hình chuyên biệt
- Đáp ứng các yêu cầu tuân thủ với các điều khiển PII và bảo mật tích hợp sẵn

### Cho người dùng

- Nhận được các phản hồi tốt hơn, chính xác hơn
- Trải nghiệm các thời gian phản hồi nhanh hơn thông qua bộ nhớ đệm
- Hưởng lợi từ cải thiện an toàn và quyền riêng tư

## Bước tiếp theo

Tìm hiểu thêm về các khái niệm cốt lõi:

- [Semantic Router là gì?](semantic-router-overview.md) - Hiểu định tuyến ngữ nghĩa
- [Collective Intelligence là gì?](collective-intelligence.md) - Cách các tín hiệu tạo ra trí thông minh
- [Signal-Driven Decision là gì?](signal-driven-decisions.md) - Tìm hiểu sâu về công cụ quyết định
