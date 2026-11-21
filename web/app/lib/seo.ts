import { appName, siteUrl } from './shared';

type SeoInput = {
  title: string;
  description: string;
  path: string;
  image?: string;
  type?: 'website' | 'article';
  imageAlt?: string;
};

// Every page shares one preview image, the herd of mammoths from the homepage,
// except Woolly Team, which has its own.
const IMAGE_ALT = 'A herd of woolly mammoths carrying notes across a snowy field under a blue sky';

// Tags for search engines and link previews. Titles never use colons or dashes.
export function seo({ title, description, path, image, type = 'website', imageAlt = IMAGE_ALT }: SeoInput) {
  const url = new URL(path, siteUrl).toString();
  const imageUrl = image ? new URL(image, siteUrl).toString() : undefined;
  return [
    { title },
    { name: 'description', content: description },
    { tagName: 'link', rel: 'canonical', href: url },
    { property: 'og:site_name', content: appName },
    { property: 'og:type', content: type },
    { property: 'og:title', content: title },
    { property: 'og:description', content: description },
    { property: 'og:url', content: url },
    { property: 'og:locale', content: 'en_US' },
    ...(imageUrl
      ? [
          { property: 'og:image', content: imageUrl },
          { property: 'og:image:type', content: 'image/png' },
          { property: 'og:image:width', content: '1200' },
          { property: 'og:image:height', content: '630' },
          { property: 'og:image:alt', content: imageAlt },
        ]
      : []),
    { name: 'twitter:card', content: imageUrl ? 'summary_large_image' : 'summary' },
    { name: 'twitter:title', content: title },
    { name: 'twitter:description', content: description },
    ...(imageUrl
      ? [
          { name: 'twitter:image', content: imageUrl },
          { name: 'twitter:image:alt', content: imageAlt },
        ]
      : []),
  ];
}
