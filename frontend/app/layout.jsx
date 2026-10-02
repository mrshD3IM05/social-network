import './globals.css'
import { themeScript } from '@/lib/theme'

export const metadata = {
  title: 'social-network',
  description: 'A social network for the people who matter',
}

export default function RootLayout({ children }) {
  return (
    // the theme script sets data-theme before React loads, hence the warning opt-out
    <html lang="en" suppressHydrationWarning>
      <head>
        <script dangerouslySetInnerHTML={{ __html: themeScript }} />
        {/* the fonts used in globals.css */}
        <link
          rel="stylesheet"
          href="https://fonts.googleapis.com/css2?family=Bricolage+Grotesque:opsz,wght@12..96,500..800&family=Onest:wght@400;500;600;700&display=swap"
        />
      </head>
      <body>{children}</body>
    </html>
  )
}
