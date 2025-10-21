import { defineConfig } from 'vitepress'

export default defineConfig({
  title: "Gokku",
  description: "Lightweight git-push deployment system for Go and multi-language applications",
  base: '/',
  ignoreDeadLinks: true,

  head: [
    ['link', { rel: 'icon', href: '/favicon.ico' }]
  ],

  themeConfig: {
    logo: '/logo.svg',

    nav: [
      { text: '1.0.53', link: '/' },
      { text: 'Home', link: '/' },
      { text: 'Guide', link: '/guide/getting-started' },
      { text: 'Examples', link: '/examples/' },
      { text: 'Reference', link: '/reference/configuration' }
    ],

    sidebar: {
      '/guide/': [
        {
          text: 'Introduction',
          items: [
            { text: 'What is Gokku?', link: '/guide/what-is-gokku' },
            { text: 'Getting Started', link: '/guide/getting-started' },
            { text: 'Installation', link: '/guide/installation' }
          ]
        },
        {
          text: 'Core Concepts',
          items: [
            { text: 'Configuration', link: '/guide/configuration' },
            { text: 'Environments', link: '/guide/environments' },
            { text: 'Deployment', link: '/guide/deployment' }
          ]
        },
        {
          text: 'Advanced',
          items: [
            { text: 'Docker Support', link: '/guide/docker' },
            { text: 'Environment Variables', link: '/guide/env-vars' },
            { text: 'Rollback', link: '/guide/rollback' }
          ]
        }
      ],
      '/examples/': [
        {
          text: 'Examples',
          items: [
            { text: 'Overview', link: '/examples/' },
            { text: 'Go Application', link: '/examples/go-app' },
            { text: 'Python Application', link: '/examples/python-app' },
            { text: 'Docker Application', link: '/examples/docker-app' },
            { text: 'Multi-App Project', link: '/examples/multi-app' }
          ]
        }
      ],
      '/reference/': [
        {
          text: 'Reference',
          items: [
            { text: 'Configuration', link: '/reference/configuration' },
            { text: 'CLI Commands', link: '/reference/cli' },

          ]
        }
      ]
    },

    socialLinks: [
      { icon: 'github', link: 'https://github.com/thadeu/gokku' }
    ],

    footer: {
      message: 'Released under the MIT License.',
      copyright: 'Copyright © 2025-present'
    },

    search: {
      provider: 'local'
    }
  }
})

