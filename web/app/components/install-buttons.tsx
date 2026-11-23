import { Button } from '@/components/sky/parts';
import { cursorInstallUrl, lmStudioInstallUrl } from '@/lib/deeplinks';
import { cn } from '@/lib/utils';

interface InstallButtonsProps {
  className?: string;
}

/** One-click buttons that open Cursor or LM Studio and add Woolly. They need the woolly command installed first. */
export function InstallButtons({ className }: InstallButtonsProps) {
  return (
    <div className={cn('flex flex-wrap gap-3', className)}>
      <Button href={cursorInstallUrl()} tone="white">
        Add to Cursor
      </Button>
      <Button href={lmStudioInstallUrl()} tone="white">
        Add to LM Studio
      </Button>
    </div>
  );
}
