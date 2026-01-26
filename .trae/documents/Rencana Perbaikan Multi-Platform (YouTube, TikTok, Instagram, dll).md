# Perbaikan Error Download YouTube, TikTok, Instagram, Vimeo, dan Dailymotion

Berdasarkan analisis log dan laporan Anda, berikut adalah rencana perbaikan menyeluruh:

## 1. Penyelesaian Instagram (Prioritas Tinggi)
Status saat ini: `chromedp_strategy.go` belum lengkap (masih ada TODO).
- **Tindakan**: Melengkapi fungsi `GetVideoInfo` di `chromedp_strategy.go`.
- **Detail**: 
  - Integrasi validasi Blob URL (mengabaikan URL `blob:` yang tidak valid).
  - Implementasi penangkapan m3u8 master playlist untuk download video yang stabil.
  - Memastikan cookies dari sesi browser dikirim ke downloader.

## 2. Perbaikan TikTok (Masalah "Empty File")
Status saat ini: `yt-dlp` gagal mengunduh konten (file kosong) dan warning cookies.
- **Tindakan**: Menambahkan flag `--impersonate` pada argumen `yt-dlp` di `downloader.go`.
- **Detail**:
  - Karena kita sudah menginstall `curl-cffi`, kita perlu memaksa `yt-dlp` untuk menggunakannya guna menghindari deteksi bot TikTok.
  - Menambahkan argumen `--impersonate="chrome110"` (atau versi yang sesuai) khusus untuk TikTok.

## 3. Perbaikan YouTube (Masalah 403/Bot)
Status saat ini: Kembali error setelah update.
- **Tindakan**: Memperketat penggunaan `cookies.txt` dan Client ID.
- **Detail**:
  - Memastikan path `cookies.txt` absolut `/app/cookies.txt` terbaca.
  - Mengubah opsi client `yt-dlp` menjadi `android` atau `tv` secara eksplisit untuk YouTube.

## 4. Perbaikan Vimeo & Dailymotion
Status saat ini: Error download.
- **Tindakan**: Menambahkan header `Referer`.
- **Detail**:
  - Vimeo dan Dailymotion sering menolak request tanpa referer yang benar.
  - Menambahkan `--referer <video_url>` ke argumen `yt-dlp`.

## 5. Verifikasi & Rebuild
- **Tindakan**: Rebuild container worker dan backend.
- **Catatan**: Mengingat error Docker sebelumnya, kita akan mencoba rebuild lagi. Jika gagal, saya akan pandu cara manual.

Apakah Anda setuju dengan rencana ini?
