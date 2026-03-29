import { defineConfig } from 'vitepress'
import { useSidebar } from 'vitepress-openapi'
import spec from '../public/openapi.json'

const sidebar = useSidebar({ spec })

export default defineConfig({
  base: '/packster/',
  title: 'Packster',
  description: 'Package version management — API documentation',
  themeConfig: {
    nav: [
      { text: 'Home', link: '/' },
      { text: 'Getting Started', link: '/getting-started' },
      { text: 'API Reference', link: '/api' },
    ],
    sidebar: [
      {
        text: 'Guide',
        items: [
          { text: 'Getting Started', link: '/getting-started' },
          { text: 'Installation', link: '/installation' },
        ],
      },
      {
        text: 'API Reference',
        link: '/api',
        items: sidebar.generateSidebarGroups(),
      },
    ],
    socialLinks: [
      { icon: 'github', link: 'https://github.com/IdanKoblik/packster' },
    ],
  },
})
