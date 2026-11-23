import { type ReactNode, useCallback, useEffect, useState } from 'react';
import { type Activity, api, type Memory } from '../api';
import { AppLogo } from './logos';
import { ago, I, Icon, Skeleton, type Toast } from './ui';

export type LibraryView = 'conflicts' | 'queue' | 'problems' | 'history' | 'trash' | 'activity';

const TITLES: Record<LibraryView, string> = {
  conflicts: 'Conflicts',
  queue: 'Still true?',
  problems: 'Files to fix',
  history: 'History',
  trash: 'Trash',
  activity: 'Activity',
};

const VERBS: Record<string, string> = {
  remember: 'Saved',
  recall: 'Searched',
  profile: 'Read your profile',
  list: 'Listed memories',
  context: 'Loaded your profile',
  update: 'Edited',
  forget: 'Forgot',
  restore: 'Restored',
  resolve: 'Kept',
  keep: 'Kept',
  empty_trash: 'Emptied the trash',
  import: 'Imported',
  learn: 'Learned',
  dream: 'Tidied memory:',
  sync: 'Synced',
};

/** Conflicts, old guesses to check, unreadable files, history, trash and the activity log, inside the app. */
export function Library({
  view,
  onBack,
  onToast,
  onChanged,
}: {
  view: LibraryView;
  onBack: () => void;
  onToast: (t: Toast) => void;
  onChanged: () => void;
}) {
  const [items, setItems] = useState<Memory[] | null>(null);
  const [all, setAll] = useState<Map<string, Memory>>(new Map());
  const [log, setLog] = useState<Activity[] | null>(null);
  const [problems, setProblems] = useState<string[] | null>(null);

  const load = useCallback(async () => {
    try {
      if (view === 'activity') {
        const [entries, mems, trash] = await Promise.all([api.activity(300), api.list(true), api.trash()]);
        setAll(new Map([...mems, ...trash].map((m) => [m.id, m])));
        setLog(entries);
        return;
      }
      if (view === 'problems') {
        setProblems(await api.problems());
        return;
      }
      if (view === 'queue') {
        setItems(await api.reviewQueue());
        return;
      }
      const mems = await api.list(true);
      setAll(new Map(mems.map((m) => [m.id, m])));
      if (view === 'conflicts') setItems(mems.filter((m) => m.status === 'conflict'));
      else if (view === 'history') setItems(mems.filter((m) => m.status === 'superseded'));
      else setItems(await api.trash());
    } catch (e) {
      onToast({ text: (e as Error).message });
    }
  }, [view, onToast]);

  useEffect(() => {
    setItems(null);
    setLog(null);
    setProblems(null);
    load();
  }, [load]);

  const act = async (f: () => Promise<unknown>, text: string) => {
    try {
      await f();
      onToast({ text });
      load();
      onChanged();
    } catch (e) {
      onToast({ text: (e as Error).message });
    }
  };

  const loading = view === 'activity' ? log === null : view === 'problems' ? problems === null : items === null;
  const listView = view !== 'activity' && view !== 'problems';

  return (
    <div className="ca-page narrow">
      <button className="back" onClick={onBack}>
        <Icon d={I.chevronLeft} size={18} />
        Settings
      </button>
      <div className="ca-title-row">
        <h1 className="ca-title big">{TITLES[view]}</h1>
        {view === 'trash' && items && items.length > 0 && (
          <button className="btn plain danger" onClick={() => act(api.emptyTrash, 'Trash emptied')}>
            Empty trash
          </button>
        )}
      </div>

      {loading && (
        <div className="cards">
          {[0, 1, 2].map((i) => (
            <div key={i} className="card">
              <Skeleton w={`${80 - i * 12}%`} h={14} />
              <Skeleton w={120} h={11} className="skel-gap" />
            </div>
          ))}
        </div>
      )}

      {!loading && view === 'queue' && items && items.length > 0 && (
        <p className="sec-sub">An AI guessed these, and nothing has used them in 90 days. Keep the ones that are still true and forget the rest.</p>
      )}

      {!loading && listView && items?.length === 0 && (
        <Empty
          icon={view === 'conflicts' || view === 'queue' ? I.check : view === 'trash' ? I.trash : I.clock}
          title={view === 'conflicts' || view === 'queue' ? 'Nothing to review' : view === 'trash' ? 'Trash is empty' : 'No history yet'}
          text={
            view === 'conflicts'
              ? 'When an app saves something that disagrees with you, it shows up here.'
              : view === 'queue'
                ? 'Every guess an AI made was used or changed in the last 90 days.'
                : view === 'trash'
                  ? 'Forgotten memories wait here until you empty the trash.'
                  : 'Older versions of facts show up here when they change.'
          }
        />
      )}

      {!loading && view === 'problems' && problems?.length === 0 && (
        <Empty icon={I.check} title="Every file reads fine" text="Memory files Woolly can't fully read show up here, so you can fix them by hand." />
      )}

      {!loading && view === 'problems' && problems && problems.length > 0 && (
        <>
          <p className="sec-sub">Woolly couldn't fully read these memory files, so they are left out of search. Open each one in a text editor and fix or remove it.</p>
          <div className="cards">
            {problems.map((p) => (
              <div key={p} className="card">
                <p className="card-text">
                  <code>{p}</code>
                </p>
              </div>
            ))}
          </div>
        </>
      )}

      {!loading && listView && items && items.length > 0 && (
        <div className="cards">
          {items.map((m) => {
            const other = m.conflicts_with ? all.get(m.conflicts_with) : undefined;
            const newer = m.superseded_by ? all.get(m.superseded_by) : undefined;
            return (
              <div key={m.id} className="card">
                {view === 'conflicts' && other && (
                  <div className="versus">
                    <span className="versus-label">You said</span>
                    <p>{other.content}</p>
                  </div>
                )}
                {view === 'conflicts' && <span className="versus-label">{m.via ? `${m.via} saved` : 'An app saved'}</span>}
                <p className="card-text">{m.content}</p>
                {view === 'history' && newer && <p className="card-meta">Replaced by “{newer.content}”</p>}
                <div className="card-foot">
                  <span className="card-meta">{[m.via, m.project, ago(m.updated_at)].filter(Boolean).join(' · ')}</span>
                  <span className="card-actions">
                    {view === 'conflicts' && (
                      <>
                        {other && (
                          <button className="btn secondary" onClick={() => act(() => api.resolve(other.id), 'Kept yours')}>
                            Keep yours
                          </button>
                        )}
                        <button className="btn primary" onClick={() => act(() => api.resolve(m.id), 'Updated')}>
                          Use this
                        </button>
                      </>
                    )}
                    {view === 'queue' && (
                      <>
                        <button className="btn secondary" onClick={() => act(() => api.forget(m.id), 'Forgotten')}>
                          Forget
                        </button>
                        <button className="btn primary" onClick={() => act(() => api.keep(m.id), 'Kept')}>
                          Still true
                        </button>
                      </>
                    )}
                    {view === 'history' && (
                      <button className="btn secondary" onClick={() => act(() => api.resolve(m.id), 'Made current')}>
                        Make current
                      </button>
                    )}
                    {view === 'trash' && (
                      <button className="btn secondary" onClick={() => act(() => api.restore(m.id), 'Restored')}>
                        Restore
                      </button>
                    )}
                  </span>
                </div>
              </div>
            );
          })}
        </div>
      )}

      {!loading && view === 'activity' && log?.length === 0 && (
        <Empty icon={I.clock} title="No activity yet" text="When an app saves, reads or forgets a memory, it shows up here." />
      )}

      {!loading && view === 'activity' && log && log.length > 0 && (
        <div className="cards">
          {runs(log).map(({ e, n }, i) => {
            const subject = e.ids?.[0] ? all.get(e.ids[0])?.content : undefined;
            const what =
              e.action === 'recall'
                ? e.query
                  ? `Searched “${e.query}”`
                  : `Read ${e.count ?? 0} memories`
                : subject && ['remember', 'update', 'forget', 'restore', 'resolve', 'keep'].includes(e.action)
                  ? `${VERBS[e.action]} “${subject}”`
                  : ['import', 'learn', 'sync', 'dream'].includes(e.action)
                    ? `${VERBS[e.action]} ${e.count ?? 0}`
                    : (VERBS[e.action] ?? e.action);
            return (
              <div key={`${e.time}-${i}`} className="card row-card">
                <AppLogo app={e.app} size={30} />
                <div className="card-body">
                  <p className="card-text one-line" title={what}>
                    {what}
                    {n > 1 && <span className="times"> ×{n}</span>}
                  </p>
                  <span className="card-meta">
                    {e.app} · {ago(e.time)}
                  </span>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}

// Back-to-back repeats, like an app searching the same thing several times,
// show as one line with a count.
function runs(log: Activity[]) {
  const out: { e: Activity; n: number }[] = [];
  for (const e of log) {
    const last = out[out.length - 1];
    if (last && last.e.app === e.app && last.e.action === e.action && last.e.query === e.query && (e.ids?.[0] ?? '') === (last.e.ids?.[0] ?? '')) {
      last.n++;
    } else {
      out.push({ e, n: 1 });
    }
  }
  return out;
}

export function Empty({ icon, title, text, children }: { icon: string; title: string; text?: string; children?: ReactNode }) {
  return (
    <div className="empty">
      <span className="empty-icon">
        <Icon d={icon} size={22} />
      </span>
      <p className="empty-title">{title}</p>
      {text && <p className="empty-text">{text}</p>}
      {children && <div className="empty-actions">{children}</div>}
    </div>
  );
}
