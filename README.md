# 🏫 Aplikasi Kantin Sekolah (UKK RPL 2025–2026)

## 📌 Tentang Proyek

Aplikasi **Kantin Sekolah** ini merupakan proyek **Uji Kompetensi Keahlian (UKK)** untuk kompetensi **Rekayasa Perangkat Lunak (RPL)**.

Proyek ini dikembangkan sebagai **backend RESTful API** menggunakan bahasa **Golang**, dengan tujuan memenuhi seluruh kebutuhan sistem pemesanan makanan dan minuman di lingkungan sekolah.

Aplikasi ini dibuat dan dikembangkan oleh:

> **Yohanes Capelliou Samudra**
> Siswa **SMK Telkom Malang**
> Angkatan kelulusan **Tahun 2026**

Proyek ini juga berfungsi sebagai **tugas akhir sekolah**, sekaligus portofolio teknis backend.

---

## 🎯 Tujuan Pengembangan

* Memenuhi seluruh **requirement UKK RPL 2025–2026**
* Mengimplementasikan sistem **pemesanan kantin online** berbasis API
* Menerapkan prinsip **clean code**, **security-aware backend**, dan **real-world flow**
* Menjadi bukti kompetensi siswa dalam pengembangan sistem backend

---

## 🧑‍🎓 Fitur Siswa

Siswa dapat:

1. Login ke sistem
2. Melihat daftar menu makanan dan minuman
3. Melihat harga asli dan harga setelah diskon
4. Melakukan pemesanan makanan/minuman
5. Melihat histori transaksi
6. Mencetak struk / nota pemesanan dalam bentuk **PDF**

---

## 🧑‍🍳 Fitur Admin Stan

Admin stan dapat:

1. Login sebagai admin stan
2. Mengelola menu makanan dan minuman (CRUD)
3. Melihat pesanan masuk dari siswa
4. Mengubah status pesanan (diproses → diantar → selesai)
5. Melihat rekap pemasukan stan

---

## 🛠️ Teknologi yang Digunakan

* **Bahasa Pemrograman**: Golang
* **Framework Web**: Gin Gonic
* **ORM**: GORM
* **Database**: MySQL
* **Authentication**: JWT (JSON Web Token)
* **PDF Generator**: gofpdf
* **Arsitektur**: RESTful API (Backend Only)

---

## 🔐 Keamanan

Beberapa aspek keamanan yang diterapkan:

* JWT Authentication & Role-based Authorization
* Validasi akses berdasarkan role (siswa / admin stan / super admin)
* Validasi kepemilikan data (order hanya bisa diakses pemiliknya)
* Rate limiting sederhana untuk endpoint sensitif
* Harga dan diskon dihitung **server-side** (tidak trust client)

---

## 📂 Struktur Proyek (Ringkas)

```
cmd/server          -> Entry point aplikasi
internal/api        -> Handler API (siswa, admin, auth)
internal/app        -> Database, models, utilities
internal/routes     -> Routing & middleware
```

---

## ▶️ Alur Penggunaan (MVP UKK)

### POV Siswa

1. Login
2. Melihat menu
3. Membuat pesanan
4. Melihat histori transaksi
5. Mencetak struk PDF

### POV Admin Stan

1. Login
2. Melihat pesanan masuk
3. Mengubah status pesanan hingga selesai

---

## 🏁 Penutup

Proyek ini dikembangkan sebagai bagian dari **Uji Kompetensi Keahlian (UKK)** dan **tugas akhir sekolah** di **SMK Telkom Malang**.

Diharapkan aplikasi ini dapat menjadi bukti kemampuan teknis, pemahaman sistem backend, serta kesiapan untuk melanjutkan ke jenjang pendidikan atau industri teknologi.

---

✍️ **Dikembangkan oleh**
**Yohanes Capelliou Samudra**
SMK Telkom Malang — Lulus 2026
