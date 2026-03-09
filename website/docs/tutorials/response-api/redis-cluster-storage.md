# Lưu Trữ Cụm Redis Cho API Phản Hồi

Hướng dẫn này bao gồm triển khai **Cụm Redis** cho API Phản Hồi, cung cấp tính sẵn sàng cao, chuyển đổi dự phòng tự động và sharding dữ liệu trên nhiều nút.

> **Lưu ý:** Để thiết lập Redis độc lập đơn giản, hãy xem [Hướng dẫn Lưu Trữ Redis](redis-storage.md).

## Redis Cụm là gì?

**Cụm Redis** cung cấp:

- ✅ **Sharding dữ liệu**: Tự động phân phối dữ liệu trên nhiều nút chính (slot hash 16384)
- ✅ **Tính sẵn sàng cao**: Chuyển đổi dự phòng tự động nếu một nút chính bị lỗi (cần bản sao)
- ✅ **Mở rộng quy mô ngang hoàng**: Thêm nhiều nút hơn để tăng dung lượng
- ✅ **Không có điểm lỗi duy nhất**: Dữ liệu được sao chép trên các nút

**so với Redis Độc lập:**

- Độc lập = 1 nút, đơn giản, tốt cho dev/triển khai nhỏ
- Cụm = 6+ nút (3 chính + 3 bản sao), sẵn sàng cho sản xuất

## Thiết Lập và Triển Khai

### 1. Khởi Động Cụm Redis

#### Bước 1: Tạo Mạng Docker

```bash
docker network create redis-cluster-net
```

#### Bước 2: Khởi Động 6 Nút Redis

```bash
for port in 7001 7002 7003 7004 7005 7006; do
  docker run -d \
    --name redis-node-$port \
    --network redis-cluster-net \
    -p $port:6379 \
    redis:7-alpine \
    redis-server --cluster-enabled yes \
                 --cluster-config-file nodes.conf \
                 --cluster-node-timeout 5000 \
                 --appendonly yes \
                 --port 6379
done
```

**Cái này làm:**

- Khởi động 6 máy chủ Redis độc lập
- Kích hoạt chế độ cụm trên mỗi máy
- Cổng 7001-7006 được tiếp xúc trên localhost

#### Bước 3: Tạo Cụm

```bash
docker run --rm --network redis-cluster-net redis:7-alpine \
  redis-cli --cluster create \
  redis-node-7001:6379 \
  redis-node-7002:6379 \
  redis-node-7003:6379 \
  redis-node-7004:6379 \
  redis-node-7005:6379 \
  redis-node-7006:6379 \
  --cluster-replicas 1 --cluster-yes
```

**Cái này làm:**

- Kết nối 6 nút vào một cụm
- Tạo 3 chính (7001, 7002, 7003)
- Tạo 3 bản sao (7004, 7005, 7006)
- Phân phối slot hash: 0-5460, 5461-10922, 10923-16383

#### Bước 4: Xác Minh Cụm Đang Chạy

```bash
docker exec redis-node-7001 redis-cli cluster info
docker exec redis-node-7001 redis-cli cluster nodes
```

### 2. Cấu Hình Bộ Định Tuyến Ngữ Nghĩa

#### Tùy Chọn 1: Cấu Hình Nội Tuyến

Chỉnh sửa `config/config.yaml`:

```yaml
response_api:
  enabled: true
  store_backend: "redis"
  ttl_seconds: 86400
  redis:
    cluster_mode: true
    cluster_addresses:
      - "127.0.0.1:7001"
      - "127.0.0.1:7002"
      - "127.0.0.1:7003"
      - "127.0.0.1:7004"
      - "127.0.0.1:7005"
      - "127.0.0.1:7006"
    db: 0  # PHẢI là 0 cho cụm
    key_prefix: "sr:"
    pool_size: 20       # Cao hơn cho cụm
    max_retries: 5      # Nhiều lần thử lại hơn cho chuyên chuyển
    dial_timeout: 10    # Lâu hơn cho cụm
```

#### Tùy Chọn 2: Tệp Cấu Hình Bên Ngoài

Chỉnh sửa `config/config.yaml`:

```yaml
response_api:
  enabled: true
  store_backend: "redis"
  ttl_seconds: 86400
  redis:
    config_path: "config/response-api/redis-cluster.yaml"
```

Sau đó chỉnh sửa `config/response-api/redis-cluster.yaml` với địa chỉ cụm.

### 3. Chạy Bộ Định Tuyến Ngữ Nghĩa

```bash
make build-router
make run-router
```

### 4. Chạy EnvoyProxy

```bash
# Bắt đầu proxy Envoy
make run-envoy
```

### 5. Xác Minh Khởi Tạo Cụm

**Kiểm tra nhật ký để khởi tạo cụm:**

```bash
tail -f /tmp/router.log | grep -i "cluster\|redis"
```

**Dự kiến:**

```
RedisStore: creating cluster client (nodes=6, pool_size=20)
RedisStore: initialized successfully (cluster_mode=true, key_prefix=sr:, ttl=24h0m0s)
Response API enabled with redis backend
```

### 6. Kiểm Tra API Phản Hồi

> **Lưu ý:** Các ví dụ dưới đây sử dụng `llm-katan` (Qwen3-0.6B) làm phụ trợ LLM. Điều chỉnh tên `model` để phù hợp với cấu hình vLLM của bạn.

#### Kiểm Tra 1: Tạo Phản Hồi

```bash
curl -X POST http://localhost:8801/v1/responses \
  -H "Content-Type: application/json" \
  -d '{
    "model": "qwen3",
    "input": "Redis Cụm là gì?",
    "instructions": "Bạn là một chuyên gia cơ sở dữ liệu.",
    "store": true
  }'
```

**Phản Hồi:**

```json
{
  "id": "resp_bb63817af32280b4a3a8fb7f",
  "object": "response",
  "status": "completed",
  "model": "qwen3",
  ...
}
```

#### Kiểm Tra 2: Xác Minh Phân Phối Dữ Liệu

Kiểm tra nút nào lưu trữ dữ liệu:

```bash
for port in 7001 7002 7003 7004 7005 7006; do
  echo "=== Node $port ==="
  docker exec redis-node-$port redis-cli KEYS "sr:*"
done
```

**Đầu ra ví dụ:**

```
=== Node 7001 ===
sr:response:resp_bb63817af32280b4a3a8fb7f
=== Node 7002 ===

=== Node 7003 ===

=== Node 7004 ===

=== Node 7005 ===
sr:response:resp_bb63817af32280b4a3a8fb7f  # Bản sao của 7001
=== Node 7006 ===
```

**Cái này cho thấy:**

- Chính 7001 có dữ liệu (slot hash phù hợp)
- Bản sao 7005 có một bản sao (sao lưu)
- Các nút khác trống (slot hash khác)

#### Kiểm Tra 3: Truy Xuất Phản Hồi

```bash
curl -X GET http://localhost:8801/v1/responses/resp_bb63817af32280b4a3a8fb7f
```

**Máy khách tự động:**

- Tính slotslot hash cho khóa
- Định tuyến yêu cầu đến nút chính (7001)
- Xử lý chuyên chuyển MOVED nếu cần

#### Kiểm Tra 4: Xâu Chuỗi Cuộc Trò Chuyện

```bash
curl -X POST http://localhost:8801/v1/responses \
  -H "Content-Type: application/json" \
  -d '{
    "model": "qwen3",
    "input": "Nói cho tôi thêm về sharding",
    "previous_response_id": "resp_bb63817af32280b4a3a8fb7f",
    "store": true
  }'
```

**Phản Hồi:**

```json
{
  "id": "resp_a4ae205a80ae7bf10edecaa3",
  "previous_response_id": "resp_bb63817af32280b4a3a8fb7f",
  "status": "completed",
  ...
}
```

#### Kiểm Tra 5: Xóa Phản Hồi

```bash
curl -X DELETE http://localhost:8801/v1/responses/resp_bb63817af32280b4a3a8fb7f
```

**Xóa hoạt động trên cụm:**

- Máy khách tìm nút chính
- Xóa từ chính
- Bản sao đồng bộ tự động

## Giám Sát Cụm

### Kiểm Tra Sức Khỏe Cụm

```bash
docker exec redis-node-7001 redis-cli cluster info
```

**Số liệu chính:**

- `cluster_state:ok` - Cụm lành mạnh
- `cluster_slots_assigned:16384` - Tất cả các slot được gán
- `cluster_known_nodes:6` - Tất cả các nút được phát hiện

### Xem Các Vai Trò Nút

```bash
docker exec redis-node-7001 redis-cli cluster nodes
```

**Đầu ra cho thấy:**

- Các nút chính với phạm vi slot hash
- Các nút bản sao và nút chính nào họ sao lưu

### Giám Sát Khóa Trên Mỗi Nút

```bash
for port in 7001 7002 7003; do
  count=$(docker exec redis-node-$port redis-cli DBSIZE)
  echo "Master $port: $count keys"
done
```

## Dọn Dẹp

### Dừng và Xóa Tất Cả Các Nút

```bash
for port in 7001 7002 7003 7004 7005 7006; do
  docker stop redis-node-$port
  docker rm redis-node-$port
done

docker network rm redis-cluster-net
```

## Tham Khảo

- [Redis Storage (Standalone)](redis-storage.md) - Thiết lập độc lập đơn giản
- Cấu Hình: `config/response-api/redis-cluster.yaml`
- Bài Kiểm Tra Tích Hợp: `pkg/responsestore/redis_store_integration_test.go`
