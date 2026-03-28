# CineGraph 3D - Frontend 🎬

A dynamic, 3D interactive React application visualization for the Film API backend. Built with **React**, **Vite**, and **Three.js** via `react-force-graph`, this application allows you to explore films, genres, and people in an Obsidian-style network graph.

## ✨ Features

- **Interactive 3D Network Graph**: Traverse relationships between films, genres, and cast/crew members.
- **Smart Search**: Filter nodes dynamically via an integrated search bar.
- **Watchlist & Recommendations**: View personalized "For You" recommendations or your watch priority list dynamically.
- **Smooth Animations & Physics**: Implements D3 force physics to simulate node gravity and interactions.
- **Modern UI**: Styled with TailwindCSS and Lucide Icons for a sleek, glass-morphic vibe.

## 🛠️ Tech Stack

- **Framework**: React 19 + TypeScript
- **Bundler**: Vite
- **Graphing**: `react-force-graph-2d` / `react-force-graph-3d` (Three.js under the hood)
- **Styling**: Tailwind CSS / Vanilla CSS
- **Animations**: Motion (framer-motion)

---

## 🚀 Running Locally

### Prerequisites
- Node.js (v18 or higher)
- Film API Backend running locally (see main backend README)

### Setup

1. **Install dependencies**
   ```bash
   npm install
   ```

2. **Configure Environment**
   Create a `.env` file in this directory with the following:
   ```bash
   VITE_API_URL="http://localhost:4000/v1"
   ```

3. **Run the Development Server**
   ```bash
   npm run dev
   ```
   Open `http://localhost:3000` to interact with the graph.

---

## 🐳 Docker Self-Hosting

The frontend can be efficiently containerized independently. We use a **multi-stage build** containing `node:alpine` for the build step and an incredibly lightweight `nginx:alpine-slim` image to serve the compiled application.

### Build and Run with Docker

1. **Build the Image**
   ```bash
   docker build -t cinegraph-frontend .
   ```

2. **Run the Container**
   ```bash
   docker run -p 3000:80 -d --name cinegraph-ui cinegraph-frontend
   ```

   The app will be available on `http://localhost:3000`.

*Note: Ensure your backend Film API is already running. If your backend is in another docker container, ensure the VITE_API_URL resolves appropriately (such as putting them on the same docker network or passing it during the build ARG phase if deploying to production).*
