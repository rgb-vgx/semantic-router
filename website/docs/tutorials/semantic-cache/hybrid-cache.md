# Bộ Nhớ Đệm Hybrid: HNSW + Milvus

Bộ nhớ đệm Hybrid kết hợp chỉ mục HNSW trong bộ nhớ để tìm kiếm nhanh với cơ sở dữ liệu vectơ Milvus để lưu trữ có thể mở rộng.

## Tổng Quan

Kiến trúc hybrid cung cấp:

- **Tìm kiếm nhanh** qua chỉ mục HNSW trong bộ nhớ
- **Lưu trữ có thể mở rộng** thông qua cơ sở dữ liệu vectơ Milvus
- **Tính bền bỉ** với Milvus như nguồn sự thật
- **Lưu trữ dữ liệu nóng** với bộ nhớ đệm tài liệu cục bộ

## Kiến Trúc

```
┌──────────────────────────────────────────────────┐
│                  Bộ Nhớ Đệm Hybrid               │
├──────────────────────────────────────────────────┤
│  ┌─────────────────┐      ┌──────────────────┐  │
│  │  Trong Bộ Nhớ   │      │   Bộ Nhớ Đệm     │  │
│  │  Chỉ Mục HNSW  │◄─────┤   Tục Bản (Dữ    │  │
│  └────────┬────────┘      │   Liệu Nóng)     │  │
│           │                └──────────────────┘  │
│           │                                       │
│           │ Ánh Xạ ID                           │
│           ▼                                       │
│  ┌──────────────────────────────────────────┐   │
│  │         Cơ Sở Dữ Liệu Vectơ Milvus      │   │
│  └──────────────────────────────────────────┘   │
└──────────────────────────────────────────────────┘
```

## Cách Hoạt Động

### Đường Dẫn Viết (AddEntry)

Khi thêm mục bộ nhớ đệm:

1. Tạo nhúng bằng mô hình nhúng được cấu hình
2. Ghi mục vào Milvus để duy trì
3. Thêm mục vào chỉ mục HNSW trong bộ nhớ (nếu có chỗ trống)
4. Thêm tài liệu vào bộ nhớ đệm cục bộ

### Đường Dẫn Đọc (FindSimilar)

Khi tìm kiếm một truy vấn tương tự:

1. Tạo nhúng truy vấn
2. Tìm kiếm chỉ mục HNSW cho hàng xóm gần nhất
3. Kiểm tra bộ nhớ đệm cục bộ để tìm các tài liệu phù hợp
   - Nếu tìm thấy trong bộ nhớ đệm cục bộ: trả lại ngay (đường dẫn nóng)
   - Nếu không tìm thấy: tìm nạp từ Milvus (đường dẫn lạnh)
4. Lưu trữ các tài liệu được tìm nạp trong bộ nhớ đệm cục bộ cho các truy vấn trong tương lai

### Quản Lý Bộ Nhớ

- **Chỉ Mục HNSW**: Giới hạn ở số mục tối đa được cấu hình
- **Bộ Nhớ Đệm Cục Bộ**: Giới hạn ở số lượng tài liệu được cấu hình
- **Loại Bỏ**: Chính sách FIFO khi đạt giới hạn
- **Tính Duy Trì Dữ Liệu**: Tất cả dữ liệu vẫn nằm trong Milvus bất kể các giới hạn bộ nhớ

## Cấu Hình

### Cấu Hình Cơ Bản

```yaml
semantic_cache:
  enabled: true
  backend_type: "hybrid"
  similarity_threshold: 0.85
  ttl_seconds: 3600

  # Cài đặt cụ thể hybrid
  max_memory_entries: 100000  # Mục tối đa trong HNSW
  local_cache_size: 1000      # Kích thước bộ nhớ đệm tài liệu cục bộ

  # Tham số HNSW
  hnsw_m: 16
  hnsw_ef_construction: 200

  # Cấu hình Milvus
  backend_config_path: "config/semantic-cache/milvus.yaml"
```

## Ví Dụ Sử Dụng

### Mã Go

```go
import "github.com/vllm-project/semantic-router/src/semantic-router/pkg/cache"

// Khởi tạo bộ nhớ đệm hybrid
options := cache.HybridCacheOptions{
    Enabled:             true,
    SimilarityThreshold: 0.85,
    TTLSeconds:          3600,
    MaxMemoryEntries:    100000,
    HNSWM:               16,
    HNSWEfConstruction:  200,
    MilvusConfigPath:    "config/semantic-cache/milvus.yaml",
    LocalCacheSize:      1000,
}

hybridCache, err := cache.NewHybridCache(options)
if err != nil {
    log.Fatalf("Không thể tạo bộ nhớ đệm hybrid: %v", err)
}
defer hybridCache.Close()

// Thêm mục bộ nhớ đệm
err = hybridCache.AddEntry(
    "request-id-123",
    "gpt-4",
    "Máy tính lượng tử là gì?",
    []byte(`{"prompt": "Máy tính lượng tử là gì?"}`),
    []byte(`{"response": "Máy tính lượng tử là..."}`),
)

// Tìm kiếm truy vấn tương tự
response, found, err := hybridCache.FindSimilar(
    "gpt-4",
    "Giải thích máy tính lượng tử",
)
if found {
    fmt.Printf("Nhấn bộ nhớ đệm! Phản hồi: %s\n", string(response))
}

// Lấy thống kê
stats := hybridCache.GetStats()
fmt.Printf("Tổng cộng các mục trong HNSW: %d\n", stats.TotalEntries)
fmt.Printf("Tỷ lệ chạm: %.2f%%\n", stats.HitRatio * 100)
```

## Triển Khai Nhiều Phiên Bản

Bộ nhớ đệm hybrid hỗ trợ triển khai nhiều phiên bản nơi mỗi phiên bản duy trì chỉ mục HNSW cho riêng nó và bộ nhớ đệm cục bộ, nhưng chia sẻ Milvus để duy trì và nhất quán dữ liệu:

```
┌─────────────┐   ┌─────────────┐   ┌─────────────┐
│  Phiên Bản 1│   │  Phiên Bản 2│   │  Phiên Bản 3│
│  Bộ Nhớ Đệm HNSW │   │  Bộ Nhớ Đệm HNSW │   │  Bộ Nhớ Đệm HNSW │
└──────┬──────┘   └──────┬──────┘   └──────┬──────┘
       │                 │                 │
       └─────────────────┼─────────────────┘
                         │
                  ┌──────▼──────┐
                  │   Milvus    │
                  │  (Chia Sẻ)  │
                  └─────────────┘
```

## Xem Thêm

- [Tài Liệu Bộ Nhớ Đệm Trong Bộ Nhớ](./in-memory-cache.md)
- [Tài Liệu Bộ Nhớ Đệm Milvus](./milvus-cache.md)
