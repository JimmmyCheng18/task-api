import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 3666,
    host: '0.0.0.0'  // 允許外部 IP 訪問
  },
  build: {
    outDir: 'dist'
  }
})