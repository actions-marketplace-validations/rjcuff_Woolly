import { readFile } from 'node:fs/promises';
import { join } from 'node:path';
import { Renderer } from '@takumi-rs/core';
import { ImageResponse } from 'takumi-js/response';
import { BRAND, ditherSvg } from '@/lib/mammoth';

// Profile pictures: the dithered mammoth on night blue. They have no text, so
// the renderer needs no fonts. Link previews use the homepage picture in
// public/og/home.png. Images are rendered at build time.

let renderer: Renderer | undefined;

function getRenderer(): Renderer {
  renderer ??= new Renderer();
  return renderer;
}

const dataUri = (svg: string) => `data:image/svg+xml;base64,${Buffer.from(svg).toString('base64')}`;
const MAMMOTH = dataUri(ditherSvg(BRAND.paper));

/** The 1200 by 630 link preview for docs pages: the same picture as the homepage. */
export async function ogImage(): Promise<Response> {
  const png = await readFile(join(process.cwd(), 'public/og/home.png'));
  return new Response(new Uint8Array(png), { headers: { 'Content-Type': 'image/png' } });
}

/** A square version for profile pictures and the home screen icon, sized to stay inside a circle crop. */
export async function profileImage(): Promise<Response> {
  return new ImageResponse(
    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', width: '100%', height: '100%', backgroundColor: BRAND.blue }}>
      <img src={MAMMOTH} width={504} height={384} />
    </div>,
    { width: 800, height: 800, format: 'png', renderer: getRenderer() },
  );
}
