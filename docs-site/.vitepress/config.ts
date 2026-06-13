import { defineConfig } from 'vitepress'

export default defineConfig({
  base: '/saldeti/',
  title: 'Saldeti',
  description: 'API Simulator for Entra ID and Google Workspace',
  themeConfig: {
    nav: [
      { text: 'Home', link: '/' },
      { text: 'CLI Reference', link: '/cli-reference' },
      { text: 'Seed Files', link: '/seed-files' },
      { text: 'Google Endpoints', link: '/google-endpoints' },
      { text: 'Entra Endpoints', link: '/entra-endpoints' },
    ],
    sidebar: [
      {
        text: 'Guide',
        items: [
          { text: 'Introduction', link: '/' },
          { text: 'CLI Reference', link: '/cli-reference' },
          { text: 'Seed Files', link: '/seed-files' },
          { text: 'Google Endpoints', link: '/google-endpoints' },
          { text: 'Entra Endpoints', link: '/entra-endpoints' },
        ],
      },
    ],
    search: {
      provider: 'local',
    },
    socialLinks: [
      { icon: 'github', link: 'https://github.com/saldeti/saldeti' },
    ],
  },
})