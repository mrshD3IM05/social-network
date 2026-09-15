import './globals.css'

export const metadata = {
  title: 'social-network',
  description: 'A social network for the people who matter',
}

export default function RootLayout({ children }) {
  return (
    <html lang="en">
      <head>
        {/* the fonts used in globals.css */}
        <link
          rel="stylesheet"
          href="https://fonts.googleapis.com/css2?family=Geist:wght@400;500;600&family=Geist+Mono:wght@400;500&family=Poppins:ital,wght@0,700;1,700&display=swap"
        />
      </head>
      <body>{children}</body>
    </html>
  )
}
