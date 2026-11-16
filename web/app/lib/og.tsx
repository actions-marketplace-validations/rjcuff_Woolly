import { readFile } from 'node:fs/promises';
import { join } from 'node:path';
import { Renderer } from '@takumi-rs/core';
import { ImageResponse } from 'takumi-js/response';

// Link previews: "woolly" in Geist, set like the homepage headline, with one line under it.
// Images are rendered at build time, so the font files are read from disk.

let renderer: Promise<Renderer> | undefined;

function getRenderer(): Promise<Renderer> {
  renderer ??= (async () => {
    const dir = join(process.cwd(), 'app/og/fonts');
    const r = new Renderer();
    await r.registerFont({ name: 'Geist', data: await readFile(join(dir, 'Geist.ttf')) });
    return r;
  })();
  return renderer;
}

export async function ogImage(subtitle: string): Promise<Response> {
  return new ImageResponse(
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        width: '100%',
        height: '100%',
        backgroundColor: '#ffffff',
        color: '#111111',
        fontFamily: 'Geist',
      }}
    >
      <div style={{ fontSize: 200, fontWeight: 400, lineHeight: 1, letterSpacing: '-0.03em' }}>woolly</div>
      <div style={{ marginTop: 44, fontSize: 44, color: '#555555', fontWeight: 500 }}>{subtitle}</div>
    </div>,
    { width: 1200, height: 630, format: 'png', renderer: await getRenderer() },
  );
}

// Square profile picture for social accounts. The wordmark is sized to stay
// inside the circle most sites crop avatars to.
export async function profileImage(): Promise<Response> {
  return new ImageResponse(
    <div
      style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        width: '100%',
        height: '100%',
        backgroundColor: '#ffffff',
        color: '#111111',
      }}
    >
      <div style={{ fontFamily: 'Geist', fontSize: 170, fontWeight: 400, lineHeight: 1, letterSpacing: '-0.03em' }}>woolly</div>
    </div>,
    { width: 800, height: 800, format: 'png', renderer: await getRenderer() },
  );
}
