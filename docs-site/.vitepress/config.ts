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
      { icon: 'github', link: 'https://github.com/alexjplant/saldeti' },
    ],
    footer: {
      message: 'Released under the <a href="https://www.gnu.org/licenses/agpl-3.0.en.html">AGPL-3.0</a> License.',
      copyright: 'Saldeti is not affiliated with, endorsed by, or sponsored by Microsoft Corporation or Google LLC. All trademarks are property of their respective owners.',
    },
  },
})