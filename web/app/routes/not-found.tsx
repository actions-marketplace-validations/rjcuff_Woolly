import type { Route } from './+types/not-found';
import { DocShell } from '@/components/doc/shell';
import { Button } from '@/components/sky/parts';

export function meta({}: Route.MetaArgs) {
  return [{ title: 'Not Found | Woolly' }];
}

export default function NotFound() {
  return (
    <DocShell
      eyebrow="404"
      title="Even a mammoth forgets sometimes"
      lead="This page doesn’t exist, or it moved. The docs or the homepage will get you back on track."
    >
      <div className="flex flex-wrap gap-3">
        <Button href="/docs">Read the docs</Button>
        <Button href="/" tone="white">
          Go home
        </Button>
      </div>
    </DocShell>
  );
}
