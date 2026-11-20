import {
  isRouteErrorResponse,
  Links,
  Meta,
  Outlet,
  Scripts,
  ScrollRestoration,
} from 'react-router';
import { RootProvider } from 'fumadocs-ui/provider/react-router';
import { Analytics } from '@vercel/analytics/react';
import type { Route } from './+types/root';
import './app.css';
import NotFound from './routes/not-found';
import { LaunchBanner } from './components/launch-banner';

export const links: Route.LinksFunction = () => [
  { rel: 'preconnect', href: 'https://fonts.googleapis.com' },
  {
    rel: 'preconnect',
    href: 'https://fonts.gstatic.com',
    crossOrigin: 'anonymous',
  },
  // Text fonts never swap in mid-load, so nothing reflows on arrival: if Geist
  // isn't ready in time the first visit keeps the fallback, and later visits use it.
  {
    rel: 'stylesheet',
    href: 'https://fonts.googleapis.com/css2?family=Geist+Mono:wght@400;500&display=optional',
  },
  // Inter Tight for every page, light and large for headings.
  {
    rel: 'stylesheet',
    href: 'https://fonts.googleapis.com/css2?family=Inter+Tight:wght@300;400;500;600&display=swap',
  },
  { rel: 'icon', type: 'image/svg+xml', href: '/favicon.svg' },
  { rel: 'apple-touch-icon', href: '/apple-touch-icon.png' },
  { rel: 'manifest', href: '/site.webmanifest' },
];

export function Layout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" suppressHydrationWarning>
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        <meta name="theme-color" content="#eef4fb" />
        <Meta />
        <Links />
      </head>
      <body className="flex flex-col min-h-screen">
        <LaunchBanner />
        <RootProvider search={{ options: { type: 'static' } }} theme={{ forcedTheme: 'light', defaultTheme: 'light', enableSystem: false, disableTransitionOnChange: true }}>{children}</RootProvider>
        <Analytics />
        <ScrollRestoration />
        <Scripts />
      </body>
    </html>
  );
}

export default function App() {
  return <Outlet />;
}

export function ErrorBoundary({ error }: Route.ErrorBoundaryProps) {
  let message = 'Oops!';
  let details = 'An unexpected error occurred.';
  let stack: string | undefined;

  if (isRouteErrorResponse(error)) {
    if (error.status === 404) return <NotFound />;
    message = 'Error';
    details = error.statusText;
  } else if (import.meta.env.DEV && error && error instanceof Error) {
    details = error.message;
    stack = error.stack;
  }

  return (
    <main className="pt-16 p-4 w-full max-w-[1400px] mx-auto">
      <h1>{message}</h1>
      <p>{details}</p>
      {stack && (
        <pre className="w-full p-4 overflow-x-auto">
          <code>{stack}</code>
        </pre>
      )}
    </main>
  );
}
