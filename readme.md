---
Title: DevOps Project
---

## Architecture

```txt

                    +-------------------+
                    |     Sensor(s)     |
                    |-------------------|
                    |      LoraWan      |
                    +-------------------+
                              |
                              |
                           LoraWan
                              |
                              v
                    +-------------------+
                    |      Gateway      |
                    |-------------------|
                    | - LoraWan         |
                    | - MQTT Publisher  |
                    | - SQLite          |
                    |  (ML Prediction)  |
                    +-------------------+
                              |
                              |
                         MQTT Publish
                              |
                              v
                  +-------------------------+
                  |       MQTT Broker       |
                  |-------------------------|
                  | (EMQX, Mosquitto, etc.) |
                  +-------------------------+
                              ^
                              |
                        Internet (TLS)
                              |
                        MQTT Subscribe
                              |
                              |
                +-----------------------------+
                |      Consumer (Cloud)       |
                |-----------------------------|
                | - MQTT Subscriber           |
                | - Lưu Data vào SQLite       |
                | - Websocket Server          |
                | - HTTP Server (HTMX + UI)   |
                | - Tailwind UI Templates     |
                +-----------------------------+
                              |
                              |
             +----------------+----------------+
             |                                 |
  WebSocket (Push realtime)            HTTP Query (HTMX)
             |                                 |
   +--------------------+           +-------------------------+
   | Realtime Live View |           | Historical Query + Alert|
   +--------------------+           +-------------------------+
```

## Triển khai Kubernetes

- Triển khai được Kubernetes thông qua công cụ minikube trên 1 node
- Triển khai được Kubernetes thông qua công cụ kubeadm hoặc kubespray
  lên 1 master node VM + 1 worker node VM

## K8S Helm Chart

- Cài đặt ArgoCD lên Kubernetes Cluster, expose được ArgoCD qua NodePort
- Cài đặt Jenkins lên Kubernetes Cluster, expose được Jenkins qua NodePort
- Viết hoặc tìm mẫu Helm Chart cho app bất kỳ, để vào 1 folder riêng trong repo app
- Tạo Repo Config cho app trên, trong repo này chứa các file values.yaml
  với nội dung của cá file values.yaml là các config cần thiết
  để chạy ứng dụng trên k8s bằng Helm Chart
- Viết 1 luồng CI/CD cho app, khi có thay đổi từ source code
  1 tag mới được tạo ra trên trên repo này thì luồng CI/CD
  tương ứng của repo đó thực hiện các công việc sau:
  - Sửa code trong source code
  - Thực hiện build source code trên Jenkin bằng docker
    với image tag là tag name đã được tạo ra trên gitlab/github
    và push docker image sau khi build xong lên Docker Hub
  - Sửa giá trị Image version trong file values.yaml trong config repo
    và push thay đổi lên config repo
  - Cấu hình ArgoCD tự động triển khai lại web Deployment và api Deployment
    khi có sự thay đổi trên config repo

## Monitor

- Expose metric của app ra 1 http path
- Sử dụng ansible playbooks để triển khai container Prometheus server.
  Sau đó cấu hình prometheus add target giám sát các metrics đã expose ở trên
- Sử dụng ansible playbooks để triển khai stack EFK (elasticsearch, fluentd, kibana)
  Sau đó cấu hình logging cho web service và api service,
  đảm bảo khi có http request gửi vào web service hoặc api service
  thì trong các log mà các service này sinh ra, có ít nhất 1 log có các thông tin

## Security

- Dựng HAProxy Loadbalancer trên 1 VM riêng
  (trong trường hợp cụm lab riêng của sinh viên)
  với mode TCP, mở port trên LB trỏ đến NodePort của App trên K8S Cluster
- Sử dụng giải pháp Ingress cho các deployment,
  đảm bảo các truy cập đến các port App sử dụng https
- Cho phép sinh viên sử dụng self-signed cert để làm bài

---

## 🏗️ **TỔNG THỂ KIẾN TRÚC**

```plaintext
+----------------+       +------------------+       +-------------------+        +-------------------+
| Sensor giả lập | ====> |   Gateway xử lý  | ====> | Cloud / AWS Layer | ====>  | Web Monitoring UI |
| (Python Script)|       | (FastAPI/Django) |       |   (API / DB)      |        |   (Django + HTMX) |
+----------------+       +------------------+       +-------------------+        +-------------------+
                               |                                                       ^
                               v                                                       |
                (Xử lý dữ liệu, làm sạch, format)                                      |
                               |                                                       |
                               +--------------------> Redis / PostgreSQL <-------------+
```

---

## 🧩 **CHI TIẾT CÔNG CỤ / STACK CHO TỪNG THÀNH PHẦN**

### 🔧 **1. Sensor mô phỏng (Python Script)**

| Thành phần     | Công nghệ                        | Vai trò                                                |
| -------------- | -------------------------------- | ------------------------------------------------------ |
| Sensor Script  | `Python`                         | Sinh dữ liệu ngẫu nhiên (giá trị mực nước, áp suất...) |
| Giao tiếp      | `HTTP` (POST) hoặc `MQTT`        | Gửi đến Gateway                                        |
| Lập lịch       | `schedule`, `time.sleep`, `cron` | Gửi dữ liệu định kỳ                                    |
| Format dữ liệu | JSON                             | Chuẩn hóa format gửi                                   |

> 📁 Ví dụ thư viện: `random`, `schedule`, `paho-mqtt`, `requests`

### ⚙️ **2. Gateway xử lý dữ liệu (Python)**

| Thành phần       | Công nghệ                   | Vai trò                        |
| ---------------- | --------------------------- | ------------------------------ |
| Web Server       | `FastAPI` **hoặc** `Django` | Nhận dữ liệu từ sensor         |
| Data Cleaning    | Python logic                | Làm sạch, check anomaly        |
| Queue (tuỳ chọn) | `Redis` + `Celery`          | Batch xử lý hoặc async         |
| Gửi lên Cloud    | `requests`, `boto3`         | Gửi tiếp lên cloud API hoặc DB |
| Ghi log          | `loguru`, `logging`         | Ghi nhật ký gửi nhận dữ liệu   |

> ✅ _FastAPI gọn nhẹ, dễ triển khai cho microservice. Django thích hợp nếu bạn dùng chung project với Web UI._

### ☁️ **3. Cloud (AWS hoặc mock server)**

| Thành phần                        | Công nghệ                                      | Vai trò                            |
| --------------------------------- | ---------------------------------------------- | ---------------------------------- |
| API Gateway                       | `AWS API Gateway` hoặc custom FastAPI endpoint | Nhận request từ Gateway            |
| Lưu trữ thời gian thực            | `AWS DynamoDB` hoặc `PostgreSQL`               | Lưu dữ liệu cảm biến               |
| Xử lý realtime (tuỳ chọn)         | `AWS Lambda`, `Kinesis`                        | Xử lý hoặc filter dữ liệu realtime |
| Sử dụng local (nếu chưa dùng AWS) | FastAPI hoặc PostgreSQL local                  | Dễ test, chưa cần deploy AWS       |

> ✅ Bạn có thể dùng local PostgreSQL ban đầu rồi chuyển sang DynamoDB hoặc RDS sau.

### 🌐 **4. Website monitoring realtime**

| Thành phần         | Công nghệ                        | Vai trò                           |
| ------------------ | -------------------------------- | --------------------------------- |
| Backend web        | `Django`                         | Hiển thị dashboard                |
| Frontend UI        | `HTMX` + `Tailwind CSS`          | Tạo giao diện động đơn giản       |
| Giao tiếp realtime | **2 lựa chọn**:                  |                                   |
|                    | ✅ `HTMX polling` (5s/lần...)    | Dễ làm, dễ debug                  |
|                    | ✅ `Django Channels` + WebSocket | Giao tiếp thời gian thực thực thụ |
| DB                 | PostgreSQL (chia sẻ với Gateway) | Lưu dữ liệu cảm biến              |

> 🔁 _HTMX là giải pháp đơn giản và hiệu quả. Nếu cần realtime "live stream", nên dùng WebSocket._

## 🧰 TỔNG HỢP CÁC CÔNG CỤ ĐỀ XUẤT

| Mục tiêu                   | Công cụ cụ thể                        |
| -------------------------- | ------------------------------------- |
| Sinh dữ liệu cảm biến      | Python, `random`, `schedule`          |
| Giao tiếp Sensor → Gateway | HTTP (POST) hoặc MQTT                 |
| Gateway xử lý dữ liệu      | FastAPI hoặc Django                   |
| Queue xử lý (tùy chọn)     | Redis + Celery                        |
| Lưu dữ liệu                | PostgreSQL hoặc DynamoDB              |
| Giao tiếp Gateway → Cloud  | HTTP hoặc AWS SDK (boto3)             |
| Web UI                     | Django + HTMX + Tailwind              |
| Realtime frontend          | HTMX polling **hoặc** Django Channels |
| Logging / Debug            | loguru, Django Debug Toolbar          |

## 🗂️ KẾ HOẠCH TRIỂN KHAI GỢI Ý

1. ✅ Giai đoạn 1: Cảm biến + Gateway + lưu local PostgreSQL
2. ✅ Giai đoạn 2: Xây website theo thời gian thực với polling HTMX
3. ✅ Giai đoạn 3: Đưa Gateway lên cloud (deploy API)
4. ✅ Giai đoạn 4: Dùng WebSocket (Django Channels) hoặc chuyển sang AWS IoT

---
