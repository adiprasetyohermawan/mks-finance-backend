

# mks-finance-backend

Backend service untuk mini PoC **MKS Finance**. Repo ini berisi aplikasi backend (umumnya Spring Boot) yang menyediakan REST API, terintegrasi dengan database (MySQL/PostgreSQL), serta (opsional) ekosistem streaming seperti Kafka/Kafka Connect.

> Catatan: Dokumentasi ini dibuat agar bisa langsung dipakai untuk deployment di VM yang sudah menjalankan Docker. Jika ada perbedaan nama service/port di `docker-compose.yml` kamu, ikuti yang ada di compose.

---

## 1) Arsitektur Ringkas

Komponen yang umum dipakai pada mini PoC:

- **Backend API**: menerima request, menjalankan business logic, dan akses DB.
- **Database**: MySQL dan/atau PostgreSQL.
- **Kafka (opsional)**: message broker untuk event/stream.
- **Kafka Connect + Debezium (opsional)**: CDC dari database ke Kafka atau sink ke DB.
- **Kafka UI**: dashboard UI untuk monitoring cluster Kafka.

---

## 2) Prasyarat

### Wajib
- Git
- Docker Engine + Docker Compose v2

### Opsional (untuk development lokal tanpa Docker)
- Java 17+ (atau versi Java sesuai project)
- Maven/Gradle (sesuai project)

---

## 3) Struktur Folder (umum)

Struktur dapat sedikit berbeda, namun biasanya seperti ini:

```
backend/
  src/                  # source code aplikasi
  pom.xml | build.gradle
  Dockerfile             # build image backend
  docker-compose.yml     # stack lokal/VM (jika disediakan)
  mysql_init/            # init script MySQL (opsional)
  postgres_init/         # init script Postgres (opsional)
  README.md
```

---

## 4) Konfigurasi Environment

### 4.1 File `.env`
Jika repo menggunakan `.env`, buat dari template:

```bash
cp .env.example .env
```

### 4.2 Variabel yang umum
Nama variabel dapat berbeda tergantung implementasi, namun berikut checklist yang umum:

- `APP_PORT` (default sering 8080)
- `SPRING_PROFILES_ACTIVE=local|dev|prod`
- `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD`
- (Opsional) `KAFKA_BOOTSTRAP_SERVERS`

> Penting: `.env` dan file secret lain **tidak** boleh di-commit. Sudah di-cover oleh `.gitignore`.

---

## 5) Menjalankan di VM (Recommended: Docker)

### 5.1 Jalankan seluruh stack
Jika repo menyediakan `docker-compose.yml`, jalankan dari folder yang sama:

```bash
docker compose up -d
```

Cek status:

```bash
docker compose ps
```

Lihat log backend:

```bash
docker logs -f --tail 200 <nama_container_backend>
```

### 5.2 Port yang umum
- Backend API: `http://<VM_IP>:<APP_PORT>` (sering 8080)
- Kafka UI: `http://<VM_IP>:8080` (jika service `kafka-ui` melakukan mapping `8080:8080`)

Jika kamu tidak membuka port publik, gunakan SSH tunnel dari laptop:

```bash
ssh -L 8080:localhost:8080 root@<VM_IP>
```
Lalu buka:

- `http://localhost:8080`

### 5.3 Rebuild image (kalau ada perubahan kode)

```bash
docker compose build --no-cache
docker compose up -d
```

---

## 6) Menjalankan Lokal Tanpa Docker (Opsional)

> Gunakan ini hanya jika kamu memang ingin menjalankan backend langsung di OS.

### 6.1 Maven (contoh)

```bash
./mvnw spring-boot:run
```

Atau build jar:

```bash
./mvnw clean package -DskipTests
java -jar target/*.jar
```

### 6.2 Gradle (contoh)

```bash
./gradlew bootRun
```

---

## 7) Database

### 7.1 Akses MySQL di container

```bash
docker exec -it mysql mysql -uapp_user -p
```

### 7.2 Akses PostgreSQL di container

```bash
docker exec -it postgres psql -U app_user -d pgsql_db
```

### 7.3 Migrasi / DDL
Jika project memakai Flyway/Liquibase, lokasi file biasanya:
- `src/main/resources/db/migration` (Flyway)
- `src/main/resources/db/changelog` (Liquibase)

---

## 8) API Docs & Health Check

Karena implementasi dapat berbeda, gunakan endpoint berikut (jika tersedia):

- Health: `GET /actuator/health`
- Info: `GET /actuator/info`
- OpenAPI/Swagger UI (umum):
  - `/swagger-ui/index.html`
  - `/v3/api-docs`

Contoh test cepat via curl:

```bash
curl -s http://localhost:8080/actuator/health | jq
```

---

## 9) Kafka (Jika Dipakai)

### 9.1 Kafka UI
Jika `docker-compose.yml` memetakan port `8080:8080` untuk service `kafka-ui`, akses:

- `http://<VM_IP>:8080`

### 9.2 Cek topic

```bash
docker exec -it <kafka_container_name> kafka-topics --bootstrap-server <broker:9092> --list
```

> Nama container dan broker tergantung image (Confluent/Bitnami) dan konfigurasi compose.

---

## 10) Troubleshooting

### 10.1 Port tidak bisa diakses dari luar VM
Checklist:
- Pastikan container up: `docker compose ps`
- Pastikan port bind: `ss -lntp | grep :8080` (atau port API)
- Pastikan firewall/security group membuka port inbound
- Pastikan service di compose menggunakan `ports:` bukan hanya `expose:`

### 10.2 Backend tidak bisa connect DB
Checklist:
- `DB_HOST` mengarah ke **nama service** di docker network (mis. `mysql`), bukan `localhost`.
- Port DB sesuai: MySQL `3306`, Postgres `5432`.
- Credential sesuai dengan environment di compose.

### 10.3 Reset stack (hapus data lokal)
Hati-hati: ini menghapus volume/bind-mount data lokal.

```bash
docker compose down -v
```

---

## 11) Konvensi Git

- Jangan commit `.env`, credential, private key, dump DB.
- Gunakan branch feature: `feature/<nama>` atau `fix/<nama>`.
- Tambahkan dokumentasi perubahan konfigurasi di README ini.

---

## 12) Lisensi

Internal/PoC.
# mks-finance-backend

Backend service untuk **MKS Finance – Customer Profile 360 (PoC)**.

Repo ini berisi aplikasi backend (Spring Boot) yang menyediakan REST API untuk:
- daftar & pencarian customer,
- customer 360 profile (customer + credit applications + vehicle ownership),
- KPI dashboard,
- evidence health/sync (untuk kebutuhan PoC).

Backend ini **umumnya membaca data dari ODS PostgreSQL** (hasil sink/CDC), sehingga aplikasi bisa mengakses **near real-time data** tanpa membebani sistem MySQL production.

---

## 1) Ringkasan Arsitektur PoC

Alur data yang direkomendasikan:

1. **MySQL (source)** → perubahan data ditangkap oleh **Debezium (Kafka Connect Source)**
2. Debezium menulis event ke **Kafka topics**
3. **Kafka Connect JDBC Sink** menulis ke **PostgreSQL (ODS)**
4. **Backend API** membaca dari PostgreSQL (ODS)
5. **Frontend** memanggil Backend API

Komponen minimal untuk demo end-to-end:
- MySQL (source)
- Kafka + Kafka Connect (Debezium)
- PostgreSQL (ODS)
- Backend API
- Frontend UI

---

## 2) Daftar API yang Disediakan

Base path: `/api/v1`

| Endpoint | Method | Deskripsi | Dipakai oleh UI |
|---|---:|---|---|
| `/health` | GET | Health check backend | (opsional) |
| `/customers` | GET | List/search customers + pagination + sort | Customers Page |
| `/customers/{customerId}/profile` | GET | Customer 360 profile (customer + credit_applications + vehicle_ownership) | Customer Profile Page |
| `/stats/kpi` | GET | KPI untuk dashboard | Dashboard Page |
| `/sync/health` | GET | Evidence sync health (status, lag, SLA target, last_success, last_error) | Dashboard Page |

Contoh test cepat:

```bash
curl -s http://localhost:8088/api/v1/health | jq
curl -s "http://localhost:8088/api/v1/customers?limit=10&offset=0" | jq
curl -s "http://localhost:8088/api/v1/customers/<CUSTOMER_ID>/profile" | jq
curl -s http://localhost:8088/api/v1/stats/kpi | jq
curl -s http://localhost:8088/api/v1/sync/health | jq
```

> Catatan: port `8088` di atas adalah **contoh host port** saat backend dijalankan via Docker Compose. Jika kamu menjalankan langsung di OS, portnya mengikuti konfigurasi aplikasi (mis. 8080).

---

## 3) Prasyarat

### Untuk menjalankan via Docker (disarankan di VM)
- Docker Engine
- Docker Compose v2

### Untuk menjalankan lokal tanpa Docker (opsional)
- Java 17+
- Maven/Gradle (atau gunakan wrapper `./mvnw` / `./gradlew`)

---

## 4) Konfigurasi (Environment)

### 4.1 Variabel yang umum
Sesuaikan dengan implementasi project kamu, namun biasanya:

- `SERVER_PORT` atau `APP_PORT` (mis. 8080 di container)
- `SPRING_PROFILES_ACTIVE=local|dev|prod`
- PostgreSQL (ODS):
  - `DB_HOST`, `DB_PORT=5432`, `DB_NAME`, `DB_USER`, `DB_PASSWORD`
- (opsional) Kafka:
  - `KAFKA_BOOTSTRAP_SERVERS=kafka:29092`

**Penting:**
- Jika backend berjalan **di dalam docker network**, `DB_HOST` gunakan **nama service** (contoh: `postgres`), bukan `localhost`.
- Jangan commit file secret (`.env`, key, credential).

### 4.2 Contoh `.env` (opsional)

```env
SPRING_PROFILES_ACTIVE=dev
DB_HOST=postgres
DB_PORT=5432
DB_NAME=ods_db
DB_USER=ods_user
DB_PASSWORD=ods_pwd
```

---

## 5) Menjalankan di VM (Recommended)

### 5.1 Cara 1: Docker Compose (paling praktis)

Umumnya backend dijalankan sebagai container di stack yang sama dengan MySQL/Kafka/Postgres.

```bash
docker compose up -d

docker compose ps

docker logs -f --tail 200 backend
```

Jika kamu ada perubahan kode dan ingin rebuild image:

```bash
docker compose build backend --no-cache
docker compose up -d backend
```

### 5.2 Port & cara akses

Jika di compose kamu menulis:

```yaml
ports:
  - "8088:8080"
```

Artinya:
- **8080** = port di **dalam container** (backend listen di container)
- **8088** = port di **host VM** (yang dipakai browser/curl dari luar)

Maka aksesnya selalu menggunakan **host port (8088)**:
- dari laptop: `http://<VM_IP>:8088/api/v1/...`
- dari VM (host): `http://localhost:8088/api/v1/...`

Sedangkan `8080` hanya relevan untuk:
- komunikasi **antar container** di network Docker (mis. `http://backend:8080/...`), atau
- debugging dari dalam container.

---

## 6) Menjalankan Lokal (Tanpa Docker)

### 6.1 Maven

```bash
./mvnw spring-boot:run
```

Atau build jar:

```bash
./mvnw clean package -DskipTests
java -jar target/*.jar
```

### 6.2 Gradle

```bash
./gradlew bootRun
```

---

## 7) Database (ODS PostgreSQL)

Masuk ke PostgreSQL container:

```bash
docker exec -it postgres psql -U ods_user -d ods_db
```

Checklist schema/table ODS (contoh):
- `customers`
- `credit_applications`
- `vehicle_ownership`

> Jika backend kamu masih membaca dari MySQL, kamu bisa tetap jalankan. Namun untuk PoC yang menekankan “tidak mengganggu production MySQL”, pattern paling aman adalah backend membaca dari ODS PostgreSQL.

---

## 8) Integrasi CDC (Debezium → Kafka → Postgres)

Yang biasanya dicek evaluator PoC:
- Connector Debezium Source untuk MySQL: status **RUNNING**
- JDBC Sink ke Postgres: status **RUNNING**
- Lag/latency: p95 < 10 detik (target PoC)

Shortcut pemeriksaan:

```bash
# list connectors
curl -s http://localhost:8083/connectors | jq

# status connector
curl -s http://localhost:8083/connectors/<connector-name>/status | jq
```

---

## 9) Observability / PoC Evidence

Untuk memenuhi success criteria terkait latency & no data loss, kamu perlu bukti:

- **Sync health** dari backend (`/api/v1/sync/health`) mengeluarkan:
  - `status` (ok/warn/bad)
  - `sla_target_seconds`
  - `lag_seconds`
  - `last_success_at`
  - `last_error`

- Validasi data:
  - sampling record antara source MySQL vs target Postgres
  - cek count atau checksum sederhana (opsional)

- Response time UI:
  - cek via DevTools Network (frontend) atau `curl -w`.

Contoh hitung waktu response sederhana:

```bash
curl -o /dev/null -s -w "time_total=%{time_total}\n" \
  "http://localhost:8088/api/v1/customers?limit=20&offset=0"
```

---

## 10) Troubleshooting

### 10.1 Port tidak bisa diakses dari luar VM
- `docker compose ps` pastikan container up
- `ss -lntp | grep 8088` pastikan port bind di host
- pastikan firewall/security group membuka inbound port
- pastikan compose pakai `ports:` (bukan hanya `expose:`)

### 10.2 Backend tidak bisa connect Postgres
- Jika backend di docker network: `DB_HOST=postgres` (nama service)
- Cek credential sesuai compose
- Cek DB ready: `docker logs -f postgres`

### 10.3 Reset environment (hapus volume)
Hati-hati: menghapus data lokal.

```bash
docker compose down -v
```

---

## 11) Konvensi Git

- Jangan commit: `.env`, credential, private key, dump DB, folder build artifacts.
- Gunakan branch: `feature/<nama>` atau `fix/<nama>`.
- Setiap perubahan konfigurasi deployment/port: update README ini.

---

## 12) Lisensi

Internal / PoC.