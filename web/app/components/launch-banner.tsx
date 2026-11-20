export const productHuntUrl = 'https://www.producthunt.com/products/woolly?utm_source=other&utm_medium=social';

/** A strip in Woolly blue above every page, pointing at the Product Hunt launch. It scrolls away with the page, so the sticky navs below it work as before. */
export function LaunchBanner() {
  return (
    <a
      href={productHuntUrl}
      target="_blank"
      rel="noopener noreferrer"
      className="group block bg-brand px-4 py-2 text-center text-[14px] leading-5 text-brand-foreground outline-none transition-colors duration-150 hover:text-white focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-brand-foreground"
    >
      <span>Woolly is live on Product Hunt.</span>{' '}
      <span className="font-medium whitespace-nowrap">
        Come say hi
        <span aria-hidden="true" className="ml-1 inline-block transition-transform duration-150 ease-ui group-hover:translate-x-0.5">
          →
        </span>
      </span>
    </a>
  );
}
