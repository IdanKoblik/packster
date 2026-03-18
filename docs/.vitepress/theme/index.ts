import DefaultTheme from 'vitepress/theme'
import { theme, useOpenapi } from 'vitepress-openapi/client'
import 'vitepress-openapi/dist/style.css'
import spec from '../../public/openapi.json'

export default {
  ...DefaultTheme,
  async enhanceApp({ app, router, siteData }: { app: any; router: any; siteData: any }) {
    useOpenapi({
      spec,
      config: {
        operation: {
          hiddenSlots: ['playground'],
        },
      },
    })
    theme.enhanceApp({ app })
  },
}
