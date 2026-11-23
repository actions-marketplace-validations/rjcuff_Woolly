import type { Route } from './+types/terms';
import { LegalPage } from '@/components/legal-page';
import { seo } from '@/lib/seo';
import { contactEmail, githubUrl } from '@/lib/shared';

export function meta(_: Route.MetaArgs) {
  return seo({
    title: 'Terms | Woolly',
    description: 'The terms for using the Woolly app and website, and for Woolly Team subscriptions.',
    path: '/terms',
    image: '/og/home.png',
  });
}

export default function Terms() {
  return (
    <LegalPage title="Terms" updated="November 23, 2026">
      <p>By using the Woolly app or this website, you agree to these terms.</p>

      <h2>The software</h2>
      <p>
        Woolly is open source under the <a href={`${githubUrl}/blob/main/LICENSE`}>MIT license</a>.
        You may use, copy, change and share it, including for commercial use, as that license
        allows.
      </p>

      <h2>No warranty</h2>
      <p>
        Woolly is provided as is, without warranty of any kind. You are responsible for your own
        data, including backups. We are not liable for lost data or any damages that come from using
        the software or this website.
      </p>

      <h2>Your data</h2>
      <p>
        Your memories belong to you and stay on your computer. See the <a href="/privacy">privacy
        policy</a> for details.
      </p>

      <h2 id="woolly-team">Woolly Team subscriptions</h2>
      <p>
        Woolly Team (team.usewoolly.com) is an optional hosted service, separate from the free
        Woolly app. These terms also apply when you subscribe to it. Woolly Team is sold by Ryan
        Cuff, a sole proprietor in Florida, United States (&quot;we&quot;).
      </p>
      <ul>
        <li>
          <strong>Plans and billing.</strong> Paid plans are billed monthly per GitHub organization,
          in US dollars, through Stripe, starting the day you subscribe. They renew each month until
          you cancel. Prices are on the <a href="/team#pricing">Woolly Team page</a>. We'll tell
          subscribers at least 30 days before a price change takes effect.
        </li>
        <li>
          <strong>Taxes.</strong> Sales tax or similar taxes are added at checkout where they apply.
        </li>
        <li>
          <strong>Cancelling.</strong> Cancel anytime from Manage billing on the Plan tab. Your plan
          keeps working until the end of the month you paid for, then moves to Free.
        </li>
        <li>
          <strong>Refunds.</strong> If you ask within 14 days of your first payment, we'll refund it
          in full. After that, payments aren't refunded, but cancelling stops the next one. If you
          were charged twice by mistake, we'll refund the extra charge.
        </li>
        <li>
          <strong>Failed payments.</strong> If a payment fails, Stripe tries again and emails you. If
          it still can't be paid, the subscription ends and the organization moves to Free.
        </li>
        <li>
          <strong>Your data.</strong> Woolly Team keeps only what the{' '}
          <a href="/privacy">privacy policy</a> lists. Uninstalling the GitHub App deletes it and
          ends a paid subscription at the end of the month already paid.
        </li>
        <li>
          <strong>Fair use.</strong> Don&apos;t use Woolly Team to break the law, to get around plan
          limits, or to harm the service or other customers. We may suspend an account that does.
        </li>
        <li>
          <strong>Availability.</strong> We work to keep Woolly Team running but don&apos;t promise
          it will always be available or free of errors. The service is provided as is.
        </li>
        <li>
          <strong>Limit of liability.</strong> To the extent the law allows, our total liability for
          any claim about Woolly Team is limited to what you paid for it in the 12 months before the
          claim, and we aren&apos;t liable for indirect or consequential damages, or lost profits or
          data.
        </li>
        <li>
          <strong>Governing law.</strong> These terms are governed by the laws of the State of
          Florida, United States, without regard to its conflict of law rules.
        </li>
      </ul>

      <h2>Trademark</h2>
      <p>
        Woolly™ and the Woolly logo are trademarks of the Woolly project. The MIT license covers the
        code, not the name or logo. You may say your project works with Woolly or is based on
        Woolly, but please don't use the name or logo in a way that suggests we made or endorse
        your project.
      </p>

      <h2>Changes</h2>
      <p>We may update these terms. If we do, we'll change the date at the top of this page.</p>

      <h2>Contact</h2>
      <p>
        Questions about the app? Open an issue on <a href={githubUrl}>GitHub</a>. For a Woolly Team
        subscription, billing or a refund, email{' '}
        <a href={`mailto:${contactEmail}`}>{contactEmail}</a>.
      </p>
    </LegalPage>
  );
}
