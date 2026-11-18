import type { BaseLayoutProps } from 'fumadocs-ui/layouts/shared';
import { DitherMammoth } from '@/components/frost/dither-mammoth';
import { githubUrl } from './shared';

export function baseOptions(): BaseLayoutProps {
  return {
    nav: {
      title: (
        <span className="flex items-center gap-2.5 text-[var(--color-brand)]">
          <DitherMammoth small className="h-[24px] w-auto" />
          <span className="text-[20px] font-medium tracking-[-0.03em]">woolly</span>
        </span>
      ),
    },
    githubUrl,
    // Light only: no theme switch anywhere on the site.
    themeSwitch: { enabled: false },
  };
}
