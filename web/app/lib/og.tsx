import { Renderer } from '@takumi-rs/core';
import { ImageResponse } from 'takumi-js/response';
import { BRAND, ditherSvg } from '@/lib/mammoth';

// Link previews: the dithered mammoth on Woolly blue. They have no text, so
// the renderer needs no fonts. Images are rendered at build time.

let renderer: Renderer | undefined;

function getRenderer(): Renderer {
  renderer ??= new Renderer();
  return renderer;
}

const dataUri = (svg: string) => `data:image/svg+xml;base64,${Buffer.from(svg).toString('base64')}`;
const MAMMOTH = dataUri(ditherSvg(BRAND.paper));

/** A 1200 by 630 link preview: the dithered mammoth in ice white on Woolly blue, and nothing else. */
export async function ogImage(): Promise<Response> {
  return new ImageResponse(
    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', width: '100%', height: '100%', backgroundColor: BRAND.blue }}>
      <img src={MAMMOTH} width={546} height={416} />
    </div>,
    { width: 1200, height: 630, format: 'png', renderer: getRenderer() },
  );
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
