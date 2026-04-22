# Kos Finance Backend

Backend API untuk sistem manajemen keuangan kos/rumah kontrakan. Aplikasi ini menyediakan API untuk mengelola data pembayaran, profil pengguna, pengeluaran, dan fitur finansial lainnya dalam sistem kos.

## 📋 Deskripsi Project

**Kos Finance Backend** adalah aplikasi server yang dibangun menggunakan Go dan Gin Framework. Project ini berfungsi sebagai REST API untuk sistem manajemen keuangan kos, dengan fitur-fitur:

- **Manajemen Pembayaran**: Track pembayaran dari penghuni kos
- **Profil Pengguna**: Kelola data penghuni dan pengelola kos
- **Pengeluaran**: Catat dan kelola pengeluaran operasional
- **Autentikasi**: Sistem autentikasi untuk melindungi endpoint sensitif
- **Penyimpanan File**: Integrasi dengan AWS S3/R2 untuk penyimpanan dokumen
- **Database**: Menggunakan Supabase (PostgreSQL) untuk penyimpanan data

## 🛠️ Tech Stack

| Komponen           | Teknologi                             |
| ------------------ | ------------------------------------- |
| **Runtime**        | Go 1.23.0                             |
| **Web Framework**  | Gin                                   |
| **Database**       | Supabase (PostgreSQL)                 |
| **Storage**        | AWS S3 / R2 (Cloudflare)              |
| **Authentication** | Custom middleware dengan JWT          |
| **Module**         | github.com/rhyssh/kos-finance-backend |

## 📦 Prerequisites

Sebelum memulai, pastikan Anda memiliki:

1. **Go** versi 1.23.0 atau lebih tinggi
   - Download: https://go.dev/dl/
   - Verifikasi: `go version`

2. **Git** untuk clone repository
   - Download: https://git-scm.com/

3. **Akun Supabase** (untuk database)
   - Daftar: https://supabase.com/
   - Buat project baru dan catat URL dan API Key

4. **AWS atau Cloudflare R2** (untuk storage file) - Opsional
   - Pastikan memiliki credentials jika ingin menggunakan fitur storage

5. **Text Editor atau IDE**
   - Rekomendasi: VS Code, GoLand, atau editor lainnya

## 🚀 Instalasi & Setup

### 1. Clone Repository

```bash
git clone https://github.com/rhyssh/kos-finance-backend.git
cd kos-finance-backend
```

### 2. Download Dependencies

```bash
go mod download
go mod tidy
```

Perintah ini akan mengunduh semua dependencies yang diperlukan sesuai `go.mod`.

### 3. Setup Environment Variable

Buat file `.env` di root directory project:

```bash
# Linux/Mac
touch .env

# Windows (PowerShell)
New-Item -Path .env -ItemType File
```

Isi file `.env` dengan konfigurasi berikut:

```env
# Server Configuration
PORT=8080
GIN_MODE=debug

# Supabase Configuration
SUPABASE_URL=https://your-project-id.supabase.co
SUPABASE_KEY=your-supabase-anon-key
SUPABASE_SERVICE_ROLE_KEY=your-supabase-service-role-key

# AWS/R2 Configuration (Optional)
AWS_REGION=auto
AWS_ENDPOINT=https://your-account-id.r2.cloudflarestorage.com
AWS_ACCESS_KEY_ID=your-access-key
AWS_SECRET_ACCESS_KEY=your-secret-key
AWS_BUCKET_NAME=your-bucket-name

# Other Configuration
LOG_LEVEL=info
```

**Cara mendapatkan Supabase credentials:**

1. Login ke https://supabase.com/
2. Buka project Anda
3. Pergi ke Settings > API
4. Copy `URL` dan `anon key`

### 4. Setup Database (Supabase)

1. Buka Supabase dashboard
2. Pergi ke SQL Editor
3. Jalankan migration/schema yang diperlukan (jika ada file `migrations/` di project)
4. Pastikan table-table seperti `profiles`, `payments`, `expenses` sudah dibuat

## 🏃 Menjalankan Project Secara Local

### Cara 1: Langsung dengan Go

```bash
# Run server
go run ./cmd/api/main.go
```

Server akan berjalan di `http://localhost:8080`

### Cara 2: Build Executable

```bash
# Build
go build -o kos-finance-backend ./cmd/api

# Run
./kos-finance-backend

# Windows
.\kos-finance-backend.exe
```

### Cara 3: Dengan Makefile (Jika sudah dikonfigurasi)

```bash
make run
```

## ✅ Verifikasi Instalasi

Setelah server berjalan, test dengan endpoint health:

```bash
# Linux/Mac
curl http://localhost:8080/health

# Windows PowerShell
Invoke-WebRequest http://localhost:8080/health
```

Response yang diharapkan:

```json
{
  "status": "ok",
  "message": "Backend Golang berjalan!",
  "env": "8080"
}
```

Test koneksi Supabase:

```bash
curl http://localhost:8080/supabase-test
```

## 📁 Struktur Project

```
kos-finance-backend/
├── cmd/
│   └── api/
│       └── main.go              # Entry point aplikasi
├── internal/
│   ├── config/                  # Konfigurasi aplikasi
│   ├── handler/                 # HTTP handlers
│   │   ├── payment.go           # Handler untuk pembayaran
│   │   └── profile.go           # Handler untuk profil
│   ├── middleware/              # Custom middleware
│   │   └── auth.go              # Autentikasi middleware
│   ├── model/                   # Data models
│   │   ├── expense.go           # Model pengeluaran
│   │   └── payment.go           # Model pembayaran
│   ├── repository/              # Data access layer
│   ├── service/                 # Business logic layer
│   ├── storage/                 # File storage integration
│   │   └── r2.go                # AWS S3/R2 configuration
│   └── supabase/                # Supabase client
│       └── client.go            # Supabase connection
├── go.mod                       # Go modules file
├── go.sum                       # Go modules checksum
├── Makefile                     # Build automation (optional)
└── README.md                    # Dokumentasi ini
```

## 🔌 API Endpoints

### Public Endpoints

| Method | Endpoint         | Deskripsi                   |
| ------ | ---------------- | --------------------------- |
| GET    | `/health`        | Check status server         |
| GET    | `/supabase-test` | Test koneksi database       |
| GET    | `/api/rekap`     | Ringkasan keuangan (public) |

### Protected Endpoints (Require Authentication)

| Method | Endpoint          | Deskripsi                |
| ------ | ----------------- | ------------------------ |
| GET    | `/api/me`         | Get info user yang login |
| \*     | `/api/payments/*` | Endpoint pembayaran      |
| \*     | `/api/profiles/*` | Endpoint profil pengguna |

\*Catatan: Endpoint protected memerlukan JWT token di header `Authorization: Bearer <token>`

## 🔐 Authentication

Endpoint yang dilindungi memerlukan JWT token:

```bash
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:8080/api/me
```

Middleware autentikasi ada di `internal/middleware/auth.go`

## 📝 Contoh Development Workflow

### 1. Clone dan Setup

```bash
git clone https://github.com/rhyssh/kos-finance-backend.git
cd kos-finance-backend
go mod download
cp .env.example .env  # Atau buat .env dengan config Anda
```

### 2. Update .env dengan credentials Anda

```bash
nano .env  # atau gunakan editor favorit Anda
```

### 3. Jalankan Server

```bash
go run ./cmd/api/main.go
```

### 4. Test API (dengan curl atau Postman)

```bash
curl http://localhost:8080/health
curl http://localhost:8080/supabase-test
```

## 🐛 Troubleshooting

### Error: "module not found"

```bash
go mod download
go mod tidy
```

### Error: "cannot find .env file"

- Periksa bahwa file `.env` sudah dibuat di root directory
- Pastikan variabel environment sudah sesuai

### Error: "connection refused to Supabase"

- Verifikasi `SUPABASE_URL` dan `SUPABASE_KEY` di `.env`
- Pastikan project Supabase sudah aktif
- Check internet connection

### Error: "port already in use"

- Ganti `PORT` di `.env`, contoh: `PORT=3000`
- Atau kill process yang menggunakan port tersebut

## � Fitur Kedepannya (Upcoming Features)

Fitur-fitur yang sedang atau akan dikembangkan untuk meningkatkan fungsionalitas sistem:

- ✨ **Verifikasi Bukti Bayar dengan AI** - Sistem otomatis menggunakan Computer Vision dan Machine Learning untuk memverifikasi kelengkapan dan keaslian bukti pembayaran (transfer bank, e-wallet, etc)

- 📊 **Dashboard Analytics** - Dashboard interaktif dengan visualisasi data pembayaran, pengeluaran, dan laporan keuangan real-time

- 📱 **Mobile App Integration** - API expansion untuk mendukung mobile application (iOS/Android)

- 🔔 **Notification System** - Sistem notifikasi otomatis untuk reminder pembayaran, konfirmasi transaksi, dan alert keuangan

- 📧 **Email & SMS Gateway** - Integrasi email dan SMS untuk notifikasi dan pemberitahuan kepada penghuni

- 📈 **Laporan Keuangan Otomatis** - Generate laporan keuangan bulanan/tahunan dalam format PDF dengan analisis mendalam

- 💳 **Payment Gateway Integration** - Integrasi dengan payment gateway (Midtrans, Xendit, dll) untuk memudahkan pembayaran online

- 👥 **Multi-Tenant Support** - Dukungan untuk multiple kos/properti dalam satu sistem

- 🔐 **Enhanced Security** - 2FA (Two Factor Authentication), encryption at rest, dan audit logging

- 📍 **Geolocation & QR Code** - QR code untuk invoice dan verifikasi lokasi penghuni

- 🤖 **AI-Powered Insights** - Prediksi pembayaran yang belum masuk, analisis pola pengeluaran, dan rekomendasi manajemen keuangan

- 🌐 **Multi-Language Support** - Dukungan bahasa Indonesia, Inggris, dan bahasa lainnya

## �📚 Resources

- **Go Documentation**: https://go.dev/doc/
- **Gin Framework**: https://github.com/gin-gonic/gin
- **Supabase**: https://supabase.com/docs
- **AWS S3 SDK**: https://aws.amazon.com/sdk-for-go/

## 👨‍💻 Development Tips

1. **Hot Reload** - Gunakan `air` untuk auto-reload saat development:

   ```bash
   go install github.com/cosmtrek/air@latest
   air
   ```

2. **Database Migrations** - Gunakan Supabase dashboard untuk manage schema

3. **Testing** - Jalankan tests dengan:

   ```bash
   go test ./...
   ```

4. **Code Format** - Format code dengan:
   ```bash
   go fmt ./...
   ```

## 👤 Author

**rhyssh**
GitHub: https://github.com/rhyssh

---

Jika ada pertanyaan atau masalah, silakan buat issue atau hubungi developer.
