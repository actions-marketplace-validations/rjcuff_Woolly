import type { Route } from './+types/og.docs';
import { source } from '@/lib/source';
import { ogImage } from '@/lib/og';

export function loader({ params }: Route.LoaderArgs) {
  const slugs = params['*']
    .split('/')
    .filter((v) => v.length > 0)
    .slice(0, -1);
  const page = source.getPage(slugs);
  if (!page) throw new Response(undefined, { status: 404 });

  // Every page shares the same simple preview: the mammoth on Woolly blue.
  return ogImage();
}
