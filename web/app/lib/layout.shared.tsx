import type { BaseLayoutProps } from 'fumadocs-ui/layouts/shared';
import { SiteThemeToggle } from '@/components/brand';
import { githubUrl } from './shared';

export function baseOptions(): BaseLayoutProps {
  return {
    nav: {
      title: <span className="font-sans text-[19px] font-medium tracking-[-0.02em]">Woolly</span>,
    },
    githubUrl,
    slots: {
      themeSwitch: SiteThemeToggle,
    },
  };
}
