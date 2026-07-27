import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  return {
    plugins: [vue()],
    resolve: {
      alias: {
        '@': resolve(__dirname, 'src')
      }
    },
    server: {
      port: 20009,
      host: '0.0.0.0',
      proxy: {
        '/api': {
          target: 'http://localhost:20001',
          changeOrigin: true,
          rewrite: (path) => path.replace(/^\/api/, '/api/v1')
        }
      }
    },
    build: {
      outDir: 'dist',
      assetsDir: 'assets'
    },
    css: {
      preprocessorOptions: {
        scss: {
          additionalData: '@import "@/styles/variables.scss";\n'
        }
      }
    }
  }
})
