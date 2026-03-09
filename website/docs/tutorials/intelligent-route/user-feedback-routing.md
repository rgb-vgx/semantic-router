# Định tuyến Dựa trên Tín hiệu Phản hồi Người dùng

Hướng dẫn này hướng dẫn bạn cách định tuyến các yêu cầu dựa trên phản hồi và tín hiệu mức độ hài lòng của người dùng. Tín hiệu user_feedback giúp xác định các tin nhắn theo dõi, các sửa chữa và mức độ hài lòng.

## Lợi ích chính

- **Định tuyến thích ứng**: Phát hiện khi người dùng không hài lòng và định tuyến đến các mô hình tốt hơn
- **Xử lý sửa chữa**: Tự động xử lý các tin nhắn "điều đó là sai" và "thử lại"
- **Phân tích mức độ hài lòng**: Xác định phản hồi tích cực so với tiêu cực
- **Cải thiện trải nghiệm người dùng**: Cung cấp các phản hồi tốt hơn khi người dùng chỉ ra sự không hài lòng

## Vấn đề nó giải quyết là gì?

Người dùng thường cung cấp phản hồi trong các tin nhắn theo dõi:

- **Sửa chữa**: "Điều đó là sai", "Không, đó không phải những gì tôi có ý"
- **Mức độ hài lòng**: "Cảm ơn", "Điều đó hữu ích", "Hoàn hảo"
- **Làm rõ**: "Bạn có thể giải thích thêm không?", "Tôi không hiểu"
- **Thử lại**: "Thử lại", "Cho tôi một câu trả lời khác"

Tín hiệu user_feedback tự động xác định các mẫu này, cho phép bạn:

1. Định tuyến các sửa chữa đến các mô hình có khả năng hơn
2. Phát hiện mức độ hài lòng để giám sát
3. Xử lý các câu hỏi theo dõi một cách thích hợp
4. Cải thiện chất lượng phản hồi dựa trên phản hồi

## Cấu hình

### Cấu hình cơ bản

Xác định các tín hiệu phản hồi người dùng trong `config.yaml` của bạn:

```yaml
signals:
  user_feedbacks:
    - name: "wrong_answer"
      description: "User indicates previous answer was incorrect"

    - name: "satisfied"
      description: "User is satisfied with the answer"

    - name: "need_clarification"
      description: "User needs more clarification on the answer"

    - name: "want_different"
      description: "User wants some other different answer"
```

### Sử dụng trong các Quy tắc Quyết định

```yaml
decisions:
  - name: wrong_answer_route
    description: "Handle user feedback indicating wrong answer - rethink and provide correct response"
    priority: 150
    rules:
      operator: "AND"
      conditions:
        - type: "user_feedback"
          name: "wrong_answer"
    modelRefs:
      - model: "openai/gpt-oss-120b"
        use_reasoning: true
    plugins:
      - type: "system_prompt"
        configuration:
          system_prompt: "The user has indicated that the previous answer was incorrect. Please carefully reconsider the question, identify what might have been wrong in the previous response, and provide a corrected and accurate answer. Think step-by-step and verify your reasoning before responding."

  - name: retry_with_different_approach
    description: "Route requests for different approach"
    priority: 100
    rules:
      operator: "AND"
      conditions:
        - type: "user_feedback"
          name: "want_different"
    modelRefs:
      - model: "openai/gpt-oss-120b"
        use_reasoning: true
    plugins:
      - type: "system_prompt"
        configuration:
          system_prompt: "The user wants a different approach or perspective. Provide an alternative solution or explanation that differs from the previous response."
```

## Các loại Phản hồi

### 1. Câu trả lời sai

**Mẫu**: "Điều đó là sai", "Không", "Không chính xác", "Thử lại"

```yaml
signals:
  user_feedbacks:
    - name: "wrong_answer"
      description: "User indicates previous answer was incorrect"

decisions:
  - name: wrong_answer_route
    description: "Handle user feedback indicating wrong answer"
    priority: 150
    rules:
      operator: "AND"
      conditions:
        - type: "user_feedback"
          name: "wrong_answer"
    modelRefs:
      - model: "openai/gpt-oss-120b"
        use_reasoning: true
    plugins:
      - type: "system_prompt"
        configuration:
          system_prompt: "The user has indicated that the previous answer was incorrect. Please carefully reconsider the question and provide a corrected answer."
```

**Truy vấn ví dụ**:

- "Điều đó là sai, câu trả lời là 42" → ✅ Phát hiện sửa chữa
- "Không, đó không phải những gì tôi có ý" → ✅ Phát hiện sửa chữa
- "Thử lại với một cách tiếp cận khác" → ✅ Phát hiện sửa chữa

### 2. Hài lòng

**Mẫu**: "Cảm ơn", "Hoàn hảo", "Điều đó hữu ích", "Tuyệt vời"

```yaml
signals:
  user_feedbacks:
    - name: "satisfied"
      description: "User is satisfied with the answer"

decisions:
  - name: track_satisfaction
    description: "Track satisfied users"
    priority: 50
    rules:
      operator: "AND"
      conditions:
        - type: "user_feedback"
          name: "satisfied"
    modelRefs:
      - model: "openai/gpt-oss-120b"
        use_reasoning: false
    plugins:
      - type: "system_prompt"
        configuration:
          system_prompt: "The user is satisfied. Continue providing helpful assistance."
```

**Truy vấn ví dụ**:

- "Cảm ơn, đó chính xác là những gì tôi cần" → ✅ Phát hiện mức độ hài lòng
- "Hoàn hảo, điều đó giúp rất nhiều" → ✅ Phát hiện mức độ hài lòng
- "Giải thích tuyệt vời" → ✅ Phát hiện mức độ hài lòng

### 3. Cần làm rõ

**Mẫu**: "Bạn có thể giải thích thêm không?", "Tôi không hiểu", "Bạn có ý gì?"

```yaml
signals:
  user_feedbacks:
    - name: "need_clarification"
      description: "User needs more clarification on the answer"

decisions:
  - name: provide_clarification
    description: "Provide more detailed explanation"
    priority: 100
    rules:
      operator: "AND"
      conditions:
        - type: "user_feedback"
          name: "need_clarification"
    modelRefs:
      - model: "openai/gpt-oss-120b"
        use_reasoning: true
    plugins:
      - type: "system_prompt"
        configuration:
          system_prompt: "The user needs more clarification. Provide a more detailed, step-by-step explanation with examples."
```

**Truy vấn ví dụ**:

- "Bạn có thể giải thích điều đó bằng những thuật ngữ đơn giản hơn không?" → ✅ Cần làm rõ
- "Tôi không hiểu phần cuối cùng" → ✅ Cần làm rõ
- "Bạn có ý gì?" → ✅ Cần làm rõ

### 4. Muốn cách tiếp cận khác

**Mẫu**: "Cho tôi một câu trả lời khác", "Thử một cách khác", "Chỉ cho tôi những lựa chọn thay thế"

```yaml
signals:
  user_feedbacks:
    - name: "want_different"
      description: "User wants some other different answer"

decisions:
  - name: retry_with_different_approach
    description: "Provide alternative solution"
    priority: 100
    rules:
      operator: "AND"
      conditions:
        - type: "user_feedback"
          name: "want_different"
    modelRefs:
      - model: "openai/gpt-oss-120b"
        use_reasoning: true
    plugins:
      - type: "system_prompt"
        configuration:
          system_prompt: "The user wants a different approach or perspective. Provide an alternative solution or explanation that differs from the previous response."
```

**Truy vấn ví dụ**:

- "Cho tôi một cách khác để giải quyết vấn đề này" → ✅ Muốn cách tiếp cận khác
- "Chỉ cho tôi một cách khác" → ✅ Muốn cách tiếp cận khác
- "Bạn có thể thử một phương pháp khác không?" → ✅ Muốn cách tiếp cận khác

## Các trường hợp sử dụng

### 1. Hỗ trợ khách hàng - Nâng cao

**Vấn đề**: Khách hàng không hài lòng cần các phản hồi tốt hơn

```yaml
signals:
  user_feedbacks:
    - name: "wrong_answer"
      description: "Customer indicates previous answer was incorrect"

decisions:
  - name: escalate_to_premium
    description: "Escalate to premium model"
    priority: 150
    rules:
      operator: "AND"
      conditions:
        - type: "user_feedback"
          name: "wrong_answer"
    modelRefs:
      - model: "openai/gpt-oss-120b"
        use_reasoning: true
    plugins:
      - type: "system_prompt"
        configuration:
          system_prompt: "The customer was not satisfied with the previous answer. Provide a better, more accurate response."
```

### 2. Giáo dục - Học tập thích ứng

**Vấn đề**: Sinh viên cần các giải thích khác nhau khi bị nhầm lẫn

```yaml
signals:
  user_feedbacks:
    - name: "need_clarification"
      description: "Student needs more clarification on the answer"

decisions:
  - name: detailed_explanation
    description: "Provide detailed explanation"
    priority: 100
    rules:
      operator: "AND"
      conditions:
        - type: "user_feedback"
          name: "need_clarification"
    modelRefs:
      - model: "openai/gpt-oss-120b"
        use_reasoning: true
    plugins:
      - type: "system_prompt"
        configuration:
          system_prompt: "The student needs more clarification. Provide a detailed, step-by-step explanation with examples."
```

## Các thực hành tốt nhất

### 1. Kết hợp với Ngữ cảnh

Sử dụng lịch sử trò chuyện để cải thiện phát hiện:

```yaml
# Track conversation state
context:
  previous_response: true
  conversation_history: 3  # Last 3 messages
```

### 2. Đặt Ưu tiên Nâng cao

Các sửa chữa nên có ưu tiên cao:

```yaml
decisions:
  - name: handle_correction
    priority: 100  # High priority for corrections
```

### 3. Giám sát Tỷ lệ Hài lòng

Theo dõi các mẫu phản hồi:

```yaml
logging:
  level: info
  user_feedback: true
  satisfaction_metrics: true
```

### 4. Sử dụng các Mô hình Thích hợp

- **Sửa chữa**: Định tuyến đến các mô hình có khả năng/tốn kém hơn
- **Làm rõ**: Định tuyến đến các mô hình tốt tại các giải thích
- **Mức độ hài lòng**: Tiếp tục với mô hình hiện tại

## Tham chiếu

Xem [Signal-Driven Decision Architecture](../../overview/signal-driven-decisions.md) để biết kiến trúc tín hiệu hoàn chỉnh.
