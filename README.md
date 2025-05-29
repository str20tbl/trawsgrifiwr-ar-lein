# Trawsgrifiwr Ar-lein

**Trawsgrifiwr Ar-lein** is an open-source web application that allows users to generate **Welsh-language subtitles**
from audio or video files using automatic speech recognition (ASR) technology. Developed by the Language Technologies
Unit at Bangor University, the platform provides a simple and accessible browser-based interface for uploading,
processing, editing, and downloading subtitles.

---

## 🚀 Features

- 🎤 Upload audio or video files (e.g. `.mp3`, `.mp4`)
- 🔊 Automatically generate Welsh-language subtitles using Bangor University's ASR API
- 📝 View and edit subtitles directly in your browser
- 💾 Export subtitles in standard formats like `.srt` and `.vtt`
- 📦 Dockerized for easy deployment

---

## 🖥️ Live Demo

[A hosted demo is made available via Bangor University.](https://trawsgrifiwr.techiaith.cymru) I

f you'd like to deploy locally, see below.

---

## 🛠️ Getting Started

### Requirements

- Docker
- Docker Compose

### Installation

1. Clone the repository:

```bash
git clone https://github.com/str20tbl/trawsgrifiwr-ar-lein.git
cd trawsgrifiwr-ar-lein

docker-compose up --build
```

Open a browser and visit [http://localhost:7070](http://localhost:7070)
